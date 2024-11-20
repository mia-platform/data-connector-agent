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

package webhook

import (
	"fmt"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	logger, _ := test.NewNullLogger()

	testCases := map[string]struct {
		rawData string
		events  *Events

		expectError           string
		expectedID            string
		expectedOperationType entities.Operation
	}{
		"without id": {
			rawData: `{"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						FieldID:   "issue.id",
						Operation: entities.Write,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectError: "missing id field in event: issue.id",
		},
		"supported write event": {
			rawData: `{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						FieldID:   "issue.id",
						Operation: entities.Write,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectedID:            "my-id",
			expectedOperationType: entities.Write,
		},
		"supported delete event": {
			rawData: `{"issue":{"id":"my-id"},"webhookEvent": "my-event"}`,
			events: &Events{
				Supported: map[string]Event{
					"my-event": {
						FieldID:   "issue.id",
						Operation: entities.Delete,
					},
				},
				EventTypeFieldPath: "webhookEvent",
			},
			expectedOperationType: entities.Delete,
			expectedID:            "my-id",
		},
		"unsupported_event": {
			rawData: `{"issue": {"id": "my-id", "key": "TEST-1"}, "webhookEvent": "unsupported"}`,
			events: &Events{
				EventTypeFieldPath: "webhookEvent",
			},

			expectError: fmt.Sprintf("%s: %s", ErrUnsupportedWebhookEvent, "unsupported"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := tc.events.getPipelineEvent(logrus.NewEntry(logger), []byte(tc.rawData))
			if tc.expectError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectError)
			} else {
				require.NoError(t, err)

				require.Equal(t, &entities.Event{
					ID:            tc.expectedID,
					OperationType: tc.expectedOperationType,

					OriginalRaw: []byte(tc.rawData),
				}, event)
			}
		})
	}
}
