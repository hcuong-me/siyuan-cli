// Package siyuan provides SQL query API methods.
package siyuan

import (
	"context"
	"encoding/json"
	"fmt"
)

// SQLQueryResult represents the result of a SQL query.
type SQLQueryResult []map[string]interface{}

// QuerySQL executes a SQL query on the SiYuan database.
// Note: This should only be used for SELECT queries for security.
func (c *Client) QuerySQL(ctx context.Context, statement string) (SQLQueryResult, error) {
	req := map[string]string{
		"stmt": statement,
	}

	resp, err := c.Post(ctx, "/api/query/sql", req)
	if err != nil {
		return nil, err
	}

	var result SQLQueryResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SQL results: %w", err)
	}
	return result, nil
}
