package cart

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	greenBg  = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBg  = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBg = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg    = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	// blueBg       = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	// magentaBg    = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	// cyanBg       = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	green        = string([]byte{27, 91, 51, 50, 109})
	white        = string([]byte{27, 91, 51, 55, 109})
	yellow       = string([]byte{27, 91, 51, 51, 109})
	red          = string([]byte{27, 91, 51, 49, 109})
	blue         = string([]byte{27, 91, 51, 52, 109})
	magenta      = string([]byte{27, 91, 51, 53, 109})
	cyan         = string([]byte{27, 91, 51, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

const ENV_CART_MODE = "CART_MODE"

const (
	DebugMode   string = "debug"
	ReleaseMode string = "release"
)
const (
	debugCode = iota
	releaseCode
)

var DefaultWriter io.Writer = os.Stdout
var DefaultErrorWriter io.Writer = os.Stderr
var cartMode = debugCode

type LogWriter struct {
}

func (writer LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format("2006-01-02 15:04:05.000"), " ", string(bytes))
}

func init() {
	mode := os.Getenv(ENV_CART_MODE)
	if len(mode) == 0 {
		SetMode(DebugMode)
	} else {
		SetMode(mode)
	}
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))
}

func SetMode(value string) {
	switch value {
	case DebugMode:
		cartMode = debugCode
	case ReleaseMode:
		cartMode = releaseCode
	default:
		panic("Cart mode unknown: " + value)
	}
}

/*
IsDebugging returns true if the framework is running in debug mode.
Use SetMode(cart.Release) to switch to disable the debug mode.
*/
func IsDebugging() bool {
	return cartMode == debugCode
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		isTerm := true

		if _, ok := DefaultWriter.(*os.File); !ok || disableColor {
			isTerm = false
		}

		var yellowColor, resetColor string
		if isTerm {
			yellowColor = yellow
			resetColor = reset
		}
		values = append([]interface{}{yellowColor, resetColor}, values...)
		log.Printf("%s[DEBUG]%s "+format, values...)
	}
}

func debugWarning() {
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export CART_MODE=release
 - using code:	cart.SetMode(cart.ReleaseMode)
  ██████╗ █████╗ ██████╗ ████████╗
 ██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
 ██║     ███████║██████╔╝   ██║
 ██║     ██╔══██║██╔══██╗   ██║
 ╚██████╗██║  ██║██║  ██║   ██║
  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
`)
}

func debugError(err error) {
	if err != nil {
		if IsDebugging() {
			isTerm := true

			if _, ok := DefaultWriter.(*os.File); !ok || disableColor {
				isTerm = false
			}
			var redColor, resetColor string
			if isTerm {
				redColor = red
				resetColor = reset
			}
			log.Printf("%s[ERROR]%s %v", redColor, resetColor, err)
		}
	}
}
