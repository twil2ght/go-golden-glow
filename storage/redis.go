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
	defaultRedisDataPath  = DefaultRedisPathRoot + "/data.json"
)

type expirableEntry struct {
	Value      string    `json:"value"`
	Expiration time.Time `json:"expiration"`
}

type expirableHash struct {
	Value      m.Hash    `json:"value"`
	Expiration time.Time `json:"expiration"`
}

type redisRepository struct {
	DataPath  string
	HDataPath string
	HData     map[string]*expirableHash
	Data      map[string]*expirableEntry
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

	r.Data[key] = &expirableEntry{
		Value:      value,
		Expiration: exp,
	}
	return nil
}

func (r *redisRepository) HSet(tag string, value m.Hash, expiration string) error {
	exp, err := parseExpiration(expiration)
	if err != nil {
		return err
	}

	r.HData[tag] = &expirableHash{
		Value:      value,
		Expiration: exp,
	}
	return nil
}

func (r *redisRepository) Get(key string) (string, error) {
	ent, ok := r.Data[key]
	if !ok {
		return "", log.NotFound(key)
	}

	if !ent.Expiration.IsZero() && time.Now().After(ent.Expiration) {
		return "", log.NotFound(key)
	}

	return ent.Value, nil
}

func (r *redisRepository) HGet(tag string) (m.Hash, error) {
	he, ok := r.HData[tag]
	if !ok {
		return nil, log.NotFound(tag)
	}

	if !he.Expiration.IsZero() && time.Now().After(he.Expiration) {
		return nil, log.NotFound(tag)
	}

	return he.Value, nil
}

func (r *redisRepository) Init() error {
	r.HData = make(map[string]*expirableHash)
	r.Data = make(map[string]*expirableEntry)

	err := os.MkdirAll(DefaultRedisPathRoot, 0755)
	if err != nil {
		return err
	}

	// Initialize HData file
	_, err = os.Stat(r.HDataPath)
	if os.IsNotExist(err) {
		emptyData := make(map[string]*expirableHash)
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

	// Initialize Data file
	_, err = os.Stat(r.DataPath)
	if os.IsNotExist(err) {
		emptyData := make(map[string]*expirableEntry)
		data, _ := json.MarshalIndent(emptyData, "", "  ")
		err = os.WriteFile(r.DataPath, data, 0644)
		if err != nil {
			return fmt.Errorf("create empty redis data file failed: %w", err)
		}
		logger.Info("Created new empty Redis data file", "path", r.DataPath)
	} else if err != nil {
		return err
	}

	dataFile, err := os.ReadFile(r.DataPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dataFile, &r.Data)
	if err != nil {
		logger.Error("Failed to unmarshal Redis data file", "error", err)
		return err
	}

	logger.Info("Initialized Redis data store", "hdata_length", len(r.HData), "data_length", len(r.Data))
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
	if err := SaveAsJson(r.DataPath, r.Data); err != nil {
		return err
	}
	return nil
}
func (r *redisRepository) Shutdown() error {
	return r.Save()
}
func NewRedisRepository(HDataPath, DataPath string) RedisRepository {
	if HDataPath == "" {
		HDataPath = defaultRedisHDataPath
	}
	if DataPath == "" {
		DataPath = defaultRedisDataPath
	}
	repo := &redisRepository{
		HDataPath: HDataPath,
		DataPath:  DataPath,
		Data:      make(map[string]*expirableEntry),
		HData:     make(map[string]*expirableHash),
	}
	return repo
}
