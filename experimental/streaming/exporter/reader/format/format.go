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

package format

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/api/core"
	"go.opentelemetry.io/api/key"
	"go.opentelemetry.io/experimental/streaming/exporter"
	"go.opentelemetry.io/experimental/streaming/exporter/reader"

	// TODO this should not be an SDK dependency; move conventional tags into the API.
	"go.opentelemetry.io/experimental/streaming/sdk"
)

var (
	parentSpanIDKey = key.New("parent_span_id")
)

func AppendEvent(buf *strings.Builder, data reader.Event) {

	f := func(skipIf bool) func(kv core.KeyValue) bool {
		return func(kv core.KeyValue) bool {
			if skipIf && data.Attributes.HasValue(kv.Key) {
				return true
			}
			buf.WriteString(" ")
			buf.WriteString(kv.Key.Name)
			buf.WriteString("=")
			buf.WriteString(kv.Value.Emit())
			return true
		}
	}

	buf.WriteString(data.Time.Format("2006/01/02 15-04-05.000000"))
	buf.WriteString(" ")

	switch data.Type {
	case exporter.START_SPAN:
		buf.WriteString("start ")
		buf.WriteString(data.Name)

		if !data.Parent.HasSpanID() {
			buf.WriteString(", a root span")
		} else {
			buf.WriteString(" <")
			f(false)(parentSpanIDKey.String(data.Parent.SpanIDString()))
			if data.ParentAttributes.Len() > 0 {
				data.ParentAttributes.Foreach(f(false))
			}
			buf.WriteString(" >")
		}

	case exporter.END_SPAN:
		buf.WriteString("end ")
		buf.WriteString(data.Name)

		buf.WriteString(" (")
		buf.WriteString(data.Duration.String())
		buf.WriteString(")")

	case exporter.ADD_EVENT:
		buf.WriteString("event: ")
		buf.WriteString(data.Message)
		buf.WriteString(" (")
		data.Attributes.Foreach(func(kv core.KeyValue) bool {
			buf.WriteString(" ")
			buf.WriteString(kv.Key.Name)
			buf.WriteString("=")
			buf.WriteString(kv.Value.Emit())
			return true
		})
		buf.WriteString(")")

	case exporter.MODIFY_ATTR:
		buf.WriteString("modify attr ")
		buf.WriteString(data.Type.String())
	case exporter.RECORD_STATS:
		buf.WriteString("record")

		for _, s := range data.Stats {
			f(false)(core.Key{
				Name: s.Measure.N(),
			}.Float64(s.Value))

			buf.WriteString(" {")
			i := 0
			s.Tags.Foreach(func(kv core.KeyValue) bool {
				if i != 0 {
					buf.WriteString(",")
				}
				i++
				buf.WriteString(kv.Key.Name)
				buf.WriteString("=")
				buf.WriteString(kv.Value.Emit())
				return true
			})
			buf.WriteString("}")
		}

	case exporter.SET_STATUS:
		buf.WriteString("set status ")
		buf.WriteString(data.Status.String())

	case exporter.SET_NAME:
		buf.WriteString("set name ")
		buf.WriteString(data.Name)

	default:
		buf.WriteString(fmt.Sprintf("WAT? %d", data.Type))
	}

	// Attach the scope (span) attributes and context tags.
	buf.WriteString(" [")
	if data.Attributes.Len() > 0 {
		data.Attributes.Foreach(f(false))
	}
	if data.Tags.Len() > 0 {
		data.Tags.Foreach(f(true))
	}
	if data.SpanContext.HasSpanID() {
		f(false)(sdk.SpanIDKey.String(data.SpanContext.SpanIDString()))
	}
	if data.SpanContext.HasTraceID() {
		f(false)(sdk.TraceIDKey.String(data.SpanContext.TraceIDString()))
	}

	buf.WriteString(" ]\n")
}

func EventToString(data reader.Event) string {
	var buf strings.Builder
	AppendEvent(&buf, data)
	return buf.String()
}
