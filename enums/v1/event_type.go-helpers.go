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

package enums

import (
	"fmt"
)

var (
	EventType_shortNameValue = map[string]int32{
		"Unspecified":                                     0,
		"WorkflowExecutionStarted":                        1,
		"WorkflowExecutionCompleted":                      2,
		"WorkflowExecutionFailed":                         3,
		"WorkflowExecutionTimedOut":                       4,
		"WorkflowTaskScheduled":                           5,
		"WorkflowTaskStarted":                             6,
		"WorkflowTaskCompleted":                           7,
		"WorkflowTaskTimedOut":                            8,
		"WorkflowTaskFailed":                              9,
		"ActivityTaskScheduled":                           10,
		"ActivityTaskStarted":                             11,
		"ActivityTaskCompleted":                           12,
		"ActivityTaskFailed":                              13,
		"ActivityTaskTimedOut":                            14,
		"ActivityTaskCancelRequested":                     15,
		"ActivityTaskCanceled":                            16,
		"TimerStarted":                                    17,
		"TimerFired":                                      18,
		"TimerCanceled":                                   19,
		"WorkflowExecutionCancelRequested":                20,
		"WorkflowExecutionCanceled":                       21,
		"RequestCancelExternalWorkflowExecutionInitiated": 22,
		"RequestCancelExternalWorkflowExecutionFailed":    23,
		"ExternalWorkflowExecutionCancelRequested":        24,
		"MarkerRecorded":                                  25,
		"WorkflowExecutionSignaled":                       26,
		"WorkflowExecutionTerminated":                     27,
		"WorkflowExecutionContinuedAsNew":                 28,
		"StartChildWorkflowExecutionInitiated":            29,
		"StartChildWorkflowExecutionFailed":               30,
		"ChildWorkflowExecutionStarted":                   31,
		"ChildWorkflowExecutionCompleted":                 32,
		"ChildWorkflowExecutionFailed":                    33,
		"ChildWorkflowExecutionCanceled":                  34,
		"ChildWorkflowExecutionTimedOut":                  35,
		"ChildWorkflowExecutionTerminated":                36,
		"SignalExternalWorkflowExecutionInitiated":        37,
		"SignalExternalWorkflowExecutionFailed":           38,
		"ExternalWorkflowExecutionSignaled":               39,
		"UpsertWorkflowSearchAttributes":                  40,
		"WorkflowExecutionUpdateAccepted":                 41,
		"WorkflowExecutionUpdateRejected":                 42,
		"WorkflowExecutionUpdateCompleted":                43,
		"WorkflowPropertiesModifiedExternally":            44,
		"ActivityPropertiesModifiedExternally":            45,
		"WorkflowPropertiesModified":                      46,
	}
)

// EventTypeFromString parses a EventType value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to EventType
func EventTypeFromString(s string) (EventType, error) {
	if v, ok := EventType_value[s]; ok {
		return EventType(v), nil
	} else if v, ok := EventType_shortNameValue[s]; ok {
		return EventType(v), nil
	}
	return EventType(0), fmt.Errorf("%s is not a valid EventType", s)
}
