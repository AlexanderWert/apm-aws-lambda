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
	"go.uber.org/zap"
)

// WithListenerAddress sets the listener address of the
// server listening for telemetry events.
func WithListenerAddress(s string) ClientOption {
	return func(c *Client) {
		c.listenerAddr = s
	}
}

// WithTelemetryAPIBaseURL sets the telemetry api base url.
func WithTelemetryAPIBaseURL(s string) ClientOption {
	return func(c *Client) {
		c.telAPIBaseURL = s
	}
}

// WithLogBuffer sets the size of the buffer
// storing queued telemetry for processing.
func WithLogBuffer(size int) ClientOption {
	return func(c *Client) {
		c.telChannel = make(chan TelEvent, size)
	}
}

// WithLogger sets the logger.
func WithLogger(logger *zap.SugaredLogger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithSubscriptionTypes sets the streams that the Telemetry API will subscribe to.
func WithSubscriptionTypes(types ...SubscriptionType) ClientOption {
	return func(c *Client) {
		c.telAPISubscriptionTypes = types
	}
}

// WithInvocationLifecycler configures a lifecycler for acting on certain
// telemetry events.
func WithInvocationLifecycler(l invocationLifecycler) ClientOption {
	return func(c *Client) {
		c.invocationLifecycler = l
	}
}
