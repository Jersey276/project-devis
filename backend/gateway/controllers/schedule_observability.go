package controllers

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type scheduleHTTPMetrics struct {
	requests atomic.Int64
	success  atomic.Int64
	failure  atomic.Int64
}

var scheduleHTTPMetricsByOperation sync.Map

func getScheduleHTTPMetrics(operation string) *scheduleHTTPMetrics {
	if value, ok := scheduleHTTPMetricsByOperation.Load(operation); ok {
		return value.(*scheduleHTTPMetrics)
	}
	metrics := &scheduleHTTPMetrics{}
	actual, _ := scheduleHTTPMetricsByOperation.LoadOrStore(operation, metrics)
	return actual.(*scheduleHTTPMetrics)
}

func recordScheduleHTTP(operation string, success bool, grpcCode int32, startedAt time.Time) {
	metrics := getScheduleHTTPMetrics(operation)
	requests := metrics.requests.Add(1)
	if success {
		metrics.success.Add(1)
	} else {
		metrics.failure.Add(1)
	}

	log.Printf(
		"gateway schedule metric op=%s success=%t grpc_code=%d duration_ms=%d requests_total=%d success_total=%d failure_total=%d",
		operation,
		success,
		grpcCode,
		time.Since(startedAt).Milliseconds(),
		requests,
		metrics.success.Load(),
		metrics.failure.Load(),
	)
}
