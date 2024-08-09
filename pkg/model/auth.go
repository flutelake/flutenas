package model

import (
	"flutelake/fluteNAS/pkg/util"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	ID        uint   `json:"ID" gorm:"uniqueIndex"`
	SessionID string `json:"session_id"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expires_at"`
}

type SessionUserInfo struct {
	Username string           `json:"username"`
	Password *util.LinkedRune `json:"password"`
	IsAdmin  bool             `json:"is_admin"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct{}

type KeyRequest struct {
}

type KeyResponse struct {
	Key string `json:"key"`
}
