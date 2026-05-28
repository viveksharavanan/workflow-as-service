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

// Blob is a simple data holder for information concerning the
// user-supplied name of the binary object, the path of the stored
// binary object, and its SHA1 checksum.
type Blob struct {
	Name    string `json:"Name"`           // User-given name to the binary object
	Path    string `json:"Path,omitempty"` // Path to the stored binary object
	SHA1Sum string `json:"SHA1sum"`        // SHA1 checksum of the binary object
}

// Document represents a task in a workflow, whose life cycle it
// tracks.
//
// Documents are central to the workflow engine and its operations. In
// the process, it accumulates various details, and tracks the times
// of its modifications.  The life cycle typically involves several
// state transitions, whose details are also tracked.
//
// `Document` is a recursive structure: it can contain other
// documents.  Therefore, when a document is created, it is
// initialised with the path that leads from its root document to its
// immediate parent.  For root documents, this path is empty.
//
// Most applications should embed `Document` in their document
// structures rather than use this directly.  That enables them to
// control their data persistence mechanisms, while delegating
// workflow management to `flow`.
type Document struct {
	ID      DocumentID `json:"ID"`      // Globally-unique identifier of this document
	DocType DocType    `json:"DocType"` // For namespacing
	Path    DocPath    `json:"Path"`    // Path leading to, but not including, this document

	AccCtx AccessContext `json:"AccessContext"` // Originating access context of this document; applicable only to a root document
	State  DocState      `json:"DocState"`      // Current state of this document; applicable only to a root document

	Group Group     `json:"Group"` // Creator of this document
	Ctime time.Time `json:"Ctime"` // Creation time of this (possibly child) document

	Title string `json:"Title"`          // Human-readable title; applicable only for root documents
	Data  string `json:"Data,omitempty"` // Primary content of the document
}

// DocumentsNewInput specifies the initial data with which a new
// document has to be created in the system.
type DocumentsNewInput struct {
	DocTypeID                  // Type of the new document; required
	AccessContextID            // Access context in which the document should be created; required
	GroupID                    // (Singleton) group of the creator; required
	ParentType      DocTypeID  // Document type of the parent document, if any
	ParentID        DocumentID // Unique identifier of the parent document, if any
	Title           string     // Title of the new document; applicable to only root (top-level) documents
	Data            string     // Body of the new document; required
}

// DocumentsListInput specifies a set of filter conditions to narrow
// down document listings.
type DocumentsListInput struct {
	DocTypeID                 // Documents of this type are listed; required
	AccessContextID           // Access context from within which to list; required
	GroupID                   // List documents created by this (singleton) group
	DocStateID                // List documents currently in this state
	CtimeStarting   time.Time // List documents created after this time
	CtimeBefore     time.Time // List documents created before this time
	TitleContains   string    // List documents whose title contains the given text; expensive operation
	RootOnly        bool      // List only root (top-level) documents
}
