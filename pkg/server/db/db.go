package db

import (
	"path/filepath"

	"gorm.io/driver/sqlite"
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
	db, err = gorm.Open(sqlite.Open(p), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}
