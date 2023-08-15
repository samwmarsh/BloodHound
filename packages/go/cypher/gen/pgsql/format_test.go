// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package pgsql_test

import (
	"bytes"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/gen/pgsql"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type KindMapper struct {
	known map[string]int16
}

func (s KindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	var (
		kindIDs      = make([]int16, 0, len(kinds))
		missingKinds = make([]graph.Kind, 0, len(kinds))
	)

	for _, kind := range kinds {
		if kindID, hasKind := s.known[kind.String()]; hasKind {
			kindIDs = append(kindIDs, kindID)
		} else {
			missingKinds = append(missingKinds, kind)
		}
	}

	return kindIDs, missingKinds
}

type TestCase struct {
	ID                 int
	Source             string
	Query              *query.Builder
	Expected           string
	ExpectedParameters map[string]any
	Exclusive          bool
	Ignored            bool
	Error              bool
}

func Suite() []TestCase {
	return []TestCase{
		{
			Source:   "match (s) return s skip 5 limit 10",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s offset 5 limit 10",
		},
		{
			Source:   "match (s) return s order by s.name, s.other_prop desc",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s order by s.properties->'name' asc, s.properties->'other_prop' desc",
		},
		{
			Source:   "match (s) where (s)-[]->() return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (select count(*) from node as n2 join edge e0 on e0.start_id = n2.id join node n1 on n1.id = e0.end_id where s.id = n2.id limit 1) > 0",
		},
		{
			Source:   "match ()-[r]->() where (s {name: 'test'})-[r]->() return r",
			Expected: "select (r.id, r.start_id, r.end_id, r.kind_id, r.properties)::edgeComposite as r from node as n1 join edge r on r.start_id = n1.id join node n2 on n2.id = r.end_id where (select count(*) from node as s join edge e3 on e3.start_id = s.id join node n0 on n0.id = e3.end_id where (s.properties->>'name')::text = 'test' and r.id = e3.id limit 1) > 0",
		},
		{
			Source:   "match (s {value: 'PII'})-[r {other: 234}]->(e {that: 456}) where s.other = 'more pii' and e.number = 411 return s, r, e",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s, (r.id, r.start_id, r.end_id, r.kind_id, r.properties)::edgeComposite as r, (e.id, e.kind_ids, e.properties)::nodeComposite as e from node as s join edge r on r.start_id = s.id join node e on e.id = r.end_id where (s.properties->>'value')::text = 'PII' and (r.properties->'other')::int8 = 234 and (e.properties->'that')::int8 = 456 and (s.properties->>'other')::text = 'more pii' and (e.properties->'number')::int8 = 411",
		},
		{
			Source:   "match (s)-[r:EdgeKindA|EdgeKindB]->(e) return s.name, e.name",
			Expected: "select s.properties->'name' as \"s.name\", e.properties->'name' as \"e.name\" from node as s join edge r on r.start_id = s.id join node e on e.id = r.end_id where r.kind_id = any(array[100, 101]::int2[])",
		},
		{
			Source:   "match (s)-[:EdgeKindA|EdgeKindB]->(e)-[:EdgeKindA]->() return s.name, e.name",
			Expected: "select s.properties->'name' as \"s.name\", e.properties->'name' as \"e.name\" from node as s join edge e0 on e0.start_id = s.id join node e on e.id = e0.end_id join edge e1 on e1.start_id = e.id join node n2 on n2.id = e1.end_id where e0.kind_id = any(array[100, 101]::int2[]) and e1.kind_id = any(array[100]::int2[])",
		},
		{
			Source:   "match (s:NodeKindA)-[r:EdgeKindA|EdgeKindB]->(e:NodeKindB) return s.name, e.name",
			Expected: "select s.properties->'name' as \"s.name\", e.properties->'name' as \"e.name\" from node as s join edge r on r.start_id = s.id join node e on e.id = r.end_id where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] and r.kind_id = any(array[100, 101]::int2[]) and e.kind_ids operator(pg_catalog.&&) array[2]::int2[]",
		},
		{
			Source:   "match (s) where s.name = '123' return s.name",
			Expected: "select s.properties->'name' as \"s.name\" from node as s where (s.properties->>'name')::text = '123'",
		},
		{
			Source:   "match (s:NodeKindA), (o:NodeKindB) where s.objectid = '123' and o.linked = s.linkid return o",
			Expected: "select (o.id, o.kind_ids, o.properties)::nodeComposite as o from node as s, node as o where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] and o.kind_ids operator(pg_catalog.&&) array[2]::int2[] and (s.properties->>'objectid')::text = '123' and o.properties->'linked' = s.properties->'linkid'",
		},
		{
			Source:   "match (s) where s.name in ['option 1', 'option 2'] return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'name')::text in array['option 1', 'option 2']",
		},
		{
			Source:   "match (s) where id(s) in [1, 2, 3, 4] return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where s.id in array[1, 2, 3, 4]",
		},
		{
			Source:   "match (s) where s.created_at = localtime() return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::time without time zone = localtime",
		},
		{
			Source:   "match (s) where s.created_at = localtime('12:12:12') return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::time without time zone = '12:12:12'::time without time zone",
		},
		{
			Source:   "match (s) where s.created_at = date() return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::date = current_date",
		},
		{
			Source:   "match (s) where s.created_at = date('2023-12-12') return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::date = '2023-12-12'::date",
		},
		{
			Source:   "match (s) where s.created_at = datetime() return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::timestamp with time zone = now()",
		},
		{
			Source:   "match (s) where s.name = '1234' return count(s) as num",
			Expected: "select count(s) as num from node as s where (s.properties->>'name')::text = '1234'",
		},
		{
			Source:   "match (s) where s.created_at = datetime('2019-06-01T18:40:32.142+0100') return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::timestamp with time zone = '2019-06-01T18:40:32.142+0100'::timestamp with time zone",
		},
		{
			Source:   "match (s) where not (s.name = '123') return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where not ((s.properties->>'name')::text = '123')",
		},
		{
			Source:   "match (s) where s.created_at = localdatetime() return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::timestamp without time zone = localtimestamp",
		},
		{
			Source:   "match (s) where s.created_at = localdatetime('2019-06-01T18:40:32.142') return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'created_at')::timestamp without time zone = '2019-06-01T18:40:32.142'::timestamp without time zone",
		},
		{
			Source: "match (n) where n.name = 'w' return n.name",
		},
		{
			Source:   "match (s) where s.created_at is null return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where s.properties ? 'created_at'",
		},
		{
			Source:   "match (s) where s.created_at is not null return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where not s.properties ? 'created_at'",
		},
		{
			Source:   "match (s) where s:NodeKindA return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where s.kind_ids operator(pg_catalog.&&) array[1]::int2[]",
		},
		{
			Source:   "match (s) where s.name starts with '123' return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'name')::text like '123%'",
		},
		{
			Source:   "match (s) where s.name contains '123' return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'name')::text like '%123%'",
		},
		{
			Source:   "match (s) where s.name ends with '123' return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where (s.properties->>'name')::text like '%123'",
		},
		{
			Source:   "match (s) where s:NodeKindA return s",
			Expected: "select (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where s.kind_ids operator(pg_catalog.&&) array[1]::int2[]",
		},
		{
			Source:   "match (s) where s:NodeKindA return distinct s",
			Expected: "select distinct (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where s.kind_ids operator(pg_catalog.&&) array[1]::int2[]",
		},
		{
			Source:   "match (s) where toLower(s.name) = '1234' return distinct s",
			Expected: "select distinct (s.id, s.kind_ids, s.properties)::nodeComposite as s from node as s where lower((s.properties->>'name')::text) = '1234'",
		},
		{
			Source:   "match (s) where s.name = '1234' return labels(s)",
			Expected: "select s.kind_ids as \"s.kind_ids\" from node as s where (s.properties->>'name')::text = '1234'",
		},
		{
			Source:   "match ()-[r]->() where r.name = '1234' return type(r)",
			Expected: "select r.kind_id as \"r.kind_id\" from node as n0 join edge r on r.start_id = n0.id join node n1 on n1.id = r.end_id where (r.properties->>'name')::text = '1234'",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.InIDs(query.Node(), 1, 2, 3),
				)),
				query.Returning(query.Node()),
			),
			Expected: "select (n.id, n.kind_ids, n.properties)::nodeComposite as n from node as n where n.id = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.In(query.NodeProperty("prop"), []string{"1", "2", "3"}),
				)),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->>'prop')::text = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.In(query.NodeProperty("prop"), []int16{1, 2, 3}),
				)),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->'prop')::int2 = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.In(query.NodeProperty("prop"), []int32{1, 2, 3}),
				)),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->'prop')::int4 = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.In(query.NodeProperty("prop"), []int64{1, 2, 3}),
				)),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->'prop')::int8 = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(query.And(
					query.In(query.NodeProperty("prop"), []float32{1, 2, 3}),
				)),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->'prop')::float4 = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(
					query.And(
						query.In(query.NodeProperty("prop"), []float64{1, 2, 3}),
					),
				),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where (n.properties->'prop')::float8 = any(@p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(
					query.And(
						query.KindIn(query.Relationship(), graph.StringKind("EdgeKindA"), graph.StringKind("EdgeKindB")),
						query.Or(
							query.Not(query.Exists(query.RelationshipProperty(common.LastSeen.String()))),
							query.Before(query.RelationshipProperty(common.LastSeen.String()), time.Date(2023, time.August, 01, 0, 0, 0, 0, time.Local)),
						),
					),
				),
				query.Returning(query.Relationship()),
			),
			Expected: "select (r.id, r.start_id, r.end_id, r.kind_id, r.properties)::edgeComposite as r from node as n0 join edge r on r.start_id = n0.id join node n1 on n1.id = r.end_id where (r.kind_id = any(array[100]::int2[]) or r.kind_id = any(array[101]::int2[])) and (not r.properties ? 'lastseen' or (r.properties->>'lastseen')::timestamp with time zone < @p0)",
		},
		{
			Query: query.NewBuilderWithCriteria(
				query.Where(
					query.And(
						query.Kind(query.Node(), graph.StringKind("NodeKindA")),
						query.Or(
							query.StringContains(query.NodeProperty("name"), "name"),
							query.StringContains(query.NodeProperty("objectid"), "name"),
						),
						query.Not(query.Equals(query.NodeProperty("name"), "name")),
						query.Not(query.Equals(query.NodeProperty("objectid"), "name")),
						query.Not(
							query.And(
								query.Kind(query.Node(), graph.StringKind("NodeKindB")),
								query.Not(query.Kind(query.Node(), graph.StringKind("NodeKindC"))),
							),
						),
					),
				),
				query.Returning(query.NodeID()),
			),
			Expected: "select n.id as \"n.id\" from node as n where n.kind_ids operator(pg_catalog.&&) array[1]::int2[] and ((n.properties->>'name')::text like @p0 or (n.properties->>'objectid')::text like @p1) and not (n.properties->>'name')::text = @p2 and not (n.properties->>'objectid')::text = @p3 and not n.kind_ids operator(pg_catalog.&&) array[2]::int2[] and not n.kind_ids operator(pg_catalog.&&) array[3]::int2[]",
			ExpectedParameters: map[string]any{
				"p0": "%name%",
				"p1": "%name%",
				"p2": "name",
				"p3": "name",
			},
		},

		// UPDATE CASES
		{
			Source:   "match (s) where s:NodeKindA set s:NodeKindB return s",
			Expected: "update node as s set kind_ids = kind_ids || @p0 where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] returning (s.id, s.kind_ids, s.properties)::nodeComposite as s",
		},
		{
			Source:   "match (s) where s:NodeKindA set s:NodeKindB remove s:NodeKindA return s",
			Expected: "update node as s set kind_ids = kind_ids - @p1 || @p0 where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] returning (s.id, s.kind_ids, s.properties)::nodeComposite as s",
		},
		{
			Source:   "match (s) set s.name = 'new name', s:NodeKindA return s",
			Expected: "update node as s set properties = properties || @p0, kind_ids = kind_ids || @p1 returning (s.id, s.kind_ids, s.properties)::nodeComposite as s",
		},
		{
			Source:   "match (s) where s:NodeKindA set s.name = 'new name' return s",
			Expected: "update node as s set properties = properties || @p0 where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] returning (s.id, s.kind_ids, s.properties)::nodeComposite as s",
		},
		{
			Source:   "match (s) where s:NodeKindA set s.name = 'lol' remove s.other return s",
			Expected: "update node as s set properties = properties - @p1 || @p0 where s.kind_ids operator(pg_catalog.&&) array[1]::int2[] returning (s.id, s.kind_ids, s.properties)::nodeComposite as s",
		},

		// ERROR CASES

		// Mixed types in a list match should fail. Once a field type is set there must be no ambiguity.
		{
			Source: "match (s) where s.name in ['option 1', 'option 2', 1234] return s",
			Error:  true,
		},

		// UNSUPPORTED CASES

		// The following queries are going to require running each match as a distinct select statements with a left
		// outer join to combine result sets. This is pretty ill-defined and a stupid feature if you ask me, so I'm
		// going to leave it out for now.
		{
			Source:  "match (s), (e)-[]->(o) where s.name = '123' and e.name = 'lol' return s.name, e, o",
			Ignored: true,
		},
		{
			Source:  "match (s) where s.name = '123' match (e) where e.name = 'lol' return s.name, e",
			Ignored: true,
		},
	}
}

func TestPGSQLEmitter(t *testing.T) {
	var (
		runnable     []TestCase
		exclusiveRun bool
	)

	for idx, testCase := range Suite() {
		testCase.ID = idx

		if testCase.Ignored {
			continue
		}

		if testCase.Exclusive {
			if !exclusiveRun {
				runnable = runnable[:0]
				exclusiveRun = true
			}

			runnable = append(runnable, testCase)
		} else if !exclusiveRun {
			runnable = append(runnable, testCase)
		}
	}

	for _, testCase := range runnable {
		var regularQuery *model.RegularQuery

		if testCase.Query != nil {
			builtQuery, err := testCase.Query.Build()
			require.Nilf(t, err, "test case %d: %v", testCase.ID, err)

			regularQuery = builtQuery

		} else {
			parsedQuery, parseErr := frontend.ParseCypher(frontend.NewContext(), testCase.Source)
			require.Nilf(t, parseErr, "test case %d: %v", testCase.ID, parseErr)

			regularQuery = parsedQuery
		}

		var (
			buffer     = &bytes.Buffer{}
			kindMapper = KindMapper{
				known: map[string]int16{
					"NodeKindA": 1,
					"NodeKindB": 2,
					"NodeKindC": 3,
					"EdgeKindA": 100,
					"EdgeKindB": 101,
					"EdgeKindC": 102,
				},
			}

			parameters, translationErr = pgsql.Translate(regularQuery, kindMapper)
		)

		if testCase.Error {
			if translationErr != nil {
				continue
			}

			var (
				emitter    = pgsql.NewEmitter(false, kindMapper)
				emitterErr = emitter.Write(regularQuery, buffer)
			)

			require.NotNilf(t, emitterErr, "test case %d: %v", testCase.ID, emitterErr)
		} else {
			require.Nilf(t, translationErr, "test case %d: %v", testCase.ID, translationErr)

			if testCase.ExpectedParameters != nil {
				require.Equal(t, testCase.ExpectedParameters, parameters)
			}

			var (
				emitter    = pgsql.NewEmitter(false, kindMapper)
				emitterErr = emitter.Write(regularQuery, buffer)
			)

			require.Nilf(t, emitterErr, "test case %d: %v", testCase.ID, emitterErr)
			require.Equalf(t, testCase.Expected, buffer.String(), "test case %d", testCase.ID)
		}
	}
}

func TestBinder(t *testing.T) {
	var (
		binder                 = pgsql.NewBinder()
		regularQuery, parseErr = frontend.ParseCypher(frontend.DefaultCypherContext(), "match (s) with s as m return s")
		binderErr              = binder.Scan(regularQuery)
	)

	require.Nil(t, parseErr)
	require.Nil(t, binderErr)

	require.True(t, binder.IsBound("s"))
	require.True(t, binder.IsPatternBinding("s"))
	require.True(t, binder.IsBound("m"))

	// TODO: This might want to be true depending on how references play out during joins
	require.False(t, binder.IsPatternBinding("m"))
}

//func TestCypherEmitter_HappyPath(t *testing.T) {
//	test.LoadFixture(t, test.PositiveTestCases).Run(t)
//}
//
//func TestCypherEmitter_NegativeCases(t *testing.T) {
//	test.LoadFixture(t, test.NegativeTestCases).Run(t)
//}
