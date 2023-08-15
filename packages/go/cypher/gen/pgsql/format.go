package pgsql

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model"
	pgModel "github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"io"
	"strconv"
)

const strippedLiteral = "$STRIPPED"

type KindMapper interface {
	MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds)
}

type Emitter struct {
	StripLiterals bool
	kindMapper    KindMapper
}

func NewEmitter(stripLiterals bool, kindMapper KindMapper) *Emitter {
	return &Emitter{
		StripLiterals: stripLiterals,
		kindMapper:    kindMapper,
	}
}

func (s *Emitter) formatMapLiteral(output io.Writer, mapLiteral model.MapLiteral) error {
	if _, err := io.WriteString(output, "{"); err != nil {
		return err
	}

	first := true
	for key, subExpression := range mapLiteral {
		if !first {
			if _, err := io.WriteString(output, ", "); err != nil {
				return err
			}
		} else {
			first = false
		}

		if _, err := io.WriteString(output, key); err != nil {
			return err
		}

		if _, err := io.WriteString(output, ": "); err != nil {
			return err
		}

		if err := s.WriteExpression(output, subExpression); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(output, "}"); err != nil {
		return err
	}

	return nil
}

func (s *Emitter) formatLiteral(output io.Writer, literal *model.Literal) error {
	const literalNullToken = "null"

	// Check for a null literal first
	if literal.Null {
		if _, err := io.WriteString(output, literalNullToken); err != nil {
			return err
		}
		return nil
	}

	// Attempt to string format the literal value
	switch typedLiteral := literal.Value.(type) {
	case string:
		if _, err := io.WriteString(output, typedLiteral); err != nil {
			return err
		}

	case int8:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int16:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int32:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case int64:
		if _, err := io.WriteString(output, strconv.FormatInt(typedLiteral, 10)); err != nil {
			return err
		}

	case int:
		if _, err := io.WriteString(output, strconv.FormatInt(int64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint8:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint16:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint32:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case uint64:
		if _, err := io.WriteString(output, strconv.FormatUint(typedLiteral, 10)); err != nil {
			return err
		}

	case uint:
		if _, err := io.WriteString(output, strconv.FormatUint(uint64(typedLiteral), 10)); err != nil {
			return err
		}

	case bool:
		if _, err := io.WriteString(output, strconv.FormatBool(typedLiteral)); err != nil {
			return err
		}

	case float32:
		if _, err := io.WriteString(output, strconv.FormatFloat(float64(typedLiteral), 'f', -1, 64)); err != nil {
			return err
		}

	case float64:
		if _, err := io.WriteString(output, strconv.FormatFloat(typedLiteral, 'f', -1, 64)); err != nil {
			return err
		}

	case model.MapLiteral:
		if err := s.formatMapLiteral(output, typedLiteral); err != nil {
			return err
		}

	case *model.ListLiteral:
		if _, err := io.WriteString(output, "array["); err != nil {
			return err
		}

		for idx, subExpression := range *typedLiteral {
			if idx > 0 {
				if _, err := io.WriteString(output, ", "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(output, subExpression); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(output, "]"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unexpected literal type for string formatting: %T", literal)
	}

	return nil
}

func (s *Emitter) writeReturn(writer io.Writer, returnClause *model.Return) error {
	if returnClause.Projection.Distinct {
		if _, err := WriteStrings(writer, "distinct "); err != nil {
			return err
		}
	}

	for idx, projectionItem := range returnClause.Projection.Items {
		if idx > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return nil
			}
		}

		if err := s.WriteExpression(writer, projectionItem); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeWhere(writer io.Writer, whereClause *model.Where) error {
	if len(whereClause.Expressions) > 0 {
		if _, err := io.WriteString(writer, " where "); err != nil {
			return err
		}
	}

	for _, expression := range whereClause.Expressions {
		if err := s.WriteExpression(writer, expression); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writePatternElements(writer io.Writer, patternElements []*model.PatternElement) error {
	for idx, patternElement := range patternElements {
		if nodePattern, isNodePattern := patternElement.AsNodePattern(); isNodePattern {
			if idx == 0 {
				if _, err := io.WriteString(writer, pgModel.NodeTable); err != nil {
					return nil
				}

				if _, err := io.WriteString(writer, " as "); err != nil {
					return nil
				}

				if err := s.WriteExpression(writer, nodePattern.Binding); err != nil {
					return nil
				}
			} else {
				previousRelationshipPattern, _ := patternElements[idx-1].AsRelationshipPattern()

				if _, err := WriteStrings(writer, " join ", pgModel.NodeTable, " "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, nodePattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, " on "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, nodePattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".id = "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, previousRelationshipPattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".end_id"); err != nil {
					return err
				}
			}
		} else {
			relationshipPattern, _ := patternElement.AsRelationshipPattern()

			if idx == 0 {
				if _, err := io.WriteString(writer, pgModel.EdgeTable); err != nil {
					return nil
				}

				if _, err := io.WriteString(writer, " as "); err != nil {
					return nil
				}

				if err := s.WriteExpression(writer, relationshipPattern.Binding); err != nil {
					return nil
				}
			} else {
				previousNodePattern, _ := patternElements[idx-1].AsNodePattern()

				if _, err := WriteStrings(writer, " join ", pgModel.EdgeTable, " "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, relationshipPattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, " on "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, relationshipPattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".start_id = "); err != nil {
					return err
				}

				if err := s.WriteExpression(writer, previousNodePattern.Binding); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".id"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Emitter) writeMatch(writer io.Writer, matchClause *model.Match) error {
	for idx, pattern := range matchClause.Pattern {
		if idx > 0 {
			if _, err := io.WriteString(writer, ", "); err != nil {
				return err
			}
		}

		if err := s.writePatternElements(writer, pattern.PatternElements); err != nil {
			return err
		}
	}

	if matchClause.Where != nil {
		if err := s.writeWhere(writer, matchClause.Where); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeSelect(writer io.Writer, singlePartQuery *model.SinglePartQuery) error {
	if _, err := io.WriteString(writer, "select "); err != nil {
		return err
	}

	if singlePartQuery.Return != nil {
		if err := s.writeReturn(writer, singlePartQuery.Return); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(writer, " from "); err != nil {
		return err
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if readingClause.Match != nil {
			if err := s.writeMatch(writer, readingClause.Match); err != nil {
				return err
			}
		}
	}

	if singlePartQuery.Return != nil {
		if order := singlePartQuery.Return.Projection.Order; order != nil {
			if _, err := WriteStrings(writer, " order by "); err != nil {
				return err
			}

			for idx, orderItem := range order.Items {
				if idx > 0 {
					if _, err := WriteStrings(writer, ", "); err != nil {
						return err
					}
				}

				if err := s.WriteExpression(writer, orderItem.Expression); err != nil {
					return err
				}

				if orderItem.Ascending {
					if _, err := WriteStrings(writer, " asc"); err != nil {
						return err
					}
				} else {
					if _, err := WriteStrings(writer, " desc"); err != nil {
						return err
					}
				}
			}
		}

		if skip := singlePartQuery.Return.Projection.Skip; skip != nil {
			if _, err := WriteStrings(writer, " offset "); err != nil {
				return err
			}

			if err := s.WriteExpression(writer, skip.Value); err != nil {
				return err
			}
		}

		if limit := singlePartQuery.Return.Projection.Limit; limit != nil {
			if _, err := WriteStrings(writer, " limit "); err != nil {
				return err
			}

			if err := s.WriteExpression(writer, limit.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Emitter) writeUpdate(writer io.Writer, singlePartQuery *model.SinglePartQuery) error {
	if _, err := io.WriteString(writer, "update "); err != nil {
		return err
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			for idx, pattern := range matchClause.Pattern {
				if idx > 0 {
					if _, err := io.WriteString(writer, ", "); err != nil {
						return err
					}
				}

				if err := s.writePatternElements(writer, pattern.PatternElements); err != nil {
					return err
				}
			}
		}
	}

	if len(singlePartQuery.UpdatingClauses) > 0 {
		if _, err := WriteStrings(writer, " set "); err != nil {
			return err
		}

		for idx, item := range singlePartQuery.UpdatingClauses {
			if idx > 0 {
				if _, err := WriteStrings(writer, ", "); err != nil {
					return err
				}
			}

			switch typedUpdateItem := item.(type) {
			case *PropertyMutation:
				// Can't use aliased names in the set clauses of the SQL statement so default to just the raw
				// column names
				if _, err := WriteStrings(writer, "properties = properties"); err != nil {
					return err
				}

				if typedUpdateItem.Additions != nil {
					if typedUpdateItem.Removals != nil {
						if _, err := WriteStrings(writer, " - "); err != nil {
							return err
						}

						if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
							return err
						}
					}

					if _, err := WriteStrings(writer, " || "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Additions); err != nil {
						return err
					}
				} else if typedUpdateItem.Removals != nil {
					if _, err := WriteStrings(writer, " - "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
						return err
					}
				}

			case *KindMutation:
				// Cypher and therefore this translation does not support kind mutation of relationships
				if typedUpdateItem.Variable.Type != Node {
					return fmt.Errorf("unsupported SQL type for kind mutation: %s", typedUpdateItem.Variable.Type)
				}

				// Can't use aliased names in the set clauses of the SQL statement so default to just the raw
				// column names
				if _, err := WriteStrings(writer, "kind_ids = kind_ids"); err != nil {
					return err
				}

				if typedUpdateItem.Additions != nil {
					if typedUpdateItem.Removals != nil {
						if _, err := WriteStrings(writer, " - "); err != nil {
							return err
						}

						if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
							return err
						}
					}

					if _, err := WriteStrings(writer, " || "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Additions); err != nil {
						return err
					}
				} else if typedUpdateItem.Removals != nil {
					if _, err := WriteStrings(writer, " - "); err != nil {
						return err
					}

					if err := s.WriteExpression(writer, typedUpdateItem.Removals); err != nil {
						return err
					}
				}

			default:
				return fmt.Errorf("unsupported update clause item: %T", item)
			}
		}
	} else {
		return fmt.Errorf("all updating clauses must be rewritten")
	}

	for _, readingClause := range singlePartQuery.ReadingClauses {
		if matchClause := readingClause.Match; matchClause != nil {
			if matchClause.Where != nil {
				if err := s.writeWhere(writer, matchClause.Where); err != nil {
					return err
				}
			}
		}
	}

	if singlePartQuery.Return != nil {
		if _, err := WriteStrings(writer, " returning "); err != nil {
			return err
		}

		if err := s.writeReturn(writer, singlePartQuery.Return); err != nil {
			return err
		}
	}

	return nil
}

func (s *Emitter) writeSinglePartQuery(writer io.Writer, singlePartQuery *model.SinglePartQuery) error {
	if len(singlePartQuery.UpdatingClauses) > 0 {
		return s.writeUpdate(writer, singlePartQuery)
	} else {
		return s.writeSelect(writer, singlePartQuery)
	}
}

func (s *Emitter) writeSubquery(writer io.Writer, subquery *Subquery) error {
	if _, err := io.WriteString(writer, "(select count(*) from "); err != nil {
		return err
	}

	if err := s.writePatternElements(writer, subquery.PatternElements); err != nil {
		return err
	}

	if subquery.Filter != nil {
		subQueryWhereClause := model.NewWhere()
		subQueryWhereClause.Add(subquery.Filter)

		if err := s.writeWhere(writer, subQueryWhereClause); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(writer, " limit 1) > 0"); err != nil {
		return err
	}

	return nil
}

func (s *Emitter) WriteExpression(writer io.Writer, expression model.Expression) error {
	switch typedExpression := expression.(type) {
	case *Subquery:
		if err := s.writeSubquery(writer, typedExpression); err != nil {
			return err
		}

	case *model.Negation:
		if _, err := io.WriteString(writer, "not "); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

	case *model.Disjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(writer, " or "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(writer, joinedExpression); err != nil {
				return err
			}
		}

	case *model.Conjunction:
		for idx, joinedExpression := range typedExpression.Expressions {
			if idx > 0 {
				if _, err := io.WriteString(writer, " and "); err != nil {
					return err
				}
			}

			if err := s.WriteExpression(writer, joinedExpression); err != nil {
				return err
			}
		}

	case *model.Comparison:
		if err := s.WriteExpression(writer, typedExpression.Left); err != nil {
			return err
		}

		for _, nextPart := range typedExpression.Partials {
			if err := s.WriteExpression(writer, nextPart); err != nil {
				return err
			}
		}

	case *model.PartialComparison:
		if _, err := WriteStrings(writer, " ", typedExpression.Operator.String(), " "); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Right); err != nil {
			return err
		}

	case *AnnotatedLiteral:
		if err := s.WriteExpression(writer, &typedExpression.Literal); err != nil {
			return err
		}

	case *model.Literal:
		if !s.StripLiterals {
			return s.formatLiteral(writer, typedExpression)
		} else {
			_, err := io.WriteString(writer, strippedLiteral)
			return err
		}

	case *model.Variable:
		if _, err := io.WriteString(writer, typedExpression.Symbol); err != nil {
			return err
		}

	case *AnnotatedVariable:
		if _, err := io.WriteString(writer, typedExpression.Symbol); err != nil {
			return err
		}

	case *Entity:
		switch typedExpression.Binding.Type {
		case Node:
			if _, err := WriteStrings(writer, "(", typedExpression.Binding.Symbol, ".id, ", typedExpression.Binding.Symbol, ".kind_ids, ", typedExpression.Binding.Symbol, ".properties)::nodeComposite"); err != nil {
				return err
			}

		case Edge:
			if _, err := WriteStrings(writer, "(", typedExpression.Binding.Symbol, ".id, ", typedExpression.Binding.Symbol, ".start_id, ", typedExpression.Binding.Symbol, ".end_id, ", typedExpression.Binding.Symbol, ".kind_id, ", typedExpression.Binding.Symbol, ".properties)::edgeComposite"); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported entity type %s", typedExpression.Binding.Type)
		}

	case *NodeKindsReference:
		if err := s.WriteExpression(writer, typedExpression.Variable); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".kind_ids"); err != nil {
			return err
		}

	case *EdgeKindReference:
		if err := s.WriteExpression(writer, typedExpression.Variable); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".kind_id"); err != nil {
			return err
		}

	case *AnnotatedPropertyLookup:
		if _, err := io.WriteString(writer, "("); err != nil {
			return nil
		}

		if err := s.WriteExpression(writer, typedExpression.Atom); err != nil {
			return err
		}

		switch typedExpression.Type {
		case
			// We can't directly cast from JSONB types to time types since they require parsing first. The '->>'
			// operator coerces the underlying JSONB value to text before type casting
			Date, TimeWithTimeZone, TimeWithoutTimeZone, TimestampWithTimeZone, TimestampWithoutTimeZone,

			// Text types also require the `->>' operator otherwise type casting clobbers itself
			Text:

			if _, err := io.WriteString(writer, ".properties->>'"); err != nil {
				return nil
			}

		default:
			if _, err := io.WriteString(writer, ".properties->'"); err != nil {
				return nil
			}
		}

		if _, err := WriteStrings(writer, typedExpression.Symbols[0], "')::", typedExpression.Type.String()); err != nil {
			return nil
		}

	case *model.PropertyLookup:
		if err := s.WriteExpression(writer, typedExpression.Atom); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".properties->'", typedExpression.Symbols[0], "'"); err != nil {
			return nil
		}

	case *AnnotatedKindMatcher:
		if err := s.WriteExpression(writer, typedExpression.Reference); err != nil {
			return err
		}

		if mappedKinds, missingKinds := s.kindMapper.MapKinds(typedExpression.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("query references the following undefined kinds: %v", missingKinds.Strings())
		} else {
			mappedKindStr := JoinInt(mappedKinds, ", ")

			switch typedExpression.Type {
			case Node:
				if _, err := WriteStrings(writer, ".kind_ids operator(pg_catalog.&&) array[", mappedKindStr, "]::int2[]"); err != nil {
					return err
				}

			case Edge:
				if _, err := WriteStrings(writer, ".kind_id = any(array[", mappedKindStr, "]::int2[])"); err != nil {
					return err
				}
			}
		}

	case *model.FunctionInvocation:
		if err := s.translateFunctionInvocation(writer, typedExpression); err != nil {
			return err
		}

	case *model.Parameter:
		if _, err := WriteStrings(writer, "@", typedExpression.Symbol); err != nil {
			return err
		}

	case *AnnotatedParameter:
		if _, err := WriteStrings(writer, "@", typedExpression.Symbol); err != nil {
			return err
		}

	case *model.Parenthetical:
		if _, err := WriteStrings(writer, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case *PropertiesReference:
		if err := s.WriteExpression(writer, typedExpression.Reference); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ".properties"); err != nil {
			return err
		}

	case *model.ProjectionItem:
		if err := s.WriteExpression(writer, typedExpression.Expression); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, " as "); err != nil {
			return err
		}

		if typedExpression.Binding != nil {
			if err := s.WriteExpression(writer, typedExpression.Binding); err != nil {
				return err
			}
		} else {
			if _, err := WriteStrings(writer, "\""); err != nil {
				return err
			}

			switch typedProjectionExpression := typedExpression.Expression.(type) {
			case *NodeKindsReference:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *EdgeKindReference:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *model.FunctionInvocation:
				if err := s.WriteExpression(writer, typedProjectionExpression); err != nil {
					return err
				}

			case *model.PropertyLookup:
				if err := s.WriteExpression(writer, typedProjectionExpression.Atom); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".", typedProjectionExpression.Symbols[0]); err != nil {
					return err
				}

			case *AnnotatedPropertyLookup:
				if err := s.WriteExpression(writer, typedProjectionExpression.Atom); err != nil {
					return err
				}

				if _, err := WriteStrings(writer, ".", typedProjectionExpression.Symbols[0]); err != nil {
					return err
				}

			case *AnnotatedVariable:
				if err := s.WriteExpression(writer, typedProjectionExpression.Symbol); err != nil {
					return err
				}

			case *Entity:
				if err := s.WriteExpression(writer, typedProjectionExpression.Binding); err != nil {
					return err
				}

			default:
				return fmt.Errorf("unexpected projection item for binding formatting: %T", typedExpression.Expression)
			}

			if _, err := WriteStrings(writer, "\""); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected expression type for string formatting: %T", expression)
	}

	return nil
}

func (s *Emitter) translateFunctionInvocation(writer io.Writer, functionInvocation *model.FunctionInvocation) error {
	switch functionInvocation.Name {
	case cypherIdentityFunction:
		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := io.WriteString(writer, ".id"); err != nil {
			return err
		}

	case cypherDateFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::date"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "current_date"); err != nil {
			return err
		}

	case cypherTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::time with time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "current_time"); err != nil {
			return err
		}

	case cypherLocalTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::time without time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "localtime"); err != nil {
			return err
		}

	case cypherToLowerFunction:
		if _, err := WriteStrings(writer, pgsqlToLowerFunction, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case cypherDateTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::timestamp with time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "now()"); err != nil {
			return err
		}

	case cypherLocalDateTimeFunction:
		if len(functionInvocation.Arguments) > 0 {
			if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
				return err
			}

			if _, err := io.WriteString(writer, "::timestamp without time zone"); err != nil {
				return err
			}
		} else if _, err := io.WriteString(writer, "localtimestamp"); err != nil {
			return err
		}

	case cypherCountFunction:
		if _, err := WriteStrings(writer, "count("); err != nil {
			return err
		}

		for _, argument := range functionInvocation.Arguments {
			if err := s.WriteExpression(writer, argument); err != nil {
				return err
			}
		}

		if _, err := WriteStrings(writer, ")"); err != nil {
			return err
		}

	case pgsqlAnyFunction:
		if _, err := WriteStrings(writer, functionInvocation.Name, "("); err != nil {
			return err
		}

		if err := s.WriteExpression(writer, functionInvocation.Arguments[0]); err != nil {
			return err
		}

		if _, err := io.WriteString(writer, ")"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported function invocation %s", functionInvocation.Name)
	}

	return nil
}

func (s *Emitter) Write(regularQuery *model.RegularQuery, writer io.Writer) error {
	if regularQuery.SingleQuery != nil {
		if regularQuery.SingleQuery.MultiPartQuery != nil {
			return fmt.Errorf("not supported yet")
		}

		if regularQuery.SingleQuery.SinglePartQuery != nil {
			if err := s.writeSinglePartQuery(writer, regularQuery.SingleQuery.SinglePartQuery); err != nil {
				return err
			}
		}
	}

	return nil
}
