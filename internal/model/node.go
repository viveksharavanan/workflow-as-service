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

// Node represents a specific logical unit of processing and routing
// in a workflow.
type Node struct {
	ID       NodeID          `json:"ID"`                      // Unique identifier of this node
	DocType  DocTypeID       `json:"DocType"`                 // Document type which this node's workflow manages
	State    DocStateID      `json:"DocState"`                // A document arriving at this node must be in this state
	AccCtx   AccessContextID `json:"AccessContext,omitempty"` // Specific access context associated with this state, if any
	Wflow    WorkflowID      `json:"Workflow"`                // Containing flow of this node
	Name     string          `json:"Name"`                    // Unique within its workflow
	NodeType NodeType        `json:"NodeType"`                // Topology type of this node
	nfunc    NodeFunc        // Processing function of this node
}

// SetFunc registers the given node function with this node.
//
// If `nil` is given, a default node function is registered instead.
// This default function sets the document title as the message
// subject, and the event's data as the message body.
func (n *Node) SetFunc(fn NodeFunc) error {
	if fn == nil {
		n.nfunc = DefNodeFunc
		return nil
	}

	n.nfunc = fn
	return nil
}

// Func answers the processing function registered in this node
// definition.
func (n *Node) Func() NodeFunc {
	return n.nfunc
}
