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
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	readOnlyTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}

	readWriteTxOptions = pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	}
)

type Config struct {
	Options       pgx.TxOptions
	QueryExecMode pgx.QueryExecMode
}

func OptionSetQueryExecMode(queryExecMode pgx.QueryExecMode) graph.TransactionOption {
	return func(config *graph.TransactionConfig) {
		if pgCfg, typeOK := config.DriverConfig.(*Config); typeOK {
			pgCfg.QueryExecMode = queryExecMode
		}
	}
}

type Driver struct {
	pool                      *pgxpool.Pool
	schemaManager             *SchemaManager
	defaultTransactionTimeout time.Duration
}

func (s *Driver) KindMapper() KindMapper {
	return s.schemaManager
}

func (s *Driver) SetDefaultGraph(ctx context.Context, graphSchema graph.Graph) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return s.schemaManager.AssertDefaultGraph(tx, graphSchema)
	})
}

func (s *Driver) SetBatchWriteSize(size int) {
}

func (s *Driver) SetWriteFlushSize(size int) {
}

func (s *Driver) BatchOperation(ctx context.Context, batchDelegate graph.BatchDelegate) error {
	if cfg, err := renderConfig(readWriteTxOptions, nil); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if batch, err := newBatch(ctx, conn, s.schemaManager, cfg); err != nil {
			return err
		} else {
			defer batch.Close()

			if err := batchDelegate(batch); err != nil {
				return err
			}

			return batch.Commit()
		}
	}
}

func (s *Driver) Close(ctx context.Context) error {
	s.pool.Close()
	return nil
}

func renderConfig(pgxOptions pgx.TxOptions, userOptions []graph.TransactionOption) (*Config, error) {
	graphCfg := graph.TransactionConfig{
		DriverConfig: &Config{
			Options:       pgxOptions,
			QueryExecMode: pgx.QueryExecModeCacheStatement,
		},
	}

	for _, option := range userOptions {
		option(&graphCfg)
	}

	if graphCfg.DriverConfig != nil {
		if pgCfg, typeOK := graphCfg.DriverConfig.(*Config); !typeOK {
			return nil, fmt.Errorf("invalid driver config type %T", graphCfg.DriverConfig)
		} else {
			return pgCfg, nil
		}
	}

	return nil, fmt.Errorf("driver config is nil")
}

const (
	errStringContextCanceled = "context canceled"
)

func (s *Driver) ReadTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(readOnlyTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if tx, err := newTransaction(ctx, conn, s.schemaManager, cfg); err != nil {
			return err
		} else {
			defer tx.Close()
			return txDelegate(tx)
		}
	}
}

func (s *Driver) WriteTransaction(ctx context.Context, txDelegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	if cfg, err := renderConfig(readWriteTxOptions, options); err != nil {
		return err
	} else if conn, err := s.pool.Acquire(ctx); err != nil {
		return err
	} else {
		defer conn.Release()

		if tx, err := newTransaction(ctx, conn, s.schemaManager, cfg); err != nil {
			return err
		} else {
			defer tx.Close()

			if err := txDelegate(tx); err != nil {
				return err
			}

			return tx.Commit()
		}
	}
}

func (s *Driver) FetchSchema(ctx context.Context) (graph.Schema, error) {
	return graph.Schema{}, nil
}

func (s *Driver) AssertSchema(ctx context.Context, schema graph.Schema) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return s.schemaManager.AssertSchema(tx, schema)
	}, OptionSetQueryExecMode(pgx.QueryExecModeSimpleProtocol))
}

func (s *Driver) Run(ctx context.Context, query string, parameters map[string]any) error {
	return s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		result := tx.Run(query, parameters)
		defer result.Close()

		return result.Error()
	})
}
