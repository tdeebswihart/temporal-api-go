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

package common

import (
	//"bytes"
	"encoding"
	gojson "encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"go.temporal.io/api/internal/protojson/json"
	"go.temporal.io/api/temporalproto"
	//	"google.golang.org/protobuf/encoding/protojson"
)

var _ temporalproto.JSONPBMaybeMarshaler = (*Payload)(nil)
var _ temporalproto.JSONPBMaybeMarshaler = (*Payloads)(nil)

// !!! This file is copied from internal/temporalcommonv1 to common/v1.
// !!! DO NOT EDIT at common/v1/payload_json.go.
func marshalSingular(enc *json.Encoder, value interface{}) error {
	switch vv := value.(type) {
	case string:
		enc.WriteString(vv)
	case bool:
		enc.WriteBool(vv)
	case int:
		enc.WriteInt(int64(vv))
	case int64:
		enc.WriteInt(vv)
	case uint:
		enc.WriteUint(uint64(vv))
	case uint64:
		enc.WriteUint(vv)
	case float32:
		enc.WriteFloat(float64(vv), 32)
	case float64:
		enc.WriteFloat(vv, 64)
	default:
		return fmt.Errorf("cannot marshal type %[1]T value %[1]v", vv)
	}
	return nil
}

func marshalStruct(enc *json.Encoder, vv reflect.Value) error {
	enc.StartObject()
	defer enc.EndObject()
	ty := vv.Type()

Loop:
	for i, n := 0, vv.NumField(); i < n; i++ {
		f := vv.Field(i)
		name := f.String()
		// lowercase. private field
		if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}

		// Handle encoding/json struct tags
		tag, present := ty.Field(i).Tag.Lookup("json")
		if present {
			opts := strings.Split(tag, ",")
			for j := range opts {
				if opts[j] == "omitempty" && vv.IsZero() {
					continue Loop
				} else if opts[j] == "-" {
					continue Loop
				}
				// name overridden
				name = opts[j]
			}
		}
		if err := enc.WriteName(name); err != nil {
			return fmt.Errorf("unable to write name %s: %w", name, err)
		}
		if err := marshalSingular(enc, f); err != nil {
			return fmt.Errorf("unable to marshal value for name %s: %w", name, err)
		}
	}
	return nil
}

type keyVal struct {
	k string
	v interface{}
}

// Map keys must be either strings or integers. We don't use encoding.TextMarshaler so we don't care
func marshalMap(enc *json.Encoder, vv reflect.Value) error {
	enc.StartObject()
	defer enc.EndObject()

	sv := make([]keyVal, vv.Len())
	iter := vv.MapRange()
	for i := 0; iter.Next(); i++ {
		k := iter.Key()
		sv[i].v = iter.Value().Interface()

		if k.Kind() == reflect.String {
			sv[i].k = k.String()
		} else if tm, ok := k.Interface().(encoding.TextMarshaler); ok {
			if k.Kind() == reflect.Pointer && k.IsNil() {
				return nil
			}
			buf, err := tm.MarshalText()
			sv[i].k = string(buf)
			if err != nil {
				return err
			}
		} else {
			switch k.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				sv[i].k = strconv.FormatInt(k.Int(), 10)
				return nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				sv[i].k = strconv.FormatUint(k.Uint(), 10)
				return nil
			default:
				return fmt.Errorf("map key type %T not supported", k)
			}
		}
	}
	// Sort keys just like encoding/json
	sort.Slice(sv, func(i, j int) bool { return sv[i].k < sv[j].k })

	for i := 0; i < len(sv); i++ {
		if err := enc.WriteName(sv[i].k); err != nil {
			return fmt.Errorf("unable to write name %s: %w", sv[i].k, err)
		}
		if err := marshalSingular(enc, sv[i].v); err != nil {
			return fmt.Errorf("unable to marshal value for name %s: %w", sv[i].k, err)
		}
	}
	return nil
}

func marshalValue(enc *json.Encoder, vv reflect.Value) error {
	switch vv.Kind() {
	case reflect.Slice, reflect.Array:
		if vv.IsNil() || vv.Len() == 0 {
			enc.WriteNull()
			return nil
		}
		enc.StartArray()
		defer enc.EndArray()
		for i := 0; i < vv.Len(); i++ {
			if err := marshalValue(enc, vv.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Interface, reflect.Pointer:
		if vv.IsNil() {
			return nil
		}
		marshalValue(enc, vv.Elem())
	case reflect.Struct:
		marshalStruct(enc, vv)
	case reflect.Map:
		if vv.IsNil() || vv.Len() == 0 {
			enc.WriteNull()
			return nil
		}
		marshalMap(enc, vv)
	case reflect.Bool,
		reflect.String,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		marshalSingular(enc, vv.Interface())
	default:
		return fmt.Errorf("cannot marshal %[1]T value %[1]v", vv.Interface())
	}

	return nil
}

func marshal(enc *json.Encoder, value interface{}) error {
	return marshalValue(enc, reflect.ValueOf(value))
}

// Key on the marshaler metadata specifying whether shorthand is disabled.
//
// WARNING: This is internal API and should not be called externally.
const DisablePayloadShorthandMetadataKey = "__temporal_disable_payload_shorthand"

// MaybeMarshalJSONPB implements
// [go.temporal.io/api/internal/temporaljsonpb.JSONPBMaybeMarshaler.MaybeMarshalJSONPB].
//
// WARNING: This is internal API and should not be called externally.
func (p *Payloads) MaybeMarshalJSONPB(meta map[string]interface{}, enc *json.Encoder) (handled bool, err error) {
	// If this is nil, ignore
	if p == nil {
		return false, nil
	}
	// If shorthand is disabled, ignore
	if disabled, _ := meta[DisablePayloadShorthandMetadataKey].(bool); disabled {
		return false, nil
	}

	if len(p.Payloads) == 0 {
		return true, nil
	}
	// We only support marshalling to shorthand if all payloads are handled or
	// there  are no payloads
	enc.StartArray()
	defer enc.EndArray()
	for i := 0; i < len(p.Payloads); i++ {
		// If any are not handled or there is an error, return
		handled, val, err := p.Payloads[i].toJSONShorthand()
		if !handled || err != nil {
			return handled, err
		}
		if err := marshal(enc, val); err != nil {
			return true, err
		}
	}
	return true, err
}

// MaybeUnmarshalJSONPB implements
// [go.temporal.io/api/internal/temporaljsonpb.JSONPBMaybeUnmarshaler.MaybeUnmarshalJSONPB].
//
// WARNING: This is internal API and should not be called externally.
// func (p *Payloads) MaybeUnmarshalJSONPB(u jsonpb.Unmarshaler, b []byte) (handled bool, err error) {
// 	// If this is nil, ignore (should never be)
// 	if p == nil {
// 		return false, nil
// 	}
// 	// If shorthand is disabled, ignore
// 	if disabled, _ := u.Metadata[DisablePayloadShorthandMetadataKey].(bool); disabled {
// 		return false, nil
// 	}
// 	// Try to deserialize into slice. If this fails, it is not shorthand and this
// 	// does not handle it. This means on invalid JSON, we let the proto JSON
// 	// handler fail instead of here.
// 	var payloadJSONs []json.RawMessage
// 	if json.Unmarshal(b, &payloadJSONs) != nil {
// 		return false, nil
// 	}
// 	// Convert each (some may be shorthand, some may not)
// 	p.Payloads = make([]*Payload, len(payloadJSONs))
// 	for i, payloadJSON := range payloadJSONs {
// 		p.Payloads[i] = &Payload{}
// 		p.Payloads[i].fromJSONMaybeShorthand(payloadJSON)
// 	}
// 	return true, nil
// }

// MaybeMarshalJSONPB implements
// [go.temporal.io/api/internal/temporaljsonpb.JSONPBMaybeMarshaler.MaybeMarshalJSONPB].
//
// WARNING: This is internal API and should not be called externally.
func (p *Payload) MaybeMarshalJSONPB(meta map[string]interface{}, enc *json.Encoder) (handled bool, err error) {
	// If this is nil, ignore
	if p == nil {
		return false, nil
	}
	// If shorthand is disabled, ignore
	if disabled, _ := meta[DisablePayloadShorthandMetadataKey].(bool); disabled {
		return false, nil
	}
	// If any are not handled or there is an error, return
	handled, val, err := p.toJSONShorthand()
	if !handled || err != nil {
		return handled, err
	}
	logfunc("marshalling %v", val)
	return true, marshal(enc, val)
}

// MaybeUnmarshalJSONPB implements
// [go.temporal.io/api/internal/temporaljsonpb.JSONPBMaybeUnmarshaler.MaybeUnmarshalJSONPB].
//
// WARNING: This is internal API and should not be called externally.
// func (p *Payload) MaybeUnmarshalJSONPB(u *jsonpb.Unmarshaler, b []byte) (handled bool, err error) {
// 	// If this is nil, ignore (should never be)
// 	if p == nil {
// 		return false, nil
// 	}
// 	// If shorthand is disabled, ignore
// 	if disabled, _ := u.Metadata[DisablePayloadShorthandMetadataKey].(bool); disabled {
// 		return false, nil
// 	}
// 	// Always considered handled, unmarshaler ignored (unknown fields always
// 	// disallowed for non-shorthand payloads at this time)
// 	p.fromJSONMaybeShorthand(b)
// 	return true, nil
// }

func (p *Payload) toJSONShorthand() (handled bool, value interface{}, err error) {
	// Only support binary null, plain JSON and proto JSON
	switch string(p.Metadata["encoding"]) {
	case "binary/null":
		// Leave value as nil
		handled = true
	case "json/plain":
		// Must only have this single metadata
		if len(p.Metadata) != 1 {
			return false, nil, nil
		}
		// We unmarshal because we may have to indent. We let this error fail the
		// marshaller.
		handled = true
		err = gojson.Unmarshal(p.Data, &value)
	case "json/protobuf":
		// Must have the message type and no other metadata
		msgType := string(p.Metadata["messageType"])
		if msgType == "" || len(p.Metadata) != 2 {
			return false, nil, nil
		}
		// Since this is a proto object, this must unmarshal to a object. We let
		// this error fail the marshaller.
		var valueMap map[string]interface{}
		handled = true
		err = gojson.Unmarshal(p.Data, &valueMap)
		// Put the message type on the object
		if valueMap != nil {
			valueMap["_protoMessageType"] = msgType
		}
		value = valueMap
	}
	return
}

// func (p *Payload) fromJSONMaybeShorthand(b []byte) {
// 	// We need to try to deserialize into the regular payload first. If it works
// 	// and there is metadata _and_ data actually present (or null with a null
// 	// metadata encoding), we assume it's a non-shorthand payload. If it fails
// 	// (which it will if not an object or there is an unknown field or if
// 	// 'metadata' is not string + base64 or if 'data' is not base64), we assume
// 	// shorthand. We are ok disallowing unknown fields for payloads here even if
// 	// the outer unmarshaler allows them.
// 	if protojson.Unmarshal(b, p) == nil && len(p.Metadata) > 0 {
// 		// A raw payload must either have data or a binary/null encoding
// 		if len(p.Data) > 0 || string(p.Metadata["encoding"]) == "binary/null" {
// 			return
// 		}
// 	}
//
// 	// If it's "null", set no data and just metadata
// 	if string(b) == "null" {
// 		p.Data = nil
// 		p.Metadata = map[string][]byte{"encoding": []byte("binary/null")}
// 		return
// 	}
//
// 	// Now that we know it is shorthand, it might be a proto JSON with a message
// 	// type. If it does have the message type, we need to remove it and
// 	// re-serialize it to data. So the quickest way to check whether it has the
// 	// message type is to search for the key.
// 	p.Data = b
// 	p.Metadata = map[string][]byte{"encoding": []byte("json/plain")}
// 	if bytes.Contains(p.Data, []byte(`"_protoMessageType"`)) {
// 		// Try to unmarshal into map, extract and remove key, and re-serialize
// 		var valueMap map[string]interface{}
// 		if gojson.Unmarshal(p.Data, &valueMap) == nil {
// 			if msgType, _ := valueMap["_protoMessageType"].(string); msgType != "" {
// 				// Now we know it's a proto JSON, so remove the key and re-serialize
// 				delete(valueMap, "_protoMessageType")
// 				// This won't error. The resulting JSON payload data may not be exactly
// 				// what user passed in sans message type (e.g. user may have indented or
// 				// did not have same field order), but that is acceptable when going
// 				// from shorthand to non-shorthand.
// 				p.Data, _ = gojson.Marshal(valueMap)
// 				p.Metadata["encoding"] = []byte("json/protobuf")
// 				p.Metadata["messageType"] = []byte(msgType)
// 			}
// 		}
// 	}
// }
