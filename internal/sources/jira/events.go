// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

package jira

import (
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	issueCreated     = "jira:issue_created"
	issueUpdated     = "jira:issue_updated"
	issueDeleted     = "jira:issue_deleted"
	issueLinkCreated = "issuelink_created"
	issueLinkDeleted = "issuelink_deleted"

	issueEventIDPath     = "issue.id"
	issuelinkEventIDPath = "issueLink.id"
)

var DefaultSupportedEvents = webhook.Events{
	Supported: map[string]webhook.Event{
		issueCreated: {
			Operation: entities.Write,
			FieldID:   issueEventIDPath,
		},
		issueUpdated: {
			Operation: entities.Write,
			FieldID:   issueEventIDPath,
		},
		issueDeleted: {
			Operation: entities.Delete,
			FieldID:   issueEventIDPath,
		},
		issueLinkCreated: {
			Operation: entities.Write,
			FieldID:   issuelinkEventIDPath,
		},
		issueLinkDeleted: {
			Operation: entities.Delete,
			FieldID:   issuelinkEventIDPath,
		},
	},
	EventTypeFieldPath: webhookEventPath,
}
