package utils

import (
	"context"
	"sync"
	"time"
)

const LocationKey contextKey = "company_location"

// Cache: tz string -> *time.Location (thread-safe, process-level caching to avoid disk I/O)
var locCache sync.Map

// LoadLocation loads the time.Location from string with a sync.Map cache fallback
func LoadLocation(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	if v, ok := locCache.Load(tz); ok {
		return v.(*time.Location)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	locCache.Store(tz, loc)
	return loc
}

// GetLocation retrieves the cached *time.Location from the Context, defaulting to UTC if not found
func GetLocation(ctx context.Context) *time.Location {
	if loc, ok := ctx.Value(LocationKey).(*time.Location); ok && loc != nil {
		return loc
	}
	return time.UTC
}

// NowIn returns the current time localized to the tenant/company timezone derived from context
func NowIn(ctx context.Context) time.Time {
	return time.Now().In(GetLocation(ctx))
}

// StartOfDay returns the beginning of the day for the given time in the company's timezone
func StartOfDay(ctx context.Context, t time.Time) time.Time {
	loc := GetLocation(ctx)
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// StartOfMonth returns the beginning of the month for the given time in the company's timezone
func StartOfMonth(ctx context.Context, t time.Time) time.Time {
	loc := GetLocation(ctx)
	y, m, _ := t.In(loc).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, loc)
}
