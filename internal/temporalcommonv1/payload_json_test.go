// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package common_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/temporalproto"
)

var marshaler temporalproto.JSONMarshalOptions
var unmarshaler temporalproto.JSONUnmarshalOptions

var binaryVal = base64.StdEncoding.EncodeToString([]byte("bytes"))

func TestMaybeMarshal(t *testing.T) {
	tests := []struct {
		name     string
		given    proto.Message
		expected string
	}{{
		name: "binary/plain",
		given: &common.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte("binary/plain"),
			},
			Data: []byte("bytes")},
		expected: `{"metadata": {"encoding": "YmluYXJ5L3BsYWlu"}, "data": "Ynl0ZXM="}`,
	}, {
		name:     "json/plain",
		expected: `{"_protoMessageType": "temporal.api.common.v1.WorkflowType", "name": "workflow-name"}`,
		given: &common.Payload{
			Metadata: map[string][]byte{
				"encoding":    []byte("json/protobuf"),
				"messageType": []byte("temporal.api.common.v1.WorkflowType"),
			},
			Data: []byte(`{"name":"workflow-name"}`),
		},
	}, {
		name: "memo with varying fields",
		expected: `{"fields": {
                      "some-string": "string",
			          "some-array": ["foo", 123, false]}}`,
		given: &common.Memo{
			Fields: map[string]*common.Payload{
				"some-string": &common.Payload{
					Metadata: map[string][]byte{
						"encoding": []byte("json/plain"),
					},
					Data: []byte(`"string"`),
				},
				"some-array": &common.Payload{
					Metadata: map[string][]byte{
						"encoding": []byte("json/plain"),
					},
					Data: []byte(`["foo", 123, false]`),
				},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts temporalproto.JSONMarshalOptions
			got, err := opts.Marshal(tt.given)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(got), tt.given)
		})
	}
}
