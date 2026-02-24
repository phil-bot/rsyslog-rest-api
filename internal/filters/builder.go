package filters

import (
	"fmt"
	"strings"
	"time"
)

// Builder constructs SQL WHERE clauses and arguments for filtering.
type Builder struct {
	conditions []string
	args       []interface{}
}

// New creates a new filter builder.
func New() *Builder {
	return &Builder{
		conditions: []string{},
		args:       []interface{}{},
	}
}

// AddDateRange adds a date range filter.
func (b *Builder) AddDateRange(start, end time.Time) {
	b.conditions = append(b.conditions, "ReceivedAt BETWEEN ? AND ?")
	b.args = append(b.args, start, end)
}

// AddSeverityFilter adds a severity filter using Priority MOD 8.
// Works for both legacy (Priority = Severity 0-7) and modern
// (Priority = Facility*8 + Severity) rsyslog formats.
func (b *Builder) AddSeverityFilter(values []int) {
	if len(values) == 0 {
		return
	}
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	b.conditions = append(b.conditions,
		fmt.Sprintf("Priority MOD 8 IN (%s)", strings.Join(placeholders, ",")))
	for _, v := range values {
		b.args = append(b.args, v)
	}
}

// AddMultiValueFilter adds a multi-value IN filter for a column.
func (b *Builder) AddMultiValueFilter(column string, values []interface{}) {
	if len(values) == 0 {
		return
	}
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	b.conditions = append(b.conditions,
		fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	b.args = append(b.args, values...)
}

// AddStringMultiValue adds a multi-value string IN filter.
func (b *Builder) AddStringMultiValue(column string, values []string) {
	if len(values) == 0 {
		return
	}
	ivals := make([]interface{}, len(values))
	for i, v := range values {
		ivals[i] = v
	}
	b.AddMultiValueFilter(column, ivals)
}

// AddIntMultiValue adds a multi-value integer IN filter.
func (b *Builder) AddIntMultiValue(column string, values []int) {
	if len(values) == 0 {
		return
	}
	ivals := make([]interface{}, len(values))
	for i, v := range values {
		ivals[i] = v
	}
	b.AddMultiValueFilter(column, ivals)
}

// AddMessageSearch adds LIKE search on Message column; multiple terms use OR.
func (b *Builder) AddMessageSearch(terms []string) {
	if len(terms) == 0 {
		return
	}
	conds := make([]string, len(terms))
	for i, term := range terms {
		conds[i] = "Message LIKE ?"
		b.args = append(b.args, "%"+term+"%")
	}
	b.conditions = append(b.conditions, "("+strings.Join(conds, " OR ")+")")
}

// Build returns the WHERE clause and args. Returns "1=1" when no filters.
func (b *Builder) Build() (string, []interface{}) {
	if len(b.conditions) == 0 {
		return "1=1", b.args
	}
	return strings.Join(b.conditions, " AND "), b.args
}
