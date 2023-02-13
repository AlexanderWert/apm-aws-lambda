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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

// SubscribeRequest is the request body that is sent to Telemetry API on subscribe
type SubscribeRequest struct {
	SchemaVersion  SchemaVersion      `json:"schemaVersion"`
	TelemetryTypes []SubscriptionType `json:"types"`
	BufferingCfg   BufferingCfg       `json:"buffering"`
	Destination    Destination        `json:"destination"`
}

// SchemaVersion is the Lambda runtime API schema version
type SchemaVersion string

const (
	SchemaVersion20210318 = "2021-03-18"
	SchemaVersionLatest   = SchemaVersion20210318
)

// BufferingCfg is the configuration set for receiving events from Telemetry API. Whichever of the conditions below is met first, the events will be sent
type BufferingCfg struct {
	// MaxItems is the maximum number of events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxItems uint32 `json:"maxItems"`
	// MaxBytes is the maximum size in bytes of the events to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	MaxBytes uint32 `json:"maxBytes"`
	// TimeoutMS is the maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
	TimeoutMS uint32 `json:"timeoutMs"`
}

// Destination is the configuration for listeners who would like to receive events with HTTP
type Destination struct {
	Protocol   string `json:"protocol"`
	URI        string `json:"URI"`
	HTTPMethod string `json:"method"`
	Encoding   string `json:"encoding"`
}

func (tc *Client) startHTTPServer() (string, error) {
	listener, err := net.Listen("tcp", tc.listenerAddr)
	if err != nil {
		return "", fmt.Errorf("failed to listen on %s: %w", tc.listenerAddr, err)
	}

	addr := listener.Addr().String()

	go func() {
		tc.logger.Infof("Extension listening for Lambda Telemetry API events on %s", addr)

		if err := tc.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			tc.logger.Errorf("Error upon Telemetry API server start : %v", err)
		}
	}()

	return addr, nil
}

func (tc *Client) subscribe(types []SubscriptionType, extensionID string, uri string) error {
	data, err := json.Marshal(&SubscribeRequest{
		SchemaVersion:  SchemaVersionLatest,
		TelemetryTypes: types,
		BufferingCfg: BufferingCfg{
			MaxItems:  10000,
			MaxBytes:  1024 * 1024,
			TimeoutMS: 100,
		},
		Destination: Destination{
			Protocol:   "HTTP",
			URI:        uri,
			HTTPMethod: http.MethodPost,
			Encoding:   "JSON",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal SubscribeRequest: %w", err)
	}

	url := fmt.Sprintf("%s/2022-07-01/telemetry", tc.telAPIBaseURL)
	resp, err := tc.sendRequest(url, data, extensionID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return errors.New("telemetry API is not supported in this environment")
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s failed: %d[%s]", url, resp.StatusCode, resp.Status)
		}

		return fmt.Errorf("%s failed: %d[%s] %s", url, resp.StatusCode, resp.Status, string(body))
	}

	return nil
}

func (tc *Client) sendRequest(url string, data []byte, extensionID string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
