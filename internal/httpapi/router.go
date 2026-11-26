package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/ashirzad313/Cloud-Resource/internal/vm"
)

// DatabasePinger describes the one database operation this HTTP layer needs.
// pgxpool.Pool satisfies this interface without any extra adapter code.
type DatabasePinger interface {
	Ping(context.Context) error
}

// NewRouter configures the application's HTTP routes.
func NewRouter(database DatabasePinger, service *vm.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", readinessHandler(database))
	handler := vmHandler{service: service}
	mux.HandleFunc("GET /vms", handler.list)
	mux.HandleFunc("POST /vms", handler.create)
	mux.HandleFunc("GET /vms/{id}", handler.get)
	mux.HandleFunc("DELETE /vms/{id}", handler.delete)
	mux.HandleFunc("POST /vms/{id}/start", handler.start)
	mux.HandleFunc("POST /vms/{id}/stop", handler.stop)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeStatus(w, http.StatusOK, "ok")
}

func readinessHandler(database DatabasePinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if database == nil {
			writeStatus(w, http.StatusServiceUnavailable, "not_ready")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			writeStatus(w, http.StatusServiceUnavailable, "not_ready")
			return
		}

		writeStatus(w, http.StatusOK, "ready")
	}
}

func writeStatus(w http.ResponseWriter, code int, status string) {
	writeJSON(w, code, struct {
		Status string `json:"status"`
	}{Status: status})
}
