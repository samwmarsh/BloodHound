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

package integration

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/test/integration/utils"
)

func OpenDatabase(t *testing.T) database.Database {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		t.Fatalf("Failed loading integration test config: %v", err)
	} else if db, err := database.OpenDatabase(cfg.Database.PostgreSQLConnectionString()); err != nil {
		t.Fatalf("Failed to open database: %v", err)
	} else {
		return database.NewBloodhoundDB(db, auth.NewIdentityResolver())
	}

	return nil
}

func OpenCache(t *testing.T) cache.Cache {
	if cache, err := cache.NewCache(cache.Config{MaxSize: 200}); err != nil {
		t.Fatalf("Failed creating cache: %e", err)
	} else {
		return cache
	}
	return cache.Cache{}
}

func Prepare(db database.Database) error {
	if err := db.Wipe(); err != nil {
		return fmt.Errorf("failed to clear database: %v", err)
	} else if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}
