package filters

import (
	"fmt"
	"strconv"
	"time"

	"github.com/phil-bot/rsyslox/internal/models"
)

// ValidateDateRange parses and validates start/end date strings (RFC3339).
func ValidateDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, models.NewAPIError(models.ErrCodeInvalidParameter,
				"invalid format").
				WithField("start_date").
				WithDetails("Expected ISO 8601/RFC3339 format (e.g., 2025-02-15T10:00:00Z)")
		}
	} else {
		startDate = time.Now().Add(-24 * time.Hour)
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, models.NewAPIError(models.ErrCodeInvalidParameter,
				"invalid format").
				WithField("end_date").
				WithDetails("Expected ISO 8601/RFC3339 format (e.g., 2025-02-15T10:00:00Z)")
		}
	} else {
		endDate = time.Now()
	}

	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, models.NewAPIError(
			models.ErrCodeInvalidDateRange,
			"start_date cannot be after end_date")
	}

	if endDate.Sub(startDate) > 90*24*time.Hour {
		return time.Time{}, time.Time{}, models.NewAPIError(
			models.ErrCodeInvalidDateRange,
			"date range cannot exceed 90 days").
			WithDetails(fmt.Sprintf("Requested range: %.1f days", endDate.Sub(startDate).Hours()/24))
	}

	return startDate, endDate, nil
}

// ValidatePagination validates limit and offset query parameters.
func ValidatePagination(limitStr, offsetStr string) (int, int, error) {
	const (
		defaultLimit = 10
		maxLimit     = 50000
	)

	offset := 0
	if offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, models.NewAPIError(models.ErrCodeInvalidParameter,
				fmt.Sprintf("'%s' is not a valid integer", offsetStr)).
				WithField("offset")
		}
		if val < 0 {
			return 0, 0, models.NewAPIError(models.ErrCodeInvalidParameter,
				"must be non-negative").
				WithField("offset")
		}
		offset = val
	}

	limit := defaultLimit
	if limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, models.NewAPIError(models.ErrCodeInvalidParameter,
				fmt.Sprintf("'%s' is not a valid integer", limitStr)).
				WithField("limit")
		}
		if val <= 0 {
			return 0, 0, models.NewAPIError(models.ErrCodeInvalidParameter,
				"must be greater than 0").
				WithField("limit")
		}
		if val > maxLimit {
			return 0, 0, models.NewAPIError(models.ErrCodeInvalidParameter,
				fmt.Sprintf("cannot exceed %d", maxLimit)).
				WithField("limit").
				WithDetails(fmt.Sprintf("Requested: %d", val))
		}
		limit = val
	}

	return limit, offset, nil
}

// ValidateSeverities parses a slice of severity string values (0-7).
// Returns an empty slice (no filter) when input is empty.
func ValidateSeverities(params []string) ([]int, error) {
	if len(params) == 0 {
		return nil, nil
	}
	result := make([]int, 0, len(params))
	for _, p := range params {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > 7 {
			return nil, models.NewAPIError(models.ErrCodeInvalidSeverity,
				fmt.Sprintf("'%s' is not a valid severity (0-7)", p)).
				WithField("Severity")
		}
		result = append(result, v)
	}
	return result, nil
}

// ValidateFacilities parses a slice of facility string values (0-23).
// Returns an empty slice (no filter) when input is empty.
func ValidateFacilities(params []string) ([]int, error) {
	if len(params) == 0 {
		return nil, nil
	}
	result := make([]int, 0, len(params))
	for _, p := range params {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > 23 {
			return nil, models.NewAPIError(models.ErrCodeInvalidFacility,
				fmt.Sprintf("'%s' is not a valid facility (0-23)", p)).
				WithField("Facility")
		}
		result = append(result, v)
	}
	return result, nil
}

// ValidateMessages returns the message search terms as-is.
// Returns nil (no filter) when input is empty.
func ValidateMessages(params []string) ([]string, error) {
	if len(params) == 0 {
		return nil, nil
	}
	return params, nil
}
