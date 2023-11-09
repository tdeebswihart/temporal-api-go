package temporalproto

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"math"
// 	"reflect"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"
//
// 	"google.golang.org/protobuf/proto"
// 	"google.golang.org/protobuf/reflect/protoreflect"
// )
//
// // JSONPBMarshaler is implemented by protobuf messages that customize the
// // way they are marshaled to JSON. Messages that implement this should
// // also implement JSONPBUnmarshaler so that the custom format can be
// // parsed.
// //
// // The JSON marshaling must follow the proto to JSON specification:
// //
// //	https://developers.google.com/protocol-buffers/docs/proto3#json
// type JSONPBMarshaler interface {
// 	MarshalJSONPB(*JSONMarshaler) ([]byte, error)
// }
//
// // JSONMarshaler is a configurable object for converting between
// // protocol buffer objects and a JSON representation for them.
// type JSONMarshaler struct {
// 	// Whether to render enum values as integers, as opposed to string values.
// 	EnumsAsInts bool
//
// 	// Whether to render fields with zero values.
// 	EmitDefaults bool
//
// 	// A string to indent each level by. The presence of this field will
// 	// also cause a space to appear between the field separator and
// 	// value, and for newlines to be appear between fields and array
// 	// elements.
// 	Indent string
//
// 	// Whether to use the original (.proto) name for fields.
// 	OrigName bool
// }
//
// // Marshal marshals a protocol buffer into JSON.
// func (m *JSONMarshaler) Marshal(out io.Writer, pb proto.Message) error {
// 	v := reflect.ValueOf(pb)
// 	if pb == nil || (v.Kind() == reflect.Ptr && v.IsNil()) {
// 		return errors.New("Marshal called with nil")
// 	}
// 	// Check for unset required fields first.
// 	if err := proto.CheckInitialized(pb); err != nil {
// 		return err
// 	}
// 	writer := &errWriter{writer: out}
// 	return m.marshalObject(writer, pb, "", "")
// }
//
// // MarshalToString converts a protocol buffer object to JSON string.
// func (m *JSONMarshaler) MarshalToString(pb proto.Message) (string, error) {
// 	var buf bytes.Buffer
// 	if err := m.Marshal(&buf, pb); err != nil {
// 		return "", err
// 	}
// 	return buf.String(), nil
// }
//
// // marshalObject writes a struct to the Writer.
// func (m *JSONMarshaler) marshalObject(out *errWriter, v proto.Message, indent, typeURL string) error {
// 	if jsm, ok := v.(JSONPBMarshaler); ok {
// 		b, err := jsm.MarshalJSONPB(m)
// 		if err != nil {
// 			return err
// 		}
// 		if typeURL != "" {
// 			// we are marshaling this object to an Any type
// 			var js map[string]*json.RawMessage
// 			if err = json.Unmarshal(b, &js); err != nil {
// 				return fmt.Errorf("type %T produced invalid JSON: %v", v, err)
// 			}
// 			turl, err := json.Marshal(typeURL)
// 			if err != nil {
// 				return fmt.Errorf("failed to marshal type URL %q to JSON: %v", typeURL, err)
// 			}
// 			js["@type"] = (*json.RawMessage)(&turl)
// 			if m.Indent != "" {
// 				b, err = json.MarshalIndent(js, indent, m.Indent)
// 			} else {
// 				b, err = json.Marshal(js)
// 			}
// 			if err != nil {
// 				return err
// 			}
// 		}
//
// 		out.write(string(b))
// 		return out.err
// 	}
//
// 	if jsm, ok := v.(JSONPBMaybeMarshaler); ok {
// 		if handled, b, err := jsm.MaybeMarshalJSONPB(m, indent); handled && err != nil {
// 			return err
// 		} else if handled {
// 			out.write(string(b))
// 			return out.err
// 		}
// 	}
//
// 	s := reflect.ValueOf(v).Elem()
//
// 	// Handle well-known types.
// 	// NOTE: currently disabled.
// 	if wkt, ok := v.(isWkt); ok {
// 		switch wkt.XXX_WellKnownType() {
// 		case "DoubleValue", "FloatValue", "Int64Value", "UInt64Value",
// 			"Int32Value", "UInt32Value", "BoolValue", "StringValue", "BytesValue":
// 			// "Wrappers use the same representation in JSON
// 			//  as the wrapped primitive type, ..."
// 			sprop := v.ProtoReflect().Descriptor()
// 			return m.marshalValue(out, sprop.Fields().Get(0), s.Field(0), indent)
// 		case "Any":
// 			// Any is a bit more involved.
// 			return m.marshalAny(out, v, indent)
// 		case "Duration":
// 			s, ns := s.Field(0).Int(), s.Field(1).Int()
// 			if s < -maxSecondsInDuration || s > maxSecondsInDuration {
// 				return fmt.Errorf("seconds out of range %v", s)
// 			}
// 			if ns <= -secondInNanos || ns >= secondInNanos {
// 				return fmt.Errorf("ns out of range (%v, %v)", -secondInNanos, secondInNanos)
// 			}
// 			if (s > 0 && ns < 0) || (s < 0 && ns > 0) {
// 				return errors.New("signs of seconds and nanos do not match")
// 			}
// 			// Generated output always contains 0, 3, 6, or 9 fractional digits,
// 			// depending on required precision, followed by the suffix "s".
// 			f := "%d.%09d"
// 			if ns < 0 {
// 				ns = -ns
// 				if s == 0 {
// 					f = "-%d.%09d"
// 				}
// 			}
// 			x := fmt.Sprintf(f, s, ns)
// 			x = strings.TrimSuffix(x, "000")
// 			x = strings.TrimSuffix(x, "000")
// 			x = strings.TrimSuffix(x, ".000")
// 			out.write(`"`)
// 			out.write(x)
// 			out.write(`s"`)
// 			return out.err
// 		case "Struct", "ListValue":
// 			// Let marshalValue handle the `Struct.fields` map or the `ListValue.values` slice.
// 			// TODO: pass the correct Properties if needed.
// 			return m.marshalValue(out, nil, s.Field(0), indent)
// 		case "Timestamp":
// 			// "RFC 3339, where generated output will always be Z-normalized
// 			//  and uses 0, 3, 6 or 9 fractional digits."
// 			s, ns := s.Field(0).Int(), s.Field(1).Int()
// 			if ns < 0 || ns >= secondInNanos {
// 				return fmt.Errorf("ns out of range [0, %v)", secondInNanos)
// 			}
// 			t := time.Unix(s, ns).UTC()
// 			// time.RFC3339Nano isn't exactly right (we need to get 3/6/9 fractional digits).
// 			x := t.Format("2006-01-02T15:04:05.000000000")
// 			x = strings.TrimSuffix(x, "000")
// 			x = strings.TrimSuffix(x, "000")
// 			x = strings.TrimSuffix(x, ".000")
// 			out.write(`"`)
// 			out.write(x)
// 			out.write(`Z"`)
// 			return out.err
// 		case "Value":
// 			// Value has a single oneof.
// 			kind := s.Field(0)
// 			if kind.IsNil() {
// 				// "absence of any variant indicates an error"
// 				return errors.New("nil Value")
// 			}
// 			// oneof -> *T -> T -> T.F
// 			x := kind.Elem().Elem().Field(0)
// 			// TODO: pass the correct Properties if needed.
// 			return m.marshalValue(out, nil, x, indent)
// 		}
// 	}
//
// 	out.write("{")
// 	if m.Indent != "" {
// 		out.write("\n")
// 	}
//
// 	firstField := true
//
// 	if typeURL != "" {
// 		if err := m.marshalTypeURL(out, indent, typeURL); err != nil {
// 			return err
// 		}
// 		firstField = false
// 	}
//
// 	sprop := v.ProtoReflect().Descriptor()
// 	fields := sprop.Fields()
// 	for i := 0; i < s.NumField(); i++ {
// 		value := s.Field(i)
// 		valueField := s.Type().Field(i)
//
// 		//this is not a protobuf field
// 		if valueField.Tag.Get("protobuf") == "" && valueField.Tag.Get("protobuf_oneof") == "" {
// 			continue
// 		}
//
// 		// IsNil will panic on most value kinds.
// 		switch value.Kind() {
// 		case reflect.Chan, reflect.Func, reflect.Interface:
// 			if value.IsNil() {
// 				continue
// 			}
// 		}
// 		if !m.EmitDefaults {
// 			switch value.Kind() {
// 			case reflect.Bool:
// 				if !value.Bool() {
// 					continue
// 				}
// 			case reflect.Int32, reflect.Int64:
// 				if value.Int() == 0 {
// 					continue
// 				}
// 			case reflect.Uint32, reflect.Uint64:
// 				if value.Uint() == 0 {
// 					continue
// 				}
// 			case reflect.Float32, reflect.Float64:
// 				if value.Float() == 0 {
// 					continue
// 				}
// 			case reflect.String:
// 				if value.Len() == 0 {
// 					continue
// 				}
// 			case reflect.Map, reflect.Ptr, reflect.Slice:
// 				if value.IsNil() {
// 					continue
// 				}
// 			}
// 		}
//
// 		// Oneof fields need special handling.
// 		if valueField.Tag.Get("protobuf_oneof") != "" {
// 			// value is an interface containing &T{real_value}.
// 			sv := value.Elem().Elem() // interface -> *T -> T
// 			value = sv.Field(0)
// 			valueField = sv.Type().Field(0)
// 		}
// 		if !firstField {
// 			m.writeSep(out)
// 		}
// 		if err := m.marshalField(out, fields.Get(i), value, indent); err != nil {
// 			return err
// 		}
// 		firstField = false
// 	}
//
// 	if m.Indent != "" {
// 		out.write("\n")
// 		out.write(indent)
// 	}
// 	out.write("}")
// 	return out.err
// }
//
// func (m *JSONMarshaler) writeSep(out *errWriter) {
// 	if m.Indent != "" {
// 		out.write(",\n")
// 	} else {
// 		out.write(",")
// 	}
// }
//
// func (m *JSONMarshaler) marshalAny(out *errWriter, any proto.Message, indent string) error {
// 	// "If the Any contains a value that has a special JSON mapping,
// 	//  it will be converted as follows: {"@type": xxx, "value": yyy}.
// 	//  Otherwise, the value will be converted into a JSON object,
// 	//  and the "@type" field will be inserted to indicate the actual data type."
// 	v := reflect.ValueOf(any).Elem()
// 	turl := v.Field(0).String()
// 	val := v.Field(1).Bytes()
//
// 	msg, err := resolveType(turl)
// 	if err != nil {
// 		return err
// 	}
//
// 	if err := proto.Unmarshal(val, msg); err != nil {
// 		return err
// 	}
//
// 	if _, ok := msg.(isWkt); ok {
// 		out.write("{")
// 		if m.Indent != "" {
// 			out.write("\n")
// 		}
// 		if err := m.marshalTypeURL(out, indent, turl); err != nil {
// 			return err
// 		}
// 		m.writeSep(out)
// 		if m.Indent != "" {
// 			out.write(indent)
// 			out.write(m.Indent)
// 			out.write(`"value": `)
// 		} else {
// 			out.write(`"value":`)
// 		}
// 		if err := m.marshalObject(out, msg, indent+m.Indent, ""); err != nil {
// 			return err
// 		}
// 		if m.Indent != "" {
// 			out.write("\n")
// 			out.write(indent)
// 		}
// 		out.write("}")
// 		return out.err
// 	}
//
// 	return m.marshalObject(out, msg, indent, turl)
// }
//
// func (m *JSONMarshaler) marshalTypeURL(out *errWriter, indent, typeURL string) error {
// 	if m.Indent != "" {
// 		out.write(indent)
// 		out.write(m.Indent)
// 	}
// 	out.write(`"@type":`)
// 	if m.Indent != "" {
// 		out.write(" ")
// 	}
// 	b, err := json.Marshal(typeURL)
// 	if err != nil {
// 		return err
// 	}
// 	out.write(string(b))
// 	return out.err
// }
//
// // marshalField writes field description and value to the Writer.
// func (m *JSONMarshaler) marshalField(out *errWriter, prop protoreflect.FieldDescriptor, v reflect.Value, indent string) error {
// 	if m.Indent != "" {
// 		out.write(indent)
// 		out.write(m.Indent)
// 	}
// 	out.write(`"`)
// 	if m.OrigName || prop.JSONName() == "" {
// 		out.write(prop.TextName())
// 	} else {
// 		out.write(prop.JSONName())
// 	}
// 	out.write(`":`)
// 	if m.Indent != "" {
// 		out.write(" ")
// 	}
// 	if err := m.marshalValue(out, prop, v, indent); err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// // marshalValue writes the value to the Writer.
// func (m *JSONMarshaler) marshalValue(out *errWriter, prop protoreflect.FieldDescriptor, v reflect.Value, indent string) error {
//
// 	v = reflect.Indirect(v)
//
// 	// Handle nil pointer
// 	if v.Kind() == reflect.Invalid {
// 		out.write("null")
// 		return out.err
// 	}
//
// 	// Handle repeated elements.
// 	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 {
// 		out.write("[")
// 		comma := ""
// 		for i := 0; i < v.Len(); i++ {
// 			sliceVal := v.Index(i)
// 			out.write(comma)
// 			if m.Indent != "" {
// 				out.write("\n")
// 				out.write(indent)
// 				out.write(m.Indent)
// 				out.write(m.Indent)
// 			}
// 			if err := m.marshalValue(out, prop, sliceVal, indent+m.Indent); err != nil {
// 				return err
// 			}
// 			comma = ","
// 		}
// 		if m.Indent != "" {
// 			out.write("\n")
// 			out.write(indent)
// 			out.write(m.Indent)
// 		}
// 		out.write("]")
// 		return out.err
// 	}
//
// 	// Handle well-known types.
// 	// Most are handled up in marshalObject (because 99% are messages).
// 	if v.Type().Implements(wktType) {
// 		wkt := v.Interface().(isWkt)
// 		switch wkt.XXX_WellKnownType() {
// 		case "NullValue":
// 			out.write("null")
// 			return out.err
// 		}
// 	}
//
// 	// Handle enumerations.
// 	if !m.EnumsAsInts && prop.Kind() == protoreflect.EnumKind {
// 		// Unknown enum values will are stringified by the proto library as their
// 		// value. Such values should _not_ be quoted or they will be interpreted
// 		// as an enum string instead of their value.
// 		enumStr := v.Interface().(fmt.Stringer).String()
// 		var valStr string
// 		if v.Kind() == reflect.Ptr {
// 			valStr = strconv.Itoa(int(v.Elem().Int()))
// 		} else {
// 			valStr = strconv.Itoa(int(v.Int()))
// 		}
//
// 		if m, ok := v.Interface().(interface {
// 			MarshalJSON() ([]byte, error)
// 		}); ok {
// 			data, err := m.MarshalJSON()
// 			if err != nil {
// 				return err
// 			}
// 			enumStr = string(data)
// 			enumStr, err = strconv.Unquote(enumStr)
// 			if err != nil {
// 				return err
// 			}
// 		}
//
// 		isKnownEnum := enumStr != valStr
//
// 		if isKnownEnum {
// 			out.write(`"`)
// 		}
// 		out.write(enumStr)
// 		if isKnownEnum {
// 			out.write(`"`)
// 		}
// 		return out.err
// 	}
//
// 	// Handle nested messages.
// 	if v.Kind() == reflect.Struct {
// 		i := v
// 		if v.CanAddr() {
// 			i = v.Addr()
// 		} else {
// 			i = reflect.New(v.Type())
// 			i.Elem().Set(v)
// 		}
// 		iface := i.Interface()
// 		if iface == nil {
// 			out.write(`null`)
// 			return out.err
// 		}
//
// 		if m, ok := v.Interface().(interface {
// 			MarshalJSON() ([]byte, error)
// 		}); ok {
// 			data, err := m.MarshalJSON()
// 			if err != nil {
// 				return err
// 			}
// 			out.write(string(data))
// 			return nil
// 		}
//
// 		pm, ok := iface.(proto.Message)
// 		if !ok {
// 			return fmt.Errorf("%v does not implement proto.Message", v.Type())
// 		}
// 		return m.marshalObject(out, pm, indent+m.Indent, "")
// 	}
//
// 	// Handle maps.
// 	// Since Go randomizes map iteration, we sort keys for stable output.
// 	if v.Kind() == reflect.Map {
// 		out.write(`{`)
// 		keys := v.MapKeys()
// 		sort.Sort(mapKeys(keys))
// 		for i, k := range keys {
// 			if i > 0 {
// 				out.write(`,`)
// 			}
// 			if m.Indent != "" {
// 				out.write("\n")
// 				out.write(indent)
// 				out.write(m.Indent)
// 				out.write(m.Indent)
// 			}
//
// 			// TODO handle map key prop properly
// 			b, err := json.Marshal(k.Interface())
// 			if err != nil {
// 				return err
// 			}
// 			s := string(b)
//
// 			// If the JSON is not a string value, encode it again to make it one.
// 			if !strings.HasPrefix(s, `"`) {
// 				b, err := json.Marshal(s)
// 				if err != nil {
// 					return err
// 				}
// 				s = string(b)
// 			}
//
// 			out.write(s)
// 			out.write(`:`)
// 			if m.Indent != "" {
// 				out.write(` `)
// 			}
//
// 			vprop := prop
// 			// TODO?
// 			// if prop != nil && prop.MapValue() != nil {
// 			// 	vprop = prop.MapValue()
// 			// }
// 			if err := m.marshalValue(out, vprop, v.MapIndex(k), indent+m.Indent); err != nil {
// 				return err
// 			}
// 		}
// 		if m.Indent != "" {
// 			out.write("\n")
// 			out.write(indent)
// 			out.write(m.Indent)
// 		}
// 		out.write(`}`)
// 		return out.err
// 	}
//
// 	// Handle non-finite floats, e.g. NaN, Infinity and -Infinity.
// 	if v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64 {
// 		f := v.Float()
// 		var sval string
// 		switch {
// 		case math.IsInf(f, 1):
// 			sval = `"Infinity"`
// 		case math.IsInf(f, -1):
// 			sval = `"-Infinity"`
// 		case math.IsNaN(f):
// 			sval = `"NaN"`
// 		}
// 		if sval != "" {
// 			out.write(sval)
// 			return out.err
// 		}
// 	}
//
// 	// Default handling defers to the encoding/json library.
// 	b, err := json.Marshal(v.Interface())
// 	if err != nil {
// 		return err
// 	}
// 	needToQuote := string(b[0]) != `"` && (v.Kind() == reflect.Int64 || v.Kind() == reflect.Uint64)
// 	if needToQuote {
// 		out.write(`"`)
// 	}
// 	out.write(string(b))
// 	if needToQuote {
// 		out.write(`"`)
// 	}
// 	return out.err
// }
//
// // Map fields may have key types of non-float scalars, strings and enums.
// // The easiest way to sort them in some deterministic order is to use fmt.
// // If this turns out to be inefficient we can always consider other options,
// // such as doing a Schwartzian transform.
// //
// // Numeric keys are sorted in numeric order per
// // https://developers.google.com/protocol-buffers/docs/proto#maps.
// type mapKeys []reflect.Value
//
// func (s mapKeys) Len() int      { return len(s) }
// func (s mapKeys) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
// func (s mapKeys) Less(i, j int) bool {
// 	if k := s[i].Kind(); k == s[j].Kind() {
// 		switch k {
// 		case reflect.String:
// 			return s[i].String() < s[j].String()
// 		case reflect.Int32, reflect.Int64:
// 			return s[i].Int() < s[j].Int()
// 		case reflect.Uint32, reflect.Uint64:
// 			return s[i].Uint() < s[j].Uint()
// 		}
// 	}
// 	return fmt.Sprint(s[i].Interface()) < fmt.Sprint(s[j].Interface())
// }

// // Writer wrapper inspired by https://blog.golang.org/errors-are-values
// type errWriter struct {
// 	writer io.Writer
// 	err    error
// }
//
// func (w *errWriter) write(str string) {
// 	if w.err != nil {
// 		return
// 	}
// 	_, w.err = w.writer.Write([]byte(str))
// }
