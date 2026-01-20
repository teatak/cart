package cart

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func BenchmarkFrameworks(b *testing.B) {
	payload := struct {
		Message string `json:"message"`
	}{
		Message: "pong",
	}
	queryPath := "/query?a=123&b=hello&c=true"
	type QueryReq struct {
		A int    `form:"a" query:"a"`
		B string `form:"b" query:"b"`
		C bool   `form:"c" query:"c"`
	}

	b.Run("Cart", func(b *testing.B) {
		e := New()
		e.Use("/", func(c *Context, next Next) {
			c.Header("X-Bench", "1")
			next()
		})
		e.Route("/ping", func(r *Router) {
			r.GET(func(c *Context) error {
				c.String(http.StatusOK, "pong")
				return nil
			})
		})
		e.Use("/chain", func(c *Context, next Next) { next() }, func(c *Context, next Next) { next() }, func(c *Context, next Next) { next() })
		e.Route("/chain", func(r *Router) {
			r.GET(func(c *Context) error {
				c.String(http.StatusOK, "pong")
				return nil
			})
		})
		e.Route("/user/:id", func(r *Router) {
			r.GET(func(c *Context) error {
				_, _ = c.Param("id")
				c.String(http.StatusOK, "ok")
				return nil
			})
		})
		e.Route("/json", func(r *Router) {
			r.GET(func(c *Context) error {
				c.JSON(http.StatusOK, payload)
				return nil
			})
		})
		e.Route("/query", func(r *Router) {
			r.GET(func(c *Context) error {
				var req QueryReq
				_ = c.BindQuery(&req)
				c.String(http.StatusOK, "ok")
				return nil
			})
		})
		b.Run("Ping", func(b *testing.B) { benchHandler(b, e, "/ping") })
		b.Run("Chain", func(b *testing.B) { benchHandler(b, e, "/chain") })
		b.Run("Params", func(b *testing.B) { benchHandler(b, e, "/user/123") })
		b.Run("JSON", func(b *testing.B) { benchHandler(b, e, "/json") })
		b.Run("QueryBind", func(b *testing.B) { benchHandler(b, e, queryPath) })
		b.Run("PingParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/ping") })
		b.Run("ChainParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/chain") })
		b.Run("ParamsParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/user/123") })
		b.Run("JSONParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/json") })
		b.Run("QueryBindParallel", func(b *testing.B) { benchHandlerParallel(b, e, queryPath) })
	})

	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Writer.Header().Set("X-Bench", "1")
			c.Next()
		})
		r.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
		r.GET("/chain",
			func(c *gin.Context) { c.Next() },
			func(c *gin.Context) { c.Next() },
			func(c *gin.Context) { c.Next() },
			func(c *gin.Context) { c.String(http.StatusOK, "pong") },
		)
		r.GET("/user/:id", func(c *gin.Context) {
			_ = c.Param("id")
			c.String(http.StatusOK, "ok")
		})
		r.GET("/json", func(c *gin.Context) {
			c.JSON(http.StatusOK, payload)
		})
		r.GET("/query", func(c *gin.Context) {
			var req QueryReq
			_ = c.ShouldBindQuery(&req)
			c.String(http.StatusOK, "ok")
		})
		b.Run("Ping", func(b *testing.B) { benchHandler(b, r, "/ping") })
		b.Run("Chain", func(b *testing.B) { benchHandler(b, r, "/chain") })
		b.Run("Params", func(b *testing.B) { benchHandler(b, r, "/user/123") })
		b.Run("JSON", func(b *testing.B) { benchHandler(b, r, "/json") })
		b.Run("QueryBind", func(b *testing.B) { benchHandler(b, r, queryPath) })
		b.Run("PingParallel", func(b *testing.B) { benchHandlerParallel(b, r, "/ping") })
		b.Run("ChainParallel", func(b *testing.B) { benchHandlerParallel(b, r, "/chain") })
		b.Run("ParamsParallel", func(b *testing.B) { benchHandlerParallel(b, r, "/user/123") })
		b.Run("JSONParallel", func(b *testing.B) { benchHandlerParallel(b, r, "/json") })
		b.Run("QueryBindParallel", func(b *testing.B) { benchHandlerParallel(b, r, queryPath) })
	})

	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Set("X-Bench", "1")
				return next(c)
			}
		})
		e.GET("/ping", func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		})
		e.GET("/chain", func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		}, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { return next(c) }
		}, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { return next(c) }
		}, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error { return next(c) }
		})
		e.GET("/user/:id", func(c echo.Context) error {
			_ = c.Param("id")
			return c.String(http.StatusOK, "ok")
		})
		e.GET("/json", func(c echo.Context) error {
			return c.JSON(http.StatusOK, payload)
		})
		e.GET("/query", func(c echo.Context) error {
			var req QueryReq
			_ = c.Bind(&req)
			return c.String(http.StatusOK, "ok")
		})
		b.Run("Ping", func(b *testing.B) { benchHandler(b, e, "/ping") })
		b.Run("Chain", func(b *testing.B) { benchHandler(b, e, "/chain") })
		b.Run("Params", func(b *testing.B) { benchHandler(b, e, "/user/123") })
		b.Run("JSON", func(b *testing.B) { benchHandler(b, e, "/json") })
		b.Run("QueryBind", func(b *testing.B) { benchHandler(b, e, queryPath) })
		b.Run("PingParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/ping") })
		b.Run("ChainParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/chain") })
		b.Run("ParamsParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/user/123") })
		b.Run("JSONParallel", func(b *testing.B) { benchHandlerParallel(b, e, "/json") })
		b.Run("QueryBindParallel", func(b *testing.B) { benchHandlerParallel(b, e, queryPath) })
	})
}

func benchHandler(b *testing.B, h http.Handler, path string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}

func benchHandlerParallel(b *testing.B, h http.Handler, path string) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
		}
	})
}
