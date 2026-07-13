package logs

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

// methodBadge renders an HTTP method as a colored badge:
// GET/POST — green, PUT/PATCH — orange, DELETE — red, everything else — blue.
func methodBadge(method string) string {
	attrs := []color.Attribute{color.Bold}

	switch method {
	case http.MethodGet, http.MethodPost:
		attrs = append(attrs, color.FgHiWhite, color.BgGreen)
	case http.MethodPut, http.MethodPatch:
		attrs = append(attrs, color.FgBlack, color.BgYellow)
	case http.MethodDelete:
		attrs = append(attrs, color.FgHiWhite, color.BgRed)
	default:
		attrs = append(attrs, color.FgHiWhite, color.BgBlue)
	}

	badge := color.New(attrs...).SprintFunc()
	return badge(fmt.Sprintf(" %-6s ", method))
}

func RouteLog(router *http.ServeMux, pattern string, handler http.Handler) {
	router.Handle(pattern, handler)

	method, path, found := strings.Cut(strings.TrimSpace(pattern), " ")

	if !found {
		method, path = "ANY", method
	}
	method = strings.ToUpper(method)

	fmt.Printf("%s %s\n", methodBadge(method), color.BlueString(strings.TrimSpace(path)))
}
