package vira

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	// "golang.org/x/net/http2"
	// "golang.org/x/net/http2/h2c"
)

// Middleware defines a function to process as middleware.
type Middleware func(ctx *Context) error

// Handler interface is used by vira.UseHandler as a middleware.
type Handler interface {
	Serve(ctx *Context) error
}

// Sender interface is used by ctx.Send.
type Sender interface {
	Send(ctx *Context, code int, data any) error
}

// Renderer interface is used by ctx.Render.
type Renderer interface {
	Render(ctx *Context, w io.Writer, name string, data any) error
}

// URLParser interface is used by ctx.ParseUrl. Default to:
//
//	vira.Set(vira.SetURLParser, vira.DefaultURLParser)
type URLParser interface {
	Parse(val map[string][]string, body any, tag string) error
}

// DefaultURLParser is default URLParser type.
type DefaultURLParser struct{}

// Parse implemented URLParser interface.
func (d DefaultURLParser) Parse(val map[string][]string, body any, tag string) error {
	return ValuesToStruct(val, body, tag)
}

// BodyParser interface is used by ctx.ParseBody. Default to:
//
//	vira.Set(vira.SetBodyParser, vira.DefaultBodyParser(1<<20))
type BodyParser interface {
	// Maximum allowed size for a request body
	MaxBytes() int64
	Parse(buf []byte, body any, mediaType, charset string) error
}

// DefaultBodyParser is default BodyParser type.
// SetBodyParser used 1MB as default:
//
//	vira.Set(vira.SetBodyParser, vira.DefaultBodyParser(1<<20))
type DefaultBodyParser int64

// MaxBytes implemented BodyParser interface.
func (d DefaultBodyParser) MaxBytes() int64 {
	return int64(d)
}

// Parse implemented BodyParser interface.
func (d DefaultBodyParser) Parse(buf []byte, body any, mediaType, charset string) error {
	if len(buf) == 0 {
		return ErrBadRequest.WithMsg("request entity empty")
	}
	switch true {
	case strings.HasPrefix(mediaType, MIMEApplicationJSON), isLikeMediaType(mediaType, "json"):
		err := json.Unmarshal(buf, body)
		if err == nil {
			return nil
		}

		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			if ute.Field == "" { // go1.11
				return fmt.Errorf("unmarshal type error: expected=%v, got=%v, offset=%v",
					ute.Type, ute.Value, ute.Offset)
			}
			return fmt.Errorf("unmarshal type error: field=%v, expected=%v, got=%v, offset=%v",
				ute.Field, ute.Type, ute.Value, ute.Offset)
		} else if se, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("syntax error: offset=%v, error=%v", se.Offset, se.Error())
		} else {
			return err
		}
	case strings.HasPrefix(mediaType, MIMEApplicationXML), isLikeMediaType(mediaType, "xml"):
		return xml.Unmarshal(buf, body)
	}

	return ErrUnsupportedMediaType.WithMsgf("unsupported media type: %s", mediaType)
}

// HTTPError interface is used to create a server error that include status code and error message.
type HTTPError interface {
	// Error returns error's message.
	Error() string
	// Status returns error's http status code.
	Status() int
}

// Vira is the top-level framework struct.
//
// Hello Vira!
//
//	package main
//
//	func main() {
//		vira := vira.New() // Create vira
//		vira.Use(func(ctx *vira.Context) error {
//			return ctx.HTML(200, "<h1>Hello, Vira!</h1>")
//		})
//		vira.Error(vira.Listen(":3000"))
//	}
type Vira struct {
	Server *http.Server
	mds    middlewares

	keys        []string
	renderer    Renderer
	sender      Sender
	bodyParser  BodyParser
	urlParser   URLParser
	compress    Compressible  // Default to nil, do not compress response content.
	timeout     time.Duration // Default to 0, no time out.
	serverName  string        // Vira/1.7.6
	logger      *log.Logger
	parseError  func(error) HTTPError
	renderError func(HTTPError) (code int, contentType string, body []byte)
	onerror     func(*Context, HTTPError)
	withContext func(*http.Request) context.Context
	settings    map[any]any
}

// NewTrie creates an instance of App.
func New() *Vira {
	vira := new(Vira)
	vira.Server = new(http.Server)
	// https://medium.com/@simonfrey/go-as-in-golang-standard-net-http-config-will-break-your-production-environment-1360871cb72b
	vira.Server.ReadHeaderTimeout = 20 * time.Second
	vira.Server.ReadTimeout = 60 * time.Second
	vira.Server.WriteTimeout = 120 * time.Second
	vira.Server.IdleTimeout = 90 * time.Second

	vira.mds = make(middlewares, 0)
	vira.settings = make(map[any]any)

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	vira.Set(SetEnv, env)
	vira.Set(SetServerName, "Vira/"+Version)
	vira.Set(SetTrustedProxy, false)
	vira.Set(SetBodyParser, DefaultBodyParser(2<<20)) // 2MB
	vira.Set(SetURLParser, DefaultURLParser{})
	vira.Set(SetLogger, log.New(os.Stderr, "", 0))
	vira.Set(SetGraceTimeout, 10*time.Second)
	vira.Set(SetParseError, func(err error) HTTPError {
		return ParseError(err)
	})
	vira.Set(SetRenderError, defaultRenderError)
	vira.Set(SetOnError, func(ctx *Context, err HTTPError) {
		ctx.Error(err)
	})
	return vira
}

// Use uses the given middleware `handle`.
func (vira *Vira) Use(handle Middleware) *Vira {
	vira.mds = append(vira.mds, handle)
	return vira
}

// UseHandler uses a instance that implemented Handler interface.
func (vira *Vira) UseHandler(h Handler) *Vira {
	vira.mds = append(vira.mds, h.Serve)
	return vira
}

type appSetting uint8

// Build-in vira settings
const (
	// It will be used by `ctx.ParseBody`, value should implements `vira.BodyParser` interface, default to:
	//  vira.Set(vira.SetBodyParser, vira.DefaultBodyParser(1<<20))
	SetBodyParser appSetting = iota

	// It will be used by `ctx.ParseURL`, value should implements `vira.URLParser` interface, default to:
	//  vira.Set(vira.SetURLParser, vira.DefaultURLParser)
	SetURLParser

	// Enable compress for response, value should implements `vira.Compressible` interface, no default value.
	// Example:
	//
	//  vira := vira.New()
	//  vira.Set(vira.SetCompress, compressible.WithThreshold(1024))
	SetCompress

	// Set secret keys for signed cookies, it will be used by `ctx.Cookies`, value should be `[]string` type,
	// no default value. More document https://github.com/go-http-utils/cookie, Example:
	//  vira.Set(vira.SetKeys, []string{"some key2", "some key1"})
	SetKeys

	// Set a logger to vira, value should be `*log.Logger` instance, default to:
	//  vira.Set(vira.SetLogger, log.New(os.Stderr, "", 0))
	// Maybe you need LoggerFilterWriter to filter some server errors in production:
	//  vira.Set(vira.SetLogger, log.New(vira.DefaultFilterWriter(), "", 0))
	// We recommand set logger flags to 0.
	SetLogger

	// Set a ParseError hook to vira that convert middleware error to HTTPError,
	// value should be `func(err error) HTTPError`, default to:
	//  vira.Set(SetParseError, func(err error) HTTPError {
	//  	return ParseError(err)
	//  })
	SetParseError

	// Set a SetRenderError hook to vira that convert error to raw response,
	// value should be `func(HTTPError) (code int, contentType string, body []byte)`, default to:
	//   vira.Set(SetRenderError, func(err HTTPError) (int, string, []byte) {
	//  	// default to render error as json
	//  	body, e := json.Marshal(err)
	//  	if e != nil {
	//  		body, _ = json.Marshal(map[string]string{"error": err.Error()})
	//  	}
	//  	return err.Status(), MIMEApplicationJSONCharsetUTF8, body
	//  })
	//
	// you can use another recommand one:
	//
	//  vira.Set(vira.SetRenderError, vira.RenderErrorResponse)
	//
	SetRenderError

	// Set a on-error hook to vira that handle middleware error.
	// value should be `func(ctx *Context, err HTTPError)`, default to:
	//  vira.Set(SetOnError, func(ctx *Context, err HTTPError) {
	//  	ctx.Error(err)
	//  })
	SetOnError

	// Set a SetSender to vira, it will be used by `ctx.Send`, value should implements `vira.Sender` interface,
	// no default value.
	SetSender

	// Set a renderer to vira, it will be used by `ctx.Render`, value should implements `vira.Renderer` interface,
	// no default value.
	SetRenderer

	// Set a timeout to for the middleware process, value should be `time.Duration`. No default.
	// Example:
	//  vira.Set(vira.SetTimeout, 3*time.Second)
	SetTimeout

	// Set a graceful timeout to for gracefully shuts down, value should be `time.Duration`. Default to 10*time.Second.
	// Example:
	//  vira.Set(vira.SetGraceTimeout, 60*time.Second)
	SetGraceTimeout

	// Set a function that Wrap the vira.Context' underlayer context.Context. No default.
	SetWithContext

	// Set a vira env string to vira, it can be retrieved by `ctx.Setting(vira.SetEnv)`.
	// Default to os process "APP_ENV" or "development".
	SetEnv

	// Set a server name that respond to client as "Server" header.
	// Default to "Vira/{version}".
	SetServerName

	// Set true and proxy header fields will be trusted
	// Default to false.
	SetTrustedProxy
)

// Set add key/value settings to vira. The settings can be retrieved by `ctx.Setting(key)`.
func (vira *Vira) Set(key, val any) *Vira {
	if k, ok := key.(appSetting); ok {
		switch key {
		case SetBodyParser:
			if bodyParser, ok := val.(BodyParser); !ok {
				panic(ViraErr.WithMsg("SetBodyParser setting must implemented `vira.BodyParser` interface"))
			} else {
				vira.bodyParser = bodyParser
			}
		case SetURLParser:
			if urlParser, ok := val.(URLParser); !ok {
				panic(ViraErr.WithMsg("SetURLParser setting must implemented `vira.URLParser` interface"))
			} else {
				vira.urlParser = urlParser
			}
		case SetCompress:
			if compress, ok := val.(Compressible); !ok {
				panic(ViraErr.WithMsg("SetCompress setting must implemented `vira.Compressible` interface"))
			} else {
				vira.compress = compress
			}
		case SetKeys:
			if keys, ok := val.([]string); !ok {
				panic(ViraErr.WithMsg("SetKeys setting must be `[]string`"))
			} else {
				vira.keys = keys
			}
		case SetLogger:
			if logger, ok := val.(*log.Logger); !ok {
				panic(ViraErr.WithMsg("SetLogger setting must be `*log.Logger` instance"))
			} else {
				vira.logger = logger
			}
		case SetParseError:
			if parseError, ok := val.(func(error) HTTPError); !ok {
				panic(ViraErr.WithMsg("SetParseError setting must be `func(error) HTTPError`"))
			} else {
				vira.parseError = parseError
			}
		case SetRenderError:
			if renderError, ok := val.(func(HTTPError) (int, string, []byte)); !ok {
				panic(ViraErr.WithMsg("SetRenderError setting must be `func(HTTPError) (int, string, []byte)`"))
			} else {
				vira.renderError = renderError
			}
		case SetOnError:
			if onerror, ok := val.(func(*Context, HTTPError)); !ok {
				panic(ViraErr.WithMsg("SetOnError setting must be `func(*Context, HTTPError)`"))
			} else {
				vira.onerror = onerror
			}
		case SetSender:
			if sender, ok := val.(Sender); !ok {
				panic(ViraErr.WithMsg("SetSender setting must implemented `vira.Sender` interface"))
			} else {
				vira.sender = sender
			}
		case SetRenderer:
			if renderer, ok := val.(Renderer); !ok {
				panic(ViraErr.WithMsg("SetRenderer setting must implemented `vira.Renderer` interface"))
			} else {
				vira.renderer = renderer
			}
		case SetTimeout:
			if timeout, ok := val.(time.Duration); !ok {
				panic(ViraErr.WithMsg("SetTimeout setting must be `time.Duration` instance"))
			} else {
				vira.timeout = timeout
			}
		case SetGraceTimeout:
			if _, ok := val.(time.Duration); !ok {
				panic(ViraErr.WithMsg("SetGraceTimeout setting must be `time.Duration` instance"))
			}
		case SetWithContext:
			if withContext, ok := val.(func(*http.Request) context.Context); !ok {
				panic(ViraErr.WithMsg("SetWithContext setting must be `func(*http.Request) context.Context`"))
			} else {
				vira.withContext = withContext
			}
		case SetEnv:
			if _, ok := val.(string); !ok {
				panic(ViraErr.WithMsg("SetEnv setting must be `string`"))
			}
		case SetServerName:
			if name, ok := val.(string); !ok {
				panic(ViraErr.WithMsg("SetServerName setting must be `string`"))
			} else {
				vira.serverName = name
			}
		case SetTrustedProxy:
			if _, ok := val.(bool); !ok {
				panic(ViraErr.WithMsg("SetTrustedProxy setting must be `bool`"))
			}
		}
		vira.settings[k] = val
		return vira
	}
	vira.settings[key] = val
	return vira
}

// Env returns vira's env. You can set vira env with `vira.Set(vira.SetEnv, "some env")`
// Default to os process "APP_ENV" or "development".
func (vira *Vira) Env() string {
	return vira.settings[SetEnv].(string)
}

// Listen starts the HTTP server.
func (vira *Vira) Listen(addr string) error {
	vira.Server.Addr = addr
	vira.Server.ErrorLog = vira.logger
	vira.Server.Handler = vira
	// vira.Server.Handler = h2c.NewHandler(vira, &http2.Server{})
	// Print the server info
	FprintWithColor(std.Out, fmt.Sprintf("%s http server started on %s \n", banner, addr), ColorCyan)
	return vira.Server.ListenAndServe()
}

// ListenTLS starts the HTTPS server.
func (vira *Vira) ListenTLS(addr, certFile, keyFile string) error {
	vira.Server.Addr = addr
	vira.Server.ErrorLog = vira.logger
	vira.Server.Handler = vira
	return vira.Server.ListenAndServeTLS(certFile, keyFile)
}

// ListenWithContext starts the HTTP server (or HTTPS server with keyPair) with a context
//
// Usage:
//
//	 func main() {
//	 	vira := vira.New() // Create vira
//	 	do some thing...
//
//	 	vira.ListenWithContext(vira.ContextWithSignal(context.Background()), addr)
//		  // starts the HTTPS server.
//		  // vira.ListenWithContext(vira.ContextWithSignal(context.Background()), addr, certFile, keyFile)
//	 }
func (vira *Vira) ListenWithContext(ctx context.Context, addr string, keyPair ...string) error {
	timeout := vira.settings[SetGraceTimeout].(time.Duration)
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := vira.Close(c); err != nil {
			vira.Error(err)
		}
	}()

	if len(keyPair) >= 2 && keyPair[0] != "" && keyPair[1] != "" {
		return vira.ListenTLS(addr, keyPair[0], keyPair[1])
	}
	return vira.Listen(addr)
}

// ServeWithContext accepts incoming connections on the Listener l, starts the HTTP server (or HTTPS server with keyPair) with a context
//
// Usage:
//
//	 func main() {
//			l, err := net.Listen("tcp", ":8080")
//			if err != nil {
//				log.Fatal(err)
//			}
//
//	 	vira := vira.New() // Create vira
//	 	do some thing...
//
//	 	vira.ServeWithContext(vira.ContextWithSignal(context.Background()), l)
//		  // starts the HTTPS server.
//		  // vira.ServeWithContext(vira.ContextWithSignal(context.Background()), l, certFile, keyFile)
//	 }
func (vira *Vira) ServeWithContext(ctx context.Context, l net.Listener, keyPair ...string) error {
	timeout := vira.settings[SetGraceTimeout].(time.Duration)
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := vira.Close(c); err != nil {
			vira.Error(err)
		}
	}()

	vira.Server.ErrorLog = vira.logger
	vira.Server.Handler = vira
	if len(keyPair) >= 2 && keyPair[0] != "" && keyPair[1] != "" {
		return vira.Server.ServeTLS(l, keyPair[0], keyPair[1])
	}
	return vira.Server.Serve(l)
}

// Start starts a non-blocking vira instance. It is useful for testing.
// If addr omit, the vira will listen on a random addr, use ServerListener.Addr() to get it.
// The non-blocking vira instance must close by ServerListener.Close().
func (vira *Vira) Start(addr ...string) *ServerListener {
	laddr := "127.0.0.1:0"
	if len(addr) > 0 && addr[0] != "" {
		laddr = addr[0]
	}
	vira.Server.ErrorLog = vira.logger
	vira.Server.Handler = vira

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		panic(ViraErr.WithMsgf("failed to listen on %v: %v", laddr, err))
	}

	c := make(chan error)
	go func() {
		c <- vira.Server.Serve(l)
	}()
	return &ServerListener{l, c}
}

// Error writes error to underlayer logging system.
func (vira *Vira) Error(err any) {
	if err := ErrorWithStack(err, 2); err != nil {
		str, e := err.Format()
		f := vira.logger.Flags() == 0
		switch {
		case f && e == nil:
			vira.logger.Printf("[%s] ERR %s\n", time.Now().UTC().Format("2006-01-02T15:04:05.999Z"), str)
		case f && e != nil:
			vira.logger.Printf("[%s] CRIT %s\n", time.Now().UTC().Format("2006-01-02T15:04:05.999Z"), err.String())
		case !f && e == nil:
			vira.logger.Printf("ERR %s\n", str)
		default:
			vira.logger.Printf("CRIT %s\n", err.String())
		}
	}
}

func (vira *Vira) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(vira, w, r)

	// TODO: handle compression writer
	// if compressWriter := ctx.handleCompress(); compressWriter != nil {
	// 	defer compressWriter.Close()
	// }

	// recover panic error
	defer catchRequest(ctx)
	go handleCtxEnd(ctx)

	// process vira middleware
	err := vira.mds.run(ctx)
	if ctx.Res.wroteHeader.isTrue() {
		if !IsNil(err) {
			vira.Error(err)
		}
		return
	}

	// if context canceled abnormally...
	if e := ctx.Err(); e != nil {
		if e == context.Canceled {
			// https://stackoverflow.com/questions/46234679/what-is-the-correct-http-status-code-for-a-cancelled-request
			// 499 Client Closed Request Used when the client has closed
			// the request before the server could send a response.
			ctx.Res.WriteHeader(ErrClientClosedRequest.Code)
			return
		}
		err = ErrGatewayTimeout.WithMsg(e.Error())
	}

	// handle middleware errors
	if !IsNil(err) {
		ctx.Res.afterHooks = nil // clear afterHooks when any error
		ctx.Res.ResetHeader()
		e := vira.parseError(err)
		vira.onerror(ctx, e)
		// try to ensure respond error if `vira.onerror` does't do it.
		ctx.respondError(e)
	} else {
		// try to ensure respond
		ctx.Res.respond(0, nil)
	}
}

// Close closes the underlying server gracefully.
// If context omit, Server.Close will be used to close immediately.
// Otherwise Server.Shutdown will be used to close gracefully.
func (vira *Vira) Close(ctx ...context.Context) error {
	if len(ctx) > 0 {
		return vira.Server.Shutdown(ctx[0])
	}
	return vira.Server.Close()
}

// ServerListener is returned by a non-blocking vira instance.
type ServerListener struct {
	l net.Listener
	c <-chan error
}

// Close closes the non-blocking vira instance.
func (s *ServerListener) Close() error {
	return s.l.Close()
}

// Addr returns the non-blocking vira instance addr.
func (s *ServerListener) Addr() net.Addr {
	return s.l.Addr()
}

// Wait make the non-blocking vira instance blocking.
func (s *ServerListener) Wait() error {
	return <-s.c
}

func catchRequest(ctx *Context) {
	if err := recover(); err != nil && err != http.ErrAbortHandler {
		ctx.Res.afterHooks = nil
		ctx.Res.ResetHeader()
		e := ErrorWithStack(err, 3)
		ctx.vira.onerror(ctx, e)
		// try to ensure respond error if `vira.onerror` does't do it.
		ctx.respondError(e)
	}
	// execute "end hooks" with LIFO order after Response.WriteHeader.
	// they run in a goroutine, in order to not block current HTTP Request/Response.
	if len(ctx.Res.endHooks) > 0 {
		go tryRunHooks(ctx.vira, ctx.Res.endHooks)
	}
}

func handleCtxEnd(ctx *Context) {
	<-ctx.done
	ctx.Res.ended.setTrue()
}

func runHooks(hooks []func()) {
	// run hooks in LIFO order
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i]()
	}
}

func tryRunHooks(vira *Vira, hooks []func()) {
	defer catchErr(vira)
	runHooks(hooks)
}

func catchErr(vira *Vira) {
	if err := recover(); err != nil && err != http.ErrAbortHandler {
		vira.Error(err)
	}
}
