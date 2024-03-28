package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vira-software/vira"
)

// CorsOptions is cors middleware options.
type CorsOptions struct {
	// AllowOrigins defines the origins which will be allowed to access
	// the resource. Default value is []string{"*"} .
	AllowOrigins []string
	// AllowMethods defines the methods which will be allowed to access
	// the resource. It is used in handling the preflighted requests.
	// Default value is []string{"GET", "HEAD", "PUT", "POST", "DELETE", "PATCH"} .
	AllowMethods []string
	// AllowOriginsValidator validates the request Origin by validator
	// function.The validator function accpects an `*vira.Context` and returns the
	// Access-Control-Allow-Origin value. If the validator is set, then
	// AllowMethods will be ignored.
	AllowOriginsValidator func(origin string, ctx *vira.Context) string
	// AllowHeaders defines the headers which will be allowed in the actual
	// request, It is used in handling the preflighted requests.
	AllowHeaders []string
	// ExposeHeaders defines the allowed headers that client could send when
	// accessing the resource.
	ExposeHeaders []string
	// MaxAge defines the max age that the preflighted requests can be cached.
	MaxAge time.Duration
	// Credentials defines whether or not the response to the request
	// can be exposed.
	Credentials bool
}

var (
	defaultAllowOrigins = []string{"*"}
	defaultAllowMethods = []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPut,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPatch,
	}
)

// NewCorsMiddleware creates a middleware to provide CORS support for vira.
func NewCorsMiddleware(options ...CorsOptions) vira.Middleware {
	opts := CorsOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	if opts.AllowOrigins == nil {
		opts.AllowOrigins = defaultAllowOrigins
	}
	if opts.AllowMethods == nil {
		opts.AllowMethods = defaultAllowMethods
	}
	if opts.AllowOriginsValidator == nil {
		opts.AllowOriginsValidator = func(origin string, _ *vira.Context) (allowOrigin string) {
			for _, o := range opts.AllowOrigins {
				if o == origin || o == "*" {
					allowOrigin = origin
					break
				}
			}
			return
		}
	}

	return func(ctx *vira.Context) (err error) {
		// Always set Vary, see https://github.com/rs/cors/issues/10
		ctx.Res.Vary(vira.HeaderOrigin)

		origin := ctx.GetHeader(vira.HeaderOrigin)
		// not a CORS request
		if origin == "" {
			return
		}

		allowOrigin := opts.AllowOriginsValidator(origin, ctx)
		if allowOrigin == "" {
			// If the request Origin header is not allowed. Just terminate the following steps.
			if ctx.Method == http.MethodOptions {
				return ctx.End(http.StatusOK)
			}
			return
		}

		ctx.SetHeader(vira.HeaderAccessControlAllowOrigin, allowOrigin)
		if opts.Credentials {
			// when responding to a credentialed request, server must specify a
			// domain, and cannot use wild carding.
			// See *important note* in https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS#Requests_with_credentials .
			ctx.SetHeader(vira.HeaderAccessControlAllowCredentials, "true")
		}

		// Handle preflighted requests (https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS#Preflighted_requests) .
		// https://stackoverflow.com/questions/46026409/what-are-proper-status-codes-for-cors-preflight-requests
		if ctx.Method == http.MethodOptions {
			ctx.Res.Vary(vira.HeaderAccessControlRequestMethod)
			ctx.Res.Vary(vira.HeaderAccessControlRequestHeaders)

			requestMethod := ctx.GetHeader(vira.HeaderAccessControlRequestMethod)
			// If there is no "Access-Control-Request-Method" request header. We just
			// treat this request as an invalid preflighted request, so terminate the
			// following steps.
			if requestMethod == "" {
				ctx.Res.Del(vira.HeaderAccessControlAllowOrigin)
				ctx.Res.Del(vira.HeaderAccessControlAllowCredentials)
				return ctx.End(http.StatusOK)
			}
			if len(opts.AllowMethods) > 0 {
				ctx.SetHeader(vira.HeaderAccessControlAllowMethods, strings.Join(opts.AllowMethods, ", "))
			}

			var allowHeaders string
			if len(opts.AllowHeaders) > 0 {
				allowHeaders = strings.Join(opts.AllowHeaders, ", ")
			} else {
				allowHeaders = ctx.GetHeader(vira.HeaderAccessControlRequestHeaders)
			}
			if allowHeaders != "" {
				ctx.SetHeader(vira.HeaderAccessControlAllowHeaders, allowHeaders)
			}

			if opts.MaxAge > 0 {
				ctx.SetHeader(vira.HeaderAccessControlMaxAge, strconv.Itoa(int(opts.MaxAge.Seconds())))
			}
			return ctx.End(http.StatusOK)
		}

		if len(opts.ExposeHeaders) > 0 {
			ctx.SetHeader(vira.HeaderAccessControlExposeHeaders, strings.Join(opts.ExposeHeaders, ", "))
		}
		return
	}
}
