package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate 按文件名顺序执行 migrations 目录下所有尚未执行的 SQL 文件
func Migrate(db *pgxpool.Pool, migrationsPath string) error {
	ctx := context.Background()

	// 确保迁移记录表存在
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     VARCHAR(255) PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("创建迁移记录表失败: %w", err)
	}

	// 读取已执行的迁移版本
	rows, err := db.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("查询迁移历史失败: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	// 读取迁移文件列表并排序
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		if applied[name] {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsPath, name))
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %w", name, err)
		}

		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("开启事务失败: %w", err)
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("执行迁移 %s 失败: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, name); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("记录迁移版本失败: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("提交迁移事务失败: %w", err)
		}
	}

	return nil
}
