package opentelemetry

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/Kashoo/go-micro-opentelemetry"

func NewCallWrapper(servicename string, opts ...Option) client.CallWrapper {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}

	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
		oteltrace.WithInstrumentationVersion(contrib.SemVersion()),
	)
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) (err error) {

			topts := []oteltrace.SpanOption{
				oteltrace.WithAttributes(label.String("service", servicename)),
				//oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(request)...),
				//oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, c.Path(), request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}

			t := newRequestTracker(req, tracer)
			ctx = t.start(ctx, false)

			spanName := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			defer func() { t.end(ctx, err) }()
			ctx, t.span = tracer.Start(
				ctx,
				spanName,
				topts...,
			)

			if err = cf(ctx, node, req, rsp, opts); err != nil {
				sentry.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetExtra("trace", t.span.SpanContext().TraceID)
				})
				t.span.AddEvent(
					spanName,
				)
				t.span.SetAttributes(label.String("error", err.Error()))
			}
			return
		}
	}
}

func NewHandlerWrapper(servicename string, opts ...Option) server.HandlerWrapper {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}

	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
		oteltrace.WithInstrumentationVersion(contrib.SemVersion()),
	)
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {

			t := newRequestTracker(req, tracer)
			topts := []oteltrace.SpanOption{
				oteltrace.WithAttributes(label.String("service", servicename)),
				//oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(request)...),
				//oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, c.Path(), request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}

			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()
			spanName := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			ctx, t.span = t.tracer.Start(ctx,
				spanName,
				topts...,
			)

			if err = fn(ctx, req, rsp); err != nil {
				sentry.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetExtra("trace", t.span.SpanContext().TraceID)
				})
				t.span.AddEvent(
					spanName,
				)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

// NewSubscriberWrapper accepts an opentelemetry Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(servicename string, opts ...Option) server.SubscriberWrapper {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}

	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	tracer := cfg.TracerProvider.Tracer(
		instrumentationName,
		oteltrace.WithInstrumentationVersion(contrib.SemVersion()),
	)
	return func(fn server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, p server.Message) (err error) {

			t := newEventTracker(p, tracer)
			topts := []oteltrace.SpanOption{
				oteltrace.WithAttributes(label.String("service", servicename)),
				//oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(request)...),
				//oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(service, c.Path(), request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()
			spanName := fmt.Sprintf("rpc/pubsub/%s", p.Topic())
			ctx, t.span = t.tracer.Start(ctx,
				spanName,
				topts...,
			)

			if err = fn(ctx, p); err != nil {
				sentry.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetExtra("trace", t.span.SpanContext().TraceID)
				})
				t.span.AddEvent(
					spanName)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
