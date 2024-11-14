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

//go:build integration
// +build integration

package mongo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMongo(t *testing.T) {
	ctx := context.Background()

	mongoURL, db := testutils.GenerateMongoURL(t)
	collection := testutils.RandomString(t, 6)

	w, err := NewMongoDBWriter[*entities.Event](ctx, &Config{
		URL:        config.SecretSource(mongoURL),
		Database:   db,
		Collection: collection,
	})
	require.NoError(t, err)

	coll := testutils.MongoCollection(t, mongoURL, collection, db)
	defer coll.Drop(ctx)

	t.Run("create", func(t *testing.T) {
		e := getTestEvent(t, "123", map[string]any{"foo": "bar", "key": "123"})
		err = w.Write(ctx, e)
		require.NoError(t, err)
		findAllDocuments(t, coll, []map[string]any{
			{"_eventId": "123", "foo": "bar", "key": "123"},
		})
	})

	t.Run("update", func(t *testing.T) {
		e := getTestEvent(t, "123", map[string]any{"foo": "taz", "key": "123", "another": "field"})
		err = w.Write(ctx, e)
		require.NoError(t, err)
		findAllDocuments(t, coll, []map[string]any{
			{"_eventId": "123", "foo": "taz", "key": "123", "another": "field"},
		})
	})

	t.Run("delete", func(t *testing.T) {
		e := getTestEvent(t, "123", nil)
		err = w.Delete(ctx, e)
		require.NoError(t, err)
		findAllDocuments(t, coll, []map[string]any{})
	})
}

func getTestEvent(t *testing.T, id string, data map[string]any) *entities.Event {
	t.Helper()

	e := &entities.Event{
		ID: id,
	}

	d, err := json.Marshal(data)
	require.NoError(t, err)

	e.WithData(d)

	return e
}

func findAllDocuments(t *testing.T, coll *mongo.Collection, expectedResults []map[string]any) {
	t.Helper()

	n, err := coll.CountDocuments(context.Background(), map[string]any{})
	require.NoError(t, err)
	require.Equal(t, int64(len(expectedResults)), n)

	ctx := context.Background()
	docs, err := coll.Find(ctx, map[string]any{})
	require.NoError(t, err)
	results := []map[string]any{}
	err = docs.All(ctx, &results)
	require.NoError(t, err)

	require.Equal(t, expectedResults, testutils.RemoveMongoID(results))
}
