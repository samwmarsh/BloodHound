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
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/drivers"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func NewValueMapper(values []any) *drivers.ValueMapper {
	return drivers.NewValueMapper(drivers.MapValue, values)
}

type queryError struct {
	err error
}

func (s queryError) Next() bool {
	return false
}

func (s queryError) Values() (graph.ValueMapper, error) {
	return nil, s.err
}

func (s queryError) Scan(targets ...any) error {
	return s.err
}

func (s queryError) Error() error {
	return s.err
}

func (s queryError) Close() {
}

type queryResult struct {
	rows       pgx.Rows
	kindMapper KindMapper
}

func (s *queryResult) Next() bool {
	return s.rows.Next()
}

func (s *queryResult) Values() (graph.ValueMapper, error) {
	if values, err := s.rows.Values(); err != nil {
		return nil, err
	} else {
		return NewValueMapper(values), nil
	}
}

func (s *queryResult) Scan(targets ...any) error {
	var translatedTargets []any

	for _, target := range targets {
		switch target.(type) {
		case *graph.Relationship:
			translatedTargets = append(translatedTargets, &edgeComposite{})

		case *graph.Node:
			translatedTargets = append(translatedTargets, &nodeComposite{})

		default:
			translatedTargets = append(translatedTargets, target)
		}
	}

	if err := s.rows.Scan(translatedTargets...); err != nil {
		return err
	}

	for idx, translatedTarget := range translatedTargets {
		switch typedTarget := translatedTarget.(type) {
		case *edgeComposite:
			if err := typedTarget.ToRelationship(s.kindMapper, targets[idx].(*graph.Relationship)); err != nil {
				return err
			}

		case *nodeComposite:
			if err := typedTarget.ToNode(s.kindMapper, targets[idx].(*graph.Node)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *queryResult) Error() error {
	return s.rows.Err()
}

func (s *queryResult) Close() {
	s.rows.Close()
}
