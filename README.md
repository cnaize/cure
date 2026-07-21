# Cure

`cure` — Pure Go API-Shield based on YARA rules

---

## Installation

```bash
go get github.com/cnaize/cure
```

---

## Web Frameworks Integration and Manual Scanning

```go
package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	curec "github.com/cnaize/cure/core"
	curem "github.com/cnaize/cure/middleware"
)

func main() {
	cure := curec.NewCure()
	// for native net/http, Chi, Echo, Fiber, etc.
	// or use ".GinHandler()" for Gin framework
	cureHandler := curem.NewCure(cure).
		WithOptions(&curem.Options{
			ScanNeeded:  curem.ScanNeededFull,
			ScanTimeout: 10 * time.Millisecond,
		}).HTTPHandler

	// auto update rules
	_ = cure.Run(context.Background())

	yourHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// you can also scan any raw data independently
		some := []byte("some data")
		err := cure.Scan(some, 0, time.Second, curec.DefaultCallback)
		if errors.Is(err, curec.ErrMatchFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// your business logic
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	// cure automatically scans incoming requests before your handler
	mux.Handle("/api/some", cureHandler(yourHandler))

	_ = http.ListenAndServe(":8080", mux)
}
```

---

## Custom Rules Loading

```go
package main

import (
	"context"

	curec "github.com/cnaize/cure/core"
	cures "github.com/cnaize/cure/source"
	curea "github.com/cnaize/cure/source/adapter"
)

func main() {
	// supports both local and remote files (yara, crs) and all kinds of archives
	cure := curec.NewCure().WithSources(
		cures.NewLocal("./my-yara-rules.zip"),
		cures.NewRemote("https://my-company/crs-rules.conf").
			WithAdapter(curea.NewCrs()),
	)

	// manual rules update
	_ = cure.Update(context.Background())
}
```

---

## Benchmarks

```text
Mode: Full
Body: ~10KB
Rules: 227 (OWASP CRS)

BenchmarkHTTPHandler/without_cure-8    8066433        428.3 ns/op     1008 B/op      7 allocs/op
BenchmarkHTTPHandler/with_cure-8          3615    1003092 ns/op      26172 B/op    128 allocs/op
```
