package opentelemetry

import (
	"context"
	"github.com/asim/go-micro/v3/metadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
)

type metadataSupplier struct {
	metadata *metadata.Metadata
}

func (s *metadataSupplier) Get(key string) string {
	values, valid := s.metadata.Get(key)
	if !valid {
		return ""
	}
	if len(values) == 0 {
		return ""
	}
	return values
}

func (s *metadataSupplier) Set(key string, value string) {
	s.metadata.Set(key, value)
}

// Inject injects correlation context and span context into the go-micro
// metadata object. This function is meant to be used on outgoing
// requests.
func Inject(ctx context.Context, metadata *metadata.Metadata) {
	otel.GetTextMapPropagator().Inject(ctx, &metadataSupplier{
		metadata: metadata,
	})
}

// Extract returns the correlation context and span context that
// another service encoded in the go-micro metadata object with Inject.
// This function is meant to be used on incoming requests.
func Extract(ctx context.Context, metadata *metadata.Metadata) ([]label.KeyValue, trace.SpanContext) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, &metadataSupplier{
		metadata: metadata,
	})
	labelSet := baggage.Set(ctx)
	return (&labelSet).ToSlice(), trace.RemoteSpanContextFromContext(ctx)
}
