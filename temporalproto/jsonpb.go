// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// This file comes from https://github.com/gogo/protobuf/blob/v1.3.2/jsonpb/jsonpb.go
// with slight changes to support JSONPBMaybeMarshaler and JSONPBMaybeUnmarshaler.

// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2015 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
Package jsonpb provides marshaling and unmarshaling between protocol buffers and JSON.
It follows the specification at https://developers.google.com/protocol-buffers/docs/proto3#json.

This package produces a different output than the standard "encoding/json" package,
which does not operate correctly on protocol buffers.
*/
package temporalproto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"go.temporal.io/api/internal/strs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const secondInNanos = int64(time.Second / time.Nanosecond)
const maxSecondsInDuration = 315576000000

// JSONMarshaler is a configurable object for converting between
// protocol buffer objects and a JSON representation for them.
type JSONMarshaler struct {
	// Whether to render enum values as integers, as opposed to string values.
	EnumsAsInts bool

	// Whether to render fields with zero values.
	EmitDefaults bool

	// A string to indent each level by. The presence of this field will
	// also cause a space to appear between the field separator and
	// value, and for newlines to be appear between fields and array
	// elements.
	Indent string

	// Whether to use the original (.proto) name for fields.
	OrigName bool
}

// AnyResolver takes a type URL, present in an Any message, and resolves it into
// an instance of the associated message.
type AnyResolver interface {
	Resolve(typeUrl string) (proto.Message, error)
}

func defaultResolveAny(typeUrl string) (proto.Message, error) {
	// Only the part of typeUrl after the last slash is relevant.
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(mname))
	if err != nil {
		return nil, fmt.Errorf("unknown message type %q: %w", mname, err)
	}
	return messageType.New().Interface(), nil
}

// JSONPBMarshaler is implemented by protobuf messages that customize the
// way they are marshaled to JSON. Messages that implement this should
// also implement JSONPBUnmarshaler so that the custom format can be
// parsed.
//
// The JSON marshaling must follow the proto to JSON specification:
//
//	https://developers.google.com/protocol-buffers/docs/proto3#json
type JSONPBMarshaler interface {
	MarshalJSONPB(*JSONMarshaler) ([]byte, error)
}

// JSONPBUnmarshaler is implemented by protobuf messages that customize
// the way they are unmarshaled from JSON. Messages that implement this
// should also implement JSONPBMarshaler so that the custom format can be
// produced.
//
// The JSON unmarshaling must follow the JSON to proto specification:
//
//	https://developers.google.com/protocol-buffers/docs/proto3#json
type JSONPBUnmarshaler interface {
	UnmarshalJSONPB(*JSONUnmarshaler, []byte) error
}

// Marshal marshals a protocol buffer into JSON.
func (m *JSONMarshaler) Marshal(out io.Writer, pb proto.Message) error {
	v := reflect.ValueOf(pb)
	if pb == nil || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return errors.New("Marshal called with nil")
	}
	// Check for unset required fields first.
	if err := checkRequiredFields(pb); err != nil {
		return err
	}
	writer := &errWriter{writer: out}
	return m.marshalObject(writer, pb, "", "")
}

// MarshalToString converts a protocol buffer object to JSON string.
func (m *JSONMarshaler) MarshalToString(pb proto.Message) (string, error) {
	var buf bytes.Buffer
	if err := m.Marshal(&buf, pb); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type extensionTypeSlice []protoreflect.ExtensionType

var nonFinite = map[string]float64{
	`"NaN"`:       math.NaN(),
	`"Infinity"`:  math.Inf(1),
	`"-Infinity"`: math.Inf(-1),
}

// For sorting extensions ids to ensure stable output.
func (s extensionTypeSlice) Len() int { return len(s) }
func (s extensionTypeSlice) Less(i, j int) bool {
	return s[i].TypeDescriptor().Descriptor().Number() < s[j].TypeDescriptor().Descriptor().Number()
}
func (s extensionTypeSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type isWkt interface {
	XXX_WellKnownType() string
}

var (
	wktType     = reflect.TypeOf((*isWkt)(nil)).Elem()
	messageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

// marshalObject writes a struct to the Writer.
func (m *JSONMarshaler) marshalObject(out *errWriter, v proto.Message, indent, typeURL string) error {
	if jsm, ok := v.(JSONPBMarshaler); ok {
		b, err := jsm.MarshalJSONPB(m)
		if err != nil {
			return err
		}
		if typeURL != "" {
			// we are marshaling this object to an Any type
			var js map[string]*json.RawMessage
			if err = json.Unmarshal(b, &js); err != nil {
				return fmt.Errorf("type %T produced invalid JSON: %v", v, err)
			}
			turl, err := json.Marshal(typeURL)
			if err != nil {
				return fmt.Errorf("failed to marshal type URL %q to JSON: %v", typeURL, err)
			}
			js["@type"] = (*json.RawMessage)(&turl)
			if m.Indent != "" {
				b, err = json.MarshalIndent(js, indent, m.Indent)
			} else {
				b, err = json.Marshal(js)
			}
			if err != nil {
				return err
			}
		}

		out.write(string(b))
		return out.err
	}

	if jsm, ok := v.(JSONPBMaybeMarshaler); ok {
		if handled, b, err := jsm.MaybeMarshalJSONPB(m, indent); handled && err != nil {
			return err
		} else if handled {
			out.write(string(b))
			return out.err
		}
	}

	s := reflect.ValueOf(v).Elem()

	// Handle well-known types.
	// NOTE: currently disabled.
	if wkt, ok := v.(isWkt); ok {
		switch wkt.XXX_WellKnownType() {
		case "DoubleValue", "FloatValue", "Int64Value", "UInt64Value",
			"Int32Value", "UInt32Value", "BoolValue", "StringValue", "BytesValue":
			// "Wrappers use the same representation in JSON
			//  as the wrapped primitive type, ..."
			sprop := v.ProtoReflect().Descriptor()
			return m.marshalValue(out, sprop.Fields().Get(0), s.Field(0), indent)
		case "Any":
			// Any is a bit more involved.
			return m.marshalAny(out, v, indent)
		case "Duration":
			s, ns := s.Field(0).Int(), s.Field(1).Int()
			if s < -maxSecondsInDuration || s > maxSecondsInDuration {
				return fmt.Errorf("seconds out of range %v", s)
			}
			if ns <= -secondInNanos || ns >= secondInNanos {
				return fmt.Errorf("ns out of range (%v, %v)", -secondInNanos, secondInNanos)
			}
			if (s > 0 && ns < 0) || (s < 0 && ns > 0) {
				return errors.New("signs of seconds and nanos do not match")
			}
			// Generated output always contains 0, 3, 6, or 9 fractional digits,
			// depending on required precision, followed by the suffix "s".
			f := "%d.%09d"
			if ns < 0 {
				ns = -ns
				if s == 0 {
					f = "-%d.%09d"
				}
			}
			x := fmt.Sprintf(f, s, ns)
			x = strings.TrimSuffix(x, "000")
			x = strings.TrimSuffix(x, "000")
			x = strings.TrimSuffix(x, ".000")
			out.write(`"`)
			out.write(x)
			out.write(`s"`)
			return out.err
		case "Struct", "ListValue":
			// Let marshalValue handle the `Struct.fields` map or the `ListValue.values` slice.
			// TODO: pass the correct Properties if needed.
			return m.marshalValue(out, nil, s.Field(0), indent)
		case "Timestamp":
			// "RFC 3339, where generated output will always be Z-normalized
			//  and uses 0, 3, 6 or 9 fractional digits."
			s, ns := s.Field(0).Int(), s.Field(1).Int()
			if ns < 0 || ns >= secondInNanos {
				return fmt.Errorf("ns out of range [0, %v)", secondInNanos)
			}
			t := time.Unix(s, ns).UTC()
			// time.RFC3339Nano isn't exactly right (we need to get 3/6/9 fractional digits).
			x := t.Format("2006-01-02T15:04:05.000000000")
			x = strings.TrimSuffix(x, "000")
			x = strings.TrimSuffix(x, "000")
			x = strings.TrimSuffix(x, ".000")
			out.write(`"`)
			out.write(x)
			out.write(`Z"`)
			return out.err
		case "Value":
			// Value has a single oneof.
			kind := s.Field(0)
			if kind.IsNil() {
				// "absence of any variant indicates an error"
				return errors.New("nil Value")
			}
			// oneof -> *T -> T -> T.F
			x := kind.Elem().Elem().Field(0)
			// TODO: pass the correct Properties if needed.
			return m.marshalValue(out, nil, x, indent)
		}
	}

	out.write("{")
	if m.Indent != "" {
		out.write("\n")
	}

	firstField := true

	if typeURL != "" {
		if err := m.marshalTypeURL(out, indent, typeURL); err != nil {
			return err
		}
		firstField = false
	}

	sprop := v.ProtoReflect().Descriptor()
	fields := sprop.Fields()
	for i := 0; i < s.NumField(); i++ {
		value := s.Field(i)
		valueField := s.Type().Field(i)
		if strings.HasPrefix(valueField.Name, "XXX_") {
			continue
		}

		//this is not a protobuf field
		if valueField.Tag.Get("protobuf") == "" && valueField.Tag.Get("protobuf_oneof") == "" {
			continue
		}

		// IsNil will panic on most value kinds.
		switch value.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface:
			if value.IsNil() {
				continue
			}
		}
		if !m.EmitDefaults {
			switch value.Kind() {
			case reflect.Bool:
				if !value.Bool() {
					continue
				}
			case reflect.Int32, reflect.Int64:
				if value.Int() == 0 {
					continue
				}
			case reflect.Uint32, reflect.Uint64:
				if value.Uint() == 0 {
					continue
				}
			case reflect.Float32, reflect.Float64:
				if value.Float() == 0 {
					continue
				}
			case reflect.String:
				if value.Len() == 0 {
					continue
				}
			case reflect.Map, reflect.Ptr, reflect.Slice:
				if value.IsNil() {
					continue
				}
			}
		}

		// Oneof fields need special handling.
		if valueField.Tag.Get("protobuf_oneof") != "" {
			// value is an interface containing &T{real_value}.
			sv := value.Elem().Elem() // interface -> *T -> T
			value = sv.Field(0)
			valueField = sv.Type().Field(0)
		}
		if !firstField {
			m.writeSep(out)
		}
		if err := m.marshalField(out, fields.Get(i), value, indent); err != nil {
			return err
		}
		firstField = false
	}

	// Handle proto2 extensions.
	if ep, ok := v.(proto.Message); ok {
		extensions := sprop.Extensions()
		// Sort extensions for stable output.
		ets := make([]protoreflect.ExtensionType, 0, extensions.Len())
		for i := 0; i < extensions.Len(); i++ {
			desc := extensions.Get(i)
			et, err := protoregistry.GlobalTypes.FindExtensionByName(desc.FullName())
			if err != nil {
				return fmt.Errorf("unknown extension %s", desc.FullName())
			}
			if !proto.HasExtension(ep, et) {
				continue
			}
			ets = append(ets, et)
		}
		sort.Sort(extensionTypeSlice(ets))
		for _, ext := range ets {
			value := reflect.ValueOf(ext.New())
			if !firstField {
				m.writeSep(out)
			}
			if err := m.marshalField(out, ext.TypeDescriptor(), value, indent); err != nil {
				return err
			}
			firstField = false
		}

	}

	if m.Indent != "" {
		out.write("\n")
		out.write(indent)
	}
	out.write("}")
	return out.err
}

func (m *JSONMarshaler) writeSep(out *errWriter) {
	if m.Indent != "" {
		out.write(",\n")
	} else {
		out.write(",")
	}
}

func (m *JSONMarshaler) marshalAny(out *errWriter, any proto.Message, indent string) error {
	// "If the Any contains a value that has a special JSON mapping,
	//  it will be converted as follows: {"@type": xxx, "value": yyy}.
	//  Otherwise, the value will be converted into a JSON object,
	//  and the "@type" field will be inserted to indicate the actual data type."
	v := reflect.ValueOf(any).Elem()
	turl := v.Field(0).String()
	val := v.Field(1).Bytes()

	msg, err := defaultResolveAny(turl)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(val, msg); err != nil {
		return err
	}

	if _, ok := msg.(isWkt); ok {
		out.write("{")
		if m.Indent != "" {
			out.write("\n")
		}
		if err := m.marshalTypeURL(out, indent, turl); err != nil {
			return err
		}
		m.writeSep(out)
		if m.Indent != "" {
			out.write(indent)
			out.write(m.Indent)
			out.write(`"value": `)
		} else {
			out.write(`"value":`)
		}
		if err := m.marshalObject(out, msg, indent+m.Indent, ""); err != nil {
			return err
		}
		if m.Indent != "" {
			out.write("\n")
			out.write(indent)
		}
		out.write("}")
		return out.err
	}

	return m.marshalObject(out, msg, indent, turl)
}

func (m *JSONMarshaler) marshalTypeURL(out *errWriter, indent, typeURL string) error {
	if m.Indent != "" {
		out.write(indent)
		out.write(m.Indent)
	}
	out.write(`"@type":`)
	if m.Indent != "" {
		out.write(" ")
	}
	b, err := json.Marshal(typeURL)
	if err != nil {
		return err
	}
	out.write(string(b))
	return out.err
}

// marshalField writes field description and value to the Writer.
func (m *JSONMarshaler) marshalField(out *errWriter, prop protoreflect.FieldDescriptor, v reflect.Value, indent string) error {
	if m.Indent != "" {
		out.write(indent)
		out.write(m.Indent)
	}
	out.write(`"`)
	if m.OrigName || prop.JSONName() == "" {
		out.write(prop.TextName())
	} else {
		out.write(prop.JSONName())
	}
	out.write(`":`)
	if m.Indent != "" {
		out.write(" ")
	}
	if err := m.marshalValue(out, prop, v, indent); err != nil {
		return err
	}
	return nil
}

// marshalValue writes the value to the Writer.
func (m *JSONMarshaler) marshalValue(out *errWriter, prop protoreflect.FieldDescriptor, v reflect.Value, indent string) error {

	v = reflect.Indirect(v)

	// Handle nil pointer
	if v.Kind() == reflect.Invalid {
		out.write("null")
		return out.err
	}

	// Handle repeated elements.
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 {
		out.write("[")
		comma := ""
		for i := 0; i < v.Len(); i++ {
			sliceVal := v.Index(i)
			out.write(comma)
			if m.Indent != "" {
				out.write("\n")
				out.write(indent)
				out.write(m.Indent)
				out.write(m.Indent)
			}
			if err := m.marshalValue(out, prop, sliceVal, indent+m.Indent); err != nil {
				return err
			}
			comma = ","
		}
		if m.Indent != "" {
			out.write("\n")
			out.write(indent)
			out.write(m.Indent)
		}
		out.write("]")
		return out.err
	}

	// Handle well-known types.
	// Most are handled up in marshalObject (because 99% are messages).
	if v.Type().Implements(wktType) {
		wkt := v.Interface().(isWkt)
		switch wkt.XXX_WellKnownType() {
		case "NullValue":
			out.write("null")
			return out.err
		}
	}

	// Handle enumerations.
	if !m.EnumsAsInts && prop.Kind() == protoreflect.EnumKind {
		// Unknown enum values will are stringified by the proto library as their
		// value. Such values should _not_ be quoted or they will be interpreted
		// as an enum string instead of their value.
		enumStr := v.Interface().(fmt.Stringer).String()
		var valStr string
		if v.Kind() == reflect.Ptr {
			valStr = strconv.Itoa(int(v.Elem().Int()))
		} else {
			valStr = strconv.Itoa(int(v.Int()))
		}

		if m, ok := v.Interface().(interface {
			MarshalJSON() ([]byte, error)
		}); ok {
			data, err := m.MarshalJSON()
			if err != nil {
				return err
			}
			enumStr = string(data)
			enumStr, err = strconv.Unquote(enumStr)
			if err != nil {
				return err
			}
		}

		isKnownEnum := enumStr != valStr

		if isKnownEnum {
			out.write(`"`)
		}
		out.write(enumStr)
		if isKnownEnum {
			out.write(`"`)
		}
		return out.err
	}

	// Handle nested messages.
	if v.Kind() == reflect.Struct {
		i := v
		if v.CanAddr() {
			i = v.Addr()
		} else {
			i = reflect.New(v.Type())
			i.Elem().Set(v)
		}
		iface := i.Interface()
		if iface == nil {
			out.write(`null`)
			return out.err
		}

		if m, ok := v.Interface().(interface {
			MarshalJSON() ([]byte, error)
		}); ok {
			data, err := m.MarshalJSON()
			if err != nil {
				return err
			}
			out.write(string(data))
			return nil
		}

		pm, ok := iface.(proto.Message)
		if !ok {
			return fmt.Errorf("%v does not implement proto.Message", v.Type())
		}
		return m.marshalObject(out, pm, indent+m.Indent, "")
	}

	// Handle maps.
	// Since Go randomizes map iteration, we sort keys for stable output.
	if v.Kind() == reflect.Map {
		out.write(`{`)
		keys := v.MapKeys()
		sort.Sort(mapKeys(keys))
		for i, k := range keys {
			if i > 0 {
				out.write(`,`)
			}
			if m.Indent != "" {
				out.write("\n")
				out.write(indent)
				out.write(m.Indent)
				out.write(m.Indent)
			}

			// TODO handle map key prop properly
			b, err := json.Marshal(k.Interface())
			if err != nil {
				return err
			}
			s := string(b)

			// If the JSON is not a string value, encode it again to make it one.
			if !strings.HasPrefix(s, `"`) {
				b, err := json.Marshal(s)
				if err != nil {
					return err
				}
				s = string(b)
			}

			out.write(s)
			out.write(`:`)
			if m.Indent != "" {
				out.write(` `)
			}

			vprop := prop
			// TODO?
			// if prop != nil && prop.MapValProp != nil {
			// 	vprop = prop.MapValProp
			// }
			if err := m.marshalValue(out, vprop, v.MapIndex(k), indent+m.Indent); err != nil {
				return err
			}
		}
		if m.Indent != "" {
			out.write("\n")
			out.write(indent)
			out.write(m.Indent)
		}
		out.write(`}`)
		return out.err
	}

	// Handle non-finite floats, e.g. NaN, Infinity and -Infinity.
	if v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64 {
		f := v.Float()
		var sval string
		switch {
		case math.IsInf(f, 1):
			sval = `"Infinity"`
		case math.IsInf(f, -1):
			sval = `"-Infinity"`
		case math.IsNaN(f):
			sval = `"NaN"`
		}
		if sval != "" {
			out.write(sval)
			return out.err
		}
	}

	// Default handling defers to the encoding/json library.
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return err
	}
	needToQuote := string(b[0]) != `"` && (v.Kind() == reflect.Int64 || v.Kind() == reflect.Uint64)
	if needToQuote {
		out.write(`"`)
	}
	out.write(string(b))
	if needToQuote {
		out.write(`"`)
	}
	return out.err
}

// JSONUnmarshaler is a configurable object for converting from a JSON
// representation to a protocol buffer object.
type JSONUnmarshaler struct {
	// Whether to allow messages to contain unknown fields, as opposed to
	// failing to unmarshal.
	AllowUnknownFields bool
}

// UnmarshalNext unmarshals the next protocol buffer from a JSON object stream.
// This function is lenient and will decode any options permutations of the
// related JSONMarshaler.
func (u *JSONUnmarshaler) UnmarshalNext(dec *json.Decoder, pb proto.Message) error {
	inputValue := json.RawMessage{}
	if err := dec.Decode(&inputValue); err != nil {
		return err
	}
	if err := u.unmarshalValue(reflect.ValueOf(pb).Elem(), inputValue, nil); err != nil {
		return err
	}
	return checkRequiredFields(pb)
}

// Unmarshal unmarshals a JSON object stream into a protocol
// buffer. This function is lenient and will decode any options
// permutations of the related JSONMarshaler.
func (u *JSONUnmarshaler) Unmarshal(r io.Reader, pb proto.Message) error {
	dec := json.NewDecoder(r)
	return u.UnmarshalNext(dec, pb)
}

// UnmarshalNext unmarshals the next protocol buffer from a JSON object stream.
// This function is lenient and will decode any options permutations of the
// related JSONMarshaler.
func UnmarshalNext(dec *json.Decoder, pb proto.Message) error {
	return new(JSONUnmarshaler).UnmarshalNext(dec, pb)
}

// Unmarshal unmarshals a JSON object stream into a protocol
// buffer. This function is lenient and will decode any options
// permutations of the related JSONMarshaler.
func Unmarshal(r io.Reader, pb proto.Message) error {
	return new(JSONUnmarshaler).Unmarshal(r, pb)
}

// UnmarshalString will populate the fields of a protocol buffer based
// on a JSON string. This function is lenient and will decode any options
// permutations of the related JSONMarshaler.
func UnmarshalString(str string, pb proto.Message) error {
	return new(JSONUnmarshaler).Unmarshal(strings.NewReader(str), pb)
}

func isNullValue(fd protoreflect.FieldDescriptor) bool {
	ed := fd.Enum()
	return ed != nil && ed.FullName() == "google.protobuf.NullValue"
}

// unmarshalValue converts/copies a value into the target.
// prop may be nil.
func (u *JSONUnmarshaler) unmarshalValue(target reflect.Value, inputValue json.RawMessage, prop protoreflect.FieldDescriptor) error {
	targetType := target.Type()

	// Allocate memory for pointer fields.
	if targetType.Kind() == reflect.Ptr {
		// If input value is "null" and target is a pointer type, then the field should be treated as not set
		// UNLESS the target is structpb.Value, in which case it should be set to structpb.NullValue.
		_, isJSONPBUnmarshaler := target.Interface().(JSONPBUnmarshaler)
		_, isMaybeJSONPBUnmarshaler := target.Interface().(JSONPBMaybeUnmarshaler)
		if string(inputValue) == "null" && targetType != reflect.TypeOf(&structpb.Value{}) && !isJSONPBUnmarshaler && !isMaybeJSONPBUnmarshaler {
			return nil
		}
		target.Set(reflect.New(targetType.Elem()))

		return u.unmarshalValue(target.Elem(), inputValue, prop)
	}

	if jsu, ok := target.Addr().Interface().(JSONPBUnmarshaler); ok {
		return jsu.UnmarshalJSONPB(u, []byte(inputValue))
	}

	if jsu, ok := target.Addr().Interface().(JSONPBMaybeUnmarshaler); ok {
		if handled, err := jsu.MaybeUnmarshalJSONPB(u, []byte(inputValue)); handled {
			return err
		}
	}

	// Handle well-known types that are not pointers.
	// FIXME update this using genid
	if w, ok := target.Addr().Interface().(isWkt); ok {
		switch w.XXX_WellKnownType() {
		case "DoubleValue", "FloatValue", "Int64Value", "UInt64Value",
			"Int32Value", "UInt32Value", "BoolValue", "StringValue", "BytesValue":
			return u.unmarshalValue(target.Field(0), inputValue, prop)
		case "Any":
			// Use json.RawMessage pointer type instead of value to support pre-1.8 version.
			// 1.8 changed RawMessage.MarshalJSON from pointer type to value type, see
			// https://github.com/golang/go/issues/14493
			var jsonFields map[string]*json.RawMessage
			if err := json.Unmarshal(inputValue, &jsonFields); err != nil {
				return err
			}

			val, ok := jsonFields["@type"]
			if !ok || val == nil {
				return errors.New("Any JSON doesn't have '@type'")
			}

			var turl string
			if err := json.Unmarshal([]byte(*val), &turl); err != nil {
				return fmt.Errorf("can't unmarshal Any's '@type': %q", *val)
			}
			target.Field(0).SetString(turl)

			m, err := defaultResolveAny(turl)
			if err != nil {
				return err
			}

			if _, ok := m.(isWkt); ok {
				val, ok := jsonFields["value"]
				if !ok {
					return errors.New("Any JSON doesn't have 'value'")
				}

				if err = u.unmarshalValue(reflect.ValueOf(m).Elem(), *val, nil); err != nil {
					return fmt.Errorf("can't unmarshal Any nested proto %T: %v", m, err)
				}
			} else {
				delete(jsonFields, "@type")
				nestedProto, uerr := json.Marshal(jsonFields)
				if uerr != nil {
					return fmt.Errorf("can't generate JSON for Any's nested proto to be unmarshaled: %v", uerr)
				}

				if err = u.unmarshalValue(reflect.ValueOf(m).Elem(), nestedProto, nil); err != nil {
					return fmt.Errorf("can't unmarshal Any nested proto %T: %v", m, err)
				}
			}

			b, err := proto.Marshal(m)
			if err != nil {
				return fmt.Errorf("can't marshal proto %T into Any.Value: %v", m, err)
			}
			target.Field(1).SetBytes(b)

			return nil
		case "Duration":
			unq, err := unquote(string(inputValue))
			if err != nil {
				return err
			}

			d, err := time.ParseDuration(unq)
			if err != nil {
				return fmt.Errorf("bad Duration: %v", err)
			}

			ns := d.Nanoseconds()
			s := ns / 1e9
			ns %= 1e9
			target.Field(0).SetInt(s)
			target.Field(1).SetInt(ns)
			return nil
		case "Timestamp":
			unq, err := unquote(string(inputValue))
			if err != nil {
				return err
			}

			t, err := time.Parse(time.RFC3339Nano, unq)
			if err != nil {
				return fmt.Errorf("bad Timestamp: %v", err)
			}

			target.Field(0).SetInt(t.Unix())
			target.Field(1).SetInt(int64(t.Nanosecond()))
			return nil
		case "Struct":
			var m map[string]json.RawMessage
			if err := json.Unmarshal(inputValue, &m); err != nil {
				return fmt.Errorf("bad StructValue: %v", err)
			}
			target.Field(0).Set(reflect.ValueOf(map[string]*structpb.Value{}))
			for k, jv := range m {
				pv := &structpb.Value{}
				if err := u.unmarshalValue(reflect.ValueOf(pv).Elem(), jv, prop); err != nil {
					return fmt.Errorf("bad value in StructValue for key %q: %v", k, err)
				}
				target.Field(0).SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(pv))
			}
			return nil
		case "ListValue":
			var s []json.RawMessage
			if err := json.Unmarshal(inputValue, &s); err != nil {
				return fmt.Errorf("bad ListValue: %v", err)
			}

			target.Field(0).Set(reflect.ValueOf(make([]*structpb.Value, len(s))))
			for i, sv := range s {
				if err := u.unmarshalValue(target.Field(0).Index(i), sv, prop); err != nil {
					return err
				}
			}
			return nil
		case "Value":
			ivStr := string(inputValue)
			if ivStr == "null" {
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_NullValue{}))
			} else if v, err := strconv.ParseFloat(ivStr, 0); err == nil {
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_NumberValue{NumberValue: v}))
			} else if v, err := unquote(ivStr); err == nil {
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_StringValue{StringValue: v}))
			} else if v, err := strconv.ParseBool(ivStr); err == nil {
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_BoolValue{BoolValue: v}))
			} else if err := json.Unmarshal(inputValue, &[]json.RawMessage{}); err == nil {
				lv := &structpb.ListValue{}
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_ListValue{ListValue: lv}))
				return u.unmarshalValue(reflect.ValueOf(lv).Elem(), inputValue, prop)
			} else if err := json.Unmarshal(inputValue, &map[string]json.RawMessage{}); err == nil {
				sv := &structpb.Struct{}
				target.Field(0).Set(reflect.ValueOf(&structpb.Value_StructValue{StructValue: sv}))
				return u.unmarshalValue(reflect.ValueOf(sv).Elem(), inputValue, prop)
			} else {
				return fmt.Errorf("unrecognized type for Value %q", ivStr)
			}
			return nil
		}
	}

	// I'm not sure either of these are necessary
	if t, ok := target.Addr().Interface().(*time.Time); ok {
		ts := &timestamppb.Timestamp{}
		if err := u.unmarshalValue(reflect.ValueOf(ts).Elem(), inputValue, prop); err != nil {
			return err
		}
		*t = ts.AsTime()
		return nil
	}

	if d, ok := target.Addr().Interface().(*time.Duration); ok {
		dur := &durationpb.Duration{}
		if err := u.unmarshalValue(reflect.ValueOf(dur).Elem(), inputValue, prop); err != nil {
			return err
		}
		*d = dur.AsDuration()
		return nil
	}

	// Handle enums, which have an underlying type of int32,
	// and may appear as strings.
	// The case of an enum appearing as a number is handled
	// at the bottom of this function.
	if inputValue[0] == '"' && prop != nil && prop.Kind() == protoreflect.EnumKind {
		ed := prop.Enum()
		typePrefix := strcase.ToScreamingSnake(string(ed.Name()))
		// Don't need to do unquoting; valid enum names
		// are from a limited character set.
		s := string(inputValue[1 : len(inputValue)-1])
		// Translate old-school temporal CamelCase enums to their canonical json forms
		if !strings.HasPrefix(s, typePrefix) {
			s = fmt.Sprintf("%s_%s", typePrefix, strcase.ToScreamingSnake(s))
		}
		enumVal := ed.Values().ByName(protoreflect.Name(s))
		if enumVal == nil {
			return fmt.Errorf("unknown value %q for enum %s", s, ed.Name())
		}
		if target.Kind() == reflect.Ptr { // proto2
			target.Set(reflect.New(targetType.Elem()))
			target = target.Elem()
		}
		if targetType.Kind() != reflect.Int32 {
			return fmt.Errorf("invalid target %q for enum %s", targetType.Kind(), prop.Enum)
		}
		target.SetInt(int64(enumVal.Number()))
		return nil
	}

	// if prop != nil && len(prop.CustomType) > 0 && target.CanAddr() {
	// 	if m, ok := target.Addr().Interface().(interface {
	// 		UnmarshalJSON([]byte) error
	// 	}); ok {
	// 		return json.Unmarshal(inputValue, m)
	// 	}
	// }

	// Handle nested messages.
	if targetType.Kind() == reflect.Struct {
		var jsonFields map[string]json.RawMessage
		if err := json.Unmarshal(inputValue, &jsonFields); err != nil {
			return err
		}

		consumeField := func(prop protoreflect.FieldDescriptor) (json.RawMessage, bool) {
			// Be liberal in what names we accept; both orig_name and camelName are okay.
			fieldNames := acceptedJSONFieldNames(prop)

			vOrig, okOrig := jsonFields[fieldNames.orig]
			vCamel, okCamel := jsonFields[fieldNames.camel]
			if !okOrig && !okCamel {
				return nil, false
			}
			// If, for some reason, both are present in the data, favour the camelName.
			var raw json.RawMessage
			if okOrig {
				raw = vOrig
				delete(jsonFields, fieldNames.orig)
			}
			if okCamel {
				raw = vCamel
				delete(jsonFields, fieldNames.camel)
			}
			return raw, true
		}

		sprops := prop.Message()
		for i := 0; i < target.NumField(); i++ {
			ft := target.Type().Field(i)
			if strings.HasPrefix(ft.Name, "XXX_") {
				continue
			}
			fi := sprops.Fields().Get(i)
			valueForField, ok := consumeField(fi)
			if !ok {
				continue
			}

			if err := u.unmarshalValue(target.Field(i), valueForField, fi); err != nil {
				return err
			}
		}
		// Check for any oneof fields.
		if len(jsonFields) > 0 {
			oneofs := sprops.Oneofs()
			for i := 0; i < oneofs.Len(); i++ {
				oop := oneofs.Get(i)
				fields := oop.Fields()
				for j := 0; j < fields.Len(); j++ {
					fp := fields.Get(j)
					raw, ok := consumeField(fp)
					if !ok {
						continue
					}
					mt, err := protoregistry.GlobalTypes.FindMessageByName(fp.FullName())
					if err != nil {
						return fmt.Errorf("unable to find message %s: %w", fp.FullName(), err)
					}
					nv := reflect.ValueOf(mt.New())
					target.FieldByName(strs.GoCamelCase(fp.TextName())).Set(nv)
					if err := u.unmarshalValue(nv.Elem().Field(0), raw, fp); err != nil {
						return err
					}
				}
			}
		}
		// Handle proto2 extensions.
		if len(jsonFields) > 0 {
			if ep, ok := target.Addr().Interface().(proto.Message); ok {
				extensions := sprops.Extensions()
				for i := 0; i < extensions.Len(); i++ {
					desc := extensions.Get(i)
					et, err := protoregistry.GlobalTypes.FindExtensionByName(desc.FullName())
					if err != nil {
						return fmt.Errorf("unknown extension %s", desc.FullName())
					}
					if !proto.HasExtension(ep, et) {
						continue
					}
					name := fmt.Sprintf("[%s]", ext.Name)
					raw, ok := jsonFields[name]
					if !ok {
						continue
					}
					delete(jsonFields, name)
					nv := reflect.New(reflect.TypeOf(ext.ExtensionType).Elem())
					if err := u.unmarshalValue(nv.Elem(), raw, nil); err != nil {
						return err
					}
					if err := proto.SetExtension(ep, ext, nv.Interface()); err != nil {
						return err
					}
				}
			}
		}
		if !u.AllowUnknownFields && len(jsonFields) > 0 {
			// Pick any field to be the scapegoat.
			var f string
			for fname := range jsonFields {
				f = fname
				break
			}
			return fmt.Errorf("unknown field %q in %v", f, targetType)
		}
		return nil
	}

	// Handle arrays
	if targetType.Kind() == reflect.Slice {
		if targetType.Elem().Kind() == reflect.Uint8 {
			outRef := reflect.New(targetType)
			outVal := outRef.Interface()
			//CustomType with underlying type []byte
			if _, ok := outVal.(interface {
				UnmarshalJSON([]byte) error
			}); ok {
				if err := json.Unmarshal(inputValue, outVal); err != nil {
					return err
				}
				target.Set(outRef.Elem())
				return nil
			}
			// Special case for encoded bytes. Pre-go1.5 doesn't support unmarshalling
			// strings into aliased []byte types.
			// https://github.com/golang/go/commit/4302fd0409da5e4f1d71471a6770dacdc3301197
			// https://github.com/golang/go/commit/c60707b14d6be26bf4213114d13070bff00d0b0a
			var out []byte
			if err := json.Unmarshal(inputValue, &out); err != nil {
				return err
			}
			target.SetBytes(out)
			return nil
		}

		var slc []json.RawMessage
		if err := json.Unmarshal(inputValue, &slc); err != nil {
			return err
		}
		if slc != nil {
			l := len(slc)
			target.Set(reflect.MakeSlice(targetType, l, l))
			for i := 0; i < l; i++ {
				if err := u.unmarshalValue(target.Index(i), slc[i], prop); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Handle maps (whose keys are always strings)
	if targetType.Kind() == reflect.Map {
		var mp map[string]json.RawMessage
		if err := json.Unmarshal(inputValue, &mp); err != nil {
			return err
		}
		if mp != nil {
			target.Set(reflect.MakeMap(targetType))
			for ks, raw := range mp {
				// Unmarshal map key. The core json library already decoded the key into a
				// string, so we handle that specially. Other types were quoted post-serialization.
				var k reflect.Value
				if targetType.Key().Kind() == reflect.String {
					k = reflect.ValueOf(ks)
				} else {
					k = reflect.New(targetType.Key()).Elem()
					var kprop protoreflect.FieldDescriptor
					if prop != nil && prop.MapKeyProp != nil {
						kprop = prop.MapKeyProp
					}
					if err := u.unmarshalValue(k, json.RawMessage(ks), kprop); err != nil {
						return err
					}
				}

				if !k.Type().AssignableTo(targetType.Key()) {
					k = k.Convert(targetType.Key())
				}

				// Unmarshal map value.
				v := reflect.New(targetType.Elem()).Elem()
				var vprop protoreflect.FieldDescriptor
				if prop != nil && prop.MapValProp != nil {
					vprop = prop.MapValProp
				}
				if err := u.unmarshalValue(v, raw, vprop); err != nil {
					return err
				}
				target.SetMapIndex(k, v)
			}
		}
		return nil
	}

	// Non-finite numbers can be encoded as strings.
	isFloat := targetType.Kind() == reflect.Float32 || targetType.Kind() == reflect.Float64
	if isFloat {
		if num, ok := nonFinite[string(inputValue)]; ok {
			target.SetFloat(num)
			return nil
		}
	}

	// integers & floats can be encoded as strings. In this case we drop
	// the quotes and proceed as normal.
	isNum := targetType.Kind() == reflect.Int64 || targetType.Kind() == reflect.Uint64 ||
		targetType.Kind() == reflect.Int32 || targetType.Kind() == reflect.Uint32 ||
		targetType.Kind() == reflect.Float32 || targetType.Kind() == reflect.Float64
	if isNum && strings.HasPrefix(string(inputValue), `"`) {
		inputValue = inputValue[1 : len(inputValue)-1]
	}

	// Use the encoding/json for parsing other value types.
	return json.Unmarshal(inputValue, target.Addr().Interface())
}

func unquote(s string) (string, error) {
	var ret string
	err := json.Unmarshal([]byte(s), &ret)
	return ret, err
}

type fieldNames struct {
	orig, camel string
}

func acceptedJSONFieldNames(prop protoreflect.FieldDescriptor) fieldNames {
	opts := fieldNames{orig: prop.TextName(), camel: prop.TextName()}
	if prop.JSONName() != "" {
		opts.camel = prop.JSONName()
	}
	return opts
}

// Writer wrapper inspired by https://blog.golang.org/errors-are-values
type errWriter struct {
	writer io.Writer
	err    error
}

func (w *errWriter) write(str string) {
	if w.err != nil {
		return
	}
	_, w.err = w.writer.Write([]byte(str))
}

// Map fields may have key types of non-float scalars, strings and enums.
// The easiest way to sort them in some deterministic order is to use fmt.
// If this turns out to be inefficient we can always consider other options,
// such as doing a Schwartzian transform.
//
// Numeric keys are sorted in numeric order per
// https://developers.google.com/protocol-buffers/docs/proto#maps.
type mapKeys []reflect.Value

func (s mapKeys) Len() int      { return len(s) }
func (s mapKeys) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s mapKeys) Less(i, j int) bool {
	if k := s[i].Kind(); k == s[j].Kind() {
		switch k {
		case reflect.String:
			return s[i].String() < s[j].String()
		case reflect.Int32, reflect.Int64:
			return s[i].Int() < s[j].Int()
		case reflect.Uint32, reflect.Uint64:
			return s[i].Uint() < s[j].Uint()
		}
	}
	return fmt.Sprint(s[i].Interface()) < fmt.Sprint(s[j].Interface())
}

// checkRequiredFields returns an error if any required field in the given proto message is not set.
// This function is used by both Marshal and Unmarshal.  While required fields only exist in a
// proto2 message, a proto3 message can contain proto2 message(s).
func checkRequiredFields(pb proto.Message) error {
	// Most well-known type messages do not contain required fields.  The "Any" type may contain
	// a message that has required fields.
	//
	// When an Any message is being marshaled, the code will invoked proto.Unmarshal on Any.Value
	// field in order to transform that into JSON, and that should have returned an error if a
	// required field is not set in the embedded message.
	//
	// When an Any message is being unmarshaled, the code will have invoked proto.Marshal on the
	// embedded message to store the serialized message in Any.Value field, and that should have
	// returned an error if a required field is not set.
	if _, ok := pb.(isWkt); ok {
		return nil
	}

	v := reflect.ValueOf(pb)
	// Skip message if it is not a struct pointer.
	if v.Kind() != reflect.Ptr {
		return nil
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}

	sprop := pb.ProtoReflect().Descriptor()
	pbfields := sprop.Fields()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		sfield := v.Type().Field(i)
		prop := pbfields.Get(i)

		if sfield.PkgPath != "" {
			// blank PkgPath means the field is exported; skip if not exported
			continue
		}

		if strings.HasPrefix(sfield.Name, "XXX_") {
			continue
		}

		// Oneof field is an interface implemented by wrapper structs containing the actual oneof
		// field, i.e. an interface containing &T{real_value}.
		if sfield.Tag.Get("protobuf_oneof") != "" {
			if field.Kind() != reflect.Interface {
				continue
			}
			v := field.Elem()
			if v.Kind() != reflect.Ptr || v.IsNil() {
				continue
			}
			v = v.Elem()
			if v.Kind() != reflect.Struct || v.NumField() < 1 {
				continue
			}
			field = v.Field(0)
			sfield = v.Type().Field(0)
		}

		protoTag := sfield.Tag.Get("protobuf")
		if protoTag == "" {
			continue
		}

		switch field.Kind() {
		case reflect.Map:
			if field.IsNil() {
				continue
			}
			// Check each map value.
			keys := field.MapKeys()
			for _, k := range keys {
				v := field.MapIndex(k)
				if err := checkRequiredFieldsInValue(v); err != nil {
					return err
				}
			}
		case reflect.Slice:
			// Handle non-repeated type, e.g. bytes.
			if prop.Cardinality() != protoreflect.Repeated {
				if prop.Cardinality() == protoreflect.Required && field.IsNil() {
					return fmt.Errorf("required field %q is not set", prop.Name)
				}
				continue
			}

			// Handle repeated type.
			if field.IsNil() {
				continue
			}
			// Check each slice item.
			for i := 0; i < field.Len(); i++ {
				v := field.Index(i)
				if err := checkRequiredFieldsInValue(v); err != nil {
					return err
				}
			}
		case reflect.Ptr:
			if field.IsNil() {
				if prop.Cardinality() == protoreflect.Required {
					return fmt.Errorf("required field %q is not set", prop.Name)
				}
				continue
			}
			if err := checkRequiredFieldsInValue(field); err != nil {
				return err
			}
		}
	}

	// Handle proto2 extensions.
	extensions := sprop.Extensions()
	for i := 0; i < extensions.Len(); i++ {
		desc := extensions.Get(i)
		et, err := protoregistry.GlobalTypes.FindExtensionByName(desc.FullName())
		if err != nil {
			return fmt.Errorf("unknown extension %s", desc.FullName())
		}
		if !proto.HasExtension(pb, et) {
			continue
		}
		err = checkRequiredFieldsInValue(reflect.ValueOf(pb))
		if err != nil {
			return err
		}
	}

	return nil
}

func checkRequiredFieldsInValue(v reflect.Value) error {
	if v.Type().Implements(messageType) {
		return checkRequiredFields(v.Interface().(proto.Message))
	}
	return nil
}
