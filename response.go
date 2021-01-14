package opentelemetry

import (
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
)

var ErrorKey = label.Key("error")

func setResponseStatus(span trace.Span, err error, msg string) {
	if err == nil {
		span.SetStatus(codes.Ok, msg)
		return
	}

	span.SetStatus(codes.Error, msg)
	span.SetAttributes(ErrorKey.String(err.Error()))
}
