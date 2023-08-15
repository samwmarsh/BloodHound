package pg

import (
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

func directionToReturnCriteria(direction graph.Direction) (graph.Criteria, error) {
	switch direction {
	case graph.DirectionInbound:
		// Select the relationship and the end node
		return query.Returning(
			query.Relationship(),
			query.End(),
		), nil

	case graph.DirectionOutbound:
		// Select the relationship and the start node
		return query.Returning(
			query.Relationship(),
			query.Start(),
		), nil

	default:
		return nil, fmt.Errorf("bad direction: %d", direction)
	}
}

type relationshipQuery struct {
	liveQuery
}

func (s *relationshipQuery) Filter(criteria graph.Criteria) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Where(criteria))
	return s
}

func (s *relationshipQuery) Filterf(criteriaDelegate graph.CriteriaProvider) graph.RelationshipQuery {
	return s.Filter(criteriaDelegate())
}

func (s *relationshipQuery) Delete() error {
	s.queryBuilder.Apply(query.Delete(
		query.Node(),
	))

	if err := s.queryBuilder.Prepare(); err != nil {
		return err
	} else if statement, err := s.queryBuilder.Render(); err != nil {
		return err
	} else {
		result := s.run(statement, s.queryBuilder.Parameters)
		return result.Error()
	}
}

func (s *relationshipQuery) Update(properties *graph.Properties) error {
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
	}

	return nil
}

func (s *relationshipQuery) OrderBy(criteria ...graph.Criteria) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.OrderBy(criteria...))
	return s
}

func (s *relationshipQuery) Offset(offset int) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Offset(offset))
	return s
}

func (s *relationshipQuery) Limit(limit int) graph.RelationshipQuery {
	s.queryBuilder.Apply(query.Limit(limit))
	return s
}

func (s *relationshipQuery) Count() (int64, error) {
	var count int64

	return count, s.Execute(func(results graph.Result) error {
		if !results.Next() {
			return graph.ErrNoResultsFound
		}

		return results.Scan(&count)
	}, query.Returning(
		query.Count(query.Relationship()),
	))
}

func (s *relationshipQuery) FetchAllShortestPaths(delegate func(cursor graph.Cursor[graph.Path]) error) error {
	panic("not supported")
	//s.queryBuilder.Apply(query.Returning(
	//	query.Path(),
	//))
	//
	//if err := s.queryBuilder.PrepareAllShortestPaths(); err != nil {
	//	return err
	//} else if statement, err := s.queryBuilder.Render(); err != nil {
	//	return err
	//} else if result := s.run(statement, s.queryBuilder.Parameters); result.Error() != nil {
	//	return result.Error()
	//} else {
	//	defer result.Close()
	//
	//	cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.Path, error) {
	//		var (
	//			nextPath graph.Path
	//			err      = scanner.Scan(&nextPath)
	//		)
	//
	//		return nextPath, err
	//	})
	//
	//	defer cursor.Close()
	//	return delegate(cursor)
	//}
}

func (s *relationshipQuery) FetchTriples(delegate func(cursor graph.Cursor[graph.RelationshipTripleResult]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.RelationshipTripleResult, error) {
			var (
				startID        graph.ID
				relationshipID graph.ID
				endID          graph.ID
				err            = scanner.Scan(&startID, &relationshipID, &endID)
			)

			return graph.RelationshipTripleResult{
				ID:      relationshipID,
				StartID: startID,
				EndID:   endID,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.ReturningDistinct(
		query.StartID(),
		query.RelationshipID(),
		query.EndID(),
	))
}

func (s *relationshipQuery) FetchKinds(delegate func(cursor graph.Cursor[graph.RelationshipKindsResult]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.RelationshipKindsResult, error) {
			var (
				startID          graph.ID
				relationshipID   graph.ID
				relationshipKind graph.Kind
				endID            graph.ID
				err              = scanner.Scan(&startID, &relationshipID, &relationshipKind, &endID)
			)

			return graph.RelationshipKindsResult{
				RelationshipTripleResult: graph.RelationshipTripleResult{
					ID:      relationshipID,
					StartID: startID,
					EndID:   endID,
				},
				Kind: relationshipKind,
			}, err
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.StartID(),
		query.RelationshipID(),
		query.KindsOf(query.Relationship()),
		query.EndID(),
	))
}

func (s *relationshipQuery) First() (*graph.Relationship, error) {
	var relationship graph.Relationship

	return &relationship, s.Execute(
		func(results graph.Result) error {
			if !results.Next() {
				return graph.ErrNoResultsFound
			}

			return results.Scan(&relationship)
		},
		query.Returning(
			query.Relationship(),
		),
		query.Limit(1),
	)
}

func (s *relationshipQuery) Fetch(delegate func(cursor graph.Cursor[*graph.Relationship]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (*graph.Relationship, error) {
			var relationship graph.Relationship
			return &relationship, scanner.Scan(&relationship)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.Relationship(),
	))
}

func (s *relationshipQuery) FetchDirection(direction graph.Direction, delegate func(cursor graph.Cursor[graph.DirectionalResult]) error) error {
	if returnCriteria, err := directionToReturnCriteria(direction); err != nil {
		return err
	} else {
		return s.Execute(func(result graph.Result) error {
			cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.DirectionalResult, error) {
				var (
					relationship graph.Relationship
					node         graph.Node
				)

				if err := scanner.Scan(&relationship, &node); err != nil {
					return graph.DirectionalResult{}, err
				}

				return graph.DirectionalResult{
					Direction:    direction,
					Relationship: &relationship,
					Node:         &node,
				}, nil
			})

			defer cursor.Close()
			return delegate(cursor)
		}, returnCriteria)
	}
}

func (s *relationshipQuery) FetchIDs(delegate func(cursor graph.Cursor[graph.ID]) error) error {
	return s.Execute(func(result graph.Result) error {
		cursor := graph.NewResultIterator(s.ctx, result, func(scanner graph.Scanner) (graph.ID, error) {
			var relationshipID graph.ID
			return relationshipID, scanner.Scan(&relationshipID)
		})

		defer cursor.Close()
		return delegate(cursor)
	}, query.Returning(
		query.RelationshipID(),
	))
}
