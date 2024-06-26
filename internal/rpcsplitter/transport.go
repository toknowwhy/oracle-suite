package rpcsplitter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/toknowwhy/theunit-oracle/pkg/log"
)

type Transport struct {
	transport   http.RoundTripper
	handler     http.Handler
	virtualHost string
}

func NewTransport(endpoints []string, host string, transport http.RoundTripper, log log.Logger) (*Transport, error) {
	rpc, err := NewHandler(endpoints, log)
	if err != nil {
		return nil, err
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Transport{
		transport:   transport,
		virtualHost: host,
		handler:     rpc,
	}, nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.isVirtualHost(req) {
		return t.transport.RoundTrip(req)
	}
	rec := newRecorder()
	t.handler.ServeHTTP(rec, req)
	return t.buildResponse(rec), nil
}

func (t *Transport) isVirtualHost(req *http.Request) bool {
	return req.Host == t.virtualHost
}

func (t *Transport) buildResponse(res *recorder) *http.Response {
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", res.code, http.StatusText(res.code)),
		StatusCode:    res.code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(res.body.Len()),
		Header:        res.headers,
		Body:          io.NopCloser(res.body),
	}
}

// recorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type recorder struct {
	code    int           // code is the HTTP status code
	headers http.Header   // headers is the list of HTTP headers
	body    *bytes.Buffer // body is the HTTP response body
}

func newRecorder() *recorder {
	return &recorder{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		code:    http.StatusOK,
	}
}

func (r *recorder) Header() http.Header {
	return r.headers
}

func (r *recorder) Write(buf []byte) (int, error) {
	return r.body.Write(buf)
}

func (r *recorder) WriteHeader(code int) {
	r.code = code
}
