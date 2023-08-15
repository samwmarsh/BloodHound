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

package harness

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	pgQuery "github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDriver_Run(t *testing.T) {
	driver, err := dawgs.Open(pg.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            "user=bhe dbname=bhe password=bhe4eva host=localhost",
	})

	require.Nil(t, err)

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return pgQuery.On(tx).DropSchema()
	}, pg.OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol)))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return pgQuery.On(tx).CreateSchema()
	}, pg.OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol)))

	require.Nil(t, driver.AssertSchema(context.Background(), CurrentSchema()))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		// Scope to an AD graph
		tx = tx.WithGraph(ActiveDirectoryGraphSchema("ad_graph"))

		if domainNode, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"name":      "user",
			"objectid":  "12345",
			"domainsid": "12345",
		}), ad.Entity, ad.Domain); err != nil {
			return err
		} else if userNode, err := tx.CreateNode(graph.AsProperties(map[string]any{
			"name":      "user",
			"objectid":  "12345",
			"domainsid": "12345",
		}), ad.Entity, ad.User); err != nil {
			return err
		} else if edge, err := tx.CreateRelationshipByIDs(domainNode.ID, userNode.ID, ad.Contains, graph.NewProperties()); err != nil {
			return err
		} else {
			domainNode.Properties.Set("other_prop", "lol")

			userNode.Properties.Set("is_bad", true)
			userNode.AddKinds(ad.Computer)

			edge.Properties.Set("thing", "yes")

			require.Nil(t, tx.UpdateNode(domainNode))
			require.Nil(t, tx.UpdateNode(userNode))
			require.Nil(t, tx.UpdateRelationship(edge))
		}

		return nil
	}))

	require.Nil(t, driver.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		if count, err := tx.Nodes().Filter(query.And(
			query.Equals(query.NodeProperty("objectid"), "12345"),
			query.Equals(query.NodeProperty("is_bad"), true),
		)).Count(); err != nil {
			t.Fatalf("Unexpected error during node count query: %s", err.Error())
		} else {
			require.Equal(t, int64(1), count)
		}

		if node, err := tx.Nodes().Filter(query.And(
			query.Equals(query.NodeProperty("objectid"), "12345"),
			query.Equals(query.NodeProperty("is_bad"), true),
		)).First(); err != nil {
			t.Fatalf("Unexpected error during node query: %s", err.Error())
		} else {
			require.Equal(t, "user", node.Properties.Get("name").Any())
			require.True(t, node.Kinds.ContainsOneOf(ad.Computer))

			if count, err := tx.Relationships().Filter(query.And(
				query.Equals(query.EndID(), node.ID),
			)).Count(); err != nil {
				t.Fatalf("Unexpected error during relationship query: %s", err.Error())
			} else {
				require.Equal(t, int64(1), count)
			}

			if edge, err := tx.Relationships().Filter(query.And(
				query.Equals(query.EndID(), node.ID),
			)).First(); err != nil {
				t.Fatalf("Unexpected error during relationship query: %s", err.Error())
			} else {
				require.Equal(t, ad.Contains, edge.Kind)
			}

			require.Nil(t, tx.Relationships().Filter(query.And(
				query.Equals(query.EndID(), node.ID),
			)).FetchDirection(graph.DirectionInbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
				for next := range cursor.Chan() {
					require.Equal(t, node.ID, next.Node.ID)
					require.True(t, next.Node.Kinds.ContainsOneOf(ad.User))
				}

				return cursor.Error()
			}))

			require.Nil(t, tx.Relationships().Filter(query.And(
				query.Equals(query.EndID(), node.ID),
			)).FetchDirection(graph.DirectionOutbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
				for next := range cursor.Chan() {
					require.True(t, next.Node.Kinds.ContainsOneOf(ad.Domain))
				}

				return cursor.Error()
			}))
		}

		return nil
	}))
}
