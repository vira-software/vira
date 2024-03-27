# VIRA
A high performance web service framework for Go.

### Features
- Trie base router with for regexp parameters and route groups.
- Integrated request body parser.
- Integrated middlewares such as CORS, Secure, Static and Logging.
- Integrated renderers such as  HTML JSON, JSONP and XML.

### Install
```go
go get "http://github.com/vira-software/vira"
```

### Usage
```go
 app := vira.New()

  // Add logging middleware
  app.UseHandler(logging.Default(true))

  // Add router middleware
  router := vira.NewRouter()

  // try: http://127.0.0.1:3000/hello
  router.Get("/hello", func(ctx *vira.Context) error {
    return ctx.HTML(200, "<h1>Hello, Vira!</h1>")
  })

  // try: http://127.0.0.1:3000/test?query=hello+world
  router.Otherwise(func(ctx *vira.Context) error {
    return ctx.JSON(200, map[string]any{
      "Host":    ctx.Host,
      "Method":  ctx.Method,
      "Path":    ctx.Path,
      "URI":     ctx.Req.RequestURI,
      "Headers": ctx.Req.Header,
    })
  })
  app.UseHandler(router)
  app.Error(app.Listen(":3000"))
```

### Router
Router is a trie base HTTP request handler:
- Support named parameter
- Support regexp
- Support suffix matching
- Support multi-router
- Support router layer middlewares
- Support fixed path automatic redirection
- Support trailing slash automatic redirection
- Automatic handle ``405 Method Not Allowed``
- Automatic handle ``OPTIONS`` method
- Made to be fast!!!!!!


The registered path can contain six types of parameters:

| Syntax                 | Description                                     |
| ---------------------- | ----------------------------------------------- |
| `:name`                | named parameter                                 |
| `:name(regexp)`        | named with regexp parameter                     |
| `:name+suffix`         | named parameter with suffix matching            |
| `:name(regexp)+suffix` | named with regexp parameter and suffix matching |
| `:name*`               | named with catch-all parameter                  |
| `::name`               | not named parameter, it is literal `:name`      |

- Named parameters are dynamic path segments. They match anything until the next '/' or the path end:

  Defined: `/api/:type/:ID`

  ```md
  /api/user                 no match
  /api/user/123             matched: type="user", ID="123"
  /api/user/123/posts       no match
  ```

- Named with regexp parameters match anything using regexp until the next '/' or the path end:

  Defined: `/api/:type/:ID(^\d+$)`

  ```md
  /api/user                 no match
  /api/user/abc             no match
  /api/user/123             matched: type="user", ID="123"
  /api/user/123/posts       no match
  ```

- Named parameters with suffix, such as [Google API Design](https://cloud.google.com/apis/design/custom_methods):

  Defined: `/api/:resource/:ID+:undelete`

  ```md
  /api/file/123                     no match
  /api/file/123:undelete            matched: resource="file", ID="123"
  /api/file/123:undelete/posts      no match
  ```

- Named with regexp parameters and suffix:

  Defined: `/api/:resource/:ID(^\d+$)+:cancel`

  ```md
  /api/task/123                   no match
  /api/task/123:cancel            matched: resource="task", ID="123"
  /api/task/abc:cancel            no match
  ```

- Named with catch-all parameters match anything until the path end, including the directory index (the '/' before the catch-all). Since they match anything until the end, catch-all parameters must always be the final path element.

  Defined: `/files/:filepath*`

  ```
  /files                           no match
  /files/file1                     matched: filepath="file1"
  /files/templates/post.html       matched: filepath="templates/post.html"
  ```

- The value of parameters is saved on the `Matched.Params`. Retrieve the value of a parameter by name:

  ```go
  type := matched.Params("type")
  id   := matched.Params("ID")
  ```