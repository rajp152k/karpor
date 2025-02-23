// Copyright The Karpor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mcp

import (

	"github.com/KusionStack/karpor/pkg/infra/search/storage"
	"github.com/KusionStack/karpor/pkg/infra/search/storage/elasticsearch"
)

// MCPServer is the main object that holds the necessary fields and components for the mcp server component.
type MCPServer struct {
	Storage storage.Storage
}

//MCPToolName is a string tag for an mcp server tool
type MCPToolName string


//MCPResourceName is a string tag for an mcp server resource
type MCPResourceName string


//MCPPromptName is a string tag for an mcp server prompt
type MCPPromptName string
