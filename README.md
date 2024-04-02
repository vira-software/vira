# VIRA Web Framework

## Getting started

### Getting Vira

With [Go module](https://github.com/golang/go/wiki/Modules) support, simply add the following import

```go
import "github.com/vira-software/vira"
```

to your code, and then `go [build|run|test]` will automatically fetch the necessary dependencies.

Otherwise, run the following Go command to install the `vira` package:

```sh
$ go get -u github.com/vira-software/vira
```

### Running Vira

First you need to import Vira package for using Vira, one simplest example likes the follow `example.go`:

```go
package main

import (
  "net/http"

  "github.com/vira-software/vira"
)

func main() {
  r := vira.Default()
  r.GET("/ping", func(c *vira.Context) {
    c.JSON(http.StatusOK, vira.H{
      "message": "pong",
    })
  })
  r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

And use the Go command to run the demo:

```go
# run example.go and visit 0.0.0.0:8080/ping on browser
$ go run example.go
```

## Usage

### Using GET, POST, PUT, PATCH, DELETE and OPTIONS

```go
func main() {
  // Creates a vira router with default middleware:
  // logger and recovery (crash-free) middleware
  router := vira.Default()

  router.GET("/someGet", getting)
  router.POST("/somePost", posting)
  router.PUT("/somePut", putting)
  router.DELETE("/someDelete", deleting)
  router.PATCH("/somePatch", patching)
  router.HEAD("/someHead", head)
  router.OPTIONS("/someOptions", options)

  // By default it serves on :8080 unless a
  // PORT environment variable was defined.
  router.Run()
  // router.Run(":3000") for a hard coded port
}
```

### Parameters in path

```go
func main() {
  router := vira.Default()

  // This handler will match /user/john but will not match /user/ or /user
  router.GET("/user/:name", func(c *vira.Context) {
    name := c.Param("name")
    c.String(http.StatusOK, "Hello %s", name)
  })

  // However, this one will match /user/john/ and also /user/john/send
  // If no other routers match /user/john, it will redirect to /user/john/
  router.GET("/user/:name/*action", func(c *vira.Context) {
    name := c.Param("name")
    action := c.Param("action")
    message := name + " is " + action
    c.String(http.StatusOK, message)
  })

  // For each matched request Context will hold the route definition
  router.POST("/user/:name/*action", func(c *vira.Context) {
    b := c.FullPath() == "/user/:name/*action" // true
    c.String(http.StatusOK, "%t", b)
  })

  // This handler will add a new router for /user/groups.
  // Exact routes are resolved before param routes, regardless of the order they were defined.
  // Routes starting with /user/groups are never interpreted as /user/:name/... routes
  router.GET("/user/groups", func(c *vira.Context) {
    c.String(http.StatusOK, "The available groups are [...]")
  })

  router.Run(":8080")
}
```

### Querystring parameters

```go
func main() {
  router := vira.Default()

  // Query string parameters are parsed using the existing underlying request object.
  // The request responds to an url matching:  /welcome?firstname=Jane&lastname=Doe
  router.GET("/welcome", func(c *vira.Context) {
    firstname := c.DefaultQuery("firstname", "Guest")
    lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

    c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
  })
  router.Run(":8080")
}
```

### Upload files

#### Single file

```go
func main() {
  router := vira.Default()
  // Set a lower memory limit for multipart forms (default is 32 MiB)
  router.MaxMultipartMemory = 8 << 20  // 8 MiB
  router.POST("/upload", func(c *vira.Context) {
    // Single file
    file, _ := c.FormFile("file")
    log.Println(file.Filename)

    // Upload the file to specific dst.
    c.SaveUploadedFile(file, dst)

    c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
  })
  router.Run(":8080")
}
```

How to `curl`:

```bash
curl -X POST http://localhost:8080/upload \
  -F "file=@/Users/appleboy/test.zip" \
  -H "Content-Type: multipart/form-data"
```

#### Multiple files

```go
func main() {
  router := vira.Default()
  // Set a lower memory limit for multipart forms (default is 32 MiB)
  router.MaxMultipartMemory = 8 << 20  // 8 MiB
  router.POST("/upload", func(c *vira.Context) {
    // Multipart form
    form, _ := c.MultipartForm()
    files := form.File["upload[]"]

    for _, file := range files {
      log.Println(file.Filename)

      // Upload the file to specific dst.
      c.SaveUploadedFile(file, dst)
    }
    c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
  })
  router.Run(":8080")
}
```

How to `curl`:

```bash
curl -X POST http://localhost:8080/upload \
  -F "upload[]=@/Users/appleboy/test1.zip" \
  -F "upload[]=@/Users/appleboy/test2.zip" \
  -H "Content-Type: multipart/form-data"
```

### Grouping routes

```go
func main() {
  router := vira.Default()

  // Simple group: v1
  v1 := router.Group("/v1")
  {
    v1.POST("/lovira", loviraEndpoint)
    v1.POST("/submit", submitEndpoint)
    v1.POST("/read", readEndpoint)
  }

  // Simple group: v2
  v2 := router.Group("/v2")
  {
    v2.POST("/lovira", loviraEndpoint)
    v2.POST("/submit", submitEndpoint)
    v2.POST("/read", readEndpoint)
  }

  router.Run(":8080")
}
```

### Blank Vira without middleware by default

Use

```go
r := vira.New()
```

instead of

```go
// Default With the Logger and Recovery middleware already attached
r := vira.Default()
```

### Using middleware

```go
func main() {
  // Creates a router without any middleware by default
  r := vira.New()

  // Global middleware
  // Logger middleware will write the logs to vira.DefaultWriter even if you set with VIRA_MODE=release.
  // By default vira.DefaultWriter = os.Stdout
  r.Use(vira.Logger())

  // Recovery middleware recovers from any panics and writes a 500 if there was one.
  r.Use(vira.Recovery())

  // Per route middleware, you can add as many as you desire.
  r.GET("/benchmark", MyBenchLogger(), benchEndpoint)

  // Authorization group
  // authorized := r.Group("/", AuthRequired())
  // exactly the same as:
  authorized := r.Group("/")
  // per group middleware! in this case we use the custom created
  // AuthRequired() middleware just in the "authorized" group.
  authorized.Use(AuthRequired())
  {
    authorized.POST("/lovira", loviraEndpoint)
    authorized.POST("/submit", submitEndpoint)
    authorized.POST("/read", readEndpoint)

    // nested group
    testing := authorized.Group("testing")
    // visit 0.0.0.0:8080/testing/analytics
    testing.GET("/analytics", analyticsEndpoint)
  }

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### Custom Recovery behavior

```go
func main() {
  // Creates a router without any middleware by default
  r := vira.New()

  // Global middleware
  // Logger middleware will write the logs to vira.DefaultWriter even if you set with VIRA_MODE=release.
  // By default vira.DefaultWriter = os.Stdout
  r.Use(vira.Logger())

  // Recovery middleware recovers from any panics and writes a 500 if there was one.
  r.Use(vira.CustomRecovery(func(c *vira.Context, recovered any) {
    if err, ok := recovered.(string); ok {
      c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
    }
    c.AbortWithStatus(http.StatusInternalServerError)
  }))

  r.GET("/panic", func(c *vira.Context) {
    // panic with a string -- the custom middleware could save this to a database or report it to the user
    panic("foo")
  })

  r.GET("/", func(c *vira.Context) {
    c.String(http.StatusOK, "ohai")
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### How to write log file

```go
func main() {
  // Disable Console Color, you don't need console color when writing the logs to file.
  vira.DisableConsoleColor()

  // Logvirag to a file.
  f, _ := os.Create("vira.log")
  vira.DefaultWriter = io.MultiWriter(f)

  // Use the following code if you need to write the logs to file and console at the same time.
  // vira.DefaultWriter = io.MultiWriter(f, os.Stdout)

  router := vira.Default()
  router.GET("/ping", func(c *vira.Context) {
      c.String(http.StatusOK, "pong")
  })

   router.Run(":8080")
}
```

### Custom Log Format

```go
func main() {
  router := vira.New()

  // LoggerWithFormatter middleware will write the logs to vira.DefaultWriter
  // By default vira.DefaultWriter = os.Stdout
  router.Use(vira.LoggerWithFormatter(func(param vira.LogFormatterParams) string {

    // your custom format
    return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
        param.ClientIP,
        param.TimeStamp.Format(time.RFC1123),
        param.Method,
        param.Path,
        param.Request.Proto,
        param.StatusCode,
        param.Latency,
        param.Request.UserAgent(),
        param.ErrorMessage,
    )
  }))
  router.Use(vira.Recovery())

  router.GET("/ping", func(c *vira.Context) {
    c.String(http.StatusOK, "pong")
  })

  router.Run(":8080")
}
```

Sample Output

```sh
::1 - [Fri, 07 Dec 2018 17:04:38 JST] "GET /ping HTTP/1.1 200 122.767µs "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36" "
```

### Skip logvirag

```go
func main() {
  router := vira.New()
  
  // skip logvirag for desired paths by setting SkipPaths in LoggerConfig
  loggerConfig := vira.LoggerConfig{SkipPaths: []string{"/metrics"}}
  
  // skip logvirag based on your logic by setting Skip func in LoggerConfig
  loggerConfig.Skip = func(c *vira.Context) bool {
      // as an example skip non server side errors
      return c.Writer.Status() < http.StatusInternalServerError
  }
  
  envirae.Use(vira.LoggerWithConfig(loggerConfig))
  router.Use(vira.Recovery())
  
  // skipped
  router.GET("/metrics", func(c *vira.Context) {
      c.Status(http.StatusNotImplemented)
  })

  // skipped
  router.GET("/ping", func(c *vira.Context) {
      c.String(http.StatusOK, "pong")
  })

  // not skipped
  router.GET("/data", func(c *vira.Context) {
    c.Status(http.StatusNotImplemented)
  })
  
  router.Run(":8080")
}

```

### Controlling Log output coloring

By default, logs output on console should be colorized depending on the detected TTY.

Never colorize logs:

```go
func main() {
  // Disable log's color
  vira.DisableConsoleColor()

  // Creates a vira router with default middleware:
  // logger and recovery (crash-free) middleware
  router := vira.Default()

  router.GET("/ping", func(c *vira.Context) {
      c.String(http.StatusOK, "pong")
  })

  router.Run(":8080")
}
```

Always colorize logs:

```go
func main() {
  // Force log's color
  vira.ForceConsoleColor()

  // Creates a vira router with default middleware:
  // logger and recovery (crash-free) middleware
  router := vira.Default()

  router.GET("/ping", func(c *vira.Context) {
      c.String(http.StatusOK, "pong")
  })

  router.Run(":8080")
}
```

### Model binding and validation

To bind a request body into a type, use model binding. We currently support binding of JSON, XML, YAML, TOML and standard form values (foo=bar&boo=baz).

Vira uses [**go-playground/validator/v10**](https://github.com/go-playground/validator) for validation. Check the full docs on tags usage [here](https://pkg.go.dev/github.com/go-playground/validator#hdr-Baked_In_Validators_and_Tags).

Note that you need to set the corresponding binding tag on all fields you want to bind. For example, when binding from JSON, set `json:"fieldname"`.

Also, Vira provides two sets of methods for binding:

- **Methods** - `Bind`, `BindJSON`, `BindXML`, `BindQuery`, `BindYAML`, `BindHeader`, `BindTOML`

You can also specify that specific fields are required. If a field is decorated with `binding:"required"` and has an empty value when binding, an error will be returned.

```go
// Binding from JSON
type Lovira struct {
  User     string `form:"user" json:"user" xml:"user"  binding:"required"`
  Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func main() {
  router := vira.Default()

  // Example for binding JSON ({"user": "vira", "password": "123"})
  router.POST("/loviraJSON", func(c *vira.Context) {
    var json Lovira
    if err := c.BindJSON(&json); err != nil {
      c.JSON(http.StatusBadRequest, vira.H{"error": err.Error()})
      return
    }

    if json.User != "vira" || json.Password != "123" {
      c.JSON(http.StatusUnauthorized, vira.H{"status": "unauthorized"})
      return
    }

    c.JSON(http.StatusOK, vira.H{"status": "you are logged in"})
  })

  // Example for binding XML (
  //  <?xml version="1.0" encoding="UTF-8"?>
  //  <root>
  //    <user>vira</user>
  //    <password>123</password>
  //  </root>)
  router.POST("/loviraXML", func(c *vira.Context) {
    var xml Lovira
    if err := c.BindXML(&xml); err != nil {
      c.JSON(http.StatusBadRequest, vira.H{"error": err.Error()})
      return
    }

    if xml.User != "vira" || xml.Password != "123" {
      c.JSON(http.StatusUnauthorized, vira.H{"status": "unauthorized"})
      return
    }

    c.JSON(http.StatusOK, vira.H{"status": "you are logged in"})
  })

  // Listen and serve on 0.0.0.0:8080
  router.Run(":8080")
}
```

Sample request

```sh
$ curl -v -X POST \
  http://localhost:8080/loviraJSON \
  -H 'content-type: application/json' \
  -d '{ "user": "vira" }'
> POST /loviraJSON HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.51.0
> Accept: */*
> content-type: application/json
> Content-Length: 18
>
* upload completely sent off: 18 out of 18 bytes
< HTTP/1.1 400 Bad Request
< Content-Type: application/json; charset=utf-8
< Date: Fri, 04 Aug 2017 03:51:31 GMT
< Content-Length: 100
<
{"error":"Key: 'Lovira.Password' Error:Field validation for 'Password' failed on the 'required' tag"}
```

Skip validate: when running the above example using the above the `curl` command, it returns error. Because the example use `binding:"required"` for `Password`. If use `binding:"-"` for `Password`, then it will not return error when running the above example again.

### Custom Validators

```go
package main

import (
  "net/http"
  "time"

  "github.com/vira-software/vira"
  "github.com/vira-software/vira/binding"
  "github.com/go-playground/validator/v10"
)

// Booking contains binded and validated data.
type Booking struct {
  CheckIn  time.Time `form:"check_in" binding:"required,bookabledate" time_format:"2006-01-02"`
  CheckOut time.Time `form:"check_out" binding:"required,gtfield=CheckIn" time_format:"2006-01-02"`
}

var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
  date, ok := fl.Field().Interface().(time.Time)
  if ok {
    today := time.Now()
    if today.After(date) {
      return false
    }
  }
  return true
}

func main() {
  route := vira.Default()

  if v, ok := binding.Validator.Envirae().(*validator.Validate); ok {
    v.RegisterValidation("bookabledate", bookableDate)
  }

  route.GET("/bookable", getBookable)
  route.Run(":8085")
}

func getBookable(c *vira.Context) {
  var b Booking
  if err := c.BindWith(&b, binding.Query); err == nil {
    c.JSON(http.StatusOK, vira.H{"message": "Booking dates are valid!"})
  } else {
    c.JSON(http.StatusBadRequest, vira.H{"error": err.Error()})
  }
}
```

```console
$ curl "localhost:8085/bookable?check_in=2030-04-16&check_out=2030-04-17"
{"message":"Booking dates are valid!"}

$ curl "localhost:8085/bookable?check_in=2030-03-10&check_out=2030-03-09"
{"error":"Key: 'Booking.CheckOut' Error:Field validation for 'CheckOut' failed on the 'gtfield' tag"}

$ curl "localhost:8085/bookable?check_in=2000-03-09&check_out=2000-03-10"
{"error":"Key: 'Booking.CheckIn' Error:Field validation for 'CheckIn' failed on the 'bookabledate' tag"}%
```

[Struct level validations](https://github.com/go-playground/validator/releases/tag/v8.7) can also be registered this way.

### Only Bind Query String

```go
package main

import (
  "log"
  "net/http"

  "github.com/vira-software/vira"
)

type Person struct {
  Name    string `form:"name"`
  Address string `form:"address"`
}

func main() {
  route := vira.Default()
  route.Any("/testing", startPage)
  route.Run(":8085")
}

func startPage(c *vira.Context) {
  var person Person
  if c.BindQuery(&person) == nil {
    log.Println("====== Only Bind By Query String ======")
    log.Println(person.Name)
    log.Println(person.Address)
  }
  c.String(http.StatusOK, "Success")
}

```

### Bind Query String or Post Data

```go
package main

import (
  "log"
  "net/http"
  "time"

  "github.com/vira-software/vira"
)

type Person struct {
  Name       string    `form:"name"`
  Address    string    `form:"address"`
  Birthday   time.Time `form:"birthday" time_format:"2006-01-02" time_utc:"1"`
  CreateTime time.Time `form:"createTime" time_format:"unixNano"`
  UnixTime   time.Time `form:"unixTime" time_format:"unix"`
}

func main() {
  route := vira.Default()
  route.GET("/testing", startPage)
  route.Run(":8085")
}

func startPage(c *vira.Context) {
  var person Person
  if c.Bind(&person) == nil {
    log.Println(person.Name)
    log.Println(person.Address)
    log.Println(person.Birthday)
    log.Println(person.CreateTime)
    log.Println(person.UnixTime)
  }

  c.String(http.StatusOK, "Success")
}
```

Test it with:

```sh
curl -X GET "localhost:8085/testing?name=appleboy&address=xyz&birthday=1992-03-15&createTime=1562400033000000123&unixTime=1562400033"
```

### Bind Uri

```go
package main

import (
  "net/http"

  "github.com/vira-software/vira"
)

type Person struct {
  ID string `uri:"id" binding:"required,uuid"`
  Name string `uri:"name" binding:"required"`
}

func main() {
  route := vira.Default()
  route.GET("/:name/:id", func(c *vira.Context) {
    var person Person
    if err := c.BindUri(&person); err != nil {
      c.JSON(http.StatusBadRequest, vira.H{"msg": err.Error()})
      return
    }
    c.JSON(http.StatusOK, vira.H{"name": person.Name, "uuid": person.ID})
  })
  route.Run(":8088")
}
```

Test it with:

```sh
curl -v localhost:8088/thinkerou/987fbc97-4bed-5078-9f07-9141ba07c9f3
curl -v localhost:8088/thinkerou/not-uuid
```

### Bind Header

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/vira-software/vira"
)

type testHeader struct {
  Rate   int    `header:"Rate"`
  Domain string `header:"Domain"`
}

func main() {
  r := vira.Default()
  r.GET("/", func(c *vira.Context) {
    h := testHeader{}

    if err := c.BindHeader(&h); err != nil {
      c.JSON(http.StatusOK, err)
    }

    fmt.Printf("%#v\n", h)
    c.JSON(http.StatusOK, vira.H{"Rate": h.Rate, "Domain": h.Domain})
  })

  r.Run()

// client
// curl -H "rate:300" -H "domain:music" 127.0.0.1:8080/
// output
// {"Domain":"music","Rate":300}
}
```

### XML, JSON, YAML, TOML and ProtoBuf rendering

```go
func main() {
  r := vira.Default()

  // vira.H is a shortcut for map[string]any
  r.GET("/someJSON", func(c *vira.Context) {
    c.JSON(http.StatusOK, vira.H{"message": "hey", "status": http.StatusOK})
  })

  r.GET("/moreJSON", func(c *vira.Context) {
    // You also can use a struct
    var msg struct {
      Name    string `json:"user"`
      Message string
      Number  int
    }
    msg.Name = "Lena"
    msg.Message = "hey"
    msg.Number = 123
    // Note that msg.Name becomes "user" in the JSON
    // Will output  :   {"user": "Lena", "Message": "hey", "Number": 123}
    c.JSON(http.StatusOK, msg)
  })

  r.GET("/someXML", func(c *vira.Context) {
    c.XML(http.StatusOK, vira.H{"message": "hey", "status": http.StatusOK})
  })

  r.GET("/someYAML", func(c *vira.Context) {
    c.YAML(http.StatusOK, vira.H{"message": "hey", "status": http.StatusOK})
  })

  r.GET("/someTOML", func(c *vira.Context) {
    c.TOML(http.StatusOK, vira.H{"message": "hey", "status": http.StatusOK})
  })

  r.GET("/someProtoBuf", func(c *vira.Context) {
    reps := []int64{int64(1), int64(2)}
    label := "test"
    // The specific definition of protobuf is written in the testdata/protoexample file.
    data := &protoexample.Test{
      Label: &label,
      Reps:  reps,
    }
    // Note that data becomes binary data in the response
    // Will output protoexample.Test protobuf serialized data
    c.ProtoBuf(http.StatusOK, data)
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

#### SecureJSON

Using SecureJSON to prevent json hijacking. Default prepends `"while(1),"` to response body if the given struct is array values.

```go
func main() {
  r := vira.Default()

  // You can also use your own secure json prefix
  // r.SecureJsonPrefix(")]}',\n")

  r.GET("/someJSON", func(c *vira.Context) {
    names := []string{"lena", "austin", "foo"}

    // Will output  :   while(1);["lena","austin","foo"]
    c.SecureJSON(http.StatusOK, names)
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### Serving static files

```go
func main() {
  router := vira.Default()
  router.Static("/assets", "./assets")
  router.StaticFS("/more_static", http.Dir("my_file_system"))
  router.StaticFile("/favicon.ico", "./resources/favicon.ico")
  router.StaticFileFS("/more_favicon.ico", "more_favicon.ico", http.Dir("my_file_system"))
  
  // Listen and serve on 0.0.0.0:8080
  router.Run(":8080")
}
```

### Serving data from file

```go
func main() {
  router := vira.Default()

  router.GET("/local/file", func(c *vira.Context) {
    c.File("local/file.go")
  })

  var fs http.FileSystem = // ...
  router.GET("/fs/file", func(c *vira.Context) {
    c.FileFromFS("fs/file.go", fs)
  })
}

```

### Serving data from reader

```go
func main() {
  router := vira.Default()
  router.GET("/someDataFromReader", func(c *vira.Context) {
    response, err := http.Get("https://raw.githubusercontent.com/vira-software/logo/master/color.png")
    if err != nil || response.StatusCode != http.StatusOK {
      c.Status(http.StatusServiceUnavailable)
      return
    }

    reader := response.Body
     defer reader.Close()
    contentLength := response.ContentLength
    contentType := response.Header.Get("Content-Type")

    extraHeaders := map[string]string{
      "Content-Disposition": `attachment; filename="gopher.png"`,
    }

    c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
  })
  router.Run(":8080")
}
```

### Redirects

Issuing a HTTP redirect is easy. Both internal and external locations are supported.

```go
r.GET("/test", func(c *vira.Context) {
  c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
})
```

Issuing a HTTP redirect from POST. Refer to issue: [#444](https://github.com/vira-software/vira/issues/444)

```go
r.POST("/test", func(c *vira.Context) {
  c.Redirect(http.StatusFound, "/foo")
})
```

Issuing a Router redirect, use `HandleContext` like below.

``` go
r.GET("/test", func(c *vira.Context) {
    c.Request.URL.Path = "/test2"
    r.HandleContext(c)
})
r.GET("/test2", func(c *vira.Context) {
    c.JSON(http.StatusOK, vira.H{"hello": "world"})
})
```

### Custom Middleware

```go
func Logger() vira.HandlerFunc {
  return func(c *vira.Context) {
    t := time.Now()

    // Set example variable
    c.Set("example", "12345")

    // before request

    c.Next()

    // after request
    latency := time.Since(t)
    log.Print(latency)

    // access the status we are sending
    status := c.Writer.Status()
    log.Println(status)
  }
}

func main() {
  r := vira.New()
  r.Use(Logger())

  r.GET("/test", func(c *vira.Context) {
    example := c.MustGet("example").(string)

    // it would print: "12345"
    log.Println(example)
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### Using BasicAuth() middleware

```go
// simulate some private data
var secrets = vira.H{
  "foo":    vira.H{"email": "foo@bar.com", "phone": "123433"},
  "lena":   vira.H{"email": "lena@aaa.com", "phone": "523443"},
}

func main() {
  r := vira.Default()

  // Group using vira.BasicAuth() middleware
  // vira.Accounts is a shortcut for map[string]string
  authorized := r.Group("/admin", vira.BasicAuth(vira.Accounts{
    "foo":    "bar",
    "lena":   "hello2",
  }))

  // /admin/secrets endpoint
  // hit "localhost:8080/admin/secrets
  authorized.GET("/secrets", func(c *vira.Context) {
    // get user, it was set by the BasicAuth middleware
    user := c.MustGet(vira.AuthUserKey).(string)
    if secret, ok := secrets[user]; ok {
      c.JSON(http.StatusOK, vira.H{"user": user, "secret": secret})
    } else {
      c.JSON(http.StatusOK, vira.H{"user": user, "secret": "NO SECRET :("})
    }
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### Goroutines inside a middleware

When starting new Goroutines inside a middleware or handler, you **SHOULD NOT** use the oriviraal context inside it, you have to use a read-only copy.

```go
func main() {
  r := vira.Default()

  r.GET("/long_async", func(c *vira.Context) {
    // create copy to be used inside the goroutine
    cCp := c.Copy()
    go func() {
      // simulate a long task with time.Sleep(). 5 seconds
      time.Sleep(5 * time.Second)

      // note that you are using the copied context "cCp", IMPORTANT
      log.Println("Done! in path " + cCp.Request.URL.Path)
    }()
  })

  r.GET("/long_sync", func(c *vira.Context) {
    // simulate a long task with time.Sleep(). 5 seconds
    time.Sleep(5 * time.Second)

    // since we are NOT using a goroutine, we do not have to copy the context
    log.Println("Done! in path " + c.Request.URL.Path)
  })

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}
```

### Custom HTTP configuration

Use `http.ListenAndServe()` directly, like this:

```go
func main() {
  router := vira.Default()
  http.ListenAndServe(":8080", router)
}
```

or

```go
func main() {
  router := vira.Default()

  s := &http.Server{
    Addr:           ":8080",
    Handler:        router,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }
  s.ListenAndServe()
}
```

### Run multiple service using Vira

```go
package main

import (
  "log"
  "net/http"
  "time"

  "github.com/vira-software/vira"
  "golang.org/x/sync/errgroup"
)

var (
  g errgroup.Group
)

func router01() http.Handler {
  e := vira.New()
  e.Use(vira.Recovery())
  e.GET("/", func(c *vira.Context) {
    c.JSON(
      http.StatusOK,
      vira.H{
        "code":  http.StatusOK,
        "error": "Welcome server 01",
      },
    )
  })

  return e
}

func router02() http.Handler {
  e := vira.New()
  e.Use(vira.Recovery())
  e.GET("/", func(c *vira.Context) {
    c.JSON(
      http.StatusOK,
      vira.H{
        "code":  http.StatusOK,
        "error": "Welcome server 02",
      },
    )
  })

  return e
}

func main() {
  server01 := &http.Server{
    Addr:         ":8080",
    Handler:      router01(),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
  }

  server02 := &http.Server{
    Addr:         ":8081",
    Handler:      router02(),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
  }

  g.Go(func() error {
    err := server01.ListenAndServe()
    if err != nil && err != http.ErrServerClosed {
      log.Fatal(err)
    }
    return err
  })

  g.Go(func() error {
    err := server02.ListenAndServe()
    if err != nil && err != http.ErrServerClosed {
      log.Fatal(err)
    }
    return err
  })

  if err := g.Wait(); err != nil {
    log.Fatal(err)
  }
}
```

### Graceful shutdown or restart

There are a few approaches you can use to perform a graceful shutdown or restart. You can make use of third-party packages specifically built for that, or you can manually do the same with the functions and methods from the built-in packages.

```go
// +build go1.8

package main

import (
  "context"
  "log"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/vira-software/vira"
)

func main() {
  router := vira.Default()
  router.GET("/", func(c *vira.Context) {
    time.Sleep(5 * time.Second)
    c.String(http.StatusOK, "Welcome Vira Server")
  })

  srv := &http.Server{
    Addr:    ":8080",
    Handler: router,
  }

  // Initializing the server in a goroutine so that
  // it won't block the graceful shutdown handling below
  go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
      log.Printf("listen: %s\n", err)
    }
  }()

  // Wait for interrupt signal to gracefully shutdown the server with
  // a timeout of 5 seconds.
  quit := make(chan os.Signal)
  // kill (no param) default send syscall.SIGTERM
  // kill -2 is syscall.SIVIRAT
  // kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
  signal.Notify(quit, syscall.SIVIRAT, syscall.SIGTERM)
  <-quit
  log.Println("Shutting down server...")

  // The context is used to inform the server it has 5 seconds to finish
  // the request it is currently handling
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  if err := srv.Shutdown(ctx); err != nil {
    log.Fatal("Server forced to shutdown:", err)
  }

  log.Println("Server exiting")
}
```

### Bind form-data request with custom struct

The follow example using custom struct:

```go
type StructA struct {
    FieldA string `form:"field_a"`
}

type StructB struct {
    NestedStruct StructA
    FieldB string `form:"field_b"`
}

type StructC struct {
    NestedStructPointer *StructA
    FieldC string `form:"field_c"`
}

type StructD struct {
    NestedAnonyStruct struct {
        FieldX string `form:"field_x"`
    }
    FieldD string `form:"field_d"`
}

func GetDataB(c *vira.Context) {
    var b StructB
    c.Bind(&b)
    c.JSON(http.StatusOK, vira.H{
        "a": b.NestedStruct,
        "b": b.FieldB,
    })
}

func GetDataC(c *vira.Context) {
    var b StructC
    c.Bind(&b)
    c.JSON(http.StatusOK, vira.H{
        "a": b.NestedStructPointer,
        "c": b.FieldC,
    })
}

func GetDataD(c *vira.Context) {
    var b StructD
    c.Bind(&b)
    c.JSON(http.StatusOK, vira.H{
        "x": b.NestedAnonyStruct,
        "d": b.FieldD,
    })
}

func main() {
    r := vira.Default()
    r.GET("/getb", GetDataB)
    r.GET("/getc", GetDataC)
    r.GET("/getd", GetDataD)

    r.Run()
}
```

Using the command `curl` command result:

```sh
$ curl "http://localhost:8080/getb?field_a=hello&field_b=world"
{"a":{"FieldA":"hello"},"b":"world"}
$ curl "http://localhost:8080/getc?field_a=hello&field_c=world"
{"a":{"FieldA":"hello"},"c":"world"}
$ curl "http://localhost:8080/getd?field_x=hello&field_d=world"
{"d":"world","x":{"FieldX":"hello"}}
```

### http2 server push

http.Pusher is supported only **go1.8+**. See the [golang blog](https://go.dev/blog/h2push) for detail information.

```go
package main

import (
  "log"
  "net/http"

  "github.com/vira-software/vira"
)

func main() {
  r := vira.Default()

  r.GET("/", func(c *vira.Context) {
    c.JSON(http.StatusOK, vira.H{
      "status": "success",
    })
  })

  // Listen and Server in https://127.0.0.1:8080
  r.RunTLS(":8080", "./testdata/server.pem", "./testdata/server.key")
}
```

### Define format for the log of routes

The default log of routes is:

```sh
[VIRA-debug] POST   /foo                      --> main.main.func1 (3 handlers)
[VIRA-debug] GET    /bar                      --> main.main.func2 (3 handlers)
[VIRA-debug] GET    /status                   --> main.main.func3 (3 handlers)
```

If you want to log this information in given format (e.g. JSON, key values or something else), then you can define this format with `vira.DebugPrintRouteFunc`.
In the example below, we log all routes with standard log package but you can use another log tools that suits of your needs.

```go
import (
  "log"
  "net/http"

  "github.com/vira-software/vira"
)

func main() {
  r := vira.Default()
  vira.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
    log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
  }

  r.POST("/foo", func(c *vira.Context) {
    c.JSON(http.StatusOK, "foo")
  })

  r.GET("/bar", func(c *vira.Context) {
    c.JSON(http.StatusOK, "bar")
  })

  r.GET("/status", func(c *vira.Context) {
    c.JSON(http.StatusOK, "ok")
  })

  // Listen and Server in http://0.0.0.0:8080
  r.Run()
}
```

### Set and get a cookie

```go
import (
  "fmt"

  "github.com/vira-software/vira"
)

func main() {
  router := vira.Default()

  router.GET("/cookie", func(c *vira.Context) {

      cookie, err := c.Cookie("vira_cookie")

      if err != nil {
          cookie = "NotSet"
          c.SetCookie("vira_cookie", "test", 3600, "/", "localhost", false, true)
      }

      fmt.Printf("Cookie value: %s \n", cookie)
  })

  router.Run()
}
```

## Don't trust all proxies

Vira lets you specify which headers to hold the real client IP (if any),
as well as specifying which proxies (or direct clients) you trust to
specify one of these headers.

Use function `SetTrustedProxies()` on your `vira.Vira` to specify network addresses
or network CIDRs from where clients which their request headers related to client
IP can be trusted. They can be IPv4 addresses, IPv4 CIDRs, IPv6 addresses or
IPv6 CIDRs.

**Attention:** Vira trust all proxies by default if you don't specify a trusted 
proxy using the function above, **this is NOT safe**. At the same time, if you don't
use any proxy, you can disable this feature by using `Vira.SetTrustedProxies(nil)`,
then `Context.ClientIP()` will return the remote address directly to avoid some
unnecessary computation.

```go
import (
  "fmt"

  "github.com/vira-software/vira"
)

func main() {
  router := vira.Default()
  router.SetTrustedProxies([]string{"192.168.1.2"})

  router.GET("/", func(c *vira.Context) {
    // If the client is 192.168.1.2, use the X-Forwarded-For
    // header to deduce the oriviraal client IP from the trust-
    // worthy parts of that header.
    // Otherwise, simply return the direct client IP
    fmt.Printf("ClientIP: %s\n", c.ClientIP())
  })
  router.Run()
}
```

**Notice:** If you are using a CDN service, you can set the `Vira.TrustedPlatform`
to skip TrustedProxies check, it has a higher priority than TrustedProxies. 
Look at the example below:

```go
import (
  "fmt"

  "github.com/vira-software/vira"
)

func main() {
  router := vira.Default()
  // Use predefined header vira.PlatformXXX
  // Google App Vira
  router.TrustedPlatform = vira.PlatformGoogleAppVira
  // Cloudflare
  router.TrustedPlatform = vira.PlatformCloudflare
  // Fly.io
  router.TrustedPlatform = vira.PlatformFlyIO
  // Or, you can set your own trusted request header. But be sure your CDN
  // prevents users from passing this header! For example, if your CDN puts
  // the client IP in X-CDN-Client-IP:
  router.TrustedPlatform = "X-CDN-Client-IP"

  router.GET("/", func(c *vira.Context) {
    // If you set TrustedPlatform, ClientIP() will resolve the
    // corresponding header and return IP directly
    fmt.Printf("ClientIP: %s\n", c.ClientIP())
  })
  router.Run()
}
```

### Support Let's Encrypt

example for 1-line LetsEncrypt HTTPS servers.

```go
package main

import (
  "log"
  "net/http"

  "github.com/vira-software/vira/tls"
  "github.com/vira-software/vira"
)

func main() {
  r := vira.Default()

  // Ping handler
  r.GET("/ping", func(c *vira.Context) {
    c.String(http.StatusOK, "pong")
  })

  log.Fatal(autotls.Run(r, "example1.com", "example2.com"))
}
```

example for custom autocert manager.

```go
package main

import (
  "log"
  "net/http"

  "github.com/vira-gonic/autotls"
  "github.com/vira-software/vira"
  "golang.org/x/crypto/acme/autocert"
)

func main() {
  r := vira.Default()

  // Ping handler
  r.GET("/ping", func(c *vira.Context) {
    c.String(http.StatusOK, "pong")
  })

  m := autocert.Manager{
    Prompt:     autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist("example1.com", "example2.com"),
    Cache:      autocert.DirCache("/var/www/.cache"),
  }

  log.Fatal(autotls.RunWithManager(r, &m))
}
```
