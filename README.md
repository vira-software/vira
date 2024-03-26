# VIRA

## Trie router
A minimal trie based url path router.

### Features
- Support named parameter
- Support regexp
- Support suffix matching
- Fixed path automatic redirection
- Trailing slash automatic redirection
- Automatic handle 405 Method Not Allowed
- Automatic handle 501 Not Implemented
- Automatic handle OPTIONS method
  
### Usage
```go
func main() {
  router := router.New()
  router.Get("/", func(w http.ResponseWriter, _ *http.Request, _ router.Params) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("<h1>Hello, Gear!</h1>"))
  })

  router.Get("/view/:view", func(w http.ResponseWriter, _ *http.Request, params router.Params) {
    view := params["view"]
    if view == "" {
      http.Error(w, "Invalid view", 400)
    } else {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      w.WriteHeader(200)
      w.Write([]byte("View: " + view))
    }
  })

  // srv := http.Server{Addr: ":3000", Handler: router}
  // srv.ListenAndServe()
  srv := httptest.NewServer(router)
  defer srv.Close()

  res, _ := http.Get(srv.URL + "/view/users")
  body, _ := ioutil.ReadAll(res.Body)
  res.Body.Close()

  fmt.Println(res.StatusCode, string(body))
  // Output: 200 View: users
}
```

### Pattern Rule

The defined pattern can contain six types of parameters:

| Syntax | Description |
|--------|------|
| `:name` | named parameter |
| `:name(regexp)` | named with regexp parameter |
| `:name+suffix` | named parameter with suffix matching |
| `:name(regexp)+suffix` | named with regexp parameter and suffix matching |
| `:name*` | named with catch-all parameter |
| `::name` | not named parameter, it is literal `:name` |

Named parameters are dynamic path segments. They match anything until the next '/' or the path end:

Defined: `/api/:type/:ID`
```
/api/user/123             matched: type="user", ID="123"
/api/user                 no match
/api/user/123/comments    no match
```

Named with regexp parameters match anything using regexp until the next '/' or the path end:

Defined: `/api/:type/:ID(^\d+$)`
```
/api/user/123             matched: type="user", ID="123"
/api/user                 no match
/api/user/abc             no match
/api/user/123/comments    no match
```

Named parameters with suffix, such as [Google API Design](https://cloud.google.com/apis/design/custom_methods):

Defined: `/api/:resource/:ID+:undelete`
```
/api/file/123                     no match
/api/file/123:undelete            matched: resource="file", ID="123"
/api/file/123:undelete/comments   no match
```

Named with regexp parameters and suffix:

Defined: `/api/:resource/:ID(^\d+$)+:cancel`
```
/api/task/123                   no match
/api/task/123:cancel            matched: resource="task", ID="123"
/api/task/abc:cancel            no match
```

Named with catch-all parameters match anything until the path end, including the directory index (the '/' before the catch-all). Since they match anything until the end, catch-all parameters must always be the final path element.

Defined: `/files/:filepath*`
```
/files                           no match
/files/LICENSE                   matched: filepath="LICENSE"
/files/templates/article.html    matched: filepath="templates/article.html"
```

The value of parameters is saved on the `Matched.Params`. Retrieve the value of a parameter by name:
```
type := matched.Params("type")
id   := matched.Params("ID")
```

Url query string with `?` can be provided when defining trie, but it will be ignored.

Defined: `/files?pageSize=&pageToken=`
Equal to: `/files`
```
/files                           matched, query string will be ignored
/files/LICENSE                   no match
```