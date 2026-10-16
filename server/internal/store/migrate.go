package store

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version int64
	name    string
	upSQL   string
}

// loadMigrations 读取内嵌的 *.up.sql，按版本号排序。
// 文件命名沿用 golang-migrate 约定：NNNNNN_name.up.sql。
func loadMigrations() ([]migration, error) {
	entries, err := fs.Glob(migrationsFS, "migrations/*.up.sql")
	if err != nil {
		return nil, err
	}
	migs := make([]migration, 0, len(entries))
	for _, name := range entries {
		base := name[strings.LastIndex(name, "/")+1:]
		digits := base[:strings.Index(base, "_")]
		v, err := strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration %q: 非法版本号: %w", base, err)
		}
		raw, err := migrationsFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		migs = append(migs, migration{version: v, name: base, upSQL: string(raw)})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })
	return migs, nil
}

// Migrate 应用未执行的迁移（schema_migrations 记录已应用版本）。
// 注：golang-migrate 的 sqlite 驱动与 glebarez 同注册 "sqlite" 驱动名会冲突，
// 故采用同约定的轻量自研 runner（见 docs/开发进度.md）。
func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
	)`).Error; err != nil {
		return err
	}
	applied := map[int64]bool{}
	var rows []struct{ Version int64 }
	if err := db.Raw("SELECT version FROM schema_migrations").Scan(&rows).Error; err != nil {
		return err
	}
	for _, r := range rows {
		applied[r.Version] = true
	}
	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		if err := db.Exec(m.upSQL).Error; err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		if err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.version, m.name).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}
	return nil
}
