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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const secondInNanos = int64(time.Second / time.Nanosecond)
const maxSecondsInDuration = 315576000000

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

// JSONUnmarshaler is a configurable object for converting from a JSON
// representation to a protocol buffer object.
type JSONUnmarshaler struct {
	// Whether to allow messages to contain unknown fields, as opposed to
	// failing to unmarshal.
	DiscardUnknown bool
}

var logfunc func(string, ...any) = func(msg string, args ...any) {
	fmt.Printf(msg, args...)
}

func SetLogfunc(l func(string, ...any)) {
	logfunc = l
}

// UnmarshalNext unmarshals the next protocol buffer from a JSON object stream.
// This function is lenient and will decode any options permutations of the
// related JSONMarshaler.
func (u *JSONUnmarshaler) UnmarshalNext(dec *json.Decoder, pb proto.Message) error {
	inputValue := json.RawMessage{}
	if err := dec.Decode(&inputValue); err != nil {
		return err
	}
	logfunc("Unmarshaling %s from %s\n", pb.ProtoReflect().Descriptor().FullName(), string(inputValue))
	if err := u.unmarshalMessage(pb.ProtoReflect(), inputValue); err != nil {
		return err
	}

	return proto.CheckInitialized(pb)
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

func (u *JSONUnmarshaler) unmarshalAny(pb protoreflect.Message, jsonFields map[string]json.RawMessage) error {
	if len(jsonFields) == 0 {
		// empty map is an empty Any
		return nil
	}
	atType, ok := jsonFields["@type"]
	if !ok || len(atType) < 3 {
		if u.DiscardUnknown {
			// Treat all fields as unknowns, similar to an empty object.
			return nil
		}
		return fmt.Errorf("invalid Any; missing @type in %v", jsonFields)
	}
	typeURL := string(atType[1 : len(atType)-1])
	emt, err := protoregistry.GlobalTypes.FindMessageByURL(typeURL)
	if err != nil {
		return err
	}

	fds := pb.Descriptor().Fields()
	fdType := fds.ByNumber(1)
	fdValue := fds.ByNumber(2)

	pb.Set(fdType, protoreflect.ValueOfString(typeURL))

	value, ok := jsonFields["value"]
	if !ok {
		return nil
	}
	// This is gross but it's how the official compiler handles Any values
	em := emt.New()
	if err := u.unmarshalMessage(em, value); err != nil {
		return err
	}
	// Serialize the embedded message and assign the resulting bytes to the
	// proto value field.
	b, err := proto.MarshalOptions{
		AllowPartial:  true, // No need to check required fields inside an Any.
		Deterministic: true,
	}.Marshal(em.Interface())
	if err != nil {
		return fmt.Errorf("error in marshaling Any.value field: %v", err)
	}

	pb.Set(fdValue, protoreflect.ValueOfBytes(b))
	return nil
}

func (u *JSONUnmarshaler) unmarshalMessage(pb protoreflect.Message, inputValue json.RawMessage) error {
	if jsu, ok := pb.Interface().(JSONPBMaybeUnmarshaler); ok {
		if handled, err := jsu.MaybeUnmarshalJSONPB(u, []byte(inputValue)); handled {
			return err
		}
	}
	fd := pb.Descriptor()
	fields := fd.Fields()

	var jsonFields map[string]json.RawMessage
	if err := json.Unmarshal(inputValue, &jsonFields); err != nil {
		return err
	}

	if fd.FullName() == "google.protobuf.Any" {
		return u.unmarshalAny(pb, jsonFields)
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

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		valueForField, ok := consumeField(fd)
		if !ok {
			if fd.Cardinality() == protoreflect.Required {
				return fmt.Errorf("required field %q is not set", fd.Name())
			}
			continue
		}
		logfunc("Unmarshal %s type %v from %v\n", fd.JSONName(), fd.FullName(), string(valueForField))

		if fd.IsList() {
			if err := u.unmarshalList(pb.Mutable(fd).List(), fd, valueForField); err != nil {
				return err
			}
		} else if fd.IsMap() {
			mmap := pb.Mutable(fd).Map()
			if err := u.unmarshalMap(mmap, fd, valueForField); err != nil {
				return err
			}
		} else {
			// TODO ensure only one oneof variant specified?
			if err := u.unmarshalSingular(pb, fd, valueForField); err != nil {
				return err
			}
		}
	}
	if !u.DiscardUnknown && len(jsonFields) > 0 {
		// Pick any field to be the scapegoat.
		var f string
		for fname := range jsonFields {
			f = fname
			break
		}
		return fmt.Errorf("unknown field %q in %v", f, reflect.TypeOf(pb))
	}
	return nil

}

func (u *JSONUnmarshaler) unmarshalScalar(fd protoreflect.FieldDescriptor, inputValue json.RawMessage) (protoreflect.Value, error) {
	const b32 int = 32
	const b64 int = 64

	logfunc("Unmarshal scalar field %s (%s) from %s\n", fd.JSONName(), fd.FullName(), string(inputValue))
	if inputValue == nil {
		return protoreflect.Value{}, nil
	}

	kind := fd.Kind()
	switch kind {
	case protoreflect.BoolKind:
		var b bool
		if err := json.Unmarshal(inputValue, &b); err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBool(b), nil

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if v, ok := unmarshalInt(inputValue, b32); ok {
			return v, nil
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if v, ok := unmarshalInt(inputValue, b64); ok {
			return v, nil
		}

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if v, ok := unmarshalUint(inputValue, b32); ok {
			return v, nil
		}

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if v, ok := unmarshalUint(inputValue, b64); ok {
			return v, nil
		}

	case protoreflect.FloatKind:
		if v, ok := unmarshalFloat(inputValue, b32); ok {
			return v, nil
		}

	case protoreflect.DoubleKind:
		if v, ok := unmarshalFloat(inputValue, b64); ok {
			return v, nil
		}

	case protoreflect.StringKind:
		var s string
		if err := json.Unmarshal(inputValue, &s); err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfString(s), nil

	case protoreflect.BytesKind:
		if v, ok := unmarshalBytes(inputValue); ok {
			return v, nil
		}

	case protoreflect.EnumKind:
		if v, ok := unmarshalEnum(fd, inputValue); ok {
			return v, nil
		}

	default:
		panic(fmt.Sprintf("unmarshalScalar: invalid scalar kind %v", kind))
	}

	return protoreflect.Value{}, fmt.Errorf("invalid value for %v type: %v", kind, string(inputValue))
}

func unmarshalInt(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	if input[0] == '"' {
		// Decode number from string.
		s := strings.TrimSpace(string(input[1 : len(input)-1]))
		if len(s) != len(input)-2 {
			return protoreflect.Value{}, false
		}
		return getInt(json.RawMessage(s), bitSize)
	}
	return getInt(input, bitSize)
}

func getInt(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	parts, ok := parseNumberParts(input)
	if !ok {
		return protoreflect.Value{}, false
	}
	intStr, ok := normalizeToIntString(parts)
	if !ok {
		return protoreflect.Value{}, false

	}
	n, err := strconv.ParseInt(intStr, 10, bitSize)
	if err != nil {
		return protoreflect.Value{}, false
	}
	if bitSize == 32 {
		return protoreflect.ValueOfInt32(int32(n)), true
	}

	return protoreflect.ValueOfInt64(n), true
}

func unmarshalUint(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	if input[0] == '"' {
		// Decode number from string.
		s := strings.TrimSpace(string(input[1 : len(input)-1]))
		if len(s) != len(input)-2 {
			return protoreflect.Value{}, false
		}
		return getUint(json.RawMessage(s), bitSize)
	}
	return getUint(input, bitSize)
}

func getUint(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	parts, ok := parseNumberParts(input)
	if !ok {
		return protoreflect.Value{}, false
	}
	intStr, ok := normalizeToIntString(parts)
	if !ok {
		return protoreflect.Value{}, false

	}
	n, err := strconv.ParseUint(intStr, 10, bitSize)
	if err != nil {
		return protoreflect.Value{}, false
	}
	if bitSize == 32 {
		return protoreflect.ValueOfUint32(uint32(n)), true
	}

	return protoreflect.ValueOfUint64(n), true
}

func unmarshalFloat(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	if input[0] == '"' {
		s := string(input[1 : len(input)-1])
		switch s {
		case "NaN":
			if bitSize == 32 {
				return protoreflect.ValueOfFloat32(float32(math.NaN())), true
			}
			return protoreflect.ValueOfFloat64(math.NaN()), true
		case "Infinity":
			if bitSize == 32 {
				return protoreflect.ValueOfFloat32(float32(math.Inf(+1))), true
			}
			return protoreflect.ValueOfFloat64(math.Inf(+1)), true
		case "-Infinity":
			if bitSize == 32 {
				return protoreflect.ValueOfFloat32(float32(math.Inf(-1))), true
			}
			return protoreflect.ValueOfFloat64(math.Inf(-1)), true
		}

		// Decode number from string.
		if len(s) != len(strings.TrimSpace(s)) {
			return protoreflect.Value{}, false
		}
		return getFloat(json.RawMessage(s), bitSize)
	}
	return getFloat(input, bitSize)
}

func getFloat(input json.RawMessage, bitSize int) (protoreflect.Value, bool) {
	f, err := strconv.ParseFloat(string(input), bitSize)
	if err != nil {
		return protoreflect.Value{}, false
	}
	if bitSize == 32 {
		return protoreflect.ValueOfFloat32(float32(f)), true
	}

	return protoreflect.ValueOfFloat64(f), true
}

func unmarshalBytes(input json.RawMessage) (protoreflect.Value, bool) {
	if input[0] != '"' {
		return protoreflect.Value{}, false
	}

	s := string(input[1 : len(input)-1])
	enc := base64.StdEncoding
	if strings.ContainsAny(s, "-_") {
		enc = base64.URLEncoding
	}
	if len(s)%4 != 0 {
		enc = enc.WithPadding(base64.NoPadding)
	}
	b, err := enc.DecodeString(s)
	if err != nil {
		return protoreflect.Value{}, false
	}
	return protoreflect.ValueOfBytes(b), true
}

func unmarshalEnum(fd protoreflect.FieldDescriptor, inputValue json.RawMessage) (protoreflect.Value, bool) {
	if bytes.Equal(inputValue, []byte("null")) {
		if isNullValue(fd) {
			return protoreflect.ValueOfEnum(0), true
		}
	} else if inputValue[0] == '"' {
		// Lookup EnumNumber based on name.
		// Don't need to do unquoting; valid enum names
		// are from a limited character set.
		ed := fd.Enum()
		s := string(inputValue[1 : len(inputValue)-1])
		logfunc("unmarshalEnum %s (%s) from %s\n", fd.JSONName(), fd.FullName(), s)
		// We need to support both the canonical JSON format (SCREAMING_SNAKE_ENUMS) as well
		// as our older CamelCaseEnums form for compatibility, so we try both variants
		if enumVal := fd.Enum().Values().ByName(protoreflect.Name(s)); enumVal != nil {
			return protoreflect.ValueOfEnum(enumVal.Number()), true
		}
		typePrefix := strcase.ToScreamingSnake(string(ed.Name()))
		if !strings.HasPrefix(s, typePrefix) {
			s = fmt.Sprintf("%s_%s", typePrefix, strcase.ToScreamingSnake(s))
		}
		if enumVal := fd.Enum().Values().ByName(protoreflect.Name(s)); enumVal != nil {
			return protoreflect.ValueOfEnum(enumVal.Number()), true
		}
	} else {
		var n int32
		if err := json.Unmarshal(inputValue, &n); err != nil {
			return protoreflect.Value{}, false
		}
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n)), true
	}
	return protoreflect.Value{}, false
}

func (u *JSONUnmarshaler) unmarshalList(list protoreflect.List, fd protoreflect.FieldDescriptor, inputValue json.RawMessage) error {
	var s []json.RawMessage
	if err := json.Unmarshal(inputValue, &s); err != nil {
		return fmt.Errorf("bad ListValue: %v", err)
	}

	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		for i := range s {
			val := list.NewElement()
			logfunc("Unmarshal list element for %s of %T from %v\n", fd.JSONName(), val, s[i])
			if err := u.unmarshalMessage(val.Message(), s[i]); err != nil {
				return err
			}
			list.Append(val)
		}
	default:
		for i := range s {
			val, err := u.unmarshalScalar(fd, s[i])
			if err != nil {
				return err
			}
			list.Append(val)
		}
	}
	return nil
}

func (u *JSONUnmarshaler) unmarshalMap(mmap protoreflect.Map, fd protoreflect.FieldDescriptor, inputValue json.RawMessage) error {
	var mp map[string]json.RawMessage
	if err := json.Unmarshal(inputValue, &mp); err != nil {
		return err
	}
	if len(mp) == 0 {
		return nil
	}

	var unmarshalMapValue func(json.RawMessage) (protoreflect.Value, error)
	switch fd.MapValue().Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		unmarshalMapValue = func(input json.RawMessage) (protoreflect.Value, error) {
			val := mmap.NewValue()
			if err := u.unmarshalMessage(val.Message(), input); err != nil {
				return protoreflect.Value{}, err
			}
			return val, nil
		}
	default:
		unmarshalMapValue = func(input json.RawMessage) (protoreflect.Value, error) {
			return u.unmarshalScalar(fd.MapValue(), input)
		}
	}

	logfunc("unmarshal map %s %T->%T from %#v\n", fd.JSONName(), fd.MapKey().FullName(), fd.MapValue().FullName(), mp)
	for ks, raw := range mp {
		// Unmarshal map key. The core json library already decoded the key into a
		// string, so we handle that specially. Other types were quoted post-serialization.
		k, err := u.unmarshalMapKey(ks, fd.MapKey())
		if err != nil {
			return err
		}
		v, err := unmarshalMapValue(raw)
		if err != nil {
			return err
		}
		mmap.Set(k, v)
	}
	return nil
}

// unmarshalMapKey converts given token of Name kind into a protoreflect.MapKey.
// A map key type is any integral or string type.
func (u *JSONUnmarshaler) unmarshalMapKey(name string, fd protoreflect.FieldDescriptor) (protoreflect.MapKey, error) {
	const b32 = 32
	const b64 = 64
	const base10 = 10

	kind := fd.Kind()
	switch kind {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(name).MapKey(), nil

	case protoreflect.BoolKind:
		switch name {
		case "true":
			return protoreflect.ValueOfBool(true).MapKey(), nil
		case "false":
			return protoreflect.ValueOfBool(false).MapKey(), nil
		}

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if n, err := strconv.ParseInt(name, base10, b32); err == nil {
			return protoreflect.ValueOfInt32(int32(n)).MapKey(), nil
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if n, err := strconv.ParseInt(name, base10, b64); err == nil {
			return protoreflect.ValueOfInt64(int64(n)).MapKey(), nil
		}

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if n, err := strconv.ParseUint(name, base10, b32); err == nil {
			return protoreflect.ValueOfUint32(uint32(n)).MapKey(), nil
		}

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if n, err := strconv.ParseUint(name, base10, b64); err == nil {
			return protoreflect.ValueOfUint64(uint64(n)).MapKey(), nil
		}

	default:
		panic(fmt.Sprintf("invalid kind for map key: %v", kind))
	}

	return protoreflect.MapKey{}, fmt.Errorf("invalid value for %v key: %s", kind, name)
}

// unmarshalSingular converts/copies a value into the target.
// prop may be nil.
func (u *JSONUnmarshaler) unmarshalSingular(m protoreflect.Message, fd protoreflect.FieldDescriptor, inputValue json.RawMessage) error {
	var val protoreflect.Value
	var err error
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		val = m.NewField(fd)
		err = u.unmarshalMessage(val.Message(), inputValue)
	default:
		val, err = u.unmarshalScalar(fd, inputValue)
	}

	if err != nil {
		return err
	}
	m.Set(fd, val)
	return nil
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

// isNotDelim returns true if given byte is a not delimiter character.
func isNotDelim(c byte) bool {
	return (c == '-' || c == '+' || c == '.' || c == '_' ||
		('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9'))
}

// parseNumber reads the given []byte for a valid JSON number. If it is valid,
// it returns the number of bytes.  Parsing logic follows the definition in
// https://tools.ietf.org/html/rfc7159#section-6, and is based off
// encoding/json.isValidNumber function.
func parseNumber(input []byte) (int, bool) {
	var n int

	s := input
	if len(s) == 0 {
		return 0, false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		n++
		if len(s) == 0 {
			return 0, false
		}
	}

	// Digits
	switch {
	case s[0] == '0':
		s = s[1:]
		n++

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		n++
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}

	default:
		return 0, false
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		n += 2
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		n++
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			n++
			if len(s) == 0 {
				return 0, false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}
	}

	// Check that next byte is a delimiter or it is at the end.
	if n < len(input) && isNotDelim(input[n]) {
		return 0, false
	}

	return n, true
}

// numberParts is the result of parsing out a valid JSON number. It contains
// the parts of a number. The parts are used for integer conversion.
type numberParts struct {
	neg  bool
	intp []byte
	frac []byte
	exp  []byte
}

// parseNumber constructs numberParts from given []byte. The logic here is
// similar to consumeNumber above with the difference of having to construct
// numberParts. The slice fields in numberParts are subslices of the input.
func parseNumberParts(input []byte) (numberParts, bool) {
	var neg bool
	var intp []byte
	var frac []byte
	var exp []byte

	s := input
	if len(s) == 0 {
		return numberParts{}, false
	}

	// Optional -
	if s[0] == '-' {
		neg = true
		s = s[1:]
		if len(s) == 0 {
			return numberParts{}, false
		}
	}

	// Digits
	switch {
	case s[0] == '0':
		// Skip first 0 and no need to store.
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		intp = s
		n := 1
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}
		intp = intp[:n]

	default:
		return numberParts{}, false
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		frac = s[1:]
		n := 1
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}
		frac = frac[:n]
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		exp = s
		n := 0
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			n++
			if len(s) == 0 {
				return numberParts{}, false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
			n++
		}
		exp = exp[:n]
	}

	return numberParts{
		neg:  neg,
		intp: intp,
		frac: bytes.TrimRight(frac, "0"), // Remove unnecessary 0s to the right.
		exp:  exp,
	}, true
}

// normalizeToIntString returns an integer string in normal form without the
// E-notation for given numberParts. It will return false if it is not an
// integer or if the exponent exceeds than max/min int value.
func normalizeToIntString(n numberParts) (string, bool) {
	intpSize := len(n.intp)
	fracSize := len(n.frac)

	if intpSize == 0 && fracSize == 0 {
		return "0", true
	}

	var exp int
	if len(n.exp) > 0 {
		i, err := strconv.ParseInt(string(n.exp), 10, 32)
		if err != nil {
			return "", false
		}
		exp = int(i)
	}

	var num []byte
	if exp >= 0 {
		// For positive E, shift fraction digits into integer part and also pad
		// with zeroes as needed.

		// If there are more digits in fraction than the E value, then the
		// number is not an integer.
		if fracSize > exp {
			return "", false
		}

		// Make sure resulting digits are within max value limit to avoid
		// unnecessarily constructing a large byte slice that may simply fail
		// later on.
		const maxDigits = 20 // Max uint64 value has 20 decimal digits.
		if intpSize+exp > maxDigits {
			return "", false
		}

		// Set cap to make a copy of integer part when appended.
		num = n.intp[:len(n.intp):len(n.intp)]
		num = append(num, n.frac...)
		for i := 0; i < exp-fracSize; i++ {
			num = append(num, '0')
		}
	} else {
		// For negative E, shift digits in integer part out.

		// If there are fractions, then the number is not an integer.
		if fracSize > 0 {
			return "", false
		}

		// index is where the decimal point will be after adjusting for negative
		// exponent.
		index := intpSize + exp
		if index < 0 {
			return "", false
		}

		num = n.intp
		// If any of the digits being shifted to the right of the decimal point
		// is non-zero, then the number is not an integer.
		for i := index; i < intpSize; i++ {
			if num[i] != '0' {
				return "", false
			}
		}
		num = num[:index]
	}

	if n.neg {
		return "-" + string(num), true
	}
	return string(num), true
}
