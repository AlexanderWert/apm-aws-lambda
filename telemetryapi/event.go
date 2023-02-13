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
	"context"
	"time"

	"github.com/elastic/apm-aws-lambda/extension"
)

// TelEventType represents the log type that is received in the log messages
type TelEventType string

const (
	// PlatformRuntimeDone event is sent when lambda function is finished it's execution
	PlatformInitStart             TelEventType = "platform.initStart"
	PlatformInitRuntimeDone       TelEventType = "platform.initRuntimeDone"
	PlatformInitReport            TelEventType = "platform.initReport"
	PlatformRuntimeDone           TelEventType = "platform.runtimeDone"
	PlatformFault                 TelEventType = "platform.fault"
	PlatformReport                TelEventType = "platform.report"
	PlatformLogsDropped           TelEventType = "platform.logsDropped"
	PlatformStart                 TelEventType = "platform.start"
	PlatformRestoreStart          TelEventType = "platform.restoreStart"
	PlatformRestoreRuntimeDone    TelEventType = "platform.restoreRuntimeDone"
	PlatformRestoreReport         TelEventType = "platform.restoreReport"
	PlatformTelemetrySubscription TelEventType = "platform.telemetrySubscription"
	FunctionLog                   TelEventType = "function"
	ExtensionLog                  TelEventType = "extension"
)

// TelEvent represents an event received from the Telemetry API
type TelEvent struct {
	Time         time.Time    `json:"time"`
	Type         TelEventType `json:"type"`
	StringRecord string
	Record       TelEventRecord
}

// TelEventRecord is a sub-object in a Telemetry API event
type TelEventRecord struct {
	RequestID string          `json:"requestId"`
	Status    string          `json:"status"`
	Metrics   PlatformMetrics `json:"metrics"`
}

// ProcessEvents consumes telemetry events until there are no more telemetry events that
// can be consumed or ctx is cancelled. For INVOKE event this state is
// reached when runtimeDone event for the current requestID is processed
// whereas for SHUTDOWN event this state is reached when the platformReport
// event for the previous requestID is processed.
func (tc *Client) ProcessEvents(
	ctx context.Context,
	requestID string,
	invokedFnArn string,
	dataChan chan []byte,
	prevEvent *extension.NextEventResponse,
	isShutdown bool,
) {
	// platformStartReqID is to identify the requestID for the function
	// events under the assumption that function events for a specific request
	// ID will be bounded by PlatformStart and PlatformEnd events.
	var platformStartReqID string
	for {
		select {
		case telEvent := <-tc.telChannel:
			tc.logger.Debugf("Received telemetry event %v for request ID %s", telEvent.Type, telEvent.Record.RequestID)
			switch telEvent.Type {
			case PlatformStart:
				platformStartReqID = telEvent.Record.RequestID
			case PlatformRuntimeDone:
				if err := tc.invocationLifecycler.OnLambdaLogRuntimeDone(
					telEvent.Record.RequestID,
					telEvent.Record.Status,
					telEvent.Time,
				); err != nil {
					tc.logger.Warnf("Failed to finalize invocation with request ID %s: %v", telEvent.Record.RequestID, err)
				}
				// For invocation events the platform.runtimeDone would be the last possible event.
				if !isShutdown && telEvent.Record.RequestID == requestID {
					tc.logger.Debugf(
						"Processed runtime done event for reqID %s as the last log event for the invocation",
						telEvent.Record.RequestID,
					)
					return
				}
			case PlatformReport:
				// TODO: @lahsivjar Refactor usage of prevEvent.RequestID (should now query the batch?)
				if prevEvent != nil && telEvent.Record.RequestID == prevEvent.RequestID {
					tc.logger.Debugf("Received platform report for %s", telEvent.Record.RequestID)
					processedMetrics, err := ProcessPlatformReport(prevEvent, telEvent)
					if err != nil {
						tc.logger.Errorf("Error processing Lambda platform metrics: %v", err)
					} else {
						select {
						case dataChan <- processedMetrics:
						case <-ctx.Done():
						}
					}
					// For shutdown event the platform report metrics for the previous log event
					// would be the last possible log event.
					if isShutdown {
						tc.logger.Debugf(
							"Processed platform report event for reqID %s as the last log event before shutdown",
							telEvent.Record.RequestID,
						)
						return
					}
				} else {
					tc.logger.Warn("Report event request id didn't match the previous event id")
				}
			case PlatformLogsDropped:
				tc.logger.Warnf("Telemetry events dropped due to extension falling behind: %v", telEvent.Record)
			case FunctionLog:
				processedEvent, err := ProcessFunctionTelemetry(
					platformStartReqID,
					invokedFnArn,
					telEvent,
				)
				if err != nil {
					tc.logger.Warnf("Error processing function event : %v", err)
				} else {
					select {
					case dataChan <- processedEvent:
					case <-ctx.Done():
					}
				}
			}
		case <-ctx.Done():
			tc.logger.Debug("Current invocation over. Interrupting telemetry events processing goroutine")
			return
		}
	}
}
