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

package app

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

type appConfig struct {
	awsLambdaRuntimeAPI                 string
	awsConfig                           aws.Config
	extensionName                       string
	disableTelemetryAPI                 bool
	enableFunctionTelemetrySubscription bool
	logLevel                            string
	telemetryapiAddr                    string
}

// ConfigOption is used to configure the lambda extension
type ConfigOption func(*appConfig)

// WithLambdaRuntimeAPI sets the AWS Lambda Runtime API
// endpoint (normally taken from $AWS_LAMBDA_RUNTIME_API),
// used by the AWS client.
func WithLambdaRuntimeAPI(api string) ConfigOption {
	return func(c *appConfig) {
		c.awsLambdaRuntimeAPI = api
	}
}

// WithExtensionName sets the extension name.
func WithExtensionName(name string) ConfigOption {
	return func(c *appConfig) {
		c.extensionName = name
	}
}

// WithoutTelemetryAPI disables the telemetry api.
func WithoutTelemetryAPI() ConfigOption {
	return func(c *appConfig) {
		c.disableTelemetryAPI = true
	}
}

// WithFunctionTelemetrySubscription enables the telemetry api subscription
// to function log stream. This option will only work if TelemetryAPI
// is not disabled by the WithoutTelemetryAPI config option.
func WithFunctionTelemetrySubscription() ConfigOption {
	return func(c *appConfig) {
		c.enableFunctionTelemetrySubscription = true
	}
}

// WithLogLevel sets the log level.
func WithLogLevel(level string) ConfigOption {
	return func(c *appConfig) {
		c.logLevel = level
	}
}

// WithTelemetryapiAddress sets the listener address of the
// server listening for telemetry event.
func WithTelemetryapiAddress(s string) ConfigOption {
	return func(c *appConfig) {
		c.telemetryapiAddr = s
	}
}

// WithAWSConfig sets the AWS config.
func WithAWSConfig(awsConfig aws.Config) ConfigOption {
	return func(c *appConfig) {
		c.awsConfig = awsConfig
	}
}
