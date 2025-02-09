package db

import (
	"path/filepath"

	// "gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func Instance() *gorm.DB {
	return db
}

// InitDB 初始化 DB
func InitDB(pStr string) (err error) {
	p := filepath.Join(pStr, "nas.db")
	// sq
	db, err = gorm.Open(sqlite.Open(p), &gorm.Config{})
	if err != nil {
		return err
	}
	// 设置wal 异步写入
	_ = db.Exec("PRAGMA journal_mode=WAL;")
	_ = db.Exec("PRAGMA synchronous=OFF")
	// 设置连接池大小
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to set database pool size")
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)

	return nil
}
