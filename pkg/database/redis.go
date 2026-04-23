package database

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"path/filepath"
	"strings"
	"time"
)

var (
	defaultRedisHDataPath = filepath.Join(utils.RootDir, "archive/Data/redis/hash_data.json")
)

type expirableHash map[string]time.Time

type redisRepository struct {
	JSON[expirableHash]
}

func parseExpiration(expiration string) (time.Time, error) {
	if expiration == "" {
		return time.Time{}, nil
	}

	// Try duration format first (e.g., "10s", "5m", "1h")
	if duration, err := time.ParseDuration(expiration); err == nil {
		return time.Now().Add(duration), nil
	}

	// Try relative date formats: "1year", "2month", "3day", "1year2month3day"
	if t, err := parseRelativeDate(expiration); err == nil {
		return t, nil
	}

	// Try absolute date formats: "2006-01-02", "2006-01-02 15:04", "2006-01-02 15:04:05"
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, expiration, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid expiration format: %s (supported: duration like '10s', '5m', '1h', relative date like '1year', '2month', '3day', or absolute date like '2026-12-31')", expiration)
}

func parseRelativeDate(expiration string) (time.Time, error) {
	now := time.Now()
	years, months, days := 0, 0, 0

	// Parse patterns like "1year", "2month", "3day", "1year2month3day"
	// Also support plural forms: "years", "months", "days"
	var num int
	var unit string
	var pos int

	for pos < len(expiration) {
		// Parse number
		num = 0
		start := pos
		for pos < len(expiration) && expiration[pos] >= '0' && expiration[pos] <= '9' {
			num = num*10 + int(expiration[pos]-'0')
			pos++
		}
		if pos == start {
			return time.Time{}, fmt.Errorf("expected number at position %d", pos)
		}

		// Parse unit (letters only)
		unitStart := pos
		for pos < len(expiration) && ((expiration[pos] >= 'a' && expiration[pos] <= 'z') || (expiration[pos] >= 'A' && expiration[pos] <= 'Z')) {
			pos++
		}
		if pos == unitStart {
			return time.Time{}, fmt.Errorf("expected unit at position %d", pos)
		}
		unit = expiration[unitStart:pos]

		// Match unit
		switch unit {
		case "year", "years", "y":
			years += num
		case "month", "months", "M":
			months += num
		case "day", "days", "d":
			days += num
		default:
			return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	return now.AddDate(years, months, days), nil
}

func (r *redisRepository) Set(key, value, expiration string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var length = len(strings.Fields(key))
	if length > 1 {
		if strings.HasPrefix(key, "if") || strings.HasPrefix(key, "then") || strings.HasPrefix(key, "[") || strings.HasPrefix(key, "check") {
			return fmt.Errorf("invalid key: %s", key)
		}
		if strings.HasPrefix(key, "Zero says") {
			return fmt.Errorf("invalid key: %s", key)
		}
		if strings.HasPrefix(key, "Susie says") {
			return fmt.Errorf("invalid key: %s", key)
		}
		if strings.HasPrefix(key, "Susie should") {
			return fmt.Errorf("invalid key: %s", key)
		}
	}
	exp, err := parseExpiration(expiration)
	if err != nil {
		return err
	}

	if r.data[key] == nil {
		r.data[key] = make(expirableHash)
	}
	if _, ok := r.data[key][value]; ok {
		if exp.IsZero() {
			return nil
		}
	}
	r.data[key][value] = exp
	return nil
}

func (r *redisRepository) HGet(tag string) (m.Hash, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	he, ok := r.data[tag]
	if !ok {
		return nil, log.NotFound(tag)
	}
	var res = make(m.Hash)
	for k, v := range he {
		// Check expiration from inside the value hash
		if !v.IsZero() && time.Now().After(v) {
			continue
		}
		res[k] = struct{}{}
	}

	return res, nil
}

func (r *redisRepository) Init() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Load()
}
func (r *redisRepository) Shutdown() error {
	return r.Save()
}
func NewRedisRepository(path string) RedisRepository {
	repo := &redisRepository{
		JSON: NewJSON[expirableHash](path),
	}
	return repo
}
