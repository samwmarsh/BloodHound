package tools

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/config"
	"sync"
	"time"
)

type MigratorState string

const (
	stateIdle      MigratorState = "idle"
	stateMigrating MigratorState = "migrating"
)

type PGMigrator struct {
	defaultGraph graph.Graph
	serverCtx    context.Context
	state        MigratorState
	lock         *sync.Mutex
	cfg          config.Configuration
}

func NewPGMigrator(serverCtx context.Context, cfg config.Configuration, graphSchema graph.Graph) *PGMigrator {
	return &PGMigrator{
		defaultGraph: graphSchema,
		serverCtx:    serverCtx,
		state:        stateIdle,
		lock:         &sync.Mutex{},
		cfg:          cfg,
	}
}

func (s *PGMigrator) advanceState(expected, next MigratorState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.state != expected {
		return fmt.Errorf("migrator state is %s but expected %s", s.state, expected)
	}

	s.state = next
	return nil
}

func migrateNodes(ctx context.Context, defaultGraph graph.Graph, neoDB, pgDB graph.Database) (map[graph.ID]graph.ID, error) {
	defer log.LogAndMeasure(log.LevelInfo, "Migrating nodes from Neo4j to PostgreSQL")()

	var (
		nodeIDMappings = map[graph.ID]graph.ID{}
		nextNodeID     = graph.ID(1)
	)

	if err := neoDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			if err := pgDB.BatchOperation(ctx, func(tx graph.Batch) error {
				tx.WithGraph(defaultGraph)

				for next := range cursor.Chan() {
					if err := tx.CreateNode(graph.NewNode(nextNodeID, next.Properties, next.Kinds...)); err != nil {
						return err
					} else {
						nodeIDMappings[next.ID] = nextNodeID
						nextNodeID++
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return cursor.Error()
		})
	}); err != nil {
		return nil, err
	}

	return nodeIDMappings, pgDB.Run(ctx, fmt.Sprintf(`alter sequence node_id_seq restart with %d`, nextNodeID), nil)
}

func migrateEdges(ctx context.Context, defaultGraph graph.Graph, neoDB, pgDB graph.Database, nodeIDMappings map[graph.ID]graph.ID) error {
	defer log.LogAndMeasure(log.LevelInfo, "Migrating edges from Neo4j to PostgreSQL")()

	return neoDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			if err := pgDB.BatchOperation(ctx, func(tx graph.Batch) error {
				tx.WithGraph(defaultGraph)

				for next := range cursor.Chan() {
					var (
						pgStartID = nodeIDMappings[next.StartID]
						pgEndID   = nodeIDMappings[next.EndID]
					)

					if err := tx.CreateRelationshipByIDs(pgStartID, pgEndID, next.Kind, next.Properties); err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				return err
			}

			return cursor.Error()
		})
	})
}

func (s *PGMigrator) Wait(timeout time.Duration) error {
	var (
		ticker  = time.NewTicker(time.Second)
		then    = time.Now()
		elapsed = time.Duration(0)
	)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.state == stateIdle {
				return nil
			}

			if elapsed += time.Since(then); elapsed > timeout {
				return fmt.Errorf("timed out")
			}

			then = time.Now()

		case <-s.serverCtx.Done():
			return fmt.Errorf("context canceled")
		}
	}
}

func (s *PGMigrator) Migrate() {
	if err := s.advanceState(stateIdle, stateMigrating); err != nil {
		log.Errorf("Database migration state management error: %v", err)
	} else if neo4jDB, err := dawgs.Open(neo4j.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Neo4J.Neo4jConnectionString(),
	}); err != nil {
		log.Errorf("Unable to connect to Neo4j database: %v", err)
	} else if pgDB, err := dawgs.Open(pg.DriverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            s.cfg.Database.PostgreSQLConnectionString(),
	}); err != nil {
		log.Fatalf("Unable to connect to PostgreSQL database: %v", err)
	} else {
		log.Infof("Dispatching live migration from Neo4j to PostgreSQL")

		go func() {
			log.Infof("Starting live migration from Neo4j to PostgreSQL")

			if err := pgDB.AssertSchema(s.serverCtx, graph.Schema{
				Graphs: []graph.Graph{s.defaultGraph},
			}); err != nil {
				log.Errorf("Unable to assert graph schema in PostgreSQL: %v", err)
			} else if nodeIDMappings, err := migrateNodes(s.serverCtx, s.defaultGraph, neo4jDB, pgDB); err != nil {
				log.Errorf("Failed importing nodes into PostgreSQL: %v", err)
			} else if err := migrateEdges(s.serverCtx, s.defaultGraph, neo4jDB, pgDB, nodeIDMappings); err != nil {
				log.Errorf("Failed importing edges into PostgreSQL: %v", err)
			}

			if err := s.advanceState(stateMigrating, stateIdle); err != nil {
				log.Errorf("Database migration state management error: %v", err)
			}
		}()
	}
}
