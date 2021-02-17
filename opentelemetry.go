package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/baggage"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// Default OT Header names.
	traceIDHeader      = "ot-tracer-traceid"
	spanIDHeader       = "ot-tracer-spanid"
	sampledHeader      = "ot-tracer-sampled"
	traceID64BitsWidth = 64 / 4 // 16 hex character Trace ID.
)

const instrumentationName = "github.com/Kashoo/go-micro-opentelemetry"

func getTraceFromCtx(ctx context.Context, opts ...Option) context.Context {
	var (
		sc  oteltrace.SpanContext
		err error
	)
	md, _ := metadata.FromContext(ctx)
	traceID, _ := md.Get(traceIDHeader)
	spanID, _ := md.Get(spanIDHeader)
	sampled, _ := md.Get(sampledHeader)

	sc, err = extract(traceID, spanID, sampled)
	if err != nil || !sc.IsValid() {
		return ctx
	}
	//_, spanCtx := Extract(ctx, &md, opts...)

	return oteltrace.ContextWithRemoteSpanContext(ctx, sc)
}

func injectTraceIntoCtx(ctx context.Context, opts ...Option) context.Context {
	sc := oteltrace.SpanFromContext(ctx).SpanContext()
	if !sc.TraceID.IsValid() || !sc.SpanID.IsValid() {
		// don't bother injecting anything if either trace or span IDs are not valid
		return ctx
	}
	md := make(metadata.Metadata)
	md.Set(traceIDHeader, sc.TraceID.String()[len(sc.TraceID.String())-traceID64BitsWidth:])
	md.Set(spanIDHeader, sc.SpanID.String())
	if sc.IsSampled() {
		md.Set(sampledHeader, "1")
	} else {
		md.Set(sampledHeader, "0")
	}
	m := baggage.Set(ctx)
	mi := m.Iter()
	for mi.Next() {
		labl := mi.Label()
		md.Set(fmt.Sprintf("ot-baggage-%s", labl.Key), labl.Value.Emit())
	}
	Inject(ctx, &md, opts...)

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

func NewCallWrapper(servicename string, options ...Option) client.CallWrapper {
	cfg := config{}
	for _, opt := range options {
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
			ctx = getTraceFromCtx(ctx, options...)

			ctx, t.span = tracer.Start(
				ctx,
				spanName,
				topts...,
			)

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

func NewHandlerWrapper(servicename string, options ...Option) server.HandlerWrapper {
	cfg := config{}
	for _, opt := range options {
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
			ctx = getTraceFromCtx(ctx, options...)

			ctx, t.span = tracer.Start(
				ctx,
				spanName,
				topts...,
			)

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
func NewSubscriberWrapper(servicename string, options ...Option) server.SubscriberWrapper {
	cfg := config{}
	for _, opt := range options {
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
			ctx = getTraceFromCtx(ctx, options...)

			ctx, t.span = tracer.Start(
				ctx,
				spanName,
				topts...,
			)

			if err = fn(ctx, p); err != nil {
				t.span.AddEvent(
					spanName)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
