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

// AccessContext is a namespace that provides an environment for
// workflow execution.
//
// It is an environment in which users are mapped into a hierarchy
// that determines certain aspects of workflow control. This
// hierarchy, usually, but not necessarily, reflects an organogram. In
// each access context, applicable groups are mapped to their
// respective intended permissions.  This mapping happens through
// roles.
//
// Each workflow that operates on a document type is given an
// associated access context.  This context is used to determine
// workflow routing, possible branching and rendezvous points.
//
// Please note that the same workflow may operate independently in
// multiple unrelated access contexts.  Thus, a workflow is not
// limited to one access context.  Conversely, an access context can
// have several workflows operating on it, for various document types.
// Therefore, the relationship between workflows and access contexts
// is M:N.
//
// For complex organisational requirements, the name of the access
// context can be made hierarchical with a suitable delimiter.  For
// example:
//
//   - IN:south:HYD:BR-101
//   - sbu-08/client-0249/prj-006348
type AccessContext struct {
	ID     AccessContextID `json:"ID"`               // Unique identifier of this access context
	Name   string          `json:"Name,omitempty"`   // Globally-unique namespace; can be a department, project, location, branch, etc.
	Active bool            `json:"Active,omitempty"` // Can a workflow be initiated in this context?
}

// AcGroupRoles holds the information of the various roles that each
// group has been assigned in this access context.
type AcGroupRoles struct {
	Group string `json:"Group"` // Group whose roles this represents
	Roles []Role `json:"Roles"` // Map holds the role assignments to groups
}

// AcGroup holds the information of a user together with the user's
// reporting authority.
type AcGroup struct {
	Group     `json:"Group"` // An assigned user
	ReportsTo Group          `json:"ReportsTo"` // Reporting authority of this user
}
