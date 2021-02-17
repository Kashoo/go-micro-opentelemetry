package opentelemetry

import (
	"errors"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

var (
	empty                   = trace.SpanContext{}
	otTraceIDPadding        = "0000000000000000"
	errInvalidSampledHeader = errors.New("invalid OT Sampled header found")
	errInvalidTraceIDHeader = errors.New("invalid OT traceID header found")
	errInvalidSpanIDHeader  = errors.New("invalid OT spanID header found")
	errInvalidScope         = errors.New("require either both traceID and spanID or none")
)

// extract reconstructs a SpanContext from header values based on OT
// headers.
func extract(traceID, spanID, sampled string) (trace.SpanContext, error) {
	var (
		err           error
		requiredCount int
		sc            = trace.SpanContext{}
	)

	switch strings.ToLower(sampled) {
	case "0", "false":
		// Zero value for TraceFlags sample bit is unset.
	case "1", "true":
		sc.TraceFlags = trace.FlagsSampled
	case "":
		sc.TraceFlags = trace.FlagsDeferred
	default:
		return empty, errInvalidSampledHeader
	}

	if traceID != "" {
		requiredCount++
		id := traceID
		if len(traceID) == 16 {
			// Pad 64-bit trace IDs.
			id = otTraceIDPadding + traceID
		}
		if sc.TraceID, err = trace.TraceIDFromHex(id); err != nil {
			return empty, errInvalidTraceIDHeader
		}
	}

	if spanID != "" {
		requiredCount++
		if sc.SpanID, err = trace.SpanIDFromHex(spanID); err != nil {
			return empty, errInvalidSpanIDHeader
		}
	}

	if requiredCount != 0 && requiredCount != 2 {
		return empty, errInvalidScope
	}

	return sc, nil
}
