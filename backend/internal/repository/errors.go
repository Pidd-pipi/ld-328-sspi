package repository

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// isPGDuplicateKey PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
func isPGDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// isSQLiteDuplicateKey SQLite 唯一约束冲突（测试用 glebarez/sqlite 时兜底）。
func isSQLiteDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "Duplicate entry")
}
