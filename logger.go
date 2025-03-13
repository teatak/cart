package cart

import (
	"io"
	"log"
	"os"
	"time"
)

func Logger() Handler {
	return LoggerWithWriter(DefaultWriter)
}

func LoggerWithWriter(out io.Writer) Handler {
	isTerm := true

	if _, ok := out.(*os.File); !ok || disableColor {
		isTerm = false
	}

	return func(c *Context, next Next) {
		start := time.Now()
		path := c.Request.URL.Path
		next()
		end := time.Now()
		latency := end.Sub(start)
		method := c.Request.Method
		clientIP := c.ClientIP()
		statusCode := c.Response.Status()
		var statusColor, methodColor string
		if isTerm {
			statusColor = colorForStatus(statusCode)
			methodColor = colorForMethod(method)
		}
		log.Printf("%s[CART]%s  |%s %3d %s| %13v | %15s |%s %7s %s| %s\n",
			blue, reset,
			statusColor, statusCode, reset,
			latency,
			clientIP,
			methodColor, method, reset,
			path)
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return greenBg
	case code >= 300 && code < 400:
		return whiteBg
	case code >= 400 && code < 500:
		return yellowBg
	default:
		return redBg
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}
