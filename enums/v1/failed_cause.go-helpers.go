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
	WorkflowTaskFailedCause_shortNameValue = map[string]int32{
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
		"BadModifyWorkflowPropertiesAttributes":               25,
		"PendingChildWorkflowsLimitExceeded":                  26,
		"PendingActivitiesLimitExceeded":                      27,
		"PendingSignalsLimitExceeded":                         28,
		"PendingRequestCancelLimitExceeded":                   29,
		"BadUpdateWorkflowExecutionMessage":                   30,
		"UnhandledUpdate":                                     31,
	}
)

// WorkflowTaskFailedCauseFromString parses a WorkflowTaskFailedCause value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to WorkflowTaskFailedCause
func WorkflowTaskFailedCauseFromString(s string) (WorkflowTaskFailedCause, error) {
	if v, ok := WorkflowTaskFailedCause_value[s]; ok {
		return WorkflowTaskFailedCause(v), nil
	} else if v, ok := WorkflowTaskFailedCause_shortNameValue[s]; ok {
		return WorkflowTaskFailedCause(v), nil
	}
	return WorkflowTaskFailedCause(0), fmt.Errorf("%s is not a valid WorkflowTaskFailedCause", s)
}

var (
	StartChildWorkflowExecutionFailedCause_shortNameValue = map[string]int32{
		"Unspecified":           0,
		"WorkflowAlreadyExists": 1,
		"NamespaceNotFound":     2,
	}
)

// StartChildWorkflowExecutionFailedCauseFromString parses a StartChildWorkflowExecutionFailedCause value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to StartChildWorkflowExecutionFailedCause
func StartChildWorkflowExecutionFailedCauseFromString(s string) (StartChildWorkflowExecutionFailedCause, error) {
	if v, ok := StartChildWorkflowExecutionFailedCause_value[s]; ok {
		return StartChildWorkflowExecutionFailedCause(v), nil
	} else if v, ok := StartChildWorkflowExecutionFailedCause_shortNameValue[s]; ok {
		return StartChildWorkflowExecutionFailedCause(v), nil
	}
	return StartChildWorkflowExecutionFailedCause(0), fmt.Errorf("%s is not a valid StartChildWorkflowExecutionFailedCause", s)
}

var (
	CancelExternalWorkflowExecutionFailedCause_shortNameValue = map[string]int32{
		"Unspecified":                       0,
		"ExternalWorkflowExecutionNotFound": 1,
		"NamespaceNotFound":                 2,
	}
)

// CancelExternalWorkflowExecutionFailedCauseFromString parses a CancelExternalWorkflowExecutionFailedCause value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to CancelExternalWorkflowExecutionFailedCause
func CancelExternalWorkflowExecutionFailedCauseFromString(s string) (CancelExternalWorkflowExecutionFailedCause, error) {
	if v, ok := CancelExternalWorkflowExecutionFailedCause_value[s]; ok {
		return CancelExternalWorkflowExecutionFailedCause(v), nil
	} else if v, ok := CancelExternalWorkflowExecutionFailedCause_shortNameValue[s]; ok {
		return CancelExternalWorkflowExecutionFailedCause(v), nil
	}
	return CancelExternalWorkflowExecutionFailedCause(0), fmt.Errorf("%s is not a valid CancelExternalWorkflowExecutionFailedCause", s)
}

var (
	SignalExternalWorkflowExecutionFailedCause_shortNameValue = map[string]int32{
		"Unspecified":                       0,
		"ExternalWorkflowExecutionNotFound": 1,
		"NamespaceNotFound":                 2,
		"SignalCountLimitExceeded":          3,
	}
)

// SignalExternalWorkflowExecutionFailedCauseFromString parses a SignalExternalWorkflowExecutionFailedCause value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to SignalExternalWorkflowExecutionFailedCause
func SignalExternalWorkflowExecutionFailedCauseFromString(s string) (SignalExternalWorkflowExecutionFailedCause, error) {
	if v, ok := SignalExternalWorkflowExecutionFailedCause_value[s]; ok {
		return SignalExternalWorkflowExecutionFailedCause(v), nil
	} else if v, ok := SignalExternalWorkflowExecutionFailedCause_shortNameValue[s]; ok {
		return SignalExternalWorkflowExecutionFailedCause(v), nil
	}
	return SignalExternalWorkflowExecutionFailedCause(0), fmt.Errorf("%s is not a valid SignalExternalWorkflowExecutionFailedCause", s)
}

var (
	ResourceExhaustedCause_shortNameValue = map[string]int32{
		"Unspecified":      0,
		"RpsLimit":         1,
		"ConcurrentLimit":  2,
		"SystemOverloaded": 3,
		"PersistenceLimit": 4,
		"BusyWorkflow":     5,
		"ApsLimit":         6,
	}
)

// ResourceExhaustedCauseFromString parses a ResourceExhaustedCause value from  either the protojson
// canonical SCREAMING_CASE enum or the traditional temporal PascalCase enum to ResourceExhaustedCause
func ResourceExhaustedCauseFromString(s string) (ResourceExhaustedCause, error) {
	if v, ok := ResourceExhaustedCause_value[s]; ok {
		return ResourceExhaustedCause(v), nil
	} else if v, ok := ResourceExhaustedCause_shortNameValue[s]; ok {
		return ResourceExhaustedCause(v), nil
	}
	return ResourceExhaustedCause(0), fmt.Errorf("%s is not a valid ResourceExhaustedCause", s)
}
