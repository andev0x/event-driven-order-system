package observability

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// DetachContext copies telemetry context to a non-cancelable context.
func DetachContext(ctx context.Context) context.Context {
	detached := context.Background()

	if bg := baggage.FromContext(ctx); bg.Len() > 0 {
		detached = baggage.ContextWithBaggage(detached, bg)
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		detached = trace.ContextWithSpanContext(detached, spanCtx)
	}

	return detached
}
