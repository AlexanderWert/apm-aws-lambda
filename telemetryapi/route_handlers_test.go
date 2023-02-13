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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelEventUnmarshalReport(t *testing.T) {
	te := new(TelEvent)
	reportJSON := []byte(`{
		    "time": "2020-08-20T12:31:32.123Z",
		    "type": "platform.report",
		    "record": {"requestId": "6f7f0961f83442118a7af6fe80b88d56",
			"metrics": {"durationMs": 101.51,
			    "billedDurationMs": 300,
			    "memorySizeMB": 512,
			    "maxMemoryUsedMB": 33,
			    "initDurationMs": 116.67
			}
		    }
		}`)

	err := te.UnmarshalJSON(reportJSON)
	require.NoError(t, err)
	assert.Equal(t, TelEventType("platform.report"), te.Type)
	assert.Equal(t, "2020-08-20T12:31:32.123Z", te.Time.Format(time.RFC3339Nano))
	rec := TelEventRecord{
		RequestID: "6f7f0961f83442118a7af6fe80b88d56",
		Status:    "", // no status was given in sample json
		Metrics: PlatformMetrics{
			DurationMs:       101.51,
			BilledDurationMs: 300,
			MemorySizeMB:     512,
			MaxMemoryUsedMB:  33,
			InitDurationMs:   116.67,
		},
	}
	assert.Equal(t, rec, te.Record)

}

func TestTelemetryEventUnmarshalFault(t *testing.T) {
	te := new(TelEvent)
	reportJSON := []byte(` {
		    "time": "2020-08-20T12:31:32.123Z",
		    "type": "platform.fault",
		    "record": "RequestId: d783b35e-a91d-4251-af17-035953428a2c Process exited before completing request"
		}`)

	err := te.UnmarshalJSON(reportJSON)
	require.NoError(t, err)
	assert.Equal(t, TelEventType("platform.fault"), te.Type)
	assert.Equal(t, "2020-08-20T12:31:32.123Z", te.Time.Format(time.RFC3339Nano))
	rec := "RequestId: d783b35e-a91d-4251-af17-035953428a2c Process exited before completing request"
	assert.Equal(t, rec, te.StringRecord)

}

func Test_unmarshalRuntimeDoneRecordObject(t *testing.T) {
	te := new(TelEvent)
	jsonBytes := []byte(`
	{
		"time": "2021-02-04T20:00:05.123Z",
		"type": "platform.runtimeDone",
		"record": {
		   "requestId":"6f7f0961f83442118a7af6fe80b88",
		   "status": "success"
		}
	}
	`)

	err := te.UnmarshalJSON(jsonBytes)
	require.NoError(t, err)
	assert.Equal(t, TelEventType("platform.runtimeDone"), te.Type)
	assert.Equal(t, "2021-02-04T20:00:05.123Z", te.Time.Format(time.RFC3339Nano))
	rec := TelEventRecord{
		RequestID: "6f7f0961f83442118a7af6fe80b88",
		Status:    "success",
	}
	assert.Equal(t, rec, te.Record)
}

func Test_unmarshalRuntimeDoneRecordString(t *testing.T) {
	te := new(TelEvent)
	jsonBytes := []byte(`
	{
		"time": "2021-02-04T20:00:05.123Z",
		"type": "platform.runtimeDone",
		"record": "Unknown application error occurred"
	}
	`)

	err := te.UnmarshalJSON(jsonBytes)
	require.NoError(t, err)
	assert.Equal(t, TelEventType("platform.runtimeDone"), te.Type)
	assert.Equal(t, "2021-02-04T20:00:05.123Z", te.Time.Format(time.RFC3339Nano))
	assert.Equal(t, "Unknown application error occurred", te.StringRecord)
}

func Test_unmarshalRuntimeDoneFaultRecordString(t *testing.T) {
	jsonBytes := []byte(`
		{
			"time": "2021-02-04T20:00:05.123Z",
			"type": "platform.fault",
			"record": "Unknown application error occurred"
		}
	`)

	var te TelEvent
	if err := json.Unmarshal(jsonBytes, &te); err != nil {
		t.Fail()
	}

	timeValue, _ := time.Parse(time.RFC3339, "2021-02-04T20:00:05.123Z")
	assert.Equal(t, timeValue, te.Time)
	assert.Equal(t, PlatformFault, te.Type)
	assert.Equal(t, "Unknown application error occurred", te.StringRecord)
}
