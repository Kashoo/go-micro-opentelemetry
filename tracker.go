package opentelemetry

import (
	"context"
	"fmt"
	apitracer "go.opentelemetry.io/otel/api/trace"
	"time"
)

type tracker struct {
	startedAt time.Time
	tracer apitracer.Tracer
	//profile *StatsProfile
	span    apitracer.Span

	method  string
	service string
}

type requestDescriptor interface {
	Service() string
	Endpoint() string
}

// newRequestTracker creates a new tracker for an RPC request (client or server).
func newRequestTracker(req requestDescriptor, tracer apitracer.Tracer) *tracker {
	return &tracker{
		tracer: tracer,
		//profile: profile,
		method:  req.Endpoint(),
		service: req.Service(),
	}
}

type publicationDescriptor interface {
	Topic() string
}

// newEventTracker creates a new tracker for a publication (client or server).
func newEventTracker(pub publicationDescriptor, tracer apitracer.Tracer) *tracker {
	return &tracker{
		tracer: tracer,
		//profile: profile,
		method:  pub.Topic(),
		service: "pubsub",
	}
}

// start monitoring a request. You can choose to let this method
// start a span for the request or attach one later.
func (t *tracker) start(ctx context.Context, startSpan bool) context.Context {
	t.startedAt = time.Now()

	//ctx, _ = tag.New(ctx, tag.Upsert(Service, t.service), tag.Upsert(Endpoint, t.method))
	//stats.Record(ctx, t.profile.CountMeasure.M(1))

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
	//status := getResponseStatus(err)

	//ctx, _ = tag.New(ctx, tag.Upsert(StatusCode, strconv.Itoa(int(status.Code))))
	//stats.Record(ctx, t.profile.LatencyMeasure.M(float64(time.Since(t.startedAt))/float64(time.Millisecond)))

	if t.span != nil {
		setResponseStatus(t.span,err, "")
		//t.span.SetStatus(status)
		t.span.End()
	}
}
