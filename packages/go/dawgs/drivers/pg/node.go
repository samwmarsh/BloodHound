package pg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/specterops/bloodhound/cypher/gen/pgsql"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

var (
	ErrAmbiguousQueryVariables = errors.New("query mixes node and relationship query variables")
)

func RemoveEmptyExpressionLists(stack *model.WalkStack, element model.Expression) error {
	var (
		shouldRemove  = false
		shouldReplace = false

		replacementExpression model.Expression
	)

	switch typedElement := element.(type) {
	case model.ExpressionList:
		shouldRemove = typedElement.Len() == 0

	case *model.Parenthetical:
		switch typedParentheticalElement := typedElement.Expression.(type) {
		case model.ExpressionList:
			numExpressions := typedParentheticalElement.Len()

			shouldRemove = numExpressions == 0
			shouldReplace = numExpressions == 1

			if shouldReplace {
				// Dump the parenthetical and the joined expression by grabbing the only element in the joined
				// expression for replacement
				replacementExpression = typedParentheticalElement.Get(0)
			}
		}
	}

	if shouldRemove {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Remove(element)
		}
	} else if shouldReplace {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Replace(typedParent.IndexOf(element), replacementExpression)
		}
	}

	return nil
}

type queryBuilder struct {
	Parameters map[string]any

	kindMapper pgsql.KindMapper
	query      *model.RegularQuery
}

func newQueryBuilder(kindMapper pgsql.KindMapper) *queryBuilder {
	return &queryBuilder{
		kindMapper: kindMapper,
		query: &model.RegularQuery{
			SingleQuery: &model.SingleQuery{
				SinglePartQuery: &model.SinglePartQuery{},
			},
		},
	}
}

func (s *queryBuilder) Prepare() error {
	if err := s.prepareMatch(); err != nil {
		return err
	}

	if parameters, err := pgsql.Translate(s.query, s.kindMapper); err != nil {
		return err
	} else {
		s.Parameters = parameters
	}

	// TODO: Move this into translation
	return model.Walk(s.query, model.NewVisitor(nil, RemoveEmptyExpressionLists), pgsql.CollectPGSQLTypes)
}

func (s *queryBuilder) Render() (string, error) {
	var (
		buffer = &bytes.Buffer{}
		err    = pgsql.NewEmitter(false, s.kindMapper).Write(s.query, buffer)
	)

	return buffer.String(), err
}

func (s *queryBuilder) prepareMatch() error {
	var (
		patternPart = &model.PatternPart{}

		singleNodeBound    = false
		creatingSingleNode = false

		startNodeBound       = false
		creatingStartNode    = false
		endNodeBound         = false
		creatingEndNode      = false
		relationshipBound    = false
		creatingRelationship = false

		isRelationshipQuery = false

		bindWalk = model.NewVisitor(func(stack *model.WalkStack, branch model.Expression) error {
			switch typedElement := branch.(type) {
			case *model.Variable:
				switch typedElement.Symbol {
				case query.NodeSymbol:
					singleNodeBound = true

				case query.RelationshipStartSymbol:
					startNodeBound = true
					isRelationshipQuery = true

				case query.RelationshipEndSymbol:
					endNodeBound = true
					isRelationshipQuery = true

				case query.RelationshipSymbol:
					relationshipBound = true
					isRelationshipQuery = true
				}
			}

			return nil
		}, nil)
	)

	// Zip through updating clauses first
	for _, updatingClause := range s.query.SingleQuery.SinglePartQuery.UpdatingClauses {
		typedUpdatingClause, isUpdatingClause := updatingClause.(*model.UpdatingClause)

		if !isUpdatingClause {
			return fmt.Errorf("unexpected type for updating clause: %T", updatingClause)
		}

		switch typedClause := typedUpdatingClause.Clause.(type) {
		case *model.Create:
			if err := model.Walk(typedClause, model.NewVisitor(func(stack *model.WalkStack, element model.Expression) error {
				switch typedElement := element.(type) {
				case *model.NodePattern:
					if patternBinding, typeOK := typedElement.Binding.(*model.Variable); !typeOK {
						return fmt.Errorf("expected variable for pattern binding but got: %T", typedElement.Binding)
					} else {
						switch patternBinding.Symbol {
						case query.NodeSymbol:
							creatingSingleNode = true

						case query.RelationshipStartSymbol:
							creatingStartNode = true

						case query.RelationshipEndSymbol:
							creatingEndNode = true
						}
					}

				case *model.RelationshipPattern:
					if patternBinding, typeOK := typedElement.Binding.(*model.Variable); !typeOK {
						return fmt.Errorf("expected variable for pattern binding but got: %T", typedElement.Binding)
					} else {
						switch patternBinding.Symbol {
						case query.RelationshipSymbol:
							creatingRelationship = true
						}
					}
				}

				return nil
			}, nil)); err != nil {
				return err
			}

		case *model.Delete:
			if err := model.Walk(typedClause, bindWalk, nil); err != nil {
				return err
			}
		}
	}

	// Is there a where clause?
	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil && firstReadingClause.Match.Where != nil {
		if err := model.Walk(firstReadingClause.Match.Where, bindWalk, nil); err != nil {
			return err
		}
	}

	// Is there a return clause
	if s.query.SingleQuery.SinglePartQuery.Return != nil {
		if err := model.Walk(s.query.SingleQuery.SinglePartQuery.Return, bindWalk, nil); err != nil {
			return err
		}
	}

	// Validate we're not mixing references
	if isRelationshipQuery && singleNodeBound {
		return ErrAmbiguousQueryVariables
	}

	if singleNodeBound && !creatingSingleNode {
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.NodeSymbol),
		})
	}

	if startNodeBound {
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.RelationshipStartSymbol),
		})
	}

	if isRelationshipQuery {
		if !startNodeBound && !creatingStartNode {
			patternPart.AddPatternElements(&model.NodePattern{})
		}

		if !creatingRelationship {
			if relationshipBound {
				patternPart.AddPatternElements(&model.RelationshipPattern{
					Binding:   model.NewVariableWithSymbol(query.RelationshipSymbol),
					Direction: graph.DirectionOutbound,
				})
			} else {
				patternPart.AddPatternElements(&model.RelationshipPattern{
					Direction: graph.DirectionOutbound,
				})
			}
		}

		if !endNodeBound && !creatingEndNode {
			patternPart.AddPatternElements(&model.NodePattern{})
		}
	}

	if endNodeBound {
		patternPart.AddPatternElements(&model.NodePattern{
			Binding: model.NewVariableWithSymbol(query.RelationshipEndSymbol),
		})
	}

	if firstReadingClause := query.GetFirstReadingClause(s.query); firstReadingClause != nil {
		firstReadingClause.Match.Pattern = []*model.PatternPart{patternPart}
	} else if len(patternPart.PatternElements) > 0 {
		s.query.SingleQuery.SinglePartQuery.AddReadingClause(&model.ReadingClause{
			Match: &model.Match{
				Pattern: []*model.PatternPart{
					patternPart,
				},
			},
		})
	}

	return nil
}

func (s *queryBuilder) Apply(criteria graph.Criteria) {
	switch typedCriteria := criteria.(type) {
	case *model.Where:
		if query.GetFirstReadingClause(s.query) == nil {
			s.query.SingleQuery.SinglePartQuery.AddReadingClause(&model.ReadingClause{
				Match: model.NewMatch(false),
			})
		}

		query.GetFirstReadingClause(s.query).Match.Where = model.Copy(typedCriteria)

	case *model.Return:
		s.query.SingleQuery.SinglePartQuery.Return = typedCriteria

	case *model.Limit:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Limit = model.Copy(typedCriteria)
		}

	case *model.Skip:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Skip = model.Copy(typedCriteria)
		}

	case *model.Order:
		if s.query.SingleQuery.SinglePartQuery.Return != nil {
			s.query.SingleQuery.SinglePartQuery.Return.Projection.Order = model.Copy(typedCriteria)
		}

	case []*model.UpdatingClause:
		for _, updatingClause := range typedCriteria {
			s.Apply(updatingClause)
		}

	case *model.UpdatingClause:
		s.query.SingleQuery.SinglePartQuery.AddUpdatingClause(model.Copy(typedCriteria))

	default:
		panic(fmt.Sprintf("invalid type for dawgs query: %T %+v", criteria, criteria))
	}
}

type liveQuery struct {
	ctx          context.Context
	tx           *transaction
	queryBuilder *queryBuilder
}

func (s *liveQuery) run(statement string, parameters map[string]any) graph.Result {
	return s.tx.Run(statement, parameters)
}

func (s *liveQuery) Execute(delegate func(results graph.Result) error, finalCriteria ...graph.Criteria) error {
	for _, criteria := range finalCriteria {
		s.queryBuilder.Apply(criteria)
	}

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else if result := s.run(statement, s.queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
		return delegate(result)
	}
}

func (s *liveQuery) Debug() (string, map[string]any) {
	statement, _ := s.queryBuilder.Render()
	return statement, s.queryBuilder.Parameters
}

type nodeQuery struct {
	liveQuery
}

func (s *nodeQuery) Filter(criteria graph.Criteria) graph.NodeQuery {
	s.queryBuilder.Apply(query.Where(criteria))
	return s
}

func (s *nodeQuery) Filterf(criteriaDelegate graph.CriteriaProvider) graph.NodeQuery {
	return s.Filter(criteriaDelegate())
}
func (s *nodeQuery) Delete() error {
	s.queryBuilder.Apply(query.Delete(
		query.Node(),
	))

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else if result := s.run(statement, s.queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
	}

	return nil
}

func (s *nodeQuery) Update(properties *graph.Properties) error {
	s.queryBuilder.Apply(query.Updatef(func() graph.Criteria {
		var updateStatements []graph.Criteria

		if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
			updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
		}

		if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
			updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
		}

		return updateStatements
	}))

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if cypherQuery, err := s.queryBuilder.Render(); err != nil {
		return graph.NewError(cypherQuery, err)
	} else if result := s.run(cypherQuery, s.queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	} else {
		defer result.Close()
	}

	return nil
}

func (s *nodeQuery) OrderBy(criteria ...graph.Criteria) graph.NodeQuery {
	s.queryBuilder.Apply(query.OrderBy(criteria...))
	return s
}

func (s *nodeQuery) Offset(offset int) graph.NodeQuery {
	s.queryBuilder.Apply(query.Offset(offset))
	return s
}

func (s *nodeQuery) Limit(limit int) graph.NodeQuery {
	s.queryBuilder.Apply(query.Limit(limit))
	return s
}

func (s *nodeQuery) Count() (int64, error) {
	var count int64

	return count, s.Execute(func(results graph.Result) error {
		if !results.Next() {
			return graph.ErrNoResultsFound
		}

		return results.Scan(&count)
	}, query.Returning(
		query.Count(query.Node()),
	))
}

func (s *nodeQuery) First() (*graph.Node, error) {
	var node graph.Node

	return &node, s.Execute(
		func(results graph.Result) error {
			if !results.Next() {
				return graph.ErrNoResultsFound
			}

			return results.Scan(&node)
		},
		query.Returning(
			query.Node(),
		),
		query.Limit(1),
	)
}

func (s *nodeQuery) Fetch(delegate func(cursor graph.Cursor[*graph.Node]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (*graph.Node, error) {
			var node graph.Node
			return &node, scanner.Scan(&node)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.Node(),
	))
}

func (s *nodeQuery) FetchIDs(delegate func(cursor graph.Cursor[graph.ID]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.ID, error) {
			var nodeID graph.ID
			return nodeID, scanner.Scan(&nodeID)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.NodeID(),
	))
}

func (s *nodeQuery) FetchKinds(delegate func(cursor graph.Cursor[graph.KindsResult]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.KindsResult, error) {
			var (
				nodeID    graph.ID
				nodeKinds graph.Kinds
				err       = scanner.Scan(&nodeID, &nodeKinds)
			)

			return graph.KindsResult{
				ID:    nodeID,
				Kinds: nodeKinds,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.NodeID(),
		query.KindsOf(query.Node()),
	))
}
