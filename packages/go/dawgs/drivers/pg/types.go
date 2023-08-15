package pg

import (
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type edgeComposite struct {
	ID         int32
	StartID    int32
	EndID      int32
	KindID     int16
	Properties map[string]any
}

func (s edgeComposite) ToRelationship(kindMapper KindMapper, relationship *graph.Relationship) error {
	if kinds, missingIDs := kindMapper.MapKindIDs(s.KindID); len(missingIDs) > 0 {
		return fmt.Errorf("edge references the following unknown kind IDs: %v", missingIDs)
	} else {
		relationship.Kind = kinds[0]
	}

	relationship.ID = graph.ID(s.ID)
	relationship.StartID = graph.ID(s.StartID)
	relationship.EndID = graph.ID(s.EndID)
	relationship.Properties = graph.AsProperties(s.Properties)

	return nil
}

type nodeComposite struct {
	ID         int32
	KindIDs    []int16
	Properties map[string]any
}

func (s nodeComposite) ToNode(kindMapper KindMapper, node *graph.Node) error {
	if kinds, missingIDs := kindMapper.MapKindIDs(s.KindIDs...); len(missingIDs) > 0 {
		return fmt.Errorf("node references the following unknown kind IDs: %v", missingIDs)
	} else {
		node.Kinds = kinds
	}

	node.ID = graph.ID(s.ID)
	node.Properties = graph.AsProperties(s.Properties)

	return nil
}
