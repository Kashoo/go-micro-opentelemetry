package opentelemetry

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/trace
	"go.opentelemetry.io/otel/label"
	"strings"
)

var tracePropagator propagators.TraceContext

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span.
func StartSpanFromContext(ctx context.Context, tracer trace.Tracer, name string, opts ...trace.SpanOption) (context.Context, trace.Span, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}

	fmt.Println(md)

	// Find parent span.
	// First try to get span within current service boundary.
	// If there doesn't exist, try to get it from go-micro metadata(which is cross boundary)
	//if parentSpan := tracer.SpanFromContext(ctx); parentSpan != nil {
	//	opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	//} else if spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
	//	opts = append(opts, opentracing.ChildOf(spanCtx))
	//}

	if parentSpan := trace.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, trace.WithAttributes(ctx, parentSpan.SpanContext()))
	} else if bag.Extract() {

	}

	// allocate new map with only one element
	nmd := make(metadata.Metadata, 1)

	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)

	ctx, sp := tracer.Start(ctx, name, opts...)

	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, otel.TextMapCarrier(nmd)); err != nil {
		return nil, nil, err
	}
	mp := otel.NewCompositeTextMapPropagator(nmd)

	for k, v := range nmd {
		md.Set(strings.Title(k), v)
	}

	ctx = trace.ContextWithSpan(ctx, sp)
	ctx = metadata.NewContext(ctx, md)
	return ctx, sp, nil
}

func NewCallWrapper(ot trace.TracerProvider) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) (err error) {
			name := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			tracer := ot.Tracer(name)
			t := newRequestTracker(req, tracer)
			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()
			ctx, t.span = tracer.Start(
				ctx,
				name,
			)

			if err = cf(ctx, node, req, rsp, opts); err != nil {
				t.span.AddEvent(
					name,
					label.String("error", err.Error()))
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

func NewHandlerWrapper(ot trace.TracerProvider) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			name := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			tracer := ot.Tracer(name)
			t := newRequestTracker(req, tracer)
			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()

			ctx, t.span = t.tracer.Start(ctx,
				name,
				trace.WithAttributes(commonLabels...))
			)

			if err = fn(ctx, req, rsp); err != nil {
				t.span.AddEvent(
					name,
					)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

// NewSubscriberWrapper accepts an opentelemetry Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(ot trace.TracerProvider) server.SubscriberWrapper {
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, p server.Message) (err error) {
			name := fmt.Sprintf("rpc/pubsub/%s", p.Topic())
			tracer := ot.Tracer(name)
			t := newEventTracker(p, tracer)
			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()

			ctx, t.span = t.tracer.Start(ctx,
				name,
			)

			if err = fn(ctx, p); err != nil {
				t.span.AddEvent(
					name,
					label.String("error", err.Error()))
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
