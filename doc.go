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

// Package flow is a tiny workflow engine written in Go (golang).
package flow

import (
	"database/sql"
	"log"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
	"github.com/viveksharavanan/workflow-as-service/internal/store"
	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
	"github.com/viveksharavanan/workflow-as-service/internal/workflow"
)

// Type aliases for backward compatibility -- ID types
type DocTypeID = model.DocTypeID
type DocStateID = model.DocStateID
type DocActionID = model.DocActionID
type DocEventID = model.DocEventID
type DocumentID = model.DocumentID
type WorkflowID = model.WorkflowID
type NodeID = model.NodeID
type UserID = model.UserID
type GroupID = model.GroupID
type RoleID = model.RoleID
type AccessContextID = model.AccessContextID
type MessageID = model.MessageID

// Type aliases for backward compatibility -- struct types
type DocType = model.DocType
type DocState = model.DocState
type DocAction = model.DocAction
type DocEvent = model.DocEvent
type Document = model.Document
type Workflow = model.Workflow
type Node = model.Node
type User = model.User
type Group = model.Group
type Role = model.Role
type AccessContext = model.AccessContext
type Transition = model.Transition
type TransitionMap = model.TransitionMap

// Type aliases for backward compatibility -- other types
type NodeType = model.NodeType
type NodeFunc = model.NodeFunc
type EventStatus = model.EventStatus
type Message = model.Message
type DocPath = model.DocPath
type AcGroupRoles = model.AcGroupRoles
type AcGroup = model.AcGroup
type Blob = model.Blob
type Notification = model.Notification

// Error type alias
type Error = model.Error

// Package-level state
var db *sql.DB
var s store.Store
var engine *workflow.Engine

func init() {
	f := log.Flags()
	log.SetFlags(f | log.Lmicroseconds | log.Lshortfile)
}

// RegisterDB provides an already initialised database handle to `flow`.
//
// N.B. This method **MUST** be called before anything else in `flow`.
func RegisterDB(sdb *sql.DB) error {
	if sdb == nil {
		log.Fatal("given database handle is `nil`")
	}
	db = sdb
	s = mysqlstore.NewMySQLStore(sdb, "")
	engine = workflow.NewEngine(s)

	return nil
}
