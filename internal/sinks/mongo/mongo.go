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

package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var (
	ErrMongoInitialization = errors.New("failed to start mongo writer")
)

type validateFunc func(context.Context, *mongo.Client) error

// Writer is a concrete implementation of a Writer that will save and delete data from a MongoDB instance.
type Writer[T entities.PipelineEvent] struct {
	client *mongo.Client

	database    string
	collection  string
	outputEvent map[string]any
	idField     string
}

// NewMongoDBWriter will construct a new MongoDB writer and validate the connection parameters via a ping request.
func NewMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config *Config) (sinks.Sink[T], error) {
	return newMongoDBWriter[T](ctx, config, func(ctx context.Context, c *mongo.Client) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return c.Ping(ctx, nil)
	})
}

func newMongoDBWriter[T entities.PipelineEvent](ctx context.Context, config *Config, validate validateFunc) (sinks.Sink[T], error) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	options, db, collection := mongoClientOptionsFromConfig(config)

	client, err := mongo.Connect(ctxWithCancel, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMongoInitialization, err)
	}

	if err := validate(ctxWithCancel, client); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrMongoInitialization, err)
	}

	return &Writer[T]{
		client:     client,
		database:   db,
		collection: collection,
		idField:    "_eventId",
	}, nil
}

// Write implement the Writer interface. The MongoDBWriter will do an upsert of data using its id as primary key
func (w *Writer[T]) Write(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := options.Replace()
	opts.SetUpsert(true)

	queryFilter, err := w.idFilter(data)
	if err != nil {
		return err
	}

	dataToUpsert, err := w.bsonData(data)
	if err != nil {
		return err
	}

	result, err := w.client.Database(w.database).
		Collection(w.collection).
		ReplaceOne(ctxWithCancel, queryFilter, dataToUpsert, opts)
	if err != nil {
		return err
	}

	if result.UpsertedCount != 1 && result.ModifiedCount != 1 {
		return fmt.Errorf("error upserting data: %d documents upserted", result.UpsertedCount)
	}

	return nil
}

// Delete implement the Writer interface
func (w *Writer[T]) Delete(ctx context.Context, data T) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	queryFilter, err := w.idFilter(data)
	if err != nil {
		return err
	}

	opts := options.Delete()
	result, err := w.client.Database(w.database).
		Collection(w.collection).
		DeleteOne(ctxWithCancel, queryFilter, opts)
	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("error deleting data: %d documents deleted", result.DeletedCount)
	}

	return nil
}

// mongoClientOptionsFromConfig return a ClientOptions, database and collection parameters parsed from a
// MongoDBConfig struct.
func mongoClientOptionsFromConfig(config *Config) (*options.ClientOptions, string, string) {
	connectionURI := config.URL.String()
	options := options.Client()
	options.ApplyURI(connectionURI)

	database := config.Database
	if len(database) == 0 {
		if cs, err := connstring.ParseAndValidate(connectionURI); err == nil {
			database = cs.Database
		}
	}

	return options, database, config.Collection
}

func (w Writer[T]) idFilter(event T) (bson.D, error) {
	id := event.GetID()
	if id == "" {
		return bson.D{}, fmt.Errorf("id is empty")
	}
	return bson.D{{Key: w.idField, Value: id}}, nil
}

func (w Writer[T]) bsonData(event T) ([]byte, error) {
	data, err := event.JSON()
	if err != nil {
		return nil, err
	}

	data[w.idField] = event.GetID()

	bsonData, err := bson.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bsonData, nil
}
