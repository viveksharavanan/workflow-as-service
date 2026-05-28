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
	"time"
)

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents and nodes, events are central to the
// workflow engine in `flow`.  Events cause documents to transition
// from one state to another, usually in response to user actions.  It
// is possible for system events to cause state transitions, as well.
type DocEvent struct {
	ID      DocEventID  `json:"ID"`        // Unique ID of this event
	DocType DocTypeID   `json:"DocType"`   // Document type of the document to which this event is to be applied
	DocID   DocumentID  `json:"DocID"`     // Document to which this event is to be applied
	State   DocStateID  `json:"DocState"`  // Current state of the document must equal this
	Action  DocActionID `json:"DocAction"` // Action performed by the user
	Group   GroupID     `json:"Group"`     // Group (singleton) who caused this action
	Text    string      `json:"Text"`      // Comment or other content
	Ctime   time.Time   `json:"Ctime"`     // Time at which the event occurred
	Status  EventStatus `json:"Status"`    // Status of this event
}

// DocEventsNewInput holds information needed to create a new document
// event in the system.
type DocEventsNewInput struct {
	DocTypeID          // Type of the document; required
	DocumentID         // Unique identifier of the document; required
	DocStateID         // Document must be in this state for this event to be applied; required
	DocActionID        // Action performed by `Group`; required
	GroupID            // Group (user) who performed the action that raised this event; required
	Text        string // Any comments or notes; required
}

// DocEventsListInput specifies a set of filter conditions to narrow
// down document listings.
type DocEventsListInput struct {
	DocTypeID                   // Events on documents of this type are listed
	AccessContextID             // Access context from within which to list
	GroupID                     // List events created by this (singleton) group
	DocStateID                  // List events acting on this state
	CtimeStarting   time.Time   // List events created after this time
	CtimeBefore     time.Time   // List events created before this time
	Status          EventStatus // List events that are in this state of application
}
