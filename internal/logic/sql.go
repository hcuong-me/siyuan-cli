// Package logic provides SQL query business logic with security controls.
package logic

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"siyuan/internal/siyuan"
)

// SQLLogic handles SQL query business logic with security controls.
type SQLLogic struct {
	client *siyuan.Client
}

// DangerousSQLKeywords contains SQL keywords that are not allowed.
var DangerousSQLKeywords = []string{
	"DELETE", "DROP", "TRUNCATE", "UPDATE", "INSERT", "REPLACE",
	"CREATE", "ALTER", "GRANT", "REVOKE", "COMMIT", "ROLLBACK",
	"ATTACH", "DETACH", "PRAGMA", "VACUUM", "REINDEX",
}

// NewSQLLogic creates a new SQLLogic.
func NewSQLLogic() (*SQLLogic, error) {
	c, err := siyuan.New()
	if err != nil {
		return nil, err
	}
	return &SQLLogic{client: c}, nil
}

// IsReadOnlyQuery checks if a SQL query is safe (read-only SELECT).
func (l *SQLLogic) IsReadOnlyQuery(query string) error {
	// Normalize the query - remove comments and extra whitespace
	query = normalizeQuery(query)

	upperQuery := strings.ToUpper(query)

	// Check for dangerous keywords
	for _, keyword := range DangerousSQLKeywords {
		// Use word boundary matching
		pattern := fmt.Sprintf("\\b%s\\b", keyword)
		matched, _ := regexp.MatchString("(?i)"+pattern, upperQuery)
		if matched {
			return fmt.Errorf("query contains forbidden keyword: %s (only SELECT queries are allowed)", keyword)
		}
	}

	// Must start with SELECT
	trimmed := strings.TrimSpace(upperQuery)
	if !strings.HasPrefix(trimmed, "SELECT") {
		return fmt.Errorf("query must start with SELECT (read-only queries only)")
	}

	return nil
}

// normalizeQuery removes SQL comments and extra whitespace.
func normalizeQuery(query string) string {
	// Remove single-line comments (-- ...)
	query = regexp.MustCompile("--[^\n]*").ReplaceAllString(query, "")

	// Remove multi-line comments (/* ... */)
	query = regexp.MustCompile("/\\*.*?\\*/").ReplaceAllString(query, "")

	// Normalize whitespace
	query = regexp.MustCompile("\\s+").ReplaceAllString(query, " ")

	return strings.TrimSpace(query)
}

// Query executes a read-only SQL query.
func (l *SQLLogic) Query(ctx context.Context, statement string) (siyuan.SQLQueryResult, error) {
	// Security check
	if err := l.IsReadOnlyQuery(statement); err != nil {
		return nil, err
	}

	return l.client.QuerySQL(ctx, statement)
}
