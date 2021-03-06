package tracing

import (
	"context"

	"github.com/eirsyl/flexit/endpoint"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// TraceServer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceServer(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			serverSpan := opentracing.SpanFromContext(ctx)
			if serverSpan == nil {
				// All we can do is create a new root span.
				serverSpan = tracer.StartSpan(operationName)
			} else {
				serverSpan.SetOperationName(operationName)
			}
			defer func() {
				if err != nil {
					serverSpan.SetTag("error", true)
					serverSpan.SetTag("message", err.Error())
				}
				serverSpan.Finish()
			}()
			ext.SpanKindRPCServer.Set(serverSpan)
			ctx = opentracing.ContextWithSpan(ctx, serverSpan)
			return next(ctx, request)
		}
	}
}

// TraceClient returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
func TraceClient(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var clientSpan opentracing.Span
			if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
				clientSpan = tracer.StartSpan(
					operationName,
					opentracing.ChildOf(parentSpan.Context()),
				)
			} else {
				clientSpan = tracer.StartSpan(operationName)
			}
			defer func() {
				if err != nil {
					clientSpan.SetTag("error", true)
					clientSpan.SetTag("message", err.Error())
				}
				clientSpan.Finish()
			}()
			ext.SpanKindRPCClient.Set(clientSpan)
			ctx = opentracing.ContextWithSpan(ctx, clientSpan)
			return next(ctx, request)
		}
	}
}
