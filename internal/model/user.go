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

// User represents any kind of a user invoking or otherwise
// participating in a defined workflow in the system.
//
// User details are expected to be provided by an external identity
// provider application or directory.  `flow` neither defines nor
// manages users.
type User struct {
	ID        UserID `json:"ID"`               // Must be globally-unique
	FirstName string `json:"FirstName"`        // For display purposes only
	LastName  string `json:"LastName"`         // For display purposes only
	Email     string `json:"Email"`            // E-mail address of this user
	Active    bool   `json:"Active,omitempty"` // Is this user account active?
}
