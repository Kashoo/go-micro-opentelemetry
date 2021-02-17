package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type tracker struct {
	startedAt time.Time
	tracer    trace.Tracer
	span      trace.Span
	method    string
	service   string
}

type requestDescriptor interface {
	Service() string
	Endpoint() string
}

// newRequestTracker creates a new tracker for an RPC request (client or server).
func newRequestTracker(req requestDescriptor, tracer trace.Tracer) *tracker {
	return &tracker{
		tracer:  tracer,
		method:  req.Endpoint(),
		service: req.Service(),
	}
}

type publicationDescriptor interface {
	Topic() string
}

// newEventTracker creates a new tracker for a publication (client or server).
func newEventTracker(pub publicationDescriptor, tracer trace.Tracer) *tracker {
	return &tracker{
		tracer:  tracer,
		method:  pub.Topic(),
		service: "pubsub",
	}
}

// start monitoring a request. You can choose to let this method
// start a span for the request or attach one later.
func (t *tracker) start(ctx context.Context, startSpan bool) context.Context {
	t.startedAt = time.Now()

	if startSpan {
		ctx, t.span = t.tracer.Start(
			ctx,
			fmt.Sprintf("rpc/%s/%s", t.service, t.method),
		)
	}

	return ctx
}

// end a request's monitoring session. If there is a span ongoing, it will
// be ended and metrics will be recorded.
func (t *tracker) end(ctx context.Context, err error) {
	if t.span != nil {
		setResponseStatus(t.span, err, "")
		t.span.End()
	}
}
