package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cnaize/cure/core"
	"github.com/cnaize/cure/source"
)

func newTestCure(t testing.TB) *core.Cure {
	cure := core.NewCure().WithSources(source.NewLocal("../testdata/test-rules.zip"))
	err := cure.Update(t.Context())
	require.NoError(t, err)

	return cure
}

func TestHTTPHandler(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		url     string
		body    []byte
		headers map[string]string
		scan    ScanNeeded
		status  int
	}{
		{
			name:   "good get request",
			method: http.MethodGet,
			url:    "/api/data?user=bob",
			headers: map[string]string{
				"Accept": "application/json",
				"Cookie": "session=clean_value",
			},
			scan:   ScanNeededFull,
			status: http.StatusOK,
		},
		{
			name:   "good post request",
			method: http.MethodPost,
			url:    "/api/data",
			body:   []byte(`{"user": "bob", "age": "21"}`),
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			scan:   ScanNeededFull,
			status: http.StatusOK,
		},
		{
			name:   "bad get request (query)",
			method: http.MethodGet,
			url:    "/api/data?user=bob&age=Valid_Rule_1",
			scan:   ScanNeededQuery,
			status: http.StatusBadRequest,
		},
		{
			name:   "bad get request (headers)",
			method: http.MethodGet,
			url:    "/api/data",
			headers: map[string]string{
				"X-Custom-Header": "Valid_Rule_2",
			},
			scan:   ScanNeededHeaders,
			status: http.StatusBadRequest,
		},
		{
			name:   "bad get request (cookies)",
			method: http.MethodGet,
			url:    "/api/data",
			headers: map[string]string{
				"Accept": "application/json",
				"cOoKiE": "session=clean_value; auth_token=Valid_Rule_3",
			},
			scan:   ScanNeededHeaders | ScanNeededCookies,
			status: http.StatusBadRequest,
		},
		{
			name:   "bad post request",
			method: http.MethodPost,
			url:    "/api/data",
			body:   []byte(`{"user": "bob", "age": "Valid_Rule_3"}`),
			scan:   ScanNeededBody,
			status: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					var data map[string]string
					err = json.Unmarshal(body, &data)
					assert.NoError(t, err)
					assert.Equal(t, "bob", data["user"])
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("success"))
			})

			handler := NewCure(newTestCure(t)).
				WithOptions(&Options{ScanNeeded: test.scan}).
				HTTPHandler(next)

			req, err := http.NewRequest(test.method, test.url, bytes.NewReader(test.body))
			assert.NoError(t, err)

			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, test.status, rec.Code)
		})
	}
}

func TestGinHandler(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		body   []byte
		status int
	}{
		{
			name:   "good request",
			url:    "/test",
			body:   []byte(`{"hello": "world"}`),
			status: http.StatusOK,
		},
		{
			name:   "bad request (bad query)",
			url:    "/test?name=Valid_Rule_1",
			body:   []byte(`{"hello": "world"}`),
			status: http.StatusBadRequest,
		},
		{
			name:   "bad request (bad body)",
			url:    "/test",
			body:   []byte(`{"age": "Valid_Rule_2"}`),
			status: http.StatusBadRequest,
		},
	}

	gin.SetMode(gin.TestMode)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := gin.New()
			r.Use(NewCure(newTestCure(t)).GinHandler())

			r.POST("/test", func(c *gin.Context) {
				body, err := io.ReadAll(c.Request.Body)
				assert.NoError(t, err)

				var data map[string]string
				err = json.Unmarshal(body, &data)
				assert.NoError(t, err)
				assert.Equal(t, "world", data["hello"])

				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			req, err := http.NewRequest(http.MethodPost, test.url, bytes.NewReader(test.body))
			assert.NoError(t, err)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, test.status, rec.Code)
		})
	}
}
