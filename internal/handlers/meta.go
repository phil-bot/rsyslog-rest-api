package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/phil-bot/rsyslox/internal/database"
	"github.com/phil-bot/rsyslox/internal/filters"
	"github.com/phil-bot/rsyslox/internal/models"
)

// MetaHandler handles GET /api/meta and GET /api/meta/{column}.
type MetaHandler struct {
	db *database.DB
}

// NewMetaHandler creates a new MetaHandler.
func NewMetaHandler(db *database.DB) *MetaHandler {
	return &MetaHandler{db: db}
}

func (h *MetaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed,
			models.NewAPIError("METHOD_NOT_ALLOWED", "Only GET method is allowed"))
		return
	}

	// Strip /api/meta prefix to get the column name
	// Handles both /api/meta  and  /api/meta/ColumnName
	column := strings.TrimPrefix(r.URL.Path, "/api/meta")
	column = strings.TrimPrefix(column, "/")
	column = strings.TrimSpace(column)

	if column == "" {
		h.handleList(w, r)
		return
	}

	h.handleColumnValues(w, r, column)
}

func (h *MetaHandler) handleList(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, models.MetaResponse{
		AvailableColumns: h.db.AvailableColumns,
		Usage:            "GET /api/meta/{column} to get distinct values for a column",
	})
}

func (h *MetaHandler) handleColumnValues(w http.ResponseWriter, r *http.Request, column string) {
	if !h.db.IsValidColumn(column) {
		respondError(w, http.StatusBadRequest,
			models.NewAPIError(models.ErrCodeInvalidColumn,
				"Invalid column: "+column).
				WithDetails("Available columns: "+strings.Join(h.db.AvailableColumns, ", ")))
		return
	}

	query := r.URL.Query()
	builder := filters.New()

	// Date range is optional for meta queries
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")
	if startDateStr != "" || endDateStr != "" {
		startDate, endDate, err := filters.ValidateDateRange(startDateStr, endDateStr)
		if err != nil {
			if apiErr, ok := err.(*models.APIError); ok {
				respondError(w, http.StatusBadRequest, apiErr)
			} else {
				respondError(w, http.StatusBadRequest,
					models.NewAPIError(models.ErrCodeInvalidParameter, err.Error()))
			}
			return
		}
		builder.AddDateRange(startDate, endDate)
	}

	severityParams := query["Severity"]
	if len(severityParams) == 0 {
		severityParams = query["Priority"]
	}
	severities, err := filters.ValidateSeverities(severityParams)
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			respondError(w, http.StatusBadRequest, apiErr)
		}
		return
	}

	facilities, err := filters.ValidateFacilities(query["Facility"])
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			respondError(w, http.StatusBadRequest, apiErr)
		}
		return
	}

	messages, err := filters.ValidateMessages(query["Message"])
	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			respondError(w, http.StatusBadRequest, apiErr)
		}
		return
	}

	builder.AddStringMultiValue("FromHost", query["FromHost"])
	builder.AddSeverityFilter(severities)
	builder.AddIntMultiValue("Facility", facilities)
	builder.AddMessageSearch(messages)
	builder.AddStringMultiValue("SysLogTag", query["SysLogTag"])

	whereClause, args := builder.Build()

	values, err := h.db.QueryDistinctValues(column, whereClause, args)
	if err != nil {
		log.Printf("Meta query error: %v", err)
		respondError(w, http.StatusInternalServerError,
			models.NewAPIError(models.ErrCodeDatabaseError, "Failed to query metadata"))
		return
	}

	respondJSON(w, http.StatusOK, values)
}
