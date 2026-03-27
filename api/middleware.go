package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ──────────────────────────────────────────────────────────
// ADVANCED FEATURE: Composable HTTP Middleware Chain
// Demonstrates: Higher-order functions, closures, concurrency
// ──────────────────────────────────────────────────────────

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

const userContextKey contextKey = "username"

// Logger is the structured JSON logger used across middleware.
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// ──────────────────────────────────────────────────────────
// Middleware 1: Structured Request Logger (slog)
// ──────────────────────────────────────────────────────────

// responseRecorder captures the status code for logging.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs every request with method, path, status, and latency.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		Logger.Info("HTTP Request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.statusCode),
			slog.Duration("latency", time.Since(start)),
			slog.String("remote", r.RemoteAddr),
			slog.String("user", getUserFromContext(r.Context())),
		)
	})
}

// ──────────────────────────────────────────────────────────
// Middleware 2: Token-Bucket Rate Limiter
// Uses goroutine for refill + sync.Mutex for thread safety
// ──────────────────────────────────────────────────────────

// RateLimiter implements a per-IP token bucket rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    int           // tokens per interval
	burst   int           // max tokens
	interval time.Duration
}

type tokenBucket struct {
	tokens   int
	lastTime time.Time
}

// NewRateLimiter creates a rate limiter with given rate and burst.
func NewRateLimiter(rate, burst int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		burst:    burst,
		interval: interval,
	}

	// Background goroutine to clean up stale buckets every minute.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[ip]
	if !exists {
		rl.buckets[ip] = &tokenBucket{tokens: rl.burst - 1, lastTime: time.Now()}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := time.Since(bucket.lastTime)
	refill := int(elapsed / rl.interval) * rl.rate
	bucket.tokens += refill
	if bucket.tokens > rl.burst {
		bucket.tokens = rl.burst
	}
	bucket.lastTime = time.Now()

	if bucket.tokens <= 0 {
		return false
	}
	bucket.tokens--
	return true
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-5 * time.Minute)
	for ip, bucket := range rl.buckets {
		if bucket.lastTime.Before(cutoff) {
			delete(rl.buckets, ip)
		}
	}
}

// RateLimitMiddleware limits requests per IP using the token bucket algorithm.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			}

			if !limiter.allow(ip) {
				Logger.Warn("Rate limit exceeded",
					slog.String("ip", ip),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ──────────────────────────────────────────────────────────
// Middleware 3: JWT Authentication
// ──────────────────────────────────────────────────────────

// JWTAuthMiddleware validates JWT tokens and injects user into context.
func JWTAuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				Logger.Warn("Invalid JWT token",
					slog.String("error", err.Error()),
					slog.String("ip", r.RemoteAddr),
				)
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			username, _ := claims["sub"].(string)
			ctx := context.WithValue(r.Context(), userContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getUserFromContext extracts the username from context.
func getUserFromContext(ctx context.Context) string {
	if user, ok := ctx.Value(userContextKey).(string); ok {
		return user
	}
	return "anonymous"
}
