package actions

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type operationMetrics struct {
	requests atomic.Int64
	success  atomic.Int64
	failure  atomic.Int64
}

var scheduleOperationMetrics sync.Map

func getOperationMetrics(operation string) *operationMetrics {
	if value, ok := scheduleOperationMetrics.Load(operation); ok {
		return value.(*operationMetrics)
	}
	metrics := &operationMetrics{}
	actual, _ := scheduleOperationMetrics.LoadOrStore(operation, metrics)
	return actual.(*operationMetrics)
}

func recordOperation(operation string, success bool, code int32, startedAt time.Time, err error) {
	metrics := getOperationMetrics(operation)
	requests := metrics.requests.Add(1)
	if success {
		metrics.success.Add(1)
	} else {
		metrics.failure.Add(1)
	}

	elapsedMs := time.Since(startedAt).Milliseconds()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	log.Printf(
		"schedule metric op=%s success=%t code=%d duration_ms=%d requests_total=%d success_total=%d failure_total=%d err=%q",
		operation,
		success,
		code,
		elapsedMs,
		requests,
		metrics.success.Load(),
		metrics.failure.Load(),
		errMsg,
	)
}
