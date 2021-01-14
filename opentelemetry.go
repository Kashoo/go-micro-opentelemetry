package opentelemetry

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// TracePropagationField is the key for the tracing context
	// that will be injected in go-micro's metadata.
	TracePropagationField = "X-Trace-Context"
	SpanPropagationField  = "X-Span-Context"
)


//// StartSpanFromContext returns a new span with the given operation name and options. If a span
//// is found in the context, it will be used as the parent of the resulting span.
//func StartSpanFromContext(ctx context.Context, tracer trace.Tracer, name string, opts ...trace.SpanOption) (context.Context, trace.Span, error) {
//	md, ok := metadata.FromContext(ctx)
//	if !ok {
//		md = make(metadata.Metadata)
//	}
//
//	// Find parent span.
//	// First try to get span within current service boundary.
//	// If there doesn't exist, try to get it from go-micro metadata(which is cross boundary)
//	if parentSpan := trace.SpanFromContext(ctx); parentSpan != nil {
//		opts = append(opts, trace.SpanOption(parentSpan))
//		//opts = append(opts, opentracing.ChildOf(parentSpan))
//	} else if spanCtx, err := trace.ChildOf(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
//		opts = append(opts, apitrace.ChildOf(spanCtx))
//	}
//
//	// allocate new map with only one element
//	nmd := make(metadata.Metadata, 1)
//
//	ctx, sp := tracer.Start(ctx, name, opts...)
//
//	//if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(nmd)); err != nil {
//	//	return nil, nil, err
//	//}
//	propagation
//
//	propagator := otel.GetTextMapPropagator()
//	propagator.Inject(ctx,nmd)
//
//	for k, v := range nmd {
//		md.Set(strings.Title(k), v)
//	}
//
//	ctx = trace.ContextWithSpan(ctx, sp)
//	ctx = metadata.NewContext(ctx, md)
//	return ctx, sp, nil
//}

//func getTraceSpanCtx(ctx context.Context) *trace.SpanContext {
//	spanID, _ := metadata.Get(ctx, SpanPropagationField)
//	traceID, _ := metadata.Get(ctx, TracePropagationField)
//
//	if len(traceID) > 0 && len(spanID) > 0 {
//		return &trace.SpanContext{
//			TraceID: stringToTraceID(traceID),
//			SpanID:  stringToSpanID(spanID),
//		}
//	}
//
//	return nil
//}

//func stringToTraceID(str string) trace.TraceID {
//	data, _ := hex.DecodeString(str)
//	var b [16]byte = [16]byte{}
//	for k, v := range data[:16] {
//		b[k] = v
//	}
//	return b
//}

//func stringToSpanID(str string) trace.SpanID {
//	data, _ := hex.DecodeString(str)
//	var b [8]byte = [8]byte{}
//	for k, v := range data[:8] {
//		b[k] = v
//	}
//	return b
//}

//func (o *otWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
//	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
//	ctx, span, err := StartSpanFromContext(ctx, o.ot, name)
//	if err != nil {
//		return err
//	}
//	defer span.End()
//	if err = o.Client.Call(ctx, req, rsp, opts...); err != nil {
//		span.AddEvent(
//			"",
//			trace.WithAttributes(label.String("error", err.Error())))
//
//		span.SetAttributes(label.String("error", err.Error()))
//	}
//	return err
//}
//
//func (o *otWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
//	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
//	ctx, span, err := StartSpanFromContext(ctx, o.ot, name)
//	if err != nil {
//		return nil, err
//	}
//	defer span.End()
//	stream, err := o.Client.Stream(ctx, req, opts...)
//	if err != nil {
//		span.AddEvent(
//			"",
//			trace.WithAttributes(label.String("error", err.Error())))
//		span.SetAttributes(label.String("error", err.Error()))
//	}
//	return stream, err
//}

//func (o *otWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
//	name := fmt.Sprintf("Pub to %s", p.Topic())
//	ctx, span, err := StartSpanFromContext(ctx, o.ot, name)
//	if err != nil {
//		return err
//	}
//	defer span.End()
//	if err = o.Client.Publish(ctx, p, opts...); err != nil {
//		span.AddEvent(
//			"",
//			trace.WithAttributes(label.String("error", err.Error())))
//		span.SetAttributes(label.String("error", err.Error()))
//	}
//	return err
//}

// NewClientWrapper accepts an open tracing Trace and returns a Client Wrapper
//func NewClientWrapper(ot trace.Tracer, name string) client.Wrapper {
//	return func(c client.Client) client.Client {
//		if ot == nil {
//			ot = otel.Tracer(name)
//		}
//		return &otWrapper{ot, c}
//	}
//}

// NewCallWrapper accepts an opentelemetry Tracer and returns a Call Wrapper
//func NewCallWrapper(ot trace.Tracer, name string) client.CallWrapper {
//	return func(cf client.CallFunc) client.CallFunc {
//		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
//			if ot == nil {
//				ot = otel.Tracer(name)
//			}
//			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
//			ctx, span, err := StartSpanFromContext(ctx, ot, name)
//			if err != nil {
//				return err
//			}
//			defer span.End()
//			if err = cf(ctx, node, req, rsp, opts); err != nil {
//				span.AddEvent(
//					"",
//					trace.WithAttributes(label.String("error", err.Error())))
//				span.SetAttributes(label.String("error", err.Error()))
//
//			}
//			return err
//		}
//	}
//}
func NewCallWrapper(ot *sdktrace.TracerProvider) client.CallWrapper {
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
					ctx,
					name,
					label.String("error", err.Error()))
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

func NewHandlerWrapper(ot *sdktrace.TracerProvider) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			name := fmt.Sprintf("rpc/%s/%s", req.Service(), req.Endpoint())
			tracer := ot.Tracer(name)
			t := newRequestTracker(req, tracer)
			ctx = t.start(ctx, false)
			defer func() { t.end(ctx, err) }()

			ctx, t.span = t.tracer.Start(ctx,
				name,
			)

			if err = fn(ctx, req, rsp); err != nil {
				t.span.AddEvent(
					ctx,
					name,
					label.String("error", err.Error()))
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

// NewSubscriberWrapper accepts an opentelemetry Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(ot *sdktrace.TracerProvider) server.SubscriberWrapper {
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
					ctx,
					name,
					label.String("error", err.Error()))
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
