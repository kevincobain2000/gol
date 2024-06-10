package pkg

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetHeadersResponsePNG(header http.Header) {
	header.Set("Cache-Control", "max-age=10")
	header.Set("Expires", "10")
	header.Set("Content-Type", "image/png")
	// security headers
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
	// content policy
	header.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'; script-src 'self';")
}

func SetHeadersResponseSvg(header http.Header) {
	header.Set("Cache-Control", "max-age=10")
	header.Set("Expires", "10")
	header.Set("Content-Type", "image/svg+xml")
	// security headers
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
}
func SetHeadersResponseJSON(header http.Header) {
	header.Set("Cache-Control", "max-age=10")
	header.Set("Expires", "10")
	header.Set("Content-Type", "application/json")
	// security headers
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
	// content policy
	header.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'; script-src 'self';")
}

func SetHeadersResponseHTML(header http.Header, cacheMS string) {
	header.Set("Cache-Control", "max-age="+cacheMS)
	header.Set("Expires", cacheMS)
	header.Set("Content-Type", "text/html; charset=utf-8")
	// security headers
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
}

func SetHeadersResponsePlainText(header http.Header, cacheMS string) {
	header.Set("Cache-Control", "max-age="+cacheMS)
	header.Set("Expires", cacheMS)
	header.Set("Content-Type", "text/plain")
	// security headers
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-XSS-Protection", "1; mode=block")
	// content policy
	header.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'self'; font-src 'self'; connect-src 'self'; script-src 'self';")
}

func ResponseHTML(c echo.Context, b []byte, cacheMS string) error {
	SetHeadersResponseHTML(c.Response().Header(), cacheMS)
	return c.Blob(http.StatusOK, "text/html", b)
}
func ResponsePlain(c echo.Context, b []byte, cacheMS string) error {
	SetHeadersResponsePlainText(c.Response().Header(), cacheMS)
	return c.Blob(http.StatusOK, "text/plain", b)
}
