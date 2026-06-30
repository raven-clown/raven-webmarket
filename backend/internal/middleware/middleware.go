package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/metrics"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	DiscordID  string `json:"discord_id"`
	Identifier string `json:"identifier"`
	IsAdmin    bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func RateLimit(cfg *config.Config, rs *redisstore.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			}
			key := "rl:" + ip + ":" + r.URL.Path
			ctx := r.Context()
			count, err := rs.RateLimit.Incr(ctx, key).Result()
			if err == nil && count == 1 {
				rs.RateLimit.Expire(ctx, key, time.Duration(cfg.RateLimitWindowSec)*time.Second)
			}
			if count > int64(cfg.RateLimitRequests) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()
		path := r.URL.Path
		metrics.HTTPRequests.WithLabelValues(r.Method, path, strconv.Itoa(rw.status)).Inc()
		metrics.HTTPDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func AuthRequired(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			claims := &UserClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*UserClaims)
		if !ok || !claims.IsAdmin {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUser(r *http.Request) *UserClaims {
	claims, _ := r.Context().Value(UserContextKey).(*UserClaims)
	return claims
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if c, err := r.Cookie("raven_token"); err == nil {
		return c.Value
	}
	return ""
}

func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}
