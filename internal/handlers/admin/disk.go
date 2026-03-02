package admin

import (
	"encoding/json"
	"net/http"
	"syscall"

	"github.com/phil-bot/rsyslox/internal/config"
	"github.com/phil-bot/rsyslox/internal/models"
)

// DiskHandler handles GET /api/admin/disk.
// It returns disk usage statistics for the path configured in cleanup.disk_path.
type DiskHandler struct {
	cfg *config.Config
}

func NewDiskHandler(cfg *config.Config) *DiskHandler { return &DiskHandler{cfg: cfg} }

type diskResponse struct {
	Path        string  `json:"path"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

func (h *DiskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only GET is allowed"))
		return
	}

	path := h.cfg.Cleanup.DiskPath
	if path == "" {
		path = "/"
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError("DISK_ERROR", "Failed to stat path: "+err.Error()))
		return
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free  := stat.Bavail * uint64(stat.Bsize)
	used  := total - free
	var pct float64
	if total > 0 {
		pct = float64(used) / float64(total) * 100
	}

	body, _ := json.Marshal(diskResponse{
		Path:        path,
		TotalBytes:  total,
		UsedBytes:   used,
		FreeBytes:   free,
		UsedPercent: pct,
	})
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
