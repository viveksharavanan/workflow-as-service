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

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DocTypeID is the type of unique identifiers of document types.
type DocTypeID int64

// DocStateID is the type of unique identifiers of document states.
type DocStateID int64

// DocActionID is the type of unique identifiers of document actions.
type DocActionID int64

// DocEventID is the type of unique document event identifiers.
type DocEventID int64

// DocumentID is the type of unique document identifiers.
type DocumentID int64

// WorkflowID is the type of unique workflow identifiers.
type WorkflowID int64

// NodeID is the type of unique identifiers of nodes.
type NodeID int64

// UserID is the type of unique user identifiers.
type UserID int64

// GroupID is the type of unique group identifiers.
type GroupID int64

// RoleID is the type of unique role identifiers.
type RoleID int64

// AccessContextID is the type unique access context identifiers.
type AccessContextID int64

// MessageID is the type of unique identifiers of messages.
type MessageID int64

// NodeType enumerates the possible types of workflow nodes.
type NodeType string

// The following constants are represented **identically** as part of
// an enumeration in the database.  DO NOT ALTER THESE WITHOUT ALSO
// ALTERING THE DATABASE; ELSE DATA COULD GET CORRUPTED!
const (
	// NodeTypeBegin : none incoming, one outgoing
	NodeTypeBegin NodeType = "begin"
	// NodeTypeEnd : one incoming, none outgoing
	NodeTypeEnd = "end"
	// NodeTypeLinear : one incoming, one outgoing
	NodeTypeLinear = "linear"
	// NodeTypeBranch : one incoming, two or more outgoing
	NodeTypeBranch = "branch"
	// NodeTypeJoinAny : two or more incoming, one outgoing
	NodeTypeJoinAny = "joinany"
	// NodeTypeJoinAll : two or more incoming, one outgoing
	NodeTypeJoinAll = "joinall"
)

// IsValidNodeType answers `true` if the given node type is a
// recognised node type in the system.
func IsValidNodeType(ntype string) bool {
	nt := NodeType(ntype)
	switch nt {
	case NodeTypeBegin, NodeTypeEnd, NodeTypeLinear, NodeTypeBranch, NodeTypeJoinAny, NodeTypeJoinAll:
		return true

	default:
		return false
	}
}

// EventStatus enumerates the query parameter values for filtering by
// event state.
type EventStatus uint8

const (
	// EventStatusAll does not filter events.
	EventStatusAll EventStatus = iota
	// EventStatusApplied selects only those events that have been successfully applied.
	EventStatusApplied
	// EventStatusPending selects only those events that are pending application.
	EventStatusPending
)

// NodeFunc defines the type of functions that generate notification
// messages in workflows.
//
// These functions are triggered by appropriate nodes, when document
// events are applied to documents to possibly transform them.
// Invocation of a `NodeFunc` should result in a message that can then
// be dispatched to applicable mailboxes.
//
// Error should be returned only when an impossible situation arises,
// and processing needs to abort.  Note that returning an error stops
// the workflow.  Manual intervention will be needed to move the
// document further.
//
// N. B. NodeFunc instances must be referentially transparent --
// stateless and not capture their environment in any manner.
// Unexpected bad things could happen otherwise!
type NodeFunc func(*Document, *DocEvent) *Message

// DefNodeFunc prepares a simple message that can be posted to
// applicable mailboxes.
func DefNodeFunc(d *Document, event *DocEvent) *Message {
	return &Message{
		DocType: DocType{
			ID: d.DocType.ID,
		},
		DocID: d.ID,
		Event: event.ID,
		Title: d.Title,
		Data:  event.Text,
	}
}

var (
	// reDocPath defines the regular expression for each component of
	// a document's path.
	reDocPath = regexp.MustCompile("[0-9]+?:[0-9]+?/")
)

// DocPath helps in managing document hierarchies.  It provides a set
// of utility methods that ease path management.
type DocPath string

// Root answers the root document information.
func (p *DocPath) Root() (DocTypeID, DocumentID, error) {
	root := reDocPath.FindString(string(*p))
	if root == "" {
		return 0, 0, nil
	}

	parts := strings.Split(root, ":")
	dtid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	did, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return DocTypeID(dtid), DocumentID(did), nil
}

// Components answers a sequence of this path's components, in order.
func (p *DocPath) Components() ([]struct {
	DocTypeID
	DocumentID
}, error) {
	comps := reDocPath.FindAllString(string(*p), -1)
	if len(comps) == 0 {
		return nil, nil
	}

	ary := []struct {
		DocTypeID
		DocumentID
	}{}
	for _, comp := range comps {
		parts := strings.Split(comp, ":")
		dtid, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		did, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}

		ary = append(ary, struct {
			DocTypeID
			DocumentID
		}{DocTypeID(dtid), DocumentID(did)})
	}

	return ary, nil
}

// Append adds the given document type-document ID pair to this path,
// updating it as a result.
func (p *DocPath) Append(dtid DocTypeID, did DocumentID) error {
	if dtid <= 0 || did <= 0 {
		return errors.New("document type ID and document ID should be positive integers")
	}

	*p = *p + DocPath(fmt.Sprintf("%d:%d/", dtid, did))
	return nil
}

const (
	// DefACRoleCount is the default number of roles a group can have
	// in an access context.
	DefACRoleCount = 1
)
