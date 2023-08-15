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

package main

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/harness"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
)

func AssertNil(value any) {
	if value != nil {
		panic(fmt.Sprintf("value %v not nil", value))
	}
}

func Measure(name string, fn func() error) {
	then := time.Now()

	AssertNil(fn())

	fmt.Printf("%s: %d ms\n", name, time.Since(then).Milliseconds())
}

func main() {
	neo4jDB, err := dawgs.Open(neo4j.DriverName, "neo4j://neo4j:neo4jj@localhost:7687")
	AssertNil(err)

	pgDB, err := dawgs.Open(pg.DriverName, "user=bhe dbname=bhe password=bhe4eva host=localhost")
	AssertNil(err)

	vetCtx, done := context.WithTimeout(context.Background(), time.Minute)
	defer done()

	kind1 := graph.StringKind("kind1")
	kind2 := graph.StringKind("kind2")
	properties := graph.AsProperties(map[string]any{
		"name":   "my name",
		"number": 1234,
		"float":  12.34,
		"date":   time.Now().UTC(),
	})

	AssertNil(neo4jDB.AssertSchema(vetCtx, graph.Schema{
		Graphs: []graph.Graph{
			harness.ActiveDirectoryGraphSchema("test"),
		},
	}))

	Measure(
		"Neo4j Write 10k Nodes",
		func() error {
			return neo4jDB.WriteTransaction(vetCtx, func(tx graph.Transaction) error {
				tx.WithGraph(harness.ActiveDirectoryGraphSchema("test"))

				for i := 0; i < 10000; i++ {
					if _, err := tx.CreateNode(properties, kind1, kind2); err != nil {
						return err
					}
				}

				return nil
			})
		},
	)

	AssertNil(pgDB.AssertSchema(vetCtx, graph.Schema{
		Graphs: []graph.Graph{
			harness.ActiveDirectoryGraphSchema("test"),
		},
	}))

	Measure(
		"PostgreSQL Write 10k Nodes",
		func() error {
			return pgDB.WriteTransaction(vetCtx, func(tx graph.Transaction) error {
				tx.WithGraph(harness.ActiveDirectoryGraphSchema("test"))

				for i := 0; i < 10000; i++ {
					if _, err := tx.CreateNode(properties, kind1, kind2); err != nil {
						return err
					}
				}

				return nil
			})
		},
	)
}
