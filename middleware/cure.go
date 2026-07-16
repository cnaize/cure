package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/sansecio/yargo/scanner"

	"github.com/cnaize/cure/core"
	"github.com/cnaize/cure/logger"
)

type Cure struct {
	cure     *core.Cure
	options  *Options
	callback scanner.ScanCallback
	logger   *zerolog.Logger

	initPool sync.Once
	buffPool sync.Pool
}

func NewCure(cure *core.Cure) *Cure {
	logger := logger.DefaultLogger.
		With().
		Str("module", "middleware").
		Logger()

	return &Cure{
		cure:     cure,
		options:  &Options{},
		callback: core.DefaultCallback,
		logger:   &logger,
	}
}

func (m *Cure) WithOptions(options *Options) *Cure {
	m.options = options

	return m
}

func (m *Cure) WithCallback(callback scanner.ScanCallback) *Cure {
	m.callback = callback

	return m
}

func (m *Cure) WithLogger(logger *zerolog.Logger) *Cure {
	m.logger = logger

	return m
}

func (m *Cure) HTTPHandler(next http.Handler) http.Handler {
	m.options.SetDefaults()
	m.initPool.Do(func() {
		m.buffPool = sync.Pool{
			New: func() any {
				buff := make([]byte, m.options.MaxBuffSize)
				return &buff
			},
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// scan headers
		if m.scanNeeded(ScanNeededHeaders) {
			if !m.scanHeaders(r.Header) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// scan query
		if m.scanNeeded(ScanNeededQuery) {
			if !m.scanQuery(r.URL) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// scan body
		if m.scanNeeded(ScanNeededBody) {
			if !m.scanBody(r) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Cure) GinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
		})).ServeHTTP(c.Writer, c.Request)

		if c.Writer.Status() != http.StatusOK {
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *Cure) scanHeaders(header http.Header) bool {
	const cookieKey = "Cookie"

	if len(header) < 1 {
		return true
	}

	for key, val := range header {
		if key == cookieKey && !m.scanNeeded(ScanNeededCookies) {
			continue
		}

		for _, v := range val {
			if len(v) < 1 {
				continue
			}

			if key == cookieKey {
				for cookie := range strings.SplitSeq(v, ";") {
					cookie = strings.TrimSpace(cookie)
					if len(cookie) < 1 {
						continue
					}

					if i := strings.IndexByte(cookie, '='); i != -1 {
						cookie = cookie[i+1:]
					}

					data := unsafe.Slice(unsafe.StringData(cookie), len(cookie))
					if err := m.cure.Scan(data, 0, m.options.ScanTimeout, m.callback); errors.Is(err, core.ErrMatchFound) {
						return false
					}
				}
			} else {
				data := unsafe.Slice(unsafe.StringData(v), len(v))
				if err := m.cure.Scan(data, 0, m.options.ScanTimeout, m.callback); errors.Is(err, core.ErrMatchFound) {
					return false
				}
			}
		}
	}

	return true
}

func (m *Cure) scanQuery(url *url.URL) bool {
	if len(url.RawQuery) < 1 {
		return true
	}

	for _, val := range url.Query() {
		for _, v := range val {
			if len(v) < 1 {
				continue
			}

			data := unsafe.Slice(unsafe.StringData(v), len(v))
			if err := m.cure.Scan(data, 0, m.options.ScanTimeout, m.callback); errors.Is(err, core.ErrMatchFound) {
				return false
			}
		}
	}

	return true
}

func (m *Cure) scanBody(r *http.Request) bool {
	if r.Body == nil {
		return true
	}

	bptr := m.buffPool.Get().(*[]byte)
	defer m.buffPool.Put(bptr)

	buff := (*bptr)[:m.options.MaxBuffSize]
	reader := io.LimitReader(r.Body, int64(m.options.MaxBuffSize))
	n, err := io.ReadFull(reader, buff)
	if (err != nil && err != io.EOF && err != io.ErrUnexpectedEOF) || n < 1 {
		return true
	}

	if err := m.cure.Scan(buff[:n], 0, m.options.ScanTimeout, m.callback); errors.Is(err, core.ErrMatchFound) {
		return false
	}

	head := make([]byte, n)
	copy(head, buff[:n])
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(head), r.Body))

	return true
}

func (m *Cure) scanNeeded(target ScanNeeded) bool {
	return m.options.ScanNeeded&target != 0
}
