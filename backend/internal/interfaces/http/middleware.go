package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Middleware envuelve un handler.
type Middleware func(http.Handler) http.Handler

// Chain aplica los middlewares en orden (el primero es el más externo).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recover convierte un panic en 500 JSON.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "path", r.URL.Path, "panic", fmt.Sprint(rec))
					writeErrorCode(w, http.StatusInternalServerError, "INTERNAL", "Error interno", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// RequestLog loguea método, ruta, status y duración (nunca el body). `now` es el reloj
// real (infrastructure/clock), inyectado para no llamar a time.Now() fuera de esa capa.
func RequestLog(log *slog.Logger, now func() time.Time) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
				"durationMs", now().Sub(start).Milliseconds())
		})
	}
}

// CORS habilita el origen del front en dev.
func CORS(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if o := r.Header.Get("Origin"); o != "" && o == origin {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Vary", "Origin")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Content-Type, X-Demo-User")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// DemoIdentity lee X-Demo-User y lo pone en el contexto. Un valor inválido → 401.
func DemoIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("X-Demo-User")
		if h == "" {
			next.ServeHTTP(w, r)
			return
		}
		id, ok := ParseIdentity(h)
		if !ok {
			writeErrorCode(w, http.StatusUnauthorized, "UNAUTHENTICATED", "X-Demo-User inválido", nil)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), id)))
	})
}

// RequireAccountant exige rol contador.
func RequireAccountant(next http.Handler) http.Handler {
	return requireRole(RoleAccountant, next)
}

// RequireClient exige rol cliente.
func RequireClient(next http.Handler) http.Handler {
	return requireRole(RoleClient, next)
}

func requireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := FromContext(r.Context())
		if !ok {
			writeErrorCode(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Falta el header X-Demo-User", nil)
			return
		}
		if id.Role != role {
			writeErrorCode(w, http.StatusForbidden, "FORBIDDEN", "No tenés permiso para este recurso", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
