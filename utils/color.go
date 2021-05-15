package utils

import (
	"fmt"
	"strings"
)

const (
	ColorBlack = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite

	ColorBold     = 1
	ColorDarkGray = 90
)

// colorize returns the string s wrapped in ANSI code c, unless disabled is true.
func Colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

func ColorizedFormatLevel(i interface{}, noColor bool) string {
	var l string
	if ll, ok := i.(string); ok {
		switch ll {
		case "trace":
			l = Colorize("TRACE", ColorMagenta, noColor)
		case "debug":
			l = Colorize("DEBUG", ColorYellow, noColor)
		case "info":
			l = Colorize("INFO", ColorGreen, noColor)
		case "warn":
			l = Colorize("WARN", ColorRed, noColor)
		case "error":
			l = Colorize(Colorize("ERROR", ColorRed, noColor), ColorBold, noColor)
		case "fatal":
			l = Colorize(Colorize("FATAL", ColorRed, noColor), ColorBold, noColor)
		case "panic":
			l = Colorize(Colorize("PANIC", ColorRed, noColor), ColorBold, noColor)
		default:
			l = Colorize("???", ColorBold, noColor)
		}
	} else {
		if i == nil {
			l = Colorize("???", ColorBold, noColor)
		} else {
			l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
		}
	}

	if noColor {
		return fmt.Sprintf(" %-6s", l)
	}
	// 14 - because terminal colors are taking up some bytes
	return fmt.Sprintf(" %-15s", l)
}
