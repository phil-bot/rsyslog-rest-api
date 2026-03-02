package admin

import (
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/phil-bot/rsyslox/internal/models"
)

// RestartHandler handles POST /api/admin/restart.
// It responds immediately, then replaces the current process with a fresh
// instance of itself via syscall.Exec — no external process manager required.
type RestartHandler struct{}

func NewRestartHandler() *RestartHandler { return &RestartHandler{} }

func (h *RestartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only POST is allowed"))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"restarting"}`))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	log.Println("Admin: restart requested — re-executing process in 600ms")
	go func() {
		time.Sleep(600 * time.Millisecond)

		exe, err := os.Executable()
		if err != nil {
			log.Printf("Admin: restart failed — could not find executable: %v", err)
			return
		}

		log.Printf("Admin: exec %s %v", exe, os.Args)
		if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
			log.Printf("Admin: restart failed — exec error: %v", err)
		}
	}()
}

