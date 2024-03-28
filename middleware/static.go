package middleware

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vira-software/vira"
)

// StaticOptions is static middleware options
type StaticOptions struct {
	Root        string            // The directory you wish to serve
	Prefix      string            // The url prefix you wish to serve as static request, default to `'/'`.
	StripPrefix bool              // Strip the prefix from URL path, default to `false`.
	Includes    []string          // Optional, a slice of file path to serve, it will ignore Prefix and StripPrefix options.
	Files       map[string][]byte // Optional, a map of File objects to serve.
	OnlyFiles   bool              // Optional, if Options.Files provided and Options.OnlyFiles is true, it will not seek files in other way.
}

// NewStaticMiddleware creates a static middleware to serves static content from the provided root directory.
//
//	package main
//
//	func main() {
//		app := vira.New()
//		app.Use(NewFaviconMiddleware("./assets/favicon.ico"))
//		app.Use(NewStaticMiddleware(static.Options{
//			Root:        "./assets",
//			Prefix:      "/assets",
//			StripPrefix: false,
//			Includes:    []string{"/robots.txt"},
//		}))
//		app.Use(func(ctx *vira.Context) error {
//			return ctx.HTML(200, "<h1>Hello, Vira!</h1>")
//		})
//		app.Error(app.Listen(":3000"))
//	}
func NewStaticMiddleware(opts StaticOptions) vira.Middleware {
	modTime := time.Now()
	if opts.Root == "" {
		opts.Root = "."
	}
	root := filepath.FromSlash(opts.Root)
	if root[0] == '.' {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		root = filepath.Join(wd, root)
	}
	info, _ := os.Stat(root)
	if info == nil || !info.IsDir() {
		panic(vira.ViraErr.WithMsgf("invalid root path: %s", root))
	}

	if opts.Prefix == "" {
		opts.Prefix = "/"
	}

	return func(ctx *vira.Context) (err error) {
		path := ctx.Path

		switch {
		case includes(opts.Includes, path): // do nothing
		case strings.HasPrefix(path, opts.Prefix):
			if opts.StripPrefix {
				path = strings.TrimPrefix(path, opts.Prefix)
			}
		default:
			return nil
		}

		if ctx.Method != http.MethodGet && ctx.Method != http.MethodHead {
			status := 200
			if ctx.Method != http.MethodOptions {
				status = 405
			}
			ctx.SetHeader(vira.HeaderContentType, "text/plain; charset=utf-8")
			ctx.SetHeader(vira.HeaderAllow, "GET, HEAD, OPTIONS")
			return ctx.End(status)
		}

		if opts.Files != nil {
			if file, ok := opts.Files[path]; ok {
				http.ServeContent(ctx.Res, ctx.Req, path, modTime, bytes.NewReader(file))
				return nil
			}
			if opts.OnlyFiles {
				return vira.ErrNotFound.WithMsgf("%s could not be found", path)
			}
		}
		path = filepath.Join(root, filepath.FromSlash(path))
		http.ServeFile(ctx.Res, ctx.Req, path)
		return nil
	}
}

func includes(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
