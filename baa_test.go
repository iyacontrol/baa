package baa

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var b = New()
var r = b.router
var c = newContext(nil, nil, b)
var f = func(c *Context) {}

type newHandler struct{}

func (t *newHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("m3 http.Handler.ServeHTTP", "true")
}

func TestNew1(t *testing.T) {
	Convey("new baa app", t, func() {
		b2 := New()
		So(b2, ShouldNotBeNil)
	})
}

func TestRun1(t *testing.T) {
	Convey("run baa app", t, func() {
		Convey("run baa app normal", func() {
			b3 := New()
			go b3.Run(":8011")
			go b3.RunServer(b3.Server(":8012"))
			// go run $GOROOT/src/crypto/tls/generate_cert.go --host localhost
			go b3.RunTLS(":8013", "_fixture/cert/cert.pem", "_fixture/cert/key.pem")
			go b3.RunTLSServer(b3.Server(":8014"), "_fixture/cert/cert.pem", "_fixture/cert/key.pem")
		})
		Convey("run baa app error", func() {
			b3 := New()
			defer func() {
				e := recover()
				So(e, ShouldNotBeNil)
			}()
			b3.run(b3.Server(":8015"), "")
		})
	})
}

func TestServeHTTP1(t *testing.T) {
	Convey("ServeHTTP", t, func() {
		Convey("normal serve", func() {
			b.Get("/ok", func(c *Context) {
				c.String(200, "ok")
			})
			w := request("GET", "/ok")
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("not found serve", func() {
			b.Get("/notfound", func(c *Context) {
				c.String(200, "ok")
			})
			w := request("GET", "/notfoundxx")
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
		Convey("error serve", func() {
			b.SetError(func(err error, c *Context) {
				c.Resp.WriteHeader(500)
				c.Resp.Write([]byte(err.Error()))
			})
			b.Get("/error", func(c *Context) {
				c.Error(fmt.Errorf("BOMB"))
			})
			w := request("GET", "/error")
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("http error serve", func() {
			b2 := New()
			b2.errorHandler = nil
			b2.Get("/error", func(c *Context) {
				c.Error(fmt.Errorf("BOMB"))
			})
			req, _ := http.NewRequest("GET", "/error", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("http error serve no debug", func() {
			b2 := New()
			b2.errorHandler = nil
			b2.SetDebug(false)
			b2.Get("/error", func(c *Context) {
				b2.Debug()
				c.Error(fmt.Errorf("BOMB"))
			})
			req, _ := http.NewRequest("GET", "/error", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("Middleware", func() {
			b2 := New()
			b2.Use(func(c *Context) {
				c.Resp.Header().Set("m1", "true")
				c.Set("Middleware", "ok")
				c.Next()
				So(c.Get("Middleware").(string), ShouldEqual, "ok")
			})
			b2.Use(HandlerFunc(func(c *Context) {
				c.Resp.Header().Set("m2", "true")
				c.Next()
			}))
			b2.Use(new(newHandler))
			b2.Use(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("m4", "true")
			}))
			b2.Use(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("m5", "true")
			})

			b2.Get("/ok", func(c *Context) {
				c.String(200, "ok")
			})
			req, _ := http.NewRequest("GET", "/ok", nil)
			w := httptest.NewRecorder()
			b2.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Unknow Middleware", func() {
			b2 := New()
			defer func() {
				e := recover()
				So(e, ShouldNotBeNil)
			}()
			b2.Use(func() {})
		})
	})
}

func request(method, uri string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, uri, nil)
	w := httptest.NewRecorder()
	b.ServeHTTP(w, req)
	return w
}
