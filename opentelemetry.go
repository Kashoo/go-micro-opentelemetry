package opentelemetry

import (
	"context"
	"fmt"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/metadata"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/Kashoo/go-micro-opentelemetry"

func getTraceFromCtx(ctx context.Context, opts ...Option) oteltrace.SpanContext {
	md, ok := metadata.FromContext(ctx)
	// if there is nothing from metadata
	if !ok {
		md = make(metadata.Metadata)
	}
	metadataCopy := metadata.Copy(md)
	_, spanCtx := Extract(ctx, &metadataCopy, opts...)

	return spanCtx
}

func injectTraceIntoCtx(ctx context.Context, opts ...Option) context.Context {
	md, ok := metadata.FromContext(ctx)
	// if there is nothing from metadata
	if !ok {
		md = make(metadata.Metadata)
	}
	metadataCopy := metadata.Copy(md)
	Inject(ctx, &metadataCopy, opts...)
	ctx = metadata.NewContext(ctx, metadataCopy)
	return ctx
}

// clientWrapper wraps an RPC client and adds tracing.
type clientWrapper struct {
	client.Client
	opts   []Option
	tracer oteltrace.Tracer
}

// Call implements client.Client.Call.
func (w *clientWrapper) Call(
	ctx context.Context,
	req client.Request,
	rsp interface{},
	opts ...client.CallOption) (err error) {
	t := newRequestTracker(req, w.tracer)
	ctx = t.start(ctx, true)

	defer func() { t.end(ctx, err) }()

	ctx = injectTraceIntoCtx(ctx, w.opts...)

	err = w.Client.Call(ctx, req, rsp, opts...)
	return
}

// Publish implements client.Client.Publish.
func (w *clientWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) (err error) {
	t := newEventTracker(p, w.tracer)
	ctx = t.start(ctx, true)

	defer func() { t.end(ctx, err) }()

	ctx = injectTraceIntoCtx(ctx, w.opts...)

	err = w.Client.Publish(ctx, p, opts...)
	return
}

// NewClientWrapper returns a client.Wrapper
// that adds monitoring to outgoing requests.
func NewClientWrapper(name string, opts ...Option) client.Wrapper {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(name)
	return func(c client.Client) client.Client {
		return &clientWrapper{c, opts, tracer}
	}
}

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
				oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			}

			t := newRequestTracker(req, tracer)
			ctx = t.start(ctx, false)

			spanName := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			defer func() { t.end(ctx, err) }()
			spanCtx := getTraceFromCtx(ctx)
			if spanCtx.IsValid() {
				logger.Errorf("CALL METHOD VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx),
					spanName,
					topts...,
				)
			} else {
				logger.Errorf("CALL METHOD NOT VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					ctx,
					spanName,
					topts...,
				)
			}

			if err = cf(ctx, node, req, rsp, opts); err != nil {

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
				oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
			}

			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()
			spanName := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			spanCtx := getTraceFromCtx(ctx)
			if spanCtx.IsValid() {
				logger.Errorf("HANDLER METHOD VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx),
					spanName,
					topts...,
				)
			} else {
				logger.Errorf("HANDLER METHOD NOT VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					ctx,
					spanName,
					topts...,
				)
			}
			if err = fn(ctx, req, rsp); err != nil {
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
			spanCtx := getTraceFromCtx(ctx)
			if spanCtx.IsValid() {
				logger.Errorf("SUBSCRIBE METHOD VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx),
					spanName,
					topts...,
				)
			} else {
				logger.Errorf("SUBSCRIBE METHOD NOT VALID, %v", spanCtx)
				ctx, t.span = tracer.Start(
					ctx,
					spanName,
					topts...,
				)
			}

			if err = fn(ctx, p); err != nil {
				t.span.AddEvent(
					spanName)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
