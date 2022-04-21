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
// source: temporal/api/enums/v1/failed_cause.proto

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

// Workflow tasks can fail for various reasons. Note that some of these reasons can only originate
// from the server, and some of them can only originate from the SDK/worker.
type WorkflowTaskFailedCause int32

const (
	WORKFLOW_TASK_FAILED_CAUSE_UNSPECIFIED WorkflowTaskFailedCause = 0
	// Between starting and completing the workflow task (with a workflow completion command), some
	// new command (like a signal) was processed into workflow history. The outstanding task will be
	// failed with this reason, and a worker must pick up a new task.
	WORKFLOW_TASK_FAILED_CAUSE_UNHANDLED_COMMAND                                         WorkflowTaskFailedCause = 1
	WORKFLOW_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES                          WorkflowTaskFailedCause = 2
	WORKFLOW_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES                    WorkflowTaskFailedCause = 3
	WORKFLOW_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES                                WorkflowTaskFailedCause = 4
	WORKFLOW_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES                               WorkflowTaskFailedCause = 5
	WORKFLOW_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES                              WorkflowTaskFailedCause = 6
	WORKFLOW_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES                WorkflowTaskFailedCause = 7
	WORKFLOW_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES                    WorkflowTaskFailedCause = 8
	WORKFLOW_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES                  WorkflowTaskFailedCause = 9
	WORKFLOW_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES WorkflowTaskFailedCause = 10
	WORKFLOW_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES                            WorkflowTaskFailedCause = 11
	WORKFLOW_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID                                  WorkflowTaskFailedCause = 12
	// The worker wishes to fail the task and have the next one be generated on a normal, not sticky
	// queue. Generally workers should prefer to use the explicit `ResetStickyTaskQueue` RPC call.
	WORKFLOW_TASK_FAILED_CAUSE_RESET_STICKY_TASK_QUEUE                  WorkflowTaskFailedCause = 13
	WORKFLOW_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE        WorkflowTaskFailedCause = 14
	WORKFLOW_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES WorkflowTaskFailedCause = 15
	WORKFLOW_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES     WorkflowTaskFailedCause = 16
	WORKFLOW_TASK_FAILED_CAUSE_FORCE_CLOSE_COMMAND                      WorkflowTaskFailedCause = 17
	WORKFLOW_TASK_FAILED_CAUSE_FAILOVER_CLOSE_COMMAND                   WorkflowTaskFailedCause = 18
	WORKFLOW_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE                    WorkflowTaskFailedCause = 19
	WORKFLOW_TASK_FAILED_CAUSE_RESET_WORKFLOW                           WorkflowTaskFailedCause = 20
	WORKFLOW_TASK_FAILED_CAUSE_BAD_BINARY                               WorkflowTaskFailedCause = 21
	WORKFLOW_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID           WorkflowTaskFailedCause = 22
	WORKFLOW_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES                    WorkflowTaskFailedCause = 23
	// The worker encountered a mismatch while replaying history between what was expected, and
	// what the workflow code actually did.
	WORKFLOW_TASK_FAILED_CAUSE_NON_DETERMINISTIC_ERROR WorkflowTaskFailedCause = 24
)

var WorkflowTaskFailedCause_name = map[int32]string{
	0:  "Unspecified",
	1:  "UnhandledCommand",
	2:  "BadScheduleActivityAttributes",
	3:  "BadRequestCancelActivityAttributes",
	4:  "BadStartTimerAttributes",
	5:  "BadCancelTimerAttributes",
	6:  "BadRecordMarkerAttributes",
	7:  "BadCompleteWorkflowExecutionAttributes",
	8:  "BadFailWorkflowExecutionAttributes",
	9:  "BadCancelWorkflowExecutionAttributes",
	10: "BadRequestCancelExternalWorkflowExecutionAttributes",
	11: "BadContinueAsNewAttributes",
	12: "StartTimerDuplicateId",
	13: "ResetStickyTaskQueue",
	14: "WorkflowWorkerUnhandledFailure",
	15: "BadSignalWorkflowExecutionAttributes",
	16: "BadStartChildExecutionAttributes",
	17: "ForceCloseCommand",
	18: "FailoverCloseCommand",
	19: "BadSignalInputSize",
	20: "ResetWorkflow",
	21: "BadBinary",
	22: "ScheduleActivityDuplicateId",
	23: "BadSearchAttributes",
	24: "NonDeterministicError",
}

var WorkflowTaskFailedCause_value = map[string]int32{
	"Unspecified":                                         0,
	"UnhandledCommand":                                    1,
	"BadScheduleActivityAttributes":                       2,
	"BadRequestCancelActivityAttributes":                  3,
	"BadStartTimerAttributes":                             4,
	"BadCancelTimerAttributes":                            5,
	"BadRecordMarkerAttributes":                           6,
	"BadCompleteWorkflowExecutionAttributes":              7,
	"BadFailWorkflowExecutionAttributes":                  8,
	"BadCancelWorkflowExecutionAttributes":                9,
	"BadRequestCancelExternalWorkflowExecutionAttributes": 10,
	"BadContinueAsNewAttributes":                          11,
	"StartTimerDuplicateId":                               12,
	"ResetStickyTaskQueue":                                13,
	"WorkflowWorkerUnhandledFailure":                      14,
	"BadSignalWorkflowExecutionAttributes":                15,
	"BadStartChildExecutionAttributes":                    16,
	"ForceCloseCommand":                                   17,
	"FailoverCloseCommand":                                18,
	"BadSignalInputSize":                                  19,
	"ResetWorkflow":                                       20,
	"BadBinary":                                           21,
	"ScheduleActivityDuplicateId":                         22,
	"BadSearchAttributes":                                 23,
	"NonDeterministicError":                               24,
}

func (WorkflowTaskFailedCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b293cf8d1d965f2d, []int{0}
}

type StartChildWorkflowExecutionFailedCause int32

const (
	START_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_UNSPECIFIED             StartChildWorkflowExecutionFailedCause = 0
	START_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_EXISTS StartChildWorkflowExecutionFailedCause = 1
	START_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_NAMESPACE_NOT_FOUND     StartChildWorkflowExecutionFailedCause = 2
)

var StartChildWorkflowExecutionFailedCause_name = map[int32]string{
	0: "Unspecified",
	1: "WorkflowAlreadyExists",
	2: "NamespaceNotFound",
}

var StartChildWorkflowExecutionFailedCause_value = map[string]int32{
	"Unspecified":           0,
	"WorkflowAlreadyExists": 1,
	"NamespaceNotFound":     2,
}

func (StartChildWorkflowExecutionFailedCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b293cf8d1d965f2d, []int{1}
}

type CancelExternalWorkflowExecutionFailedCause int32

const (
	CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNSPECIFIED                           CancelExternalWorkflowExecutionFailedCause = 0
	CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_EXTERNAL_WORKFLOW_EXECUTION_NOT_FOUND CancelExternalWorkflowExecutionFailedCause = 1
	CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_NAMESPACE_NOT_FOUND                   CancelExternalWorkflowExecutionFailedCause = 2
)

var CancelExternalWorkflowExecutionFailedCause_name = map[int32]string{
	0: "Unspecified",
	1: "ExternalWorkflowExecutionNotFound",
	2: "NamespaceNotFound",
}

var CancelExternalWorkflowExecutionFailedCause_value = map[string]int32{
	"Unspecified":                       0,
	"ExternalWorkflowExecutionNotFound": 1,
	"NamespaceNotFound":                 2,
}

func (CancelExternalWorkflowExecutionFailedCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b293cf8d1d965f2d, []int{2}
}

type SignalExternalWorkflowExecutionFailedCause int32

const (
	SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNSPECIFIED                           SignalExternalWorkflowExecutionFailedCause = 0
	SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_EXTERNAL_WORKFLOW_EXECUTION_NOT_FOUND SignalExternalWorkflowExecutionFailedCause = 1
	SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_NAMESPACE_NOT_FOUND                   SignalExternalWorkflowExecutionFailedCause = 2
)

var SignalExternalWorkflowExecutionFailedCause_name = map[int32]string{
	0: "Unspecified",
	1: "ExternalWorkflowExecutionNotFound",
	2: "NamespaceNotFound",
}

var SignalExternalWorkflowExecutionFailedCause_value = map[string]int32{
	"Unspecified":                       0,
	"ExternalWorkflowExecutionNotFound": 1,
	"NamespaceNotFound":                 2,
}

func (SignalExternalWorkflowExecutionFailedCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b293cf8d1d965f2d, []int{3}
}

type ResourceExhaustedCause int32

const (
	RESOURCE_EXHAUSTED_CAUSE_UNSPECIFIED ResourceExhaustedCause = 0
	// Caller exceeds request per second limit.
	RESOURCE_EXHAUSTED_CAUSE_RPS_LIMIT ResourceExhaustedCause = 1
	// Caller exceeds max concurrent request limit.
	RESOURCE_EXHAUSTED_CAUSE_CONCURRENT_LIMIT ResourceExhaustedCause = 2
	// System overloaded.
	RESOURCE_EXHAUSTED_CAUSE_SYSTEM_OVERLOADED ResourceExhaustedCause = 3
)

var ResourceExhaustedCause_name = map[int32]string{
	0: "Unspecified",
	1: "RpsLimit",
	2: "ConcurrentLimit",
	3: "SystemOverloaded",
}

var ResourceExhaustedCause_value = map[string]int32{
	"Unspecified":      0,
	"RpsLimit":         1,
	"ConcurrentLimit":  2,
	"SystemOverloaded": 3,
}

func (ResourceExhaustedCause) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b293cf8d1d965f2d, []int{4}
}

func init() {
	proto.RegisterEnum("temporal.api.enums.v1.WorkflowTaskFailedCause", WorkflowTaskFailedCause_name, WorkflowTaskFailedCause_value)
	proto.RegisterEnum("temporal.api.enums.v1.StartChildWorkflowExecutionFailedCause", StartChildWorkflowExecutionFailedCause_name, StartChildWorkflowExecutionFailedCause_value)
	proto.RegisterEnum("temporal.api.enums.v1.CancelExternalWorkflowExecutionFailedCause", CancelExternalWorkflowExecutionFailedCause_name, CancelExternalWorkflowExecutionFailedCause_value)
	proto.RegisterEnum("temporal.api.enums.v1.SignalExternalWorkflowExecutionFailedCause", SignalExternalWorkflowExecutionFailedCause_name, SignalExternalWorkflowExecutionFailedCause_value)
	proto.RegisterEnum("temporal.api.enums.v1.ResourceExhaustedCause", ResourceExhaustedCause_name, ResourceExhaustedCause_value)
}

func init() {
	proto.RegisterFile("temporal/api/enums/v1/failed_cause.proto", fileDescriptor_b293cf8d1d965f2d)
}

var fileDescriptor_b293cf8d1d965f2d = []byte{
	// 890 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x56, 0xcf, 0x72, 0xdb, 0x44,
	0x1c, 0xb6, 0x5c, 0x28, 0xb0, 0x14, 0x10, 0x0b, 0x69, 0x73, 0xd2, 0x0c, 0x0c, 0x64, 0x5a, 0x0f,
	0x38, 0x75, 0x4b, 0x9b, 0xa9, 0x0d, 0x13, 0xd6, 0xab, 0x9f, 0xe3, 0x9d, 0x48, 0x2b, 0x77, 0x77,
	0x95, 0xd8, 0xbd, 0xec, 0x88, 0x54, 0x6d, 0x35, 0x75, 0x2d, 0x8f, 0xff, 0x94, 0x1c, 0x79, 0x84,
	0xf2, 0x16, 0x0c, 0x4f, 0xc0, 0x23, 0xc0, 0x2d, 0xc7, 0x1e, 0x89, 0x73, 0x61, 0x38, 0x75, 0x86,
	0x17, 0x60, 0x64, 0x62, 0xa3, 0x50, 0x5b, 0xb2, 0xb9, 0x49, 0xda, 0xef, 0xfb, 0x56, 0xbf, 0x6f,
	0xbf, 0xfd, 0xed, 0xa2, 0xeb, 0xa3, 0xf0, 0x59, 0x3f, 0x1e, 0x04, 0xdd, 0xed, 0xa0, 0x1f, 0x6d,
	0x87, 0xbd, 0xf1, 0xb3, 0xe1, 0xf6, 0xf3, 0xca, 0xf6, 0xa3, 0x20, 0xea, 0x86, 0x0f, 0xf5, 0x51,
	0x30, 0x1e, 0x86, 0xe5, 0xfe, 0x20, 0x1e, 0xc5, 0x78, 0x63, 0x86, 0x2c, 0x07, 0xfd, 0xa8, 0x3c,
	0x45, 0x96, 0x9f, 0x57, 0x4a, 0x2f, 0xae, 0xa0, 0x6b, 0x87, 0xf1, 0xe0, 0xe9, 0xa3, 0x6e, 0xfc,
	0xbd, 0x0a, 0x86, 0x4f, 0x1b, 0x53, 0x26, 0x4d, 0x88, 0xb8, 0x84, 0xb6, 0x0e, 0x3d, 0xb1, 0xdf,
	0x70, 0xbc, 0x43, 0xad, 0x88, 0xdc, 0xd7, 0x0d, 0xc2, 0x1c, 0xb0, 0x35, 0x25, 0xbe, 0x04, 0xed,
	0x73, 0xd9, 0x02, 0xca, 0x1a, 0x0c, 0x6c, 0xb3, 0x80, 0x6f, 0xa2, 0x2f, 0x32, 0xb1, 0x4d, 0xc2,
	0xed, 0xe9, 0xbb, 0xe7, 0xba, 0x84, 0xdb, 0xa6, 0x81, 0x77, 0x51, 0x2d, 0x83, 0x51, 0x27, 0xb6,
	0x96, 0xb4, 0x09, 0xb6, 0xef, 0x80, 0x26, 0x54, 0xb1, 0x03, 0xa6, 0x3a, 0x9a, 0x28, 0x25, 0x58,
	0xdd, 0x57, 0x20, 0xcd, 0x22, 0x06, 0x44, 0x72, 0x04, 0x04, 0xdc, 0xf7, 0x41, 0x2a, 0x4d, 0x09,
	0xa7, 0xe0, 0x2c, 0x94, 0xb9, 0x84, 0xef, 0xa1, 0x3b, 0x79, 0xff, 0xa1, 0x88, 0x50, 0x5a, 0x31,
	0x17, 0x44, 0x9a, 0xfa, 0x06, 0xae, 0xa2, 0xbb, 0x39, 0xd4, 0xf3, 0x99, 0x5f, 0xe3, 0xbe, 0x89,
	0x6b, 0x68, 0x27, 0xf7, 0xef, 0xa9, 0x27, 0x6c, 0xed, 0x12, 0xb1, 0x7f, 0x91, 0x7c, 0x19, 0x33,
	0x04, 0x79, 0x13, 0x7b, 0x6e, 0xcb, 0x01, 0x05, 0x7a, 0x8e, 0x83, 0x36, 0x50, 0x5f, 0x31, 0x8f,
	0xa7, 0xa5, 0xde, 0x5a, 0xc1, 0xc5, 0xe4, 0x43, 0x8e, 0xcc, 0xdb, 0x78, 0x0f, 0xd1, 0xd5, 0xac,
	0xc8, 0x16, 0x7a, 0x07, 0xb7, 0x91, 0x5a, 0x6f, 0x55, 0xa1, 0xad, 0x40, 0x70, 0x92, 0xa7, 0x8c,
	0xf0, 0x37, 0xe8, 0x5e, 0xae, 0x69, 0x5c, 0x31, 0xee, 0x83, 0x26, 0x52, 0x73, 0x38, 0x4c, 0xd3,
	0xdf, 0xc5, 0x3b, 0xe8, 0x76, 0x06, 0x3d, 0x9d, 0x11, 0xdb, 0x6f, 0x39, 0x8c, 0x12, 0x05, 0x9a,
	0xd9, 0xe6, 0x15, 0x7c, 0x17, 0xdd, 0xca, 0x20, 0x0a, 0x90, 0xa0, 0xb4, 0x54, 0x8c, 0xee, 0x77,
	0xfe, 0x19, 0xbe, 0xef, 0x83, 0x0f, 0xe6, 0x7b, 0xf8, 0x5b, 0xf4, 0x75, 0x06, 0x6f, 0x3e, 0x94,
	0x3c, 0x80, 0x48, 0x6d, 0xb1, 0x04, 0xe6, 0x0b, 0x30, 0xdf, 0x5f, 0x61, 0x51, 0x24, 0xdb, 0xcb,
	0xb7, 0xee, 0x03, 0x4c, 0xd1, 0xee, 0x4a, 0x7b, 0x84, 0x36, 0x99, 0x63, 0x2f, 0x16, 0x31, 0xf1,
	0x2d, 0x54, 0xce, 0x10, 0x69, 0x78, 0x82, 0x82, 0xa6, 0x8e, 0x27, 0x61, 0xde, 0x24, 0x3e, 0xc4,
	0x77, 0x50, 0x25, 0x8b, 0x43, 0x98, 0xe3, 0x1d, 0x80, 0xf8, 0x0f, 0x0d, 0xe3, 0xaf, 0xd0, 0xcd,
	0xd5, 0x0a, 0x67, 0xbc, 0xe5, 0x2b, 0x2d, 0xd9, 0x03, 0x30, 0x3f, 0xc2, 0x5f, 0xa2, 0x1b, 0xb9,
	0x0b, 0x35, 0x03, 0x98, 0x1f, 0xe3, 0x1b, 0xe8, 0xf3, 0x9c, 0x49, 0xea, 0x8c, 0x13, 0xd1, 0x31,
	0x37, 0x72, 0xa2, 0xf7, 0x7a, 0x9f, 0xbb, 0x90, 0xa0, 0xab, 0xab, 0x94, 0x03, 0x44, 0xd0, 0x66,
	0xda, 0xef, 0x6b, 0x39, 0xb9, 0xe3, 0x1e, 0xd7, 0x36, 0x28, 0x10, 0x2e, 0xe3, 0x2c, 0x89, 0x9f,
	0x06, 0x21, 0x3c, 0x61, 0x6e, 0x96, 0xfe, 0x32, 0xd0, 0x96, 0x1c, 0x05, 0x83, 0x11, 0x7d, 0x12,
	0x75, 0x1f, 0xce, 0x0e, 0x07, 0x38, 0x0e, 0x8f, 0xc6, 0xa3, 0x28, 0xee, 0xa5, 0x4f, 0x88, 0x1a,
	0xda, 0x49, 0x2f, 0xfc, 0x82, 0x18, 0x65, 0x1c, 0x19, 0x7b, 0x88, 0xae, 0x43, 0x9e, 0x8f, 0x13,
	0x47, 0x00, 0xb1, 0x3b, 0x1a, 0xda, 0x4c, 0x2a, 0x69, 0x1a, 0x49, 0x3a, 0xd7, 0x11, 0xe2, 0xc4,
	0x05, 0xd9, 0x22, 0x34, 0xf1, 0x40, 0xe9, 0x86, 0xe7, 0x73, 0xdb, 0x2c, 0x96, 0x7e, 0x2c, 0xa2,
	0x12, 0x0d, 0x7a, 0x47, 0x61, 0x17, 0x8e, 0x47, 0xe1, 0xa0, 0x17, 0x74, 0x33, 0x2b, 0xdf, 0x45,
	0xb5, 0x15, 0xfa, 0x4f, 0x46, 0xf5, 0x1d, 0xe4, 0xaf, 0x2b, 0x90, 0x05, 0xfc, 0xb7, 0x14, 0x23,
	0x31, 0x76, 0x5d, 0xe9, 0xe5, 0x9e, 0xc8, 0xe8, 0x71, 0x2f, 0x58, 0xd9, 0x93, 0xf3, 0x6d, 0xf5,
	0xff, 0x3d, 0x59, 0x57, 0x60, 0x0d, 0x4f, 0xd6, 0x95, 0x5e, 0xec, 0xc9, 0x6f, 0x06, 0xba, 0x2a,
	0xc2, 0x61, 0x3c, 0x1e, 0x1c, 0x85, 0x70, 0xfc, 0x24, 0x18, 0x0f, 0x47, 0xb3, 0xfa, 0xaf, 0xa3,
	0xcf, 0x04, 0x48, 0xcf, 0x4f, 0x1a, 0x19, 0xb4, 0x9b, 0xc4, 0x97, 0x6a, 0x49, 0xa1, 0x5b, 0xe8,
	0xd3, 0xa5, 0x48, 0xd1, 0x92, 0xda, 0x61, 0x2e, 0x53, 0xa6, 0x91, 0x74, 0xa4, 0xa5, 0x38, 0xea,
	0x71, 0xea, 0x0b, 0x01, 0x5c, 0x9d, 0xc3, 0x8b, 0xb8, 0x8c, 0x4a, 0x4b, 0xe1, 0xb2, 0x23, 0x15,
	0xb8, 0x3a, 0x69, 0x97, 0x8e, 0x47, 0x6c, 0xb0, 0xcd, 0x4b, 0xf5, 0x5f, 0x8c, 0x93, 0x53, 0xab,
	0xf0, 0xf2, 0xd4, 0x2a, 0xbc, 0x3a, 0xb5, 0x8c, 0x1f, 0x26, 0x96, 0xf1, 0xd3, 0xc4, 0x32, 0x7e,
	0x9d, 0x58, 0xc6, 0xc9, 0xc4, 0x32, 0x7e, 0x9f, 0x58, 0xc6, 0x1f, 0x13, 0xab, 0xf0, 0x6a, 0x62,
	0x19, 0x2f, 0xce, 0xac, 0xc2, 0xc9, 0x99, 0x55, 0x78, 0x79, 0x66, 0x15, 0xd0, 0x66, 0x14, 0x97,
	0x17, 0xde, 0x26, 0xeb, 0x66, 0x2a, 0x0e, 0xad, 0xe4, 0xda, 0xd9, 0x32, 0x1e, 0x7c, 0xf2, 0x38,
	0x85, 0x8e, 0xe2, 0x0b, 0x17, 0xd5, 0xda, 0xf4, 0xe1, 0xe7, 0xe2, 0x86, 0x9a, 0x01, 0x48, 0x3f,
	0x2a, 0xc3, 0x54, 0xee, 0xa0, 0xf2, 0x67, 0x71, 0x73, 0xf6, 0xbd, 0x5a, 0x25, 0xfd, 0xa8, 0x5a,
	0x9d, 0x8e, 0x54, 0xab, 0x07, 0x95, 0xef, 0x2e, 0x4f, 0x6f, 0xb5, 0xb7, 0xff, 0x0e, 0x00, 0x00,
	0xff, 0xff, 0xcc, 0x52, 0xa6, 0xea, 0x01, 0x0b, 0x00, 0x00,
}

func (x WorkflowTaskFailedCause) String() string {
	s, ok := WorkflowTaskFailedCause_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (x StartChildWorkflowExecutionFailedCause) String() string {
	s, ok := StartChildWorkflowExecutionFailedCause_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (x CancelExternalWorkflowExecutionFailedCause) String() string {
	s, ok := CancelExternalWorkflowExecutionFailedCause_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (x SignalExternalWorkflowExecutionFailedCause) String() string {
	s, ok := SignalExternalWorkflowExecutionFailedCause_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (x ResourceExhaustedCause) String() string {
	s, ok := ResourceExhaustedCause_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
