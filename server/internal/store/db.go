package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 打开 SQLite（WAL + 外键 + 单连接）。个人规模下单写者连接是最稳配置。
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return nil, fmt.Errorf("%s: %w", pragma, err)
		}
	}
	return db, nil
}

// Backup 将数据库文件复制到 backupsDir，仅保留最近 keep 份（SYS-02）。
func Backup(dbPath, backupsDir string, keep int) error {
	if _, err := os.Stat(dbPath); err != nil {
		return nil
	}
	if err := os.MkdirAll(backupsDir, 0o755); err != nil {
		return err
	}
	stamp := time.Now().Format("20060102-150405")
	dst := filepath.Join(backupsDir, "app-"+stamp+".db")
	src, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, src); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if len(n) > len("app-.db") && n[:4] == "app-" && filepath.Ext(n) == ".db" {
			names = append(names, n)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for i := keep; i < len(names); i++ {
		_ = os.Remove(filepath.Join(backupsDir, names[i]))
	}
	return nil
}

// ---- settings ----

type modelSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

func (modelSetting) TableName() string { return "settings" }

func GetSetting(db *gorm.DB, key, def string) string {
	var s modelSetting
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		return def
	}
	return s.Value
}

func GetSettingInt(db *gorm.DB, key string, def int) int {
	v := GetSetting(db, key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func GetSettingFloat(db *gorm.DB, key string, def float64) float64 {
	v := GetSetting(db, key, "")
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func SetSetting(db *gorm.DB, key, value string) error {
	return db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	).Error
}

func AllSettings(db *gorm.DB) (map[string]string, error) {
	var rows []modelSetting
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.Key] = r.Value
	}
	return out, nil
}
