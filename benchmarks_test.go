package cure

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/sansecio/yargo/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cnaize/cure/core"
	"github.com/cnaize/cure/logger"
	"github.com/cnaize/cure/middleware"
)

type discardResponseWriter struct{}

func (m discardResponseWriter) Header() http.Header        { return nil }
func (m discardResponseWriter) Write([]byte) (int, error)  { return 0, nil }
func (m discardResponseWriter) WriteHeader(statusCode int) {}

type callback struct {
}

func (c callback) RuleMatching(rule *scanner.MatchRule) (bool, error) {
	logger.DefaultLogger.Err(core.ErrMatchFound).Any("rule", rule).Msg("match")
	panic(rule)
}

func BenchmarkHTTPHandler(b *testing.B) {
	ctx := context.Background()
	url := "/api/data?age=21&role=admin&session=xyz123456789&theme=dark&lang=en"
	some := []byte(`{"status": "success", "data": {"user": "bob"}}`)
	body := bytes.Repeat(some, 200) // ~10КБ
	headers := map[string]string{
		"User-Agent":        "Mozilla/5.0 (Linux; Android 16; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.12.45 Mobile Safari/537.36",
		"Accept":            "application/json",
		"X-Custom-Header-1": "value1",
		"X-Custom-Header-2": "value2",
	}
	cookies := []*http.Cookie{
		{Name: "auth_token", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"},
		{Name: "settings", Value: "volume=100,notifications=true,experimental=false"},
		{Name: "tracking_id", Value: "track_99887766554433221100abcde"},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	assert.NoError(b, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	var w discardResponseWriter

	b.Run("without cure", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			r := req.Clone(ctx)
			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(w, r)
		}
	})

	cure := core.NewCure()
	err = cure.Update(b.Context())
	require.NoError(b, err)

	handler := middleware.NewCure(cure).
		WithCallback(callback{}).
		HTTPHandler(next)

	b.Run("with cure", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			r := req.Clone(ctx)
			r.Body = io.NopCloser(bytes.NewReader(body))

			handler.ServeHTTP(w, r)
		}
	})
}
