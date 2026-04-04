package storage

import (
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"time"
)

var (
	DefaultRedisPathRoot  = filepath.Join(utils.RootDir, "archive/Data/redis")
	defaultRedisHDataPath = DefaultRedisPathRoot + "/hash_data.json"
)

type expirableHash map[string]time.Time

type redisRepository struct {
	HDataPath string
	HData     map[string]expirableHash
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
	exp, err := parseExpiration(expiration)
	if err != nil {
		return err
	}

	if r.HData[key] == nil {
		r.HData[key] = make(expirableHash)
	}
	r.HData[key][value] = exp
	return nil
}

func (r *redisRepository) Get(key string) (string, error) {
	return "", nil
}

func (r *redisRepository) HGet(tag string) (m.Hash, error) {
	he, ok := r.HData[tag]
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
	r.HData = make(map[string]expirableHash)

	err := os.MkdirAll(DefaultRedisPathRoot, 0755)
	if err != nil {
		return err
	}

	// Initialize HData file
	_, err = os.Stat(r.HDataPath)
	if os.IsNotExist(err) {
		emptyData := make(map[string]expirableHash)
		data, _ := json.MarshalIndent(emptyData, "", "  ")
		err = os.WriteFile(r.HDataPath, data, 0644)
		if err != nil {
			return fmt.Errorf("create empty redis hash file failed: %w", err)
		}
		logger.Info("Created new empty Redis hash file", "path", r.HDataPath)
	} else if err != nil {
		return err
	}

	file, err := os.ReadFile(r.HDataPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &r.HData)
	if err != nil {
		logger.Error("Failed to unmarshal Redis hash file", "error", err)
		return err
	}

	logger.Info("Initialized Redis data store", "hdata_length", len(r.HData))
	return nil
}

func (r *redisRepository) Save() error {
	err := os.MkdirAll(DefaultRedisPathRoot, 0755)
	if err != nil {
		return err
	}
	if err := SaveAsJson(r.HDataPath, r.HData); err != nil {
		return err
	}
	return nil
}
func (r *redisRepository) Shutdown() error {
	return r.Save()
}
func NewRedisRepository(HDataPath string) RedisRepository {
	if HDataPath == "" {
		HDataPath = defaultRedisHDataPath
	}
	repo := &redisRepository{
		HDataPath: HDataPath,
		HData:     make(map[string]expirableHash),
	}
	return repo
}
