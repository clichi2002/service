package mid

import (
	"context"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/business/sys/metrics"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.web.mid.panics")
			defer span.End()

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					err = errors.Errorf("PANIC: %v", rec)

					// Don't count anything on /debug routes towards metrics.
					if !strings.HasPrefix(r.URL.Path, "/debug") {
						if v, ok := ctx.Value(metrics.Key).(*metrics.Metrics); ok {
							v.Panics.Add(1)
						}
					}
				}
			}()

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}