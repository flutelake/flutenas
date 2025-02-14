package model

import (
	"gorm.io/gorm"
)

type SambaUser struct {
	gorm.Model
	ID       uint   `json:"ID" gorm:"uniqueIndex"`
	HostIP   string `json:"HostIP" gorm:"not null"`
	Username string `json:"Username" gorm:"unique;not null" validate:"required"`
	Password string `json:"Password" gorm:"not null" validate:"required"`
	Status   string `json:"Status" gorm:"default:active"`
}

func (s *SambaUser) TableName() string {
	return "samba_users"
}

const (
	SambaUserStatus_Active      = "active"
	SambaUserStatus_Init        = "init"
	SambaUserStatus_ChangingPWD = "changing_pwd"
	SambaUserStatus_Deleting    = "deleting"
)

// CreateSambaUserResponse represents the response for creating a Samba user
type CreateSambaUserResponse struct {
	ID       uint   `json:"ID"`
	Username string `json:"Username"`
}

// ListSambaUsersResponse represents the response for listing Samba users
type ListSambaUsersResponse struct {
	Users []SambaUser `json:"Users"`
}

type UpdateSambaUserRequest struct {
	ID       uint   `json:"ID"`
	Password string `json:"Password"`
}

// UpdateSambaUserResponse represents the response for updating a Samba user
type UpdateSambaUserResponse struct {
	ID uint `json:"ID"`
}

type DeleteSambaUserRequest struct {
	ID uint `json:"ID"`
}

// DeleteSambaUserResponse represents the response for deleting a Samba user
type DeleteSambaUserResponse struct {
	ID uint `json:"ID"`
}
