package vira

import "net/http"

// MIME types
const (
	// Got from https://github.com/labstack/echo
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = "application/json; charset=utf-8"
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = "application/javascript; charset=utf-8"
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = "application/xml; charset=utf-8"
	MIMEApplicationYAML                  = "application/yaml"
	MIMEApplicationTOML                  = "application/toml" // https://github.com/toml-lang/toml
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf" // https://tools.ietf.org/html/draft-rfernando-protocol-buffers-00
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = "text/html; charset=utf-8"
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = "text/plain; charset=utf-8"
	MIMEMarkdown                         = "text/markdown"
	MIMEMarkdownCharsetUTF8              = "text/markdown; charset=utf-8"
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
	MIMEApplicationSchemaJSON            = "application/schema+json"
	MIMEApplicationSchemaInstanceJSON    = "application/schema-instance+json"
	MIMEApplicationSchemaJSONLD          = "application/ld+json"
	MIMEApplicationSchemaGraphQL         = "application/graphql"
	MIMEApplicationCBOR                  = "application/cbor"
)

// HTTP Header Fields
const (
	HeaderAccept             = "Accept"              // Requests, Responses
	HeaderAcceptCharset      = "Accept-Charset"      // Requests
	HeaderAcceptEncoding     = "Accept-Encoding"     // Requests
	HeaderAcceptLanguage     = "Accept-Language"     // Requests
	HeaderAuthorization      = "Authorization"       // Requests
	HeaderCacheControl       = "Cache-Control"       // Requests, Responses
	HeaderContentLength      = "Content-Length"      // Requests, Responses
	HeaderContentMD5         = "Content-MD5"         // Requests, Responses
	HeaderContentType        = "Content-Type"        // Requests, Responses
	HeaderIfMatch            = "If-Match"            // Requests
	HeaderIfModifiedSince    = "If-Modified-Since"   // Requests
	HeaderIfNoneMatch        = "If-None-Match"       // Requests
	HeaderIfRange            = "If-Range"            // Requests
	HeaderIfUnmodifiedSince  = "If-Unmodified-Since" // Requests
	HeaderMaxForwards        = "Max-Forwards"        // Requests
	HeaderProxyAuthorization = "Proxy-Authorization" // Requests
	HeaderPragma             = "Pragma"              // Requests, Responses
	HeaderRange              = "Range"               // Requests
	HeaderReferer            = "Referer"             // Requests
	HeaderUserAgent          = "User-Agent"          // Requests
	HeaderTE                 = "TE"                  // Requests
	HeaderVia                = "Via"                 // Requests
	HeaderWarning            = "Warning"             // Requests, Responses
	HeaderCookie             = "Cookie"              // Requests
	HeaderOrigin             = "Origin"              // Requests
	HeaderAcceptDatetime     = "Accept-Datetime"     // Requests
	HeaderXRequestedWith     = "X-Requested-With"    // Requests
	HeaderXRequestID         = "X-Request-Id"        // Requests
	HeaderXCanary            = "X-Canary"            // Requests, Responses
	HeaderXForwardedScheme   = "X-Forwarded-Scheme"  // Requests
	HeaderXForwardedProto    = "X-Forwarded-Proto"   // Requests
	HeaderXForwardedFor      = "X-Forwarded-For"     // Requests
	HeaderXForwardedHost     = "X-Forwarded-Host"    // Requests
	HeaderXForwardedServer   = "X-Forwarded-Server"  // Requests
	HeaderXRealIP            = "X-Real-Ip"           // Requests
	HeaderXRealScheme        = "X-Real-Scheme"       // Requests

	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"      // Responses
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"     // Responses
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"     // Responses
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials" // Responses
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"    // Responses
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"           // Responses
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"    // Responses
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"   // Responses
	HeaderAcceptPatch                   = "Accept-Patch"                     // Responses
	HeaderAcceptRanges                  = "Accept-Ranges"                    // Responses
	HeaderAllow                         = "Allow"                            // Responses
	HeaderContentEncoding               = "Content-Encoding"                 // Responses
	HeaderContentLanguage               = "Content-Language"                 // Responses
	HeaderContentLocation               = "Content-Location"                 // Responses
	HeaderContentDisposition            = "Content-Disposition"              // Responses
	HeaderContentRange                  = "Content-Range"                    // Responses
	HeaderETag                          = "ETag"                             // Responses
	HeaderExpires                       = "Expires"                          // Responses
	HeaderLastModified                  = "Last-Modified"                    // Responses
	HeaderLink                          = "Link"                             // Responses
	HeaderLocation                      = "Location"                         // Responses
	HeaderP3P                           = "P3P"                              // Responses
	HeaderProxyAuthenticate             = "Proxy-Authenticate"               // Responses
	HeaderRefresh                       = "Refresh"                          // Responses
	HeaderRetryAfter                    = "Retry-After"                      // Responses
	HeaderServer                        = "Server"                           // Responses
	HeaderSetCookie                     = "Set-Cookie"                       // Responses
	HeaderStrictTransportSecurity       = "Strict-Transport-Security"        // Responses
	HeaderTransferEncoding              = "Transfer-Encoding"                // Responses
	HeaderUpgrade                       = "Upgrade"                          // Responses
	HeaderVary                          = "Vary"                             // Responses
	HeaderWWWAuthenticate               = "WWW-Authenticate"                 // Responses
	HeaderPublicKeyPins                 = "Public-Key-Pins"                  // Responses
	HeaderPublicKeyPinsReportOnly       = "Public-Key-Pins-Report-Only"      // Responses
	HeaderRefererPolicy                 = "Referrer-Policy"                  // Responses

	// Common Non-Standard Response Headers
	HeaderXFrameOptions                   = "X-Frame-Options"                     // Responses
	HeaderXXSSProtection                  = "X-XSS-Protection"                    // Responses
	HeaderContentSecurityPolicy           = "Content-Security-Policy"             // Responses
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only" // Responses
	HeaderXContentSecurityPolicy          = "X-Content-Security-Policy"           // Responses
	HeaderXWebKitCSP                      = "X-WebKit-CSP"                        // Responses
	HeaderXContentTypeOptions             = "X-Content-Type-Options"              // Responses
	HeaderXPoweredBy                      = "X-Powered-By"                        // Responses
	HeaderXUACompatible                   = "X-UA-Compatible"                     // Responses
	HeaderXCSRFToken                      = "X-CSRF-Token"                        // Responses
	HeaderXHTTPMethodOverride             = "X-HTTP-Method-Override"              // Responses
	HeaderXDNSPrefetchControl             = "X-DNS-Prefetch-Control"              // Responses
	HeaderXDownloadOptions                = "X-Download-Options"                  // Responses
)

// Predefined errors
var (
	ViraErr = &Error{Code: http.StatusInternalServerError, Err: "Error"}

	// https://golang.org/pkg/net/http/#pkg-constants
	ErrBadRequest                    = ViraErr.WithCode(http.StatusBadRequest).WithErr("BadRequest")
	ErrUnauthorized                  = ViraErr.WithCode(http.StatusUnauthorized).WithErr("Unauthorized")
	ErrPaymentRequired               = ViraErr.WithCode(http.StatusPaymentRequired).WithErr("PaymentRequired")
	ErrForbidden                     = ViraErr.WithCode(http.StatusForbidden).WithErr("Forbidden")
	ErrNotFound                      = ViraErr.WithCode(http.StatusNotFound).WithErr("NotFound")
	ErrMethodNotAllowed              = ViraErr.WithCode(http.StatusMethodNotAllowed).WithErr("MethodNotAllowed")
	ErrNotAcceptable                 = ViraErr.WithCode(http.StatusNotAcceptable).WithErr("NotAcceptable")
	ErrProxyAuthRequired             = ViraErr.WithCode(http.StatusProxyAuthRequired).WithErr("ProxyAuthenticationRequired")
	ErrRequestTimeout                = ViraErr.WithCode(http.StatusRequestTimeout).WithErr("RequestTimeout")
	ErrConflict                      = ViraErr.WithCode(http.StatusConflict).WithErr("Conflict")
	ErrGone                          = ViraErr.WithCode(http.StatusGone).WithErr("Gone")
	ErrLengthRequired                = ViraErr.WithCode(http.StatusLengthRequired).WithErr("LengthRequired")
	ErrPreconditionFailed            = ViraErr.WithCode(http.StatusPreconditionFailed).WithErr("PreconditionFailed")
	ErrRequestEntityTooLarge         = ViraErr.WithCode(http.StatusRequestEntityTooLarge).WithErr("RequestEntityTooLarge")
	ErrRequestURITooLong             = ViraErr.WithCode(http.StatusRequestURITooLong).WithErr("RequestURITooLong")
	ErrUnsupportedMediaType          = ViraErr.WithCode(http.StatusUnsupportedMediaType).WithErr("UnsupportedMediaType")
	ErrRequestedRangeNotSatisfiable  = ViraErr.WithCode(http.StatusRequestedRangeNotSatisfiable).WithErr("RequestedRangeNotSatisfiable")
	ErrExpectationFailed             = ViraErr.WithCode(http.StatusExpectationFailed).WithErr("ExpectationFailed")
	ErrTeapot                        = ViraErr.WithCode(http.StatusTeapot).WithErr("Teapot")
	ErrMisdirectedRequest            = ViraErr.WithCode(421).WithErr("MisdirectedRequest")
	ErrUnprocessableEntity           = ViraErr.WithCode(http.StatusUnprocessableEntity).WithErr("UnprocessableEntity")
	ErrLocked                        = ViraErr.WithCode(http.StatusLocked).WithErr("Locked")
	ErrFailedDependency              = ViraErr.WithCode(http.StatusFailedDependency).WithErr("FailedDependency")
	ErrUpgradeRequired               = ViraErr.WithCode(http.StatusUpgradeRequired).WithErr("UpgradeRequired")
	ErrPreconditionRequired          = ViraErr.WithCode(http.StatusPreconditionRequired).WithErr("PreconditionRequired")
	ErrTooManyRequests               = ViraErr.WithCode(http.StatusTooManyRequests).WithErr("TooManyRequests")
	ErrRequestHeaderFieldsTooLarge   = ViraErr.WithCode(http.StatusRequestHeaderFieldsTooLarge).WithErr("RequestHeaderFieldsTooLarge")
	ErrUnavailableForLegalReasons    = ViraErr.WithCode(http.StatusUnavailableForLegalReasons).WithErr("UnavailableForLegalReasons")
	ErrClientClosedRequest           = ViraErr.WithCode(499).WithErr("ClientClosedRequest")
	ErrInternalServerError           = ViraErr.WithCode(http.StatusInternalServerError).WithErr("InternalServerError")
	ErrNotImplemented                = ViraErr.WithCode(http.StatusNotImplemented).WithErr("NotImplemented")
	ErrBadGateway                    = ViraErr.WithCode(http.StatusBadGateway).WithErr("BadGateway")
	ErrServiceUnavailable            = ViraErr.WithCode(http.StatusServiceUnavailable).WithErr("ServiceUnavailable")
	ErrGatewayTimeout                = ViraErr.WithCode(http.StatusGatewayTimeout).WithErr("GatewayTimeout")
	ErrHTTPVersionNotSupported       = ViraErr.WithCode(http.StatusHTTPVersionNotSupported).WithErr("HTTPVersionNotSupported")
	ErrVariantAlsoNegotiates         = ViraErr.WithCode(http.StatusVariantAlsoNegotiates).WithErr("VariantAlsoNegotiates")
	ErrInsufficientStorage           = ViraErr.WithCode(http.StatusInsufficientStorage).WithErr("InsufficientStorage")
	ErrLoopDetected                  = ViraErr.WithCode(http.StatusLoopDetected).WithErr("LoopDetected")
	ErrNotExtended                   = ViraErr.WithCode(http.StatusNotExtended).WithErr("NotExtended")
	ErrNetworkAuthenticationRequired = ViraErr.WithCode(http.StatusNetworkAuthenticationRequired).WithErr("NetworkAuthenticationRequired")
)

// ErrByStatus returns a vira.Error by http status.
func ErrByStatus(status int) *Error {
	switch status {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthorized
	case 402:
		return ErrPaymentRequired
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 405:
		return ErrMethodNotAllowed
	case 406:
		return ErrNotAcceptable
	case 407:
		return ErrProxyAuthRequired
	case 408:
		return ErrRequestTimeout
	case 409:
		return ErrConflict
	case 410:
		return ErrGone
	case 411:
		return ErrLengthRequired
	case 412:
		return ErrPreconditionFailed
	case 413:
		return ErrRequestEntityTooLarge
	case 414:
		return ErrRequestURITooLong
	case 415:
		return ErrUnsupportedMediaType
	case 416:
		return ErrRequestedRangeNotSatisfiable
	case 417:
		return ErrExpectationFailed
	case 418:
		return ErrTeapot
	case 421:
		return ErrMisdirectedRequest
	case 422:
		return ErrUnprocessableEntity
	case 423:
		return ErrLocked
	case 424:
		return ErrFailedDependency
	case 426:
		return ErrUpgradeRequired
	case 428:
		return ErrPreconditionRequired
	case 429:
		return ErrTooManyRequests
	case 431:
		return ErrRequestHeaderFieldsTooLarge
	case 451:
		return ErrUnavailableForLegalReasons
	case 499:
		return ErrClientClosedRequest
	case 500:
		return ErrInternalServerError
	case 501:
		return ErrNotImplemented
	case 502:
		return ErrBadGateway
	case 503:
		return ErrServiceUnavailable
	case 504:
		return ErrGatewayTimeout
	case 505:
		return ErrHTTPVersionNotSupported
	case 506:
		return ErrVariantAlsoNegotiates
	case 507:
		return ErrInsufficientStorage
	case 508:
		return ErrLoopDetected
	case 510:
		return ErrNotExtended
	case 511:
		return ErrNetworkAuthenticationRequired
	default:
		return ViraErr.WithCode(status)
	}
}

const banner = `       __                
___  _|__|___________    
\  \/ /  \_  __ \__  \   
 \   /|  ||  | \// __ \_ 
  \_/ |__||__|  (____  / 
                     \/
`
