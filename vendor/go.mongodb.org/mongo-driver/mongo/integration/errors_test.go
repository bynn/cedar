// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// +build go1.13

package integration

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/internal/testutil/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestErrors(t *testing.T) {
	mt := mtest.New(t, noClientOpts)
	defer mt.Close()

	mt.RunOpts("errors are wrapped", noClientOpts, func(mt *mtest.T) {
		mt.Run("network error during application operation", func(mt *mtest.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := mt.Client.Ping(ctx, mtest.PrimaryRp)
			assert.True(mt, errors.Is(err, context.Canceled), "expected error %v, got %v", context.Canceled, err)
		})

		authOpts := mtest.NewOptions().Auth(true).Topologies(mtest.ReplicaSet, mtest.Single).MinServerVersion("4.0")
		mt.RunOpts("network error during auth", authOpts, func(mt *mtest.T) {
			mt.SetFailPoint(mtest.FailPoint{
				ConfigureFailPoint: "failCommand",
				Mode: mtest.FailPointMode{
					Times: 1,
				},
				Data: mtest.FailPointData{
					// Set the fail point for saslContinue rather than saslStart because the driver will use speculative
					// auth on 4.4+ so there won't be an explicit saslStart command.
					FailCommands:    []string{"saslContinue"},
					CloseConnection: true,
				},
			})

			client, err := mongo.Connect(mtest.Background, options.Client().ApplyURI(mtest.ClusterURI()))
			assert.Nil(mt, err, "Connect error: %v", err)
			defer client.Disconnect(mtest.Background)

			// A connection getting closed should manifest as an io.EOF error.
			err = client.Ping(mtest.Background, mtest.PrimaryRp)
			assert.True(mt, errors.Is(err, io.EOF), "expected error %v, got %v", io.EOF, err)
		})
	})

	mt.RunOpts("network timeouts", noClientOpts, func(mt *mtest.T) {
		mt.Run("context timeouts return DeadlineExceeded", func(mt *mtest.T) {
			_, err := mt.Coll.InsertOne(mtest.Background, bson.D{{"x", 1}})
			assert.Nil(mt, err, "InsertOne error: %v", err)

			mt.ClearEvents()
			filter := bson.M{
				"$where": "function() { sleep(1000); return false; }",
			}
			timeoutCtx, cancel := context.WithTimeout(mtest.Background, 100*time.Millisecond)
			defer cancel()
			_, err = mt.Coll.Find(timeoutCtx, filter)

			evt := mt.GetStartedEvent()
			assert.Equal(mt, "find", evt.CommandName, "expected command 'find', got %q", evt.CommandName)
			assert.True(mt, errors.Is(err, context.DeadlineExceeded),
				"errors.Is failure: expected error %v to be %v", err, context.DeadlineExceeded)
		})

		mt.Run("socketTimeoutMS timeouts return network errors", func(mt *mtest.T) {
			_, err := mt.Coll.InsertOne(mtest.Background, bson.D{{"x", 1}})
			assert.Nil(mt, err, "InsertOne error: %v", err)

			// Reset the test client to have a 100ms socket timeout. We do this here rather than passing it in as a
			// test option using mt.RunOpts because that could cause the collection creation or InsertOne to fail.
			resetClientOpts := options.Client().
				SetSocketTimeout(100 * time.Millisecond)
			mt.ResetClient(resetClientOpts)

			mt.ClearEvents()
			filter := bson.M{
				"$where": "function() { sleep(1000); return false; }",
			}
			_, err = mt.Coll.Find(mtest.Background, filter)

			evt := mt.GetStartedEvent()
			assert.Equal(mt, "find", evt.CommandName, "expected command 'find', got %q", evt.CommandName)

			assert.False(mt, errors.Is(err, context.DeadlineExceeded),
				"errors.Is failure: expected error %v to not be %v", err, context.DeadlineExceeded)
			var netErr net.Error
			ok := errors.As(err, &netErr)
			assert.True(mt, ok, "errors.As failure: expected error %v to be a net.Error", err)
			assert.True(mt, netErr.Timeout(), "expected error %v to be a network timeout", err)
		})
	})
	mt.Run("ServerError", func(mt *mtest.T) {
		matchWrapped := errors.New("wrapped err")
		otherWrapped := errors.New("other err")
		const matchCode = 100
		const otherCode = 120
		const label = "testError"
		testCases := []struct {
			name               string
			err                mongo.ServerError
			hasCode            bool
			hasLabel           bool
			hasMessage         bool
			hasCodeWithMessage bool
			isResult           bool
		}{
			{
				"CommandError all true",
				mongo.CommandError{matchCode, "foo", []string{label}, "name", matchWrapped},
				true,
				true,
				true,
				true,
				true,
			},
			{
				"CommandError all false",
				mongo.CommandError{otherCode, "bar", []string{"otherError"}, "name", otherWrapped},
				false,
				false,
				false,
				false,
				false,
			},
			{
				"CommandError has code not message",
				mongo.CommandError{matchCode, "bar", []string{}, "name", nil},
				true,
				false,
				false,
				false,
				false,
			},
			{
				"WriteException all in writeConcernError",
				mongo.WriteException{
					&mongo.WriteConcernError{"name", matchCode, "foo", nil},
					nil,
					[]string{label},
				},
				true,
				true,
				true,
				true,
				false,
			},
			{
				"WriteException all in writeError",
				mongo.WriteException{
					nil,
					mongo.WriteErrors{
						mongo.WriteError{0, otherCode, "bar"},
						mongo.WriteError{0, matchCode, "foo"},
					},
					[]string{"otherError"},
				},
				true,
				false,
				true,
				true,
				false,
			},
			{
				"WriteException all false",
				mongo.WriteException{
					&mongo.WriteConcernError{"name", otherCode, "bar", nil},
					mongo.WriteErrors{
						mongo.WriteError{0, otherCode, "baz"},
					},
					[]string{"otherError"},
				},
				false,
				false,
				false,
				false,
				false,
			},
			{
				"WriteException HasErrorCodeAndMessage false",
				mongo.WriteException{
					&mongo.WriteConcernError{"name", matchCode, "bar", nil},
					mongo.WriteErrors{
						mongo.WriteError{0, otherCode, "foo"},
					},
					[]string{"otherError"},
				},
				true,
				false,
				true,
				false,
				false,
			},
			{
				"BulkWriteException all in writeConcernError",
				mongo.BulkWriteException{
					&mongo.WriteConcernError{"name", matchCode, "foo", nil},
					nil,
					[]string{label},
				},
				true,
				true,
				true,
				true,
				false,
			},
			{
				"BulkWriteException all in writeError",
				mongo.BulkWriteException{
					nil,
					[]mongo.BulkWriteError{
						{mongo.WriteError{0, matchCode, "foo"}, &mongo.InsertOneModel{}},
						{mongo.WriteError{0, otherCode, "bar"}, &mongo.InsertOneModel{}},
					},
					[]string{"otherError"},
				},
				true,
				false,
				true,
				true,
				false,
			},
			{
				"BulkWriteException all false",
				mongo.BulkWriteException{
					&mongo.WriteConcernError{"name", otherCode, "bar", nil},
					[]mongo.BulkWriteError{
						{mongo.WriteError{0, otherCode, "baz"}, &mongo.InsertOneModel{}},
					},
					[]string{"otherError"},
				},
				false,
				false,
				false,
				false,
				false,
			},
			{
				"BulkWriteException HasErrorCodeAndMessage false",
				mongo.BulkWriteException{
					&mongo.WriteConcernError{"name", matchCode, "bar", nil},
					[]mongo.BulkWriteError{
						{mongo.WriteError{0, otherCode, "foo"}, &mongo.InsertOneModel{}},
					},
					[]string{"otherError"},
				},
				true,
				false,
				true,
				false,
				false,
			},
		}
		for _, tc := range testCases {
			mt.Run(tc.name, func(mt *mtest.T) {
				has := tc.err.HasErrorCode(matchCode)
				assert.Equal(mt, has, tc.hasCode, "expected HasErrorCode to return %v, got %v", tc.hasCode, has)
				has = tc.err.HasErrorLabel(label)
				assert.Equal(mt, has, tc.hasLabel, "expected HasErrorLabel to return %v, got %v", tc.hasLabel, has)

				// Check for full message and substring
				has = tc.err.HasErrorMessage("foo")
				assert.Equal(mt, has, tc.hasMessage, "expected HasErrorMessage to return %v, got %v", tc.hasMessage, has)
				has = tc.err.HasErrorMessage("fo")
				assert.Equal(mt, has, tc.hasMessage, "expected HasErrorMessage to return %v, got %v", tc.hasMessage, has)
				has = tc.err.HasErrorCodeWithMessage(matchCode, "foo")
				assert.Equal(mt, has, tc.hasCodeWithMessage, "expected HasErrorCodeWithMessage to return %v, got %v", tc.hasCodeWithMessage, has)
				has = tc.err.HasErrorCodeWithMessage(matchCode, "fo")
				assert.Equal(mt, has, tc.hasCodeWithMessage, "expected HasErrorCodeWithMessage to return %v, got %v", tc.hasCodeWithMessage, has)

				assert.Equal(mt, errors.Is(tc.err, matchWrapped), tc.isResult, "expected errors.Is result to be %v", tc.isResult)
			})
		}
	})
}
