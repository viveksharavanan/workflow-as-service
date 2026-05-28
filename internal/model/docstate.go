// (c) Copyright 2015-2017 JONNALAGADDA Srinivas
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package model

// DocState is one of a set of enumerated states for a top-level
// document, as defined by the consuming application.
//
// `flow`, therefore, does not assume anything about the specifics of
// any state.  Applications can, for example, embed `DocState` in a
// struct that provides more context.  Document states should be
// loaded during application initialisation.
//
// N.B. A `DocState` once defined and used, should *NEVER* be removed.
// At best, it can be deprecated by defining a new one, and then
// altering the corresponding workflow definition to use the new one
// instead.
type DocState struct {
	ID   DocStateID `json:"ID"`             // Unique identifier of this document state
	Name string     `json:"Name,omitempty"` // Unique identifier of this state in its workflow
}
