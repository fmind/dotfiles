package <slug>

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"

	"<import_path>/config"
	"<import_path>/templates"
)

// Middleware defines the standard function signature for http middleware wrappers
type Middleware func(http.Handler) http.Handler

// Chain links multiple Middleware functions around a root http.Handler
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// responseWriter intercepts http writes to capture the status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Unwrap exposes the wrapped ResponseWriter so http.ResponseController and
// optional interfaces (Flusher, Hijacker) keep working through this wrapper.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// RequestLogger returns a Middleware that logs incoming and outgoing HTTP request information
func RequestLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			logger.InfoContext(
				r.Context(), "http request processed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Int("size_bytes", rw.bytesWritten),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// SecurityHeaders injects standard, robust HTTP security headers into all responses
func SecurityHeaders(env config.Environment) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")

			// Revoke access to sensitive browser features by default
			w.Header().Set("Permissions-Policy", "accelerometer=(), autoplay=(), camera=(), display-capture=(), encrypted-media=(), fullscreen=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), midi=(), payment=(), sync-xhr=(), usb=()")

			// Generate a cryptographically secure per-request nonce for the CSP.
			// crypto/rand.Read never returns an error on supported platforms (Go 1.24+).
			var nonceBytes [16]byte
			_, _ = rand.Read(nonceBytes[:])
			nonce := base64.StdEncoding.EncodeToString(nonceBytes[:])
			ctx := context.WithValue(r.Context(), templates.NonceKey, nonce)
			r = r.WithContext(ctx)

			csp := "default-src 'self'; " +
				"script-src 'self' 'nonce-" + nonce + "' 'unsafe-eval'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data:; " +
				"font-src 'self' data:; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'none'; " +
				"object-src 'none';"
			w.Header().Set("Content-Security-Policy", csp)

			if env == config.Production {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}
