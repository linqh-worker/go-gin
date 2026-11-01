package logx

import (
	"context"
	"go-gin/internal/traceid"

	"github.com/rs/zerolog"
)

type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {

	e.Str("module", Module(e.GetCtx()))
	traceId := getTraceIdFromContext(e.GetCtx())
	if traceId != "" {
		e.Str(traceid.TraceIdFieldName, getTraceIdFromContext(e.GetCtx()))
	}
}

func getTraceIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if trace_id, ok := ctx.Value(traceid.TraceIdFieldName).(string); ok {
		return trace_id
	}
	return ""
}
