package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type SambaShare struct {
	gorm.Model
	ID              uint                 `json:"ID" gorm:"uniqueIndex"`
	HostIP          string               `json:"HostIP" gorm:"not null"`
	Name            string               `json:"Name" gorm:"unique;not null" validate:"required"`
	Path            string               `json:"Path" gorm:"not null" validate:"required"`
	Pseudo          string               `json:"Pseudo"`
	UserPermissions UserPermissionString `json:"UserPermissions" gorm:"foreignKey:SambaShareID"`
	Status          string               `json:"Status" gorm:"default:init"`
}

type UserPermissionString string

func (u UserPermissionString) Get() []UserPermission {
	var arr []UserPermission
	err := json.Unmarshal([]byte(u), &arr)
	if err != nil {
		return []UserPermission{}
	}
	return arr
}

type UserPermission struct {
	Username   string `json:"Username" validate:"required"`
	Permission string `json:"Permission" validate:"required"`
}

const (
	SambaACL_ReadyOnly  = "r"
	SambaACL_ReadWrite  = "rw"
	SambaUser_Anonymous = "everyone"
)

func (s *SambaShare) TableName() string {
	return "samba_shares"
}

const (
	SambaShareStatus_Init     = "init"
	SambaShareStatus_Active   = "active"
	SambaShareStatus_Updating = "updating"
	SambaShareStatus_Deleting = "deleting"
)

// CreateSambaShareResponse represents the response for creating a Samba share
type CreateSambaShareResponse struct {
	ID   uint   `json:"ID"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// ListSambaSharesResponse represents the response for listing Samba shares
type ListSambaSharesResponse struct {
	Shares []SambaShare `json:"Shares"`
}

type UpdateSambaShareRequest struct {
	ID   uint   `json:"ID"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// UpdateSambaShareResponse represents the response for updating a Samba share
type UpdateSambaShareResponse struct {
	ID uint `json:"ID"`
}

type DeleteSambaShareRequest struct {
	ID uint `json:"ID"`
}

// DeleteSambaShareResponse represents the response for deleting a Samba share
type DeleteSambaShareResponse struct {
	ID uint `json:"ID"`
}
