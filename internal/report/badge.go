package report

import (
	"fmt"
	"html"
	"strings"

	"github.com/soulteary/cppreportcard-action/internal/model"
)

// gradeColor shares the palette of soulteary/goreportcard-action.
func gradeColor(grade string) string {
	switch grade {
	case "A+":
		return "44cc11"
	case "A":
		return "97ca00"
	case "B":
		return "a4a61d"
	case "C":
		return "dfb317"
	case "D":
		return "fe7d37"
	case "E", "F":
		return "e05d44"
	default:
		return "9f9f9f"
	}
}

// badgeTextWidth approximates 11px Verdana/DejaVu Sans, matching the original
// Go Report Card badge without requiring fonts or an external rendering service.
func badgeTextWidth(text string) int {
	width := 0
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z':
			width += 8
		case r == 'i' || r == 'l' || r == 'j' || r == '.' || r == ' ':
			width += 3
		case r == '+' || r == '%':
			width += 8
		default:
			width += 6
		}
	}
	return width
}

// SVG renders the original Go Report Card flat style: text-sized segments,
// rounded clipping, a subtle highlight and a shadow for readable small text.
func SVG(r model.Result) string {
	label := "c++ report"
	if r.Mode != "full" {
		label += " (" + r.Mode + ")"
	}
	message := r.Grade
	if !r.Complete {
		message = "incomplete"
	}
	color := gradeColor(message)
	labelWidth, messageWidth := badgeTextWidth(label)+20, badgeTextWidth(message)+20
	width := labelWidth + messageWidth
	title := html.EscapeString(fmt.Sprintf("%s: %s (%s analysis)", label, message, r.Mode))
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s"><title>%s</title>`, width, title, title)
	b.WriteString(`<linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>`)
	fmt.Fprintf(&b, `<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="%d" height="20" fill="#555"/><rect x="%d" width="%d" height="20" fill="#%s"/><rect width="%d" height="20" fill="url(#s)"/></g>`, width, labelWidth, labelWidth, messageWidth, color, width)
	b.WriteString(`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">`)
	for _, part := range []struct {
		text string
		x    int
	}{{label, labelWidth * 5}, {message, labelWidth*10 + messageWidth*5}} {
		text := html.EscapeString(part.text)
		fmt.Fprintf(&b, `<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text><text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>`, part.x, badgeTextWidth(part.text)*10, text, part.x, badgeTextWidth(part.text)*10, text)
	}
	b.WriteString("</g></svg>\n")
	return b.String()
}
