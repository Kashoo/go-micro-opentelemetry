package opentelemetry

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"go.opentelemetry.io/otel/label"
	apitracer "go.opentelemetry.io/otel/trace"
)

func NewCallWrapper(ot apitracer.TracerProvider) client.CallWrapper {
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
				)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

func NewHandlerWrapper(ot apitracer.TracerProvider) server.HandlerWrapper {
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
					name,
				)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}

// NewSubscriberWrapper accepts an opentelemetry Tracer and returns a Subscriber Wrapper
func NewSubscriberWrapper(ot apitracer.TracerProvider) server.SubscriberWrapper {
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
					name)
				t.span.SetAttributes(label.String("error", err.Error()))

			}
			return
		}
	}
}
