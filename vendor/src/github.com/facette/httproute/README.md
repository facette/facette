[![GoDoc](https://godoc.org/github.com/facette/httproute?status.svg)](https://godoc.org/github.com/facette/httproute)

# httproute: HTTP router

Basic HTTP router for Go.

## Example

The following code:

```go
package main

import (
        "context"
        "fmt"
        "log"
        "net/http"

        "github.com/facette/httproute"
)

func main() {
        r := httproute.NewRouter()

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

func handleBar(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(rw, "Received %q\n", ctx.Value("baz").(string))
}

func handleDefault(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(rw, "Default here!")
}

func handleFoo(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
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
