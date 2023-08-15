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

package schema

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
)

const (
	ActiveDirectoryGraphPrefix = "ad"
	AzureGraphPrefix           = "az"
)

func ActiveDirectoryGraphName(suffix string) string {
	return ActiveDirectoryGraphPrefix + "_" + suffix
}

func AzureGraphName(suffix string) string {
	return AzureGraphPrefix + "_" + suffix
}

func DefaultGraphSchema() graph.Graph {
	return CombinedGraphSchema("default_graph")
}

func CombinedGraphSchema(name string) graph.Graph {
	return graph.Graph{
		Name:  name,
		Nodes: append(common.NodeKinds(), append(azure.NodeKinds(), ad.NodeKinds()...)...),
		Edges: append(common.Relationships(), append(azure.Relationships(), ad.Relationships()...)...),
		NodeConstraints: []graph.Constraint{{
			Field: common.ObjectID.String(),
			Type:  graph.TextSearchIndex,
		}},
		NodeIndexes: []graph.Index{
			{
				Field: common.Name.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.SystemTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.UserTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: azure.TenantID.String(),
				Type:  graph.BTreeIndex,
			},
		},
	}
}

func AzureGraphSchema(name string) graph.Graph {
	return graph.Graph{
		Name:  name,
		Nodes: azure.NodeKinds(),
		Edges: azure.Relationships(),
		NodeConstraints: []graph.Constraint{{
			Field: common.ObjectID.String(),
			Type:  graph.TextSearchIndex,
		}},
		NodeIndexes: []graph.Index{
			{
				Field: common.Name.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.SystemTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.UserTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: azure.TenantID.String(),
				Type:  graph.BTreeIndex,
			},
		},
	}
}

func ActiveDirectoryGraphSchema(name string) graph.Graph {
	return graph.Graph{
		Name:  name,
		Nodes: ad.NodeKinds(),
		Edges: ad.Relationships(),
		NodeConstraints: []graph.Constraint{{
			Field: common.ObjectID.String(),
			Type:  graph.TextSearchIndex,
		}},
		NodeIndexes: []graph.Index{
			{
				Field: common.Name.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.SystemTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: common.UserTags.String(),
				Type:  graph.TextSearchIndex,
			},
			{
				Field: ad.DistinguishedName.String(),
				Type:  graph.BTreeIndex,
			},
			{
				Field: ad.DomainFQDN.String(),
				Type:  graph.BTreeIndex,
			},
			{
				Field: ad.DomainSID.String(),
				Type:  graph.BTreeIndex,
			},
		},
	}
}

func CurrentSchema() graph.Schema {
	return graph.Schema{
		Graphs: []graph.Graph{
			{
				Nodes: common.Nodes(),
				Edges: common.Relationships(),
			},
			ActiveDirectoryGraphSchema(""),
			AzureGraphSchema(""),
		},
	}
}
