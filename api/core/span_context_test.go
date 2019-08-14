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

package core_test

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-go/api/core"
)

func TestIsValid(t *testing.T) {
	for _, testcase := range []struct {
		name string
		tid  core.TraceID
		sid  uint64
		want bool
	}{
		{
			name: "bothTrue",
			tid:  core.TraceID{High: uint64(42)},
			sid:  uint64(42),
			want: true,
		}, {
			name: "bothFalse",
			tid:  core.TraceID{High: uint64(0)},
			sid:  uint64(0),
			want: false,
		}, {
			name: "oneTrue",
			tid:  core.TraceID{High: uint64(0)},
			sid:  uint64(42),
			want: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			sc := core.SpanContext{
				TraceID: testcase.tid,
				SpanID:  testcase.sid,
			}
			have := sc.IsValid()
			if have != testcase.want {
				t.Errorf("Want: %v, but have: %v", testcase.want, have)
			}
		})
	}
}

func TestHasTraceID(t *testing.T) {
	for _, testcase := range []struct {
		name string
		tid  core.TraceID
		want bool
	}{
		{
			name: "both",
			tid:  core.TraceID{High: uint64(42), Low: uint64(42)},
			want: true,
		}, {
			name: "neither",
			tid:  core.TraceID{},
			want: false,
		}, {
			name: "high",
			tid:  core.TraceID{High: uint64(42)},
			want: true,
		}, {
			name: "low",
			tid:  core.TraceID{Low: uint64(42)},
			want: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			//proto: func (sc SpanContext) HasTraceID() bool{}
			sc := core.SpanContext{TraceID: testcase.tid}
			have := sc.HasTraceID()
			if have != testcase.want {
				t.Errorf("Want: %v, but have: %v", testcase.want, have)
			}
		})
	}
}

func TestHasSpanID(t *testing.T) {
	for _, testcase := range []struct {
		name string
		sc   core.SpanContext
		want bool
	}{
		{
			name: "has",
			sc:   core.SpanContext{SpanID: uint64(42)},
			want: true,
		}, {
			name: "hasnt",
			sc:   core.SpanContext{},
			want: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			//proto: func (sc SpanContext) HasSpanID() bool {}
			have := testcase.sc.HasSpanID()
			if have != testcase.want {
				t.Errorf("Want: %v, but have: %v", testcase.want, have)
			}
		})
	}
}

func TestSpanIDString(t *testing.T) {
	for _, testcase := range []struct {
		name string
		sc   core.SpanContext
		want string
	}{
		{
			name: "fourtytwo",
			sc:   core.SpanContext{SpanID: uint64(42)},
			want: `000000000000002a`,
		}, {
			name: "empty",
			sc:   core.SpanContext{},
			want: `0000000000000000`,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			//proto: func (sc SpanContext) SpanIDString() string {}
			have := testcase.sc.SpanIDString()
			if have != testcase.want {
				t.Errorf("Want: %s, but have: %s", testcase.want, have)
			}
		})
	}
}

func TestTraceIDString(t *testing.T) {
	for _, testcase := range []struct {
		name string
		sc   core.SpanContext
		want string
	}{
		{
			name: "fourtytwo",
			sc: core.SpanContext{
				TraceID: core.TraceID{
					High: uint64(42),
					Low:  uint64(42),
				},
			},
			want: `000000000000002a000000000000002a`,
		}, {
			name: "empty",
			sc:   core.SpanContext{TraceID: core.TraceID{}},
			want: `00000000000000000000000000000000`,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			//proto: func (sc SpanContext) TraceIDString() string {}
			have := testcase.sc.TraceIDString()
			if have != testcase.want {
				t.Errorf("Want: %s, but have: %s", testcase.want, have)
			}
		})
	}
}

func TestSpanContextIsSampled(t *testing.T) {
	for _, testcase := range []struct {
		name string
		sc   core.SpanContext
		want bool
	}{
		{
			name: "sampled",
			sc: core.SpanContext{
				TraceID: core.TraceID{
					High: uint64(42),
					Low:  uint64(42),
				},
				TraceOptions: core.TraceOptionSampled,
			},
			want: true,
		}, {
			name: "sampled plus unused",
			sc: core.SpanContext{
				TraceID: core.TraceID{
					High: uint64(42),
					Low:  uint64(42),
				},
				TraceOptions: core.TraceOptionSampled | traceOptionBitMaskUnused,
			},
			want: true,
		}, {
			name: "not sampled/default",
			sc:   core.SpanContext{TraceID: core.TraceID{}},
			want: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			//proto: func (sc SpanContext) TraceIDString() string {}
			have := testcase.sc.IsSampled()
			if have != testcase.want {
				t.Errorf("Want: %v, but have: %v", testcase.want, have)
			}
		})
	}
}
