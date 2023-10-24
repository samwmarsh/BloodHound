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
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"time"
)

const (
	DriverName = "pg"

	poolInitConnectionTimeout = time.Second * 10
	defaultTransactionTimeout = time.Minute * 15
)

func afterPooledConnectionEstablished(ctx context.Context, conn *pgx.Conn) error {
	log.Debugf("Established a new database connection.")

	for _, dataType := range pgsql.CompositeTypes {
		if definition, err := conn.LoadType(ctx, dataType.String()); err != nil {
			if !StateObjectDoesNotExist.ErrorMatches(err) {
				return err
			}
		} else {
			conn.TypeMap().RegisterType(definition)
		}
	}

	return nil
}

func afterPooledConnectionRelease(conn *pgx.Conn) bool {
	for _, dataType := range pgsql.CompositeTypes {
		if _, hasType := conn.TypeMap().TypeForName(dataType.String()); !hasType {
			// This connection should be destroyed since it does not contain information regarding the schema's
			// composite types
			log.Warnf("Unable to find expected data type: %s. This database connection will not be pooled.", dataType)
			return false
		}
	}

	return true
}

func newDatabase(connectionString string) (graph.Database, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	if poolCfg, err := pgxpool.ParseConfig(connectionString); err != nil {
		return nil, err
	} else {
		poolCfg.MinConns = 5
		poolCfg.MaxConns = 30
		poolCfg.AfterConnect = afterPooledConnectionEstablished
		poolCfg.AfterRelease = afterPooledConnectionRelease

		if pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg); err != nil {
			return nil, err
		} else {
			return &Driver{
				pool:                      pool,
				schemaManager:             NewSchemaManager(),
				defaultTransactionTimeout: defaultTransactionTimeout,
			}, nil
		}
	}
}

func init() {
	dawgs.Register(DriverName, func(cfg dawgs.Config) (graph.Database, error) {
		if connectionString, typeOK := cfg.DriverCfg.(string); !typeOK {
			return nil, fmt.Errorf("expected string for configuration type but got %T", cfg)
		} else {
			return newDatabase(connectionString)
		}
	})
}
