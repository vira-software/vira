package middleware

import (
	"net/http"
	"strings"

	"github.com/vira-software/vira"
)

// NewGrpcMiddleware creates a middleware with gRPC server to Handle gRPC requests.
func NewGrpcMiddleware(srv http.Handler) vira.Middleware {
	return func(ctx *vira.Context) error {
		// "application/grpc", "application/grpc+proto"
		if strings.HasPrefix(ctx.GetHeader(vira.HeaderContentType), "application/grpc") {
			srv.ServeHTTP(ctx.Res, ctx.Req)
			ctx.End(204) // Must end with 204 to handle rpc error
		}
		return nil
	}
}
