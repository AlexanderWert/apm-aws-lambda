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

package accumulator

import (
	cryptorand "crypto/rand"
	"fmt"
	"time"

	"go.elastic.co/apm/v2/model"
	"go.elastic.co/fastjson"
)

type InitSpan struct {
	SpanData model.Span
}

func NewInitSpan(parentTxID model.SpanID, traceID model.TraceID, timestamp int64, duration float32) (*InitSpan, error) {
	var spanID model.SpanID
	if _, err := cryptorand.Read(spanID[:]); err != nil {
		return nil, fmt.Errorf("failed generating span ID for init span")
	}

	var sampleRate float64 = 1.0

	initSpan := &model.Span{
		Name:          "AWS Lambda Init",
		Timestamp:     model.Time(time.UnixMicro(timestamp)),
		Duration:      float64(duration),
		Type:          "awslambda",
		Subtype:       "init",
		ID:            spanID,
		TransactionID: parentTxID,
		TraceID:       traceID,
		ParentID:      parentTxID,
		SampleRate:    &sampleRate,
	}

	return &InitSpan{
		SpanData: *initSpan,
	}, nil
}

func (s *InitSpan) GetBytes() ([]byte, error) {
	var json fastjson.Writer
	json.RawString(`{"span":`)
	if err := s.SpanData.MarshalFastJSON(&json); err != nil {
		return nil, err
	}
	json.RawByte('}')
	return json.Bytes(), nil
}
