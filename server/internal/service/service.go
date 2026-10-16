package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"studypilot/server/internal/store"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
	ErrConflict = errors.New("conflict")
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// DB 暴露给测试使用。
func (s *Service) DB() *gorm.DB { return s.db }

const dateLayout = "2006-01-02"

func today() string {
	return time.Now().Format(dateLayout)
}

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

func validDate(s string) bool {
	_, err := parseDate(s)
	return err == nil
}

func weekEndDate(weekStart string) (string, error) {
	d, err := parseDate(weekStart)
	if err != nil {
		return "", err
	}
	return d.AddDate(0, 0, 6).Format(dateLayout), nil
}

// isWeekStart 按设置项 week_start_day 校验日期是否为配置的周起始日（默认周一）。
func (s *Service) isWeekStart(v string) bool {
	d, err := parseDate(v)
	if err != nil {
		return false
	}
	day := store.GetSettingInt(s.db, "week_start_day", 1)
	if day < 0 || day > 6 {
		day = 1
	}
	return int(d.Weekday()) == day
}

func ptr[T any](v T) *T { return &v }

// ---- settings 转发（存储于 store 包） ----

func (s *Service) AllSettings() (map[string]string, error) { return store.AllSettings(s.db) }

func (s *Service) SetSetting(key, value string) error { return store.SetSetting(s.db, key, value) }
