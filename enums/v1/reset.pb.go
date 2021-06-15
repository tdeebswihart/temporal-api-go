// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
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

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: temporal/api/enums/v1/reset.proto

package enums

import (
	fmt "fmt"
	math "math"
	strconv "strconv"

	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ResetReapplyType int32

const (
	RESET_REAPPLY_TYPE_UNSPECIFIED ResetReapplyType = 0
	RESET_REAPPLY_TYPE_ALL         ResetReapplyType = 1
	RESET_REAPPLY_TYPE_NONE        ResetReapplyType = 2
)

var ResetReapplyType_name = map[int32]string{
	0: "Unspecified",
	1: "All",
	2: "None",
}

var ResetReapplyType_value = map[string]int32{
	"Unspecified": 0,
	"All":         1,
	"None":        2,
}

func (ResetReapplyType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0bce7088bd4da0dd, []int{0}
}

func init() {
	proto.RegisterEnum("temporal.api.enums.v1.ResetReapplyType", ResetReapplyType_name, ResetReapplyType_value)
}

func init() { proto.RegisterFile("temporal/api/enums/v1/reset.proto", fileDescriptor_0bce7088bd4da0dd) }

var fileDescriptor_0bce7088bd4da0dd = []byte{
	// 277 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x2c, 0x49, 0xcd, 0x2d,
	0xc8, 0x2f, 0x4a, 0xcc, 0xd1, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f, 0xcd, 0x2b, 0xcd, 0x2d, 0xd6, 0x2f,
	0x33, 0xd4, 0x2f, 0x4a, 0x2d, 0x4e, 0x2d, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x85,
	0x29, 0xd1, 0x4b, 0x2c, 0xc8, 0xd4, 0x03, 0x2b, 0xd1, 0x2b, 0x33, 0xd4, 0xca, 0xe7, 0x12, 0x08,
	0x02, 0xa9, 0x0a, 0x4a, 0x4d, 0x2c, 0x28, 0xc8, 0xa9, 0x0c, 0xa9, 0x2c, 0x48, 0x15, 0x52, 0xe2,
	0x92, 0x0b, 0x72, 0x0d, 0x76, 0x0d, 0x89, 0x0f, 0x72, 0x75, 0x0c, 0x08, 0xf0, 0x89, 0x8c, 0x0f,
	0x89, 0x0c, 0x70, 0x8d, 0x0f, 0xf5, 0x0b, 0x0e, 0x70, 0x75, 0xf6, 0x74, 0xf3, 0x74, 0x75, 0x11,
	0x60, 0x10, 0x92, 0xe2, 0x12, 0xc3, 0xa2, 0xc6, 0xd1, 0xc7, 0x47, 0x80, 0x51, 0x48, 0x9a, 0x4b,
	0x1c, 0x8b, 0x9c, 0x9f, 0xbf, 0x9f, 0xab, 0x00, 0x93, 0xd3, 0x26, 0xc6, 0x0b, 0x0f, 0xe5, 0x18,
	0x6e, 0x3c, 0x94, 0x63, 0xf8, 0xf0, 0x50, 0x8e, 0xb1, 0xe1, 0x91, 0x1c, 0xe3, 0x8a, 0x47, 0x72,
	0x8c, 0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x8b, 0x47,
	0x72, 0x0c, 0x1f, 0x1e, 0xc9, 0x31, 0x4e, 0x78, 0x2c, 0xc7, 0x70, 0xe1, 0xb1, 0x1c, 0xc3, 0x8d,
	0xc7, 0x72, 0x0c, 0x5c, 0x12, 0x99, 0xf9, 0x7a, 0x58, 0x7d, 0xe0, 0xc4, 0x05, 0x76, 0x7f, 0x00,
	0xc8, 0x93, 0x01, 0x8c, 0x51, 0x8a, 0xe9, 0x48, 0xea, 0x32, 0xf3, 0x51, 0xc2, 0xc3, 0x1a, 0xcc,
	0x58, 0xc5, 0x24, 0x1a, 0x02, 0x53, 0xe0, 0x58, 0x90, 0xa9, 0xe7, 0x0a, 0x36, 0x28, 0xcc, 0xf0,
	0x15, 0x93, 0x04, 0x4c, 0xdc, 0xca, 0xca, 0xb1, 0x20, 0xd3, 0xca, 0x0a, 0x2c, 0x63, 0x65, 0x15,
	0x66, 0x98, 0xc4, 0x06, 0x0e, 0x43, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa0, 0xdd, 0xb8,
	0x14, 0x68, 0x01, 0x00, 0x00,
}

func (x ResetReapplyType) String() string {
	s, ok := ResetReapplyType_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
