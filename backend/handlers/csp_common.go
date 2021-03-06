package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func contentSecurityPolicy() string {
	directives := map[string][]string{
		"script-src": []string{
			"'self'",
			"https://www.google-analytics.com",
			"https://www.googletagmanager.com",
			// URLs for /login route (UserKit)
			"https://widget.userkit.io",
			"https://api.userkit.io",
			"https://www.google.com/recaptcha/",
			"https://www.gstatic.com/recaptcha/",
			"https://apis.google.com",
		},
		"style-src": []string{
			"'self'",
			// URLs for /login route (UserKit)
			"https://widget.userkit.io/css/",
			"https://fonts.googleapis.com",
			"https://fonts.gstatic.com",
			// Google auth requires this, and I can't figure out any way to avoid it.
			"'unsafe-inline'",
		},
		"frame-src": []string{
			// URLs for /login route (UserKit)
			"https://www.google.com/recaptcha/",
			"https://accounts.google.com",
		},
		"img-src": []string{
			"'self'",
			// For bootstrap navbar images
			"data:",
			// For Google Analytics
			"https://www.google-analytics.com",
			// For Google Sign In
			"https://*.googleusercontent.com",
		},
	}
	directives["script-src"] = append(directives["script-src"], extraScriptSrcSources()...)
	directives["style-src"] = append(directives["style-src"], extraStyleSrcSources()...)
	policyParts := []string{}
	for directive, sources := range directives {
		policyParts = append(policyParts, fmt.Sprintf("%s %s", directive, strings.Join(sources, " ")))
	}
	return strings.Join(policyParts, "; ")
}

func (s defaultServer) enableCsp(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy())
		h(w, r)
	}
}
