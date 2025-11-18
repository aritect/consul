package utils

import (
	"fmt"
	"html"
	"math"
	"strings"
)

func FormatChange(change float64) string {
	return fmt.Sprintf("%.2f%%", change)
}

func FormatPercentage(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df%%%%", decimals)
	return fmt.Sprintf(format, value)
}

func FormatNumber(value float64, sign string) string {
	const (
		TRILLION = 1_000_000_000_000
		BILLION  = 1_000_000_000
		MILLION  = 1_000_000
		THOUSAND = 1_000
	)

	absValue := math.Abs(value)

	var formatted string
	if absValue >= TRILLION {
		formatted = fmt.Sprintf("%.2f", value/TRILLION)
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
		formatted += "T"
	} else if absValue >= BILLION {
		formatted = fmt.Sprintf("%.2f", value/BILLION)
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
		formatted += "B"
	} else if absValue >= MILLION {
		formatted = fmt.Sprintf("%.2f", value/MILLION)
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
		formatted += "M"
	} else if absValue >= THOUSAND {
		formatted = fmt.Sprintf("%.2f", value/THOUSAND)
		formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
		formatted += "K"
	} else {
		formatted = fmt.Sprintf("%.2f", value)
	}

	if sign != "" {
		formatted += " " + sign
	}

	return formatted
}

func FormatUSD(value float64) string {
	return "$" + FormatNumber(value, "")
}

func EscapeHTML(text string) string {
	text = html.EscapeString(text)
	return text
}

func Bold(text string) string {
	return fmt.Sprintf("<b>%s</b>", EscapeHTML(text))
}
func Link(url, text string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, EscapeHTML(text))
}

func FormatTokenSymbol(symbol string) string {
	if symbol == "" {
		return ""
	}
	return "#" + strings.ToUpper(symbol)
}

func FormatTimeframe(frame string) string {
	switch frame {
	case "1h":
		return "1 hour"
	case "6h":
		return "6 hours"
	case "24h":
		return "24 hours"
	default:
		return frame
	}
}
