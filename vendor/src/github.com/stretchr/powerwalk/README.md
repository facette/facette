# Powerwalk

Go package for walking files and concurrently calling user code to handle each file.  This package walks the file system in the same way `filepath.Walk` does, except instead of calling the `walkFn` inline, it uses goroutines to allow the files to be handled concurrently.

Powerwalk functions by walking concurrently over many files. In order to realize any benefits from this approach, you must tell the runtime to use multiple CPUs. For example:

```
runtime.GOMAXPROCS(runtime.NumCPU())
```

## Usage

Powerwalk is a drop-in replacement for the `filepath.Walk` method ([read about that for more details](http://golang.org/pkg/path/filepath/#Walk)), and so has the same signature, even using the `filepath.WalkFunc` too.

```
powerwalk.Walk(root string, walkFn filepath.WalkFunc) error
```

By default, Powerwalk will call the `walkFn` for `powerwalk.DefaultConcurrentWalks` (currently `100`) files at a time.  To be specific about the number of concurrent files to walk, use the `WalkLimit` alternative.

```
powerwalk.WalkLimit(root string, walkFn filepath.WalkFunc, limit int) error
```

The `WalkLimit` function does the same as `Walk`, except allows you to specify the number of files to concurrently walk using the `limit` argument.  The `limit` argument must be one or higher (i.e. `>0`).  Specificying a limit that's too high, causes unnecessary overhead so sensible numbers are encouraged but not enforced.

See the [godoc documentation](http://godoc.org/github.com/stretchr/powerwalk) for more information.

