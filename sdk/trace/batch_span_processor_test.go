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

package trace_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/api/core"
	apitrace "go.opentelemetry.io/api/trace"
	sdktrace "go.opentelemetry.io/sdk/trace"
)

type testBatchExporter struct {
	mu         sync.Mutex
	spans      []*sdktrace.SpanData
	sizes      []int
	batchCount int
}

func (t *testBatchExporter) ExportSpans(sds []*sdktrace.SpanData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, sds...)
	t.sizes = append(t.sizes, len(sds))
	t.batchCount++
}

func (t *testBatchExporter) len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.spans)
}

func (t *testBatchExporter) getBatchCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.batchCount
}

func (t *testBatchExporter) get(idx int) *sdktrace.SpanData {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.spans[idx]
}

var _ sdktrace.BatchExporter = (*testBatchExporter)(nil)

func init() {
	sdktrace.Register()
	sdktrace.ApplyConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()})
}

func TestNewBatchSpanProcessorWithNilExporter(t *testing.T) {
	_, err := sdktrace.NewBatchSpanProcessor(nil)
	if err == nil {
		t.Errorf("Expected error while creating processor with nil exporter")
	}
}

type testOption struct {
	name           string
	o              []sdktrace.BatchSpanProcessorOption
	wantNumSpans   int
	wantBatchCount int
	genNumSpans    int
	waitTime       time.Duration
}

func TestNewBatchSpanProcessorWithOptions(t *testing.T) {
	schDelay := time.Duration(200 * time.Millisecond)
	waitTime := schDelay + time.Duration(100*time.Millisecond)
	options := []testOption{
		{
			name:           "default BatchSpanProcessorOptions",
			wantNumSpans:   2048,
			wantBatchCount: 4,
			genNumSpans:    2053,
			waitTime:       time.Duration(5100 * time.Millisecond),
		},
		{
			name: "non-default ScheduledDelayMillis",
			o: []sdktrace.BatchSpanProcessorOption{
				sdktrace.WithScheduleDelayMillis(schDelay),
			},
			wantNumSpans:   2048,
			wantBatchCount: 4,
			genNumSpans:    2053,
			waitTime:       waitTime,
		},
		{
			name: "non-default MaxQueueSize and ScheduledDelayMillis",
			o: []sdktrace.BatchSpanProcessorOption{
				sdktrace.WithScheduleDelayMillis(schDelay),
				sdktrace.WithMaxQueueSize(200),
			},
			wantNumSpans:   200,
			wantBatchCount: 1,
			genNumSpans:    205,
			waitTime:       waitTime,
		},
		{
			name: "non-default MaxQueueSize, ScheduledDelayMillis and MaxExportBatchSize",
			o: []sdktrace.BatchSpanProcessorOption{
				sdktrace.WithScheduleDelayMillis(schDelay),
				sdktrace.WithMaxQueueSize(205),
				sdktrace.WithMaxExportBatchSize(20),
			},
			wantNumSpans:   205,
			wantBatchCount: 11,
			genNumSpans:    210,
			waitTime:       waitTime,
		},
		{
			name: "blocking option",
			o: []sdktrace.BatchSpanProcessorOption{
				sdktrace.WithScheduleDelayMillis(schDelay),
				sdktrace.WithMaxQueueSize(200),
				sdktrace.WithMaxExportBatchSize(20),
				sdktrace.WithBlocking(),
			},
			wantNumSpans:   205,
			wantBatchCount: 11,
			genNumSpans:    205,
			waitTime:       waitTime,
		},
	}
	for _, option := range options {
		te := testBatchExporter{}
		ssp := createAndRegisterBatchSP(t, option, &te)
		if ssp == nil {
			t.Errorf("%s: Error creating new instance of BatchSpanProcessor\n", option.name)
		}
		sdktrace.RegisterSpanProcessor(ssp)

		generateSpan(t, option)

		time.Sleep(option.waitTime)

		gotNumOfSpans := te.len()
		if option.wantNumSpans != gotNumOfSpans {
			t.Errorf("%s: number of exported span: got %+v, want %+v\n", option.name, gotNumOfSpans, option.wantNumSpans)
		}

		gotBatchCount := te.getBatchCount()
		if gotBatchCount < option.wantBatchCount {
			t.Errorf("%s: number batches: got %+v, want >= %+v\n", option.name, gotBatchCount, option.wantBatchCount)
			t.Errorf("Batches %v\n", te.sizes)
		}

		// Check first Span is reported. Most recent one is dropped.
		sc := getSpanContext()
		wantTraceID := sc.TraceID
		wantTraceID.High = 1
		gotTraceID := te.get(0).SpanContext.TraceID
		if wantTraceID != gotTraceID {
			t.Errorf("%s: first exported span: got %+v, want %+v\n", option.name, gotTraceID, wantTraceID)
		}
		sdktrace.UnregisterSpanProcessor(ssp)
	}
}

func createAndRegisterBatchSP(t *testing.T, option testOption, te *testBatchExporter) *sdktrace.BatchSpanProcessor {
	ssp, err := sdktrace.NewBatchSpanProcessor(te, option.o...)
	if ssp == nil {
		t.Errorf("%s: Error creating new instance of BatchSpanProcessor, error: %v\n", option.name, err)
	}
	sdktrace.RegisterSpanProcessor(ssp)
	return ssp
}

func generateSpan(t *testing.T, option testOption) {
	sc := getSpanContext()

	for i := 0; i < option.genNumSpans; i++ {
		sc.TraceID.High = uint64(i + 1)
		_, span := apitrace.GlobalTracer().Start(context.Background(), option.name, apitrace.ChildOf(sc))
		span.End()
	}
}

func getSpanContext() core.SpanContext {
	tid := core.TraceID{High: 0x0102030405060708, Low: 0x0102040810203040}
	sid := uint64(0x0102040810203040)
	return core.SpanContext{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: 0x1,
	}
}