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

package query

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_PGIndexRegex(t *testing.T) {
	captureGroups := pgPropertyIndexRegex.FindStringSubmatch("CREATE INDEX edge_1_kind_id_idx ON public.edge_1 USING btree (kind_id)")

	require.Equal(t, pgIndexRegexNumExpectedGroups, len(captureGroups))
	require.Equal(t, "", captureGroups[pgIndexRegexGroupUnique])
	require.Equal(t, "edge_1_kind_id_idx", captureGroups[pgIndexRegexGroupName])
	require.Equal(t, "btree", captureGroups[pgIndexRegexGroupIndexType])
	require.Equal(t, "kind_id", captureGroups[pgIndexRegexGroupFields])

	captureGroups = pgPropertyIndexRegex.FindStringSubmatch("create UNIQUE index edge_1_unique_col_constraint ON public.edge_1 USING btree (unique_col)")

	require.Equal(t, pgIndexRegexNumExpectedGroups, len(captureGroups))
	require.Equal(t, "UNIQUE", captureGroups[pgIndexRegexGroupUnique])
	require.Equal(t, "edge_1_unique_col_constraint", captureGroups[pgIndexRegexGroupName])
	require.Equal(t, "btree", captureGroups[pgIndexRegexGroupIndexType])
	require.Equal(t, "unique_col", captureGroups[pgIndexRegexGroupFields])
}
