// Copyright 2019, OpenTelemetry Authors
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

// Command jaeger is an example program that creates spans
// and uploads to Jaeger.
package main

import (
	"context"
	"log"

	apitrace "go.opentelemetry.io/api/trace"
	"go.opentelemetry.io/exporter/trace/jaeger"
	"go.opentelemetry.io/sdk/trace"
)

func main() {
	trace.Register()
	ctx := context.Background()

	// Create Jaeger Exporter
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: "http://localhost:14268/api/traces",
		Process: jaeger.Process{
			ServiceName: "trace-demo",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Wrap exporter with SimpleSpanProcessor and register the processor.
	ssp := trace.NewSimpleSpanProcessor(exporter)
	trace.RegisterSpanProcessor(ssp)

	// For demoing purposes, always sample. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	ctx, span := apitrace.GlobalTracer().Start(ctx, "/foo")
	bar(ctx)
	span.End()

	exporter.Flush()
}

func bar(ctx context.Context) {
	_, span := apitrace.GlobalTracer().Start(ctx, "/bar")
	defer span.End()

	// Do bar...
}
