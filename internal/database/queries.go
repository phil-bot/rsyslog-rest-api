package database

import (
	"fmt"

	"github.com/phil-bot/rsyslox/internal/models"
)

// QueryLogs executes a paginated log query with the given WHERE clause and args.
func (db *DB) QueryLogs(whereClause string, args []interface{}, limit, offset int) ([]models.LogEntry, error) {
	query := fmt.Sprintf(`
		SELECT ID, CustomerID, ReceivedAt, DeviceReportedTime, Facility, Priority,
		       FromHost, Message, NTSeverity, Importance, EventSource, EventUser,
		       EventCategory, EventID, EventBinaryData, MaxAvailable, CurrUsage,
		       MinUsage, MaxUsage, InfoUnitID, SysLogTag, EventLogType,
		       GenericFileName, SystemID
		FROM SystemEvents
		WHERE %s
		ORDER BY ReceivedAt DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	entries := []models.LogEntry{}
	for rows.Next() {
		var entry models.LogEntry
		if err := entry.ScanFromRows(rows); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// CountLogs counts the total number of rows matching the given WHERE clause.
func (db *DB) CountLogs(whereClause string, args []interface{}) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM SystemEvents WHERE %s", whereClause)
	var total int
	if err := db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count query failed: %v", err)
	}
	return total, nil
}

// QueryDistinctValues returns distinct values for a column, with optional filters.
// "Severity" is a virtual column computed from Priority MOD 8.
func (db *DB) QueryDistinctValues(column, whereClause string, args []interface{}) (interface{}, error) {
	if column == "Severity" {
		return db.queryDistinctSeverity(whereClause, args)
	}

	query := fmt.Sprintf(
		"SELECT DISTINCT %s FROM SystemEvents WHERE %s AND %s IS NOT NULL ORDER BY %s ASC",
		column, whereClause, column, column,
	)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("meta query failed: %v", err)
	}
	defer rows.Close()

	if column == "Facility" {
		return scanMetaFacilityValues(rows)
	}
	if db.isIntegerColumn(column) {
		return scanIntValues(rows)
	}
	return scanStringValues(rows)
}

// queryDistinctSeverity returns distinct Severity values derived from Priority MOD 8.
func (db *DB) queryDistinctSeverity(whereClause string, args []interface{}) (interface{}, error) {
	query := fmt.Sprintf(
		"SELECT DISTINCT Priority MOD 8 AS Severity FROM SystemEvents WHERE %s ORDER BY Severity ASC",
		whereClause,
	)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("severity meta query failed: %v", err)
	}
	defer rows.Close()

	var result []models.MetaValue
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			continue
		}
		label := ""
		if v >= 0 && v < len(models.SeverityLabels) {
			label = models.SeverityLabels[v]
		}
		result = append(result, models.MetaValue{Val: v, Label: label})
	}
	if result == nil {
		result = []models.MetaValue{}
	}
	return result, nil
}

// scanMetaFacilityValues scans facility integer values and attaches RFC labels.
func scanMetaFacilityValues(rows interface{ Scan(...interface{}) error; Next() bool }) (interface{}, error) {
	var result []models.MetaValue
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			continue
		}
		label := ""
		if v >= 0 && v < len(models.FacilityLabels) {
			label = models.FacilityLabels[v]
		}
		result = append(result, models.MetaValue{Val: v, Label: label})
	}
	if result == nil {
		result = []models.MetaValue{}
	}
	return result, nil
}

// scanIntValues scans a result set of single integer values.
func scanIntValues(rows interface{ Scan(...interface{}) error; Next() bool }) (interface{}, error) {
	var result []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			continue
		}
		result = append(result, v)
	}
	if result == nil {
		result = []int{}
	}
	return result, nil
}

// scanStringValues scans a result set of single string values.
func scanStringValues(rows interface{ Scan(...interface{}) error; Next() bool }) (interface{}, error) {
	var result []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			continue
		}
		result = append(result, v)
	}
	if result == nil {
		result = []string{}
	}
	return result, nil
}

// isIntegerColumn returns true for columns known to hold integer values.
func (db *DB) isIntegerColumn(column string) bool {
	intCols := map[string]bool{
		"Facility": true, "Priority": true, "NTSeverity": true,
		"Importance": true, "EventCategory": true, "EventID": true,
		"MaxAvailable": true, "CurrUsage": true, "MinUsage": true,
		"MaxUsage": true, "InfoUnitID": true, "SystemID": true,
	}
	return intCols[column]
}
