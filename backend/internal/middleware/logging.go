package middleware

import (
	"log/slog"
	"path"
	"time"

	"github.com/gin-gonic/gin"
)

// matchesExcluded checks if the given request path matches any of the exclude patterns.
// Patterns can include '*' wildcards for path segments (e.g., "/api/containers/*/stats/stream").
func matchesExcluded(p string, patterns []string) bool {
	for _, pat := range patterns {
		if pat == "" {
			continue
		}
		if pat == p {
			return true
		}
		if ok, err := path.Match(pat, p); err == nil && ok {
			return true
		}
	}
	return false
}

// LoggingMiddleware is a JSON slog-based request logger with robust path ignoring.
func LoggingMiddleware(excludePatterns ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if matchesExcluded(reqPath, excludePatterns) {
			// Skip logging entirely for excluded paths
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		method := c.Request.Method
		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		query := c.Request.URL.RawQuery
		msg := "HTTP request"

		// Log errors as error level if any were recorded in the context
		if len(c.Errors) > 0 || status >= 500 {
			slog.ErrorContext(
				c.Request.Context(),
				msg,
				slog.Int("status", status),
				slog.String("method", method),
				slog.String("path", reqPath),
				slog.String("query", query),
				slog.String("ip", ip),
				slog.String("user_agent", ua),
				slog.Duration("latency", latency),
				slog.Int("bytes_out", size),
				slog.String("errors", c.Errors.String()),
			)
			return
		}

		slog.InfoContext(
			c.Request.Context(),
			msg,
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", reqPath),
			slog.String("query", query),
			slog.String("ip", ip),
			slog.String("user_agent", ua),
			slog.Duration("latency", latency),
			slog.Int("bytes_out", size),
		)
	}
}
