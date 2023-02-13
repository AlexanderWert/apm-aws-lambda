// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package telemetryapi

import (
	"go.elastic.co/apm/v2/model"
	"go.elastic.co/fastjson"
)

type logContainer struct {
	Log *logLine
}

type logLine struct {
	Timestamp model.Time
	Message   string
	FAAS      *faas
}

func (t *logLine) MarshalFastJSON(w *fastjson.Writer) error {
	var firstErr error
	w.RawString("{\"message\":")
	w.String(t.Message)
	w.RawString(",\"@timestamp\":")
	if err := t.Timestamp.MarshalFastJSON(w); err != nil && firstErr == nil {
		firstErr = err
	}
	if t.FAAS != nil {
		w.RawString(",\"faas\":")
		if err := t.FAAS.MarshalFastJSON(w); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	w.RawByte('}')
	return firstErr
}

// faas struct is a subset of go.elastic.co/apm/v2/model#FAAS
//
// The purpose of having a separate struct is to have a custom
// marshalling logic that is targeted for the faas fields
// available for function events. For example: `coldstart` value
// cannot be inferred for function events so this struct drops
// the field entirely.
type faas struct {
	// ID holds a unique identifier of the invoked serverless function.
	ID string `json:"id,omitempty"`
	// Execution holds the request ID of the function invocation.
	Execution string `json:"execution,omitempty"`
}

func (f *faas) MarshalFastJSON(w *fastjson.Writer) error {
	w.RawString("{\"id\":")
	w.String(f.ID)
	w.RawString(",\"execution\":")
	w.String(f.Execution)
	w.RawByte('}')
	return nil
}

func (lc logContainer) MarshalFastJSON(json *fastjson.Writer) error {
	json.RawString(`{"log":`)
	if err := lc.Log.MarshalFastJSON(json); err != nil {
		return err
	}
	json.RawByte('}')
	return nil
}

// ProcessFunctionTelemetry processes the `function` telemetry line from Telemetry API and returns
// a byte array containing the JSON body for the extracted telemetry along with the timestamp.
// A non nil error is returned when marshaling of the event into JSON fails.
func ProcessFunctionLog(
	requestID string,
	invokedFnArn string,
	telemetry TelEvent,
) ([]byte, error) {
	lc := logContainer{
		Log: &logLine{
			Timestamp: model.Time(telemetry.Time),
			Message:   telemetry.StringRecord,
		},
	}

	lc.Log.FAAS = &faas{
		ID:        invokedFnArn,
		Execution: requestID,
	}

	var jsonWriter fastjson.Writer
	if err := lc.MarshalFastJSON(&jsonWriter); err != nil {
		return nil, err
	}

	return jsonWriter.Bytes(), nil
}
