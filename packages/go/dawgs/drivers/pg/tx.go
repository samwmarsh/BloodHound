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

package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cypher/gen/pgsql"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type transaction struct {
	schemaManager   *SchemaManager
	queryExecMode   pgx.QueryExecMode
	ctx             context.Context
	tx              pgx.Tx
	targetSchema    graph.Graph
	targetSchemaSet bool
}

func newTransaction(ctx context.Context, conn *pgxpool.Conn, schemaManager *SchemaManager, cfg *Config) (*transaction, error) {
	if pgxTx, err := conn.BeginTx(ctx, cfg.Options); err != nil {
		return nil, err
	} else {
		return &transaction{
			schemaManager:   schemaManager,
			queryExecMode:   cfg.QueryExecMode,
			ctx:             ctx,
			tx:              pgxTx,
			targetSchemaSet: false,
		}, nil
	}
}

func (s *transaction) TraversalMemoryLimit() size.Size {
	return size.Gibibyte
}

func (s *transaction) WithGraph(schema graph.Graph) graph.Transaction {
	s.targetSchema = schema
	s.targetSchemaSet = true

	return s
}

func (s *transaction) Close() {
	if s.tx != nil {
		s.tx.Rollback(s.ctx)
		s.tx = nil
	}
}

func (s *transaction) getTargetGraph() (model.Graph, error) {
	if !s.targetSchemaSet {
		// Look for a default graph target
		if defaultGraph, hasDefaultGraph := s.schemaManager.DefaultGraph(); !hasDefaultGraph {
			return model.Graph{}, fmt.Errorf("driver operation requires a graph target to be set")
		} else {
			return defaultGraph, nil
		}
	}

	return s.schemaManager.AssertGraph(s, s.targetSchema)
}

func (s *transaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, kinds); err != nil {
		return nil, err
	} else if propertiesJSONB, err := pgsql.PropertiesToJSONB(properties); err != nil {
		return nil, err
	} else {
		var (
			nodeID int32
			result = s.tx.QueryRow(s.ctx, createNodeStatement, s.queryExecMode, graphTarget.ID, kindIDSlice, propertiesJSONB)
		)

		if err := result.Scan(&nodeID); err != nil {
			return nil, err
		}

		return graph.NewNode(graph.ID(nodeID), properties, kinds...), nil
	}
}

func (s *transaction) UpdateNode(node *graph.Node) error {
	builder := newQueryBuilder(s.schemaManager)

	builder.Apply(
		query.Where(
			query.Equals(query.NodeID(), node.ID),
		),
	)

	builder.Apply(
		query.Updatef(func() graph.Criteria {
			var (
				properties       = node.Properties
				updateStatements []graph.Criteria
			)

			if addedKinds := node.AddedKinds; len(addedKinds) > 0 {
				updateStatements = append(updateStatements, query.AddKinds(query.Node(), addedKinds))
			}

			if deletedKinds := node.DeletedKinds; len(deletedKinds) > 0 {
				updateStatements = append(updateStatements, query.DeleteKinds(query.Node(), deletedKinds))
			}

			if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
				updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
			}

			if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
				updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
			}

			return updateStatements
		}),
	)

	if err := builder.Prepare(); err != nil {
		return err
	} else if sqlQuery, err := builder.Render(); err != nil {
		return graph.NewError(sqlQuery, err)
	} else if result := s.Run(sqlQuery, builder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
	}

	return nil
}

func (s *transaction) Nodes() graph.NodeQuery {
	return &nodeQuery{
		liveQuery{
			ctx:          s.ctx,
			tx:           s,
			queryBuilder: newQueryBuilder(s.schemaManager),
		},
	}
}

func (s *transaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, graph.Kinds{kind}); err != nil {
		return nil, err
	} else if propertiesJSONB, err := pgsql.PropertiesToJSONB(properties); err != nil {
		return nil, err
	} else {
		var (
			edgeID int32
			result = s.tx.QueryRow(s.ctx, createEdgeStatement, s.queryExecMode, graphTarget.ID, startNodeID, endNodeID, kindIDSlice[0], propertiesJSONB)
		)

		if err := result.Scan(&edgeID); err != nil {
			return nil, err
		}

		return graph.NewRelationship(graph.ID(edgeID), startNodeID, endNodeID, properties, kind), nil
	}
}

func (s *transaction) UpdateRelationship(relationship *graph.Relationship) error {
	var (
		modifiedProperties    = relationship.Properties.ModifiedProperties()
		deletedProperties     = relationship.Properties.DeletedProperties()
		numModifiedProperties = len(modifiedProperties)
		numDeletedProperties  = len(deletedProperties)

		statement string
		arguments []any
	)

	if numModifiedProperties > 0 {
		if jsonbArgument, err := pgsql.ValueToJSONB(modifiedProperties); err != nil {
			return err
		} else {
			arguments = append(arguments, jsonbArgument)
		}

		if numDeletedProperties > 0 {
			if textArrayArgument, err := pgsql.StringSliceToTextArray(deletedProperties); err != nil {
				return err
			} else {
				arguments = append(arguments, textArrayArgument)
			}

			statement = edgePropertySetAndDeleteStatement
		} else {
			statement = edgePropertySetOnlyStatement
		}
	} else if numDeletedProperties > 0 {
		if textArrayArgument, err := pgsql.StringSliceToTextArray(deletedProperties); err != nil {
			return err
		} else {
			arguments = append(arguments, textArrayArgument)
		}

		statement = edgePropertyDeleteOnlyStatement
	}

	_, err := s.tx.Exec(s.ctx, statement, append(arguments, relationship.ID)...)
	return err
}

func (s *transaction) Relationships() graph.RelationshipQuery {
	return &relationshipQuery{
		liveQuery{
			ctx:          s.ctx,
			tx:           s,
			queryBuilder: newQueryBuilder(s.schemaManager),
		},
	}
}

func (s *transaction) query(query string, parameters map[string]any) (pgx.Rows, error) {
	if parameters == nil || len(parameters) == 0 {
		return s.tx.Query(s.ctx, query, s.queryExecMode)
	}

	return s.tx.Query(s.ctx, query, s.queryExecMode, pgx.NamedArgs(parameters))
}

func (s *transaction) Run(query string, parameters map[string]any) graph.Result {
	if rows, err := s.query(query, parameters); err != nil {
		return queryError{
			err: err,
		}
	} else {
		return &queryResult{
			rows:       rows,
			kindMapper: s.schemaManager,
		}
	}
}

func (s *transaction) Commit() error {
	return s.tx.Commit(s.ctx)
}
