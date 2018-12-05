# httprouter: HTTP router [![GoDoc][godoc-badge]][godoc-url] [![Travis CI][travis-badge]][travis-url]

Basic HTTP router for Go.

## Example

The following code:

```go
package main

import (
        "fmt"
        "log"
        "net/http"

        "github.com/vbatoufflet/httprouter"
)

func main() {
        r := httprouter.NewRouter()

        r.Endpoint("/foo").
                Get(handleFoo).
                Post(handleFoo)

        r.Endpoint("/bar/:baz").
                Get(handleBar)

        r.Endpoint("/*").
                Get(handleDefault)

        s := &http.Server{
                Addr:    ":8080",
                Handler: r,
        }

        if err := s.ListenAndServe(); err != nil {
                log.Fatal(err)
        }
}

func handleBar(rw http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(rw, "Received %q\n", httprouter.ContextParam(r, "baz").(string))
}

func handleDefault(rw http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(rw, "Default here!")
}

func handleFoo(rw http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(rw, "Received %q request\n", r.Method)
}
```

will give you:

```
# curl http://localhost:8080/foo
Received "GET" request

# curl -X POST http://localhost:8080/foo
Received "POST" request

# curl http://localhost:8080/bar/baz
Received "baz"

# curl http://localhost:8080/
Default here!
```

[godoc-badge]: https://godoc.org/github.com/vbatoufflet/httprouter?status.svg
[godoc-url]: https://godoc.org/github.com/vbatoufflet/httprouter
[travis-badge]: https://api.travis-ci.org/vbatoufflet/httprouter.svg
[travis-url]: https://travis-ci.org/vbatoufflet/httprouter
