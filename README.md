# Cure

`cure` — Pure Go Web Application Firewall (WAF) based on YARA rules (powered by [Yargo](https://github.com/sansecio/yargo))

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
	// or use r.Use(curem.NewCure(cure).GinHandler()) for Gin framework
	cwaf := curem.NewCure(cure).HTTPHandler

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
	mux.Handle("/api/some", cwaf(yourHandler))

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
Mode: Full (Headers, Cookies, Query, Body)
Body: ~10KB
Rules: 318 (OWASP CRS)

BenchmarkHTTPHandler/without_cure-8    85122314        37.94 ns/op       64 B/op     2 allocs/op
BenchmarkHTTPHandler/with_cure-8           3709    994004 ns/op       18986 B/op    32 allocs/op
```
