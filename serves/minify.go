package serves

import (
	"bytes"
	"net/http"
	"regexp"

	"github.com/gorilla/handlers"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

type cacheResp struct {
	content []byte
	header  http.Header
	code    int
}

type cacheRW struct {
	w      http.ResponseWriter
	buf    *bytes.Buffer
	code   int
	header http.Header
}

func (c *cacheRW) Header() http.Header {
	if c.header == nil {
		c.header = c.w.Header()
	}
	return c.header
}

func (c cacheRW) Write(b []byte) (int, error) {
	l, err := c.w.Write(b)
	if err != nil {
		return l, err
	}
	c.buf.Write(b[:l])
	return l, err
}

func (c *cacheRW) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
	c.code = statusCode
}

type MinifyMiddleware struct {
	next  http.Handler
	cache map[string]cacheResp
}

func NewMinifyMiddleware(next http.Handler) *MinifyMiddleware {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	return &MinifyMiddleware{
		next:  handlers.CompressHandler(m.Middleware(next)),
		cache: make(map[string]cacheResp),
	}
}

func (m *MinifyMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if cached, ok := m.cache[req.URL.Path]; ok {
		for k, v := range cached.header {
			w.Header()[k] = v
		}
		w.WriteHeader(cached.code)
		w.Write(cached.content)
		return
	}
	crw := &cacheRW{
		w:   w,
		buf: &bytes.Buffer{},
	}
	m.next.ServeHTTP(crw, req)
	m.cache[req.URL.Path] = cacheResp{
		content: crw.buf.Bytes(),
		header:  crw.header,
		code:    crw.code,
	}

}
