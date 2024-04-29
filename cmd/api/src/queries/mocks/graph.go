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

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/specterops/bloodhound/src/queries (interfaces: Graph)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	graph "github.com/specterops/bloodhound/dawgs/graph"
	model "github.com/specterops/bloodhound/src/model"
	queries "github.com/specterops/bloodhound/src/queries"
	agi "github.com/specterops/bloodhound/src/services/agi"
	gomock "go.uber.org/mock/gomock"
)

// MockGraph is a mock of Graph interface.
type MockGraph struct {
	ctrl     *gomock.Controller
	recorder *MockGraphMockRecorder
}

// MockGraphMockRecorder is the mock recorder for MockGraph.
type MockGraphMockRecorder struct {
	mock *MockGraph
}

// NewMockGraph creates a new mock instance.
func NewMockGraph(ctrl *gomock.Controller) *MockGraph {
	mock := &MockGraph{ctrl: ctrl}
	mock.recorder = &MockGraphMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGraph) EXPECT() *MockGraphMockRecorder {
	return m.recorder
}

// BatchNodeUpdate mocks base method.
func (m *MockGraph) BatchNodeUpdate(arg0 context.Context, arg1 graph.NodeUpdate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchNodeUpdate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchNodeUpdate indicates an expected call of BatchNodeUpdate.
func (mr *MockGraphMockRecorder) BatchNodeUpdate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchNodeUpdate", reflect.TypeOf((*MockGraph)(nil).BatchNodeUpdate), arg0, arg1)
}

// FetchNodesByObjectIDs mocks base method.
func (m *MockGraph) FetchNodesByObjectIDs(arg0 context.Context, arg1 ...string) (graph.NodeSet, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchNodesByObjectIDs", varargs...)
	ret0, _ := ret[0].(graph.NodeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchNodesByObjectIDs indicates an expected call of FetchNodesByObjectIDs.
func (mr *MockGraphMockRecorder) FetchNodesByObjectIDs(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchNodesByObjectIDs", reflect.TypeOf((*MockGraph)(nil).FetchNodesByObjectIDs), varargs...)
}

// GetADEntityQueryResult mocks base method.
func (m *MockGraph) GetADEntityQueryResult(arg0 context.Context, arg1 queries.EntityQueryParameters, arg2 bool) (interface{}, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetADEntityQueryResult", arg0, arg1, arg2)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetADEntityQueryResult indicates an expected call of GetADEntityQueryResult.
func (mr *MockGraphMockRecorder) GetADEntityQueryResult(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetADEntityQueryResult", reflect.TypeOf((*MockGraph)(nil).GetADEntityQueryResult), arg0, arg1, arg2)
}

// GetAllShortestPaths mocks base method.
func (m *MockGraph) GetAllShortestPaths(arg0 context.Context, arg1, arg2 string, arg3 graph.Criteria) (graph.PathSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllShortestPaths", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(graph.PathSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllShortestPaths indicates an expected call of GetAllShortestPaths.
func (mr *MockGraphMockRecorder) GetAllShortestPaths(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllShortestPaths", reflect.TypeOf((*MockGraph)(nil).GetAllShortestPaths), arg0, arg1, arg2, arg3)
}

// GetAssetGroupComboNode mocks base method.
func (m *MockGraph) GetAssetGroupComboNode(arg0 context.Context, arg1, arg2 string) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAssetGroupComboNode", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAssetGroupComboNode indicates an expected call of GetAssetGroupComboNode.
func (mr *MockGraphMockRecorder) GetAssetGroupComboNode(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssetGroupComboNode", reflect.TypeOf((*MockGraph)(nil).GetAssetGroupComboNode), arg0, arg1, arg2)
}

// GetAssetGroupNodes mocks base method.
func (m *MockGraph) GetAssetGroupNodes(arg0 context.Context, arg1 string) (graph.NodeSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAssetGroupNodes", arg0, arg1)
	ret0, _ := ret[0].(graph.NodeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAssetGroupNodes indicates an expected call of GetAssetGroupNodes.
func (mr *MockGraphMockRecorder) GetAssetGroupNodes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssetGroupNodes", reflect.TypeOf((*MockGraph)(nil).GetAssetGroupNodes), arg0, arg1)
}

// GetEntityByObjectId mocks base method.
func (m *MockGraph) GetEntityByObjectId(arg0 context.Context, arg1 string, arg2 ...graph.Kind) (*graph.Node, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEntityByObjectId", varargs...)
	ret0, _ := ret[0].(*graph.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEntityByObjectId indicates an expected call of GetEntityByObjectId.
func (mr *MockGraphMockRecorder) GetEntityByObjectId(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEntityByObjectId", reflect.TypeOf((*MockGraph)(nil).GetEntityByObjectId), varargs...)
}

// GetEntityCountResults mocks base method.
func (m *MockGraph) GetEntityCountResults(arg0 context.Context, arg1 *graph.Node, arg2 map[string]interface{}) map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEntityCountResults", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// GetEntityCountResults indicates an expected call of GetEntityCountResults.
func (mr *MockGraphMockRecorder) GetEntityCountResults(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEntityCountResults", reflect.TypeOf((*MockGraph)(nil).GetEntityCountResults), arg0, arg1, arg2)
}

// GetFilteredAndSortedNodes mocks base method.
func (m *MockGraph) GetFilteredAndSortedNodes(arg0 model.OrderCriteria, arg1 graph.Criteria) (graph.NodeSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFilteredAndSortedNodes", arg0, arg1)
	ret0, _ := ret[0].(graph.NodeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFilteredAndSortedNodes indicates an expected call of GetFilteredAndSortedNodes.
func (mr *MockGraphMockRecorder) GetFilteredAndSortedNodes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFilteredAndSortedNodes", reflect.TypeOf((*MockGraph)(nil).GetFilteredAndSortedNodes), arg0, arg1)
}

// GetNodesByKind mocks base method.
func (m *MockGraph) GetNodesByKind(arg0 context.Context, arg1 ...graph.Kind) (graph.NodeSet, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetNodesByKind", varargs...)
	ret0, _ := ret[0].(graph.NodeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodesByKind indicates an expected call of GetNodesByKind.
func (mr *MockGraphMockRecorder) GetNodesByKind(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodesByKind", reflect.TypeOf((*MockGraph)(nil).GetNodesByKind), varargs...)
}

// PrepareCypherQuery mocks base method.
func (m *MockGraph) PrepareCypherQuery(arg0 string) (queries.PreparedQuery, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareCypherQuery", arg0)
	ret0, _ := ret[0].(queries.PreparedQuery)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PrepareCypherQuery indicates an expected call of PrepareCypherQuery.
func (mr *MockGraphMockRecorder) PrepareCypherQuery(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareCypherQuery", reflect.TypeOf((*MockGraph)(nil).PrepareCypherQuery), arg0)
}

// RawCypherSearch mocks base method.
func (m *MockGraph) RawCypherSearch(arg0 context.Context, arg1 queries.PreparedQuery, arg2 bool) (model.UnifiedGraph, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawCypherSearch", arg0, arg1, arg2)
	ret0, _ := ret[0].(model.UnifiedGraph)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RawCypherSearch indicates an expected call of RawCypherSearch.
func (mr *MockGraphMockRecorder) RawCypherSearch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawCypherSearch", reflect.TypeOf((*MockGraph)(nil).RawCypherSearch), arg0, arg1, arg2)
}

// SearchByNameOrObjectID mocks base method.
func (m *MockGraph) SearchByNameOrObjectID(arg0 context.Context, arg1, arg2 string) (graph.NodeSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchByNameOrObjectID", arg0, arg1, arg2)
	ret0, _ := ret[0].(graph.NodeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchByNameOrObjectID indicates an expected call of SearchByNameOrObjectID.
func (mr *MockGraphMockRecorder) SearchByNameOrObjectID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchByNameOrObjectID", reflect.TypeOf((*MockGraph)(nil).SearchByNameOrObjectID), arg0, arg1, arg2)
}

// SearchNodesByName mocks base method.
func (m *MockGraph) SearchNodesByName(arg0 context.Context, arg1 graph.Kinds, arg2 string, arg3, arg4 int) ([]model.SearchResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchNodesByName", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].([]model.SearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchNodesByName indicates an expected call of SearchNodesByName.
func (mr *MockGraphMockRecorder) SearchNodesByName(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchNodesByName", reflect.TypeOf((*MockGraph)(nil).SearchNodesByName), arg0, arg1, arg2, arg3, arg4)
}

// UpdateSelectorTags mocks base method.
func (m *MockGraph) UpdateSelectorTags(arg0 context.Context, arg1 agi.AgiData, arg2 model.UpdatedAssetGroupSelectors) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSelectorTags", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSelectorTags indicates an expected call of UpdateSelectorTags.
func (mr *MockGraphMockRecorder) UpdateSelectorTags(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSelectorTags", reflect.TypeOf((*MockGraph)(nil).UpdateSelectorTags), arg0, arg1, arg2)
}

// ValidateOUs mocks base method.
func (m *MockGraph) ValidateOUs(arg0 context.Context, arg1 []string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateOUs", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateOUs indicates an expected call of ValidateOUs.
func (mr *MockGraphMockRecorder) ValidateOUs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateOUs", reflect.TypeOf((*MockGraph)(nil).ValidateOUs), arg0, arg1)
}
