package pgsql

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model"
)

func CollectPGSQLTypes(nextCursor *model.WalkCursor, expression model.Expression) bool {
	switch typedExpression := expression.(type) {
	case *PropertiesReference:
		model.Collect(nextCursor, typedExpression.Reference)

	case *AnnotatedPropertyLookup:
		model.CollectExpression(nextCursor, typedExpression.Atom)

	case *AnnotatedKindMatcher:
		model.CollectExpression(nextCursor, typedExpression.Reference)

	case *Entity:
		model.Collect(nextCursor, typedExpression.Binding)

	case *Subquery:
		model.CollectSlice(nextCursor, typedExpression.PatternElements)
		model.CollectExpression(nextCursor, typedExpression.Filter)

	case *PropertyMutation:
		model.Collect(nextCursor, typedExpression.Reference)
		model.Collect(nextCursor, typedExpression.Removals)
		model.Collect(nextCursor, typedExpression.Additions)

	case *KindMutation:
		model.Collect(nextCursor, typedExpression.Variable)
		model.Collect(nextCursor, typedExpression.Removals)
		model.Collect(nextCursor, typedExpression.Additions)

	case *NodeKindsReference:
		model.CollectExpression(nextCursor, typedExpression.Variable)

	case *EdgeKindReference:
		model.CollectExpression(nextCursor, typedExpression.Variable)

	case *AnnotatedLiteral, *AnnotatedVariable, *AnnotatedParameter:
		// Valid types but no descent

	default:
		return false
	}

	return true
}

func rewrite(stack *model.WalkStack, original, rewritten model.Expression) error {
	switch typedTrunk := stack.Trunk().(type) {
	case model.ExpressionList:
		for idx, expression := range typedTrunk.GetAll() {
			if expression == original {
				typedTrunk.Replace(idx, rewritten)
			}
		}

	case *model.ProjectionItem:
		typedTrunk.Expression = rewritten

	case *model.SetItem:
		if typedTrunk.Right == original {
			typedTrunk.Right = rewritten
		} else if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		} else {
			return fmt.Errorf("unable to match original expression against SetItem left and right operands")
		}

	case *model.PartialComparison:
		typedTrunk.Right = rewritten

	case *model.RemoveItem:
		switch typedRewritten := rewritten.(type) {
		case *model.KindMatcher:
			typedTrunk.KindMatcher = typedRewritten
		}

	case *model.Projection:
		for idx, projectionItem := range typedTrunk.Items {
			if projectionItem == original {
				typedTrunk.Items[idx] = rewritten
			}
		}

	case *model.Negation:
		typedTrunk.Expression = rewritten

	case *model.Comparison:
		if typedTrunk.Left == original {
			typedTrunk.Left = rewritten
		}

	default:
		return fmt.Errorf("unable to replace expression for trunk type %T", stack.Trunk())
	}

	return nil
}
