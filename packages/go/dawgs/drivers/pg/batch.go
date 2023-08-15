package pg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cypher/gen/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
)

const (
	DefaultBatchWriteSize = 20_000
	DefaultWriteFlushSize = DefaultBatchWriteSize * 5

	// DefaultConcurrentConnections defines the default number of concurrent graph database connections allowed.
	DefaultConcurrentConnections = 50
)

type batch struct {
	ctx                        context.Context
	innerTransaction           *transaction
	schemaManager              *SchemaManager
	nodeDeletionBuffer         []graph.ID
	relationshipDeletionBuffer []graph.ID
	nodeCreateBuffer           []*graph.Node
	nodeUpdateByBuffer         []graph.NodeUpdate
	relationshipCreateBuffer   []*graph.Relationship
	relationshipUpdateByBuffer []graph.RelationshipUpdate
	batchWriteSize             int
	targetSchema               graph.Graph
	targetSchemaSet            bool
}

func newBatch(ctx context.Context, conn *pgxpool.Conn, schemaManager *SchemaManager, cfg *Config) (*batch, error) {
	if tx, err := newTransaction(ctx, conn, schemaManager, cfg); err != nil {
		return nil, err
	} else {
		return &batch{
			ctx:              ctx,
			schemaManager:    schemaManager,
			innerTransaction: tx,
			batchWriteSize:   DefaultBatchWriteSize,
		}, nil
	}
}

func (s *batch) WithGraph(schema graph.Graph) graph.Batch {
	s.targetSchema = schema
	s.targetSchemaSet = true

	return s
}

func (s *batch) CreateNode(node *graph.Node) error {
	s.nodeCreateBuffer = append(s.nodeCreateBuffer, node)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) Nodes() graph.NodeQuery {
	return s.innerTransaction.Nodes()
}

func (s *batch) Relationships() graph.RelationshipQuery {
	return s.innerTransaction.Relationships()
}

func (s *batch) UpdateNodeBy(update graph.NodeUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (s *batch) flushNodeCreateBuffer() error {
	var (
		withoutIDs = false
		withIDs    = false
	)

	for _, node := range s.nodeCreateBuffer {
		if node.ID == 0 || node.ID == graph.UnregisteredNodeID {
			withoutIDs = true
		} else {
			withIDs = true
		}

		if withIDs && withoutIDs {
			return fmt.Errorf("batch may not mix preset node IDs with entries that require an auto-generated ID")
		}
	}

	if withoutIDs {
		return s.flushNodeCreateBufferWithoutIDs()
	}

	return s.flushNodeCreateBufferWithIDs()
}

type Int2ArrayEncoder struct {
	buffer *bytes.Buffer
}

func (s *Int2ArrayEncoder) Encode(values []int16) string {
	s.buffer.Reset()
	s.buffer.WriteRune('{')

	for idx, value := range values {
		if idx > 0 {
			s.buffer.WriteRune(',')
		}

		s.buffer.WriteString(strconv.Itoa(int(value)))
	}

	s.buffer.WriteRune('}')
	return s.buffer.String()
}

func (s *batch) flushNodeCreateBufferWithIDs() error {
	var (
		numCreates    = len(s.nodeCreateBuffer)
		nodeIDs       = make([]uint32, numCreates)
		kindIDSlices  = make([]string, numCreates)
		kindIDEncoder = Int2ArrayEncoder{
			buffer: &bytes.Buffer{},
		}
		properties = make([]pgtype.JSONB, numCreates)
	)

	for idx, nextNode := range s.nodeCreateBuffer {
		nodeIDs[idx] = nextNode.ID.Uint32()

		if mappedKindIDs, missingKinds := s.schemaManager.MapKinds(nextNode.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("unable to map kinds %v", missingKinds)
		} else {
			kindIDSlices[idx] = kindIDEncoder.Encode(mappedKindIDs)
		}

		if propertiesJSONB, err := pgsql.PropertiesToJSONB(nextNode.Properties); err != nil {
			return err
		} else {
			properties[idx] = propertiesJSONB
		}
	}

	s.innerTransaction.WithGraph(s.targetSchema)

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createNodeWithIDBatchStatement, graphTarget.ID, nodeIDs, kindIDSlices, properties); err != nil {
		return err
	}

	s.nodeCreateBuffer = s.nodeCreateBuffer[:0]
	return nil
}

func (s *batch) flushNodeCreateBufferWithoutIDs() error {
	var (
		numCreates   = len(s.nodeCreateBuffer)
		kindIDSlices = make([][]int16, numCreates)
		properties   = make([]pgtype.JSONB, numCreates)
	)

	for idx, nextNode := range s.nodeCreateBuffer {
		if mappedKindIDs, missingKinds := s.schemaManager.MapKinds(nextNode.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("unable to map kinds %v", missingKinds)
		} else {
			kindIDSlices[idx] = mappedKindIDs
		}

		if propertiesJSONB, err := pgsql.PropertiesToJSONB(nextNode.Properties); err != nil {
			return err
		} else {
			properties[idx] = propertiesJSONB
		}
	}

	s.innerTransaction.WithGraph(s.targetSchema)

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createNodeWithoutIDBatchStatement, graphTarget.ID, kindIDSlices, properties); err != nil {
		return err
	}

	s.nodeCreateBuffer = s.nodeCreateBuffer[:0]
	return nil
}

func (s *batch) flushRelationshipCreateBuffer() error {
	var (
		numCreates   = len(s.relationshipCreateBuffer)
		startIDs     = make([]uint32, numCreates)
		endIDs       = make([]uint32, numCreates)
		kindIDs      = make([]int16, numCreates)
		properties   = make([]pgtype.JSONB, numCreates)
		kindMapSlice = make(graph.Kinds, 1)
	)

	for idx, nextRel := range s.relationshipCreateBuffer {
		startIDs[idx] = nextRel.StartID.Uint32()
		endIDs[idx] = nextRel.EndID.Uint32()
		kindMapSlice[0] = nextRel.Kind

		if mappedKindIDs, missingKinds := s.schemaManager.MapKinds(kindMapSlice); len(missingKinds) > 0 {
			return fmt.Errorf("unable to map kind %s", nextRel.Kind)
		} else {
			kindIDs[idx] = mappedKindIDs[0]
		}

		if propertiesJSONB, err := pgsql.PropertiesToJSONB(nextRel.Properties); err != nil {
			return err
		} else {
			properties[idx] = propertiesJSONB
		}
	}

	s.innerTransaction.WithGraph(s.targetSchema)

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createEdgeBatchStatement, graphTarget.ID, startIDs, endIDs, kindIDs, properties); err != nil {
		return err
	}

	s.relationshipCreateBuffer = s.relationshipCreateBuffer[:0]
	return nil
}

func (s *batch) tryFlush(batchWriteSize int) error {
	if len(s.relationshipCreateBuffer) > batchWriteSize {
		if err := s.flushRelationshipCreateBuffer(); err != nil {
			return err
		}
	}

	if len(s.nodeCreateBuffer) > batchWriteSize {
		if err := s.flushNodeCreateBuffer(); err != nil {
			return err
		}
	}

	return nil
}

func (s *batch) CreateRelationship(relationship *graph.Relationship) error {
	s.relationshipCreateBuffer = append(s.relationshipCreateBuffer, relationship)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) error {
	return s.CreateRelationship(&graph.Relationship{
		StartID:    startNodeID,
		EndID:      endNodeID,
		Kind:       kind,
		Properties: properties,
	})
}

func (s *batch) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (s *batch) Commit() error {
	if err := s.tryFlush(0); err != nil {
		return err
	}

	return s.innerTransaction.Commit()
}

func (s *batch) DeleteNode(id graph.ID) error {
	_, err := s.innerTransaction.tx.Exec(s.innerTransaction.ctx, deleteNodeStatement, id)
	return err
}

func (s *batch) DeleteRelationship(id graph.ID) error {
	_, err := s.innerTransaction.tx.Exec(s.innerTransaction.ctx, deleteEdgeStatement, id)
	return err
}

func (s *batch) Close() {
	s.innerTransaction.Close()
}
