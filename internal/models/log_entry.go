package models

import (
	"database/sql"
	"time"
)

// Severity and Facility label maps (RFC 5424).
var SeverityLabels = [8]string{"Emergency", "Alert", "Critical", "Error", "Warning", "Notice", "Info", "Debug"}
var FacilityLabels = [24]string{
	"kern", "user", "mail", "daemon", "auth", "syslog", "lpr", "news",
	"uucp", "cron", "authpriv", "ftp", "ntp", "audit", "alert", "clock",
	"local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7",
}

// LogEntry represents a single log entry from the SystemEvents table.
type LogEntry struct {
	ID                 int        `json:"ID"`
	CustomerID         *int64     `json:"CustomerID"`
	ReceivedAt         time.Time  `json:"ReceivedAt"`
	DeviceReportedTime *time.Time `json:"DeviceReportedTime"`
	Facility           int        `json:"Facility"`
	FacilityLabel      string     `json:"Facility_Label"`
	Priority           int        `json:"Priority"`
	Severity           int        `json:"Severity"`
	SeverityLabel      string     `json:"Severity_Label"`
	FromHost           string     `json:"FromHost"`
	Message            string     `json:"Message"`
	NTSeverity         *int       `json:"NTSeverity"`
	Importance         *int       `json:"Importance"`
	EventSource        *string    `json:"EventSource"`
	EventUser          *string    `json:"EventUser"`
	EventCategory      *int       `json:"EventCategory"`
	EventID            *int       `json:"EventID"`
	EventBinaryData    *string    `json:"EventBinaryData"`
	MaxAvailable       *int       `json:"MaxAvailable"`
	CurrUsage          *int       `json:"CurrUsage"`
	MinUsage           *int       `json:"MinUsage"`
	MaxUsage           *int       `json:"MaxUsage"`
	InfoUnitID         *int       `json:"InfoUnitID"`
	SysLogTag          *string    `json:"SysLogTag"`
	EventLogType       *string    `json:"EventLogType"`
	GenericFileName    *string    `json:"GenericFileName"`
	SystemID           *int       `json:"SystemID"`
}

// ScanFromRows scans a database row into a LogEntry.
// Handles both legacy (Priority = Severity 0-7) and modern
// (Priority = Facility*8 + Severity) rsyslog formats.
func (e *LogEntry) ScanFromRows(rows *sql.Rows) error {
	var rawPriority int
	err := rows.Scan(
		&e.ID, &e.CustomerID, &e.ReceivedAt, &e.DeviceReportedTime,
		&e.Facility, &rawPriority,
		&e.FromHost, &e.Message, &e.NTSeverity, &e.Importance,
		&e.EventSource, &e.EventUser, &e.EventCategory, &e.EventID,
		&e.EventBinaryData, &e.MaxAvailable, &e.CurrUsage,
		&e.MinUsage, &e.MaxUsage, &e.InfoUnitID, &e.SysLogTag,
		&e.EventLogType, &e.GenericFileName, &e.SystemID,
	)
	if err != nil {
		return err
	}

	// Normalize priority format
	if rawPriority > 7 {
		// Modern format: PRI = Facility*8 + Severity
		e.Priority = rawPriority
		e.Severity = rawPriority % 8
		e.Facility = rawPriority / 8
	} else {
		// Legacy format: Priority column stores Severity (0-7) directly
		e.Severity = rawPriority
		e.Priority = e.Facility*8 + rawPriority
	}

	// Set labels
	if e.Severity >= 0 && e.Severity < 8 {
		e.SeverityLabel = SeverityLabels[e.Severity]
	}
	if e.Facility >= 0 && e.Facility < 24 {
		e.FacilityLabel = FacilityLabels[e.Facility]
	}

	return nil
}
