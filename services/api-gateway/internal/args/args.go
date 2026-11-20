// Package args validates and normalises client-supplied GraphQL arguments.
package args

import (
	"fmt"
	"strings"
)

const (
	// MaxLimit caps every client-supplied `limit` so one query cannot ask a
	// downstream service for an unbounded page.
	MaxLimit = 100
	// MaxIDs caps ID-list arguments (users, profiles, postsByIds).
	MaxIDs = 200
)

// Limit applies a default to a nil limit and rejects out-of-range values.
func Limit(limit *int32, def int32) (int32, error) {
	if limit == nil {
		return def, nil
	}
	if *limit < 1 || *limit > MaxLimit {
		return 0, fmt.Errorf("limit must be between 1 and %d", MaxLimit)
	}
	return *limit, nil
}

// CheckIDCount rejects oversized ID-list arguments.
func CheckIDCount(name string, ids []string) error {
	if len(ids) > MaxIDs {
		return fmt.Errorf("%s accepts at most %d ids", name, MaxIDs)
	}
	return nil
}

// UniqueIDs trims, drops blanks, and de-duplicates while preserving order.
func UniqueIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
