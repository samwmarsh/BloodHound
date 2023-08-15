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

package drivers

import (
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
)

func AsStringSlice(rawValues []any) ([]string, error) {
	strings := make([]string, len(rawValues))

	for idx, rawValue := range rawValues {
		switch typedValue := rawValue.(type) {
		case string:
			strings[idx] = typedValue
		default:
			return nil, fmt.Errorf("unexpected type %T will not negotiate to string", rawValue)
		}
	}

	return strings, nil
}

func AsKinds(rawValues []any) (graph.Kinds, error) {
	if stringValues, err := AsStringSlice(rawValues); err != nil {
		return nil, err
	} else {
		return graph.StringsToKinds(stringValues), nil
	}
}

func AsUint8(value any) (uint8, error) {
	switch typedValue := value.(type) {
	case uint8:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint8", value)
	}
}

func AsUint16(value any) (uint16, error) {
	switch typedValue := value.(type) {
	case uint8:
		return uint16(typedValue), nil
	case uint16:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint16", value)
	}
}

func AsUint32(value any) (uint32, error) {
	switch typedValue := value.(type) {
	case uint8:
		return uint32(typedValue), nil
	case uint16:
		return uint32(typedValue), nil
	case uint32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint32", value)
	}
}

func AsUint64(value any) (uint64, error) {
	switch typedValue := value.(type) {
	case uint:
		return uint64(typedValue), nil
	case uint8:
		return uint64(typedValue), nil
	case uint16:
		return uint64(typedValue), nil
	case uint32:
		return uint64(typedValue), nil
	case uint64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint64", value)
	}
}

func AsUint(value any) (uint, error) {
	switch typedValue := value.(type) {
	case uint:
		return typedValue, nil
	case uint8:
		return uint(typedValue), nil
	case uint16:
		return uint(typedValue), nil
	case uint32:
		return uint(typedValue), nil
	case uint64:
		return uint(typedValue), nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint", value)
	}
}

func AsInt8(value any) (int8, error) {
	switch typedValue := value.(type) {
	case int8:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int8", value)
	}
}

func AsInt16(value any) (int16, error) {
	switch typedValue := value.(type) {
	case int8:
		return int16(typedValue), nil
	case int16:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int16", value)
	}
}

func AsInt32(value any) (int32, error) {
	switch typedValue := value.(type) {
	case int8:
		return int32(typedValue), nil
	case int16:
		return int32(typedValue), nil
	case int32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int32", value)
	}
}

func AsInt64(value any) (int64, error) {
	switch typedValue := value.(type) {
	case graph.ID:
		return int64(typedValue), nil
	case int:
		return int64(typedValue), nil
	case int8:
		return int64(typedValue), nil
	case int16:
		return int64(typedValue), nil
	case int32:
		return int64(typedValue), nil
	case int64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func AsInt(value any) (int, error) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, nil
	case int8:
		return int(typedValue), nil
	case int16:
		return int(typedValue), nil
	case int32:
		return int(typedValue), nil
	case int64:
		return int(typedValue), nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int", value)
	}
}

func AsFloat32(value any) (float32, error) {
	switch typedValue := value.(type) {
	case float32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func AsFloat64(value any) (float64, error) {
	switch typedValue := value.(type) {
	case float32:
		return float64(typedValue), nil
	case float64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func AsTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case string:
		if parsedTime, err := time.Parse(time.RFC3339Nano, typedValue); err != nil {
			return time.Time{}, err
		} else {
			return parsedTime, nil
		}

	case float64:
		return time.Unix(int64(typedValue), 0), nil

	case int64:
		return time.Unix(typedValue, 0), nil

	case time.Time:
		return typedValue, nil

	default:
		return time.Time{}, fmt.Errorf("unexecpted type %T will not negotiate to time.Time", value)
	}
}

func MapValue(rawValue, target any) error {
	switch typedTarget := target.(type) {
	case *uint:
		if value, err := AsUint(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint8:
		if value, err := AsUint8(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint16:
		if value, err := AsUint16(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint32:
		if value, err := AsUint32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint64:
		if value, err := AsUint64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int:
		if value, err := AsInt(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int8:
		if value, err := AsInt8(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int16:
		if value, err := AsInt16(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int32:
		if value, err := AsInt32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int64:
		if value, err := AsInt64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *graph.ID:
		if value, err := AsInt64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = graph.ID(value)
		}

	case *float32:
		if value, err := AsFloat32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *float64:
		if value, err := AsFloat64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *bool:
		if value, typeOK := rawValue.(bool); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to bool", value)
		} else {
			*typedTarget = value
		}

	case *graph.Kind:
		if strValue, typeOK := rawValue.(string); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to string", rawValue)
		} else {
			*typedTarget = graph.StringKind(strValue)
		}

	case *string:
		if value, typeOK := rawValue.(string); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to string", value)
		} else {
			*typedTarget = value
		}

	case *[]graph.Kind:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if kindValues, err := AsKinds(rawValues); err != nil {
			return err
		} else {
			*typedTarget = kindValues
		}

	case *graph.Kinds:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if kindValues, err := AsKinds(rawValues); err != nil {
			return err
		} else {
			*typedTarget = kindValues
		}

	case *[]string:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if stringValues, err := AsStringSlice(rawValues); err != nil {
			return err
		} else {
			*typedTarget = stringValues
		}

	case *time.Time:
		if value, err := AsTime(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	default:
		return fmt.Errorf("unsupported scan type %T", target)
	}

	return nil
}

type MapFunc func(rawValue, target any) error

type ValueMapper struct {
	mapFunc MapFunc
	values  []any
	idx     int
}

func NewValueMapper(mapFunc MapFunc, values []any) *ValueMapper {
	return &ValueMapper{
		mapFunc: mapFunc,
		values:  values,
		idx:     0,
	}
}

func (s *ValueMapper) Next() (any, error) {
	if s.idx >= len(s.values) {
		return nil, fmt.Errorf("attempting to get more values than returned - saw %d but wanted %d", len(s.values), s.idx+1)
	}

	nextValue := s.values[s.idx]
	s.idx++

	return nextValue, nil
}

func (s *ValueMapper) Map(target any) error {
	if rawValue, err := s.Next(); err != nil {
		return err
	} else {
		return s.mapFunc(rawValue, target)
	}
}

func (s *ValueMapper) MapOptions(targets ...any) (any, error) {
	if rawValue, err := s.Next(); err != nil {
		return nil, err
	} else {
		for _, target := range targets {
			if s.mapFunc(target, rawValue) == nil {
				return target, nil
			}
		}

		return nil, fmt.Errorf("no matching target given for type: %T", rawValue)
	}
}

func (s *ValueMapper) Scan(targets ...any) error {
	for idx, mapValue := range targets {
		if err := s.Map(mapValue); err != nil {
			return err
		} else {
			targets[idx] = mapValue
		}
	}

	return nil
}
