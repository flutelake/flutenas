package v1

import (
	"encoding/json"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"strings"
	"time"
)

type SambaShareServer struct{}

// CreateSambaShareRequest defines the request payload for creating a Samba share
type CreateSambaShareRequest struct {
	HostIP          string                 `json:"HostIP" validate:"required"`
	Name            string                 `json:"Name" validate:"required"`
	Path            string                 `json:"Path" validate:"required"`
	Pseudo          string                 `json:"Pseudo"`
	UserPermissions []model.UserPermission `json:"Users" validate:"required"`
}

// CreateSambaShareResponse defines the response payload for creating a Samba share
type CreateSambaShareResponse struct {
	ID   uint   `json:"ID"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// ListSambaSharesResponse defines the response payload for listing Samba shares
type ListSambaSharesResponse struct {
	Shares []SambaShare `json:"Shares"`
}

type SambaShare struct {
	ID              uint                   `json:"ID" gorm:"uniqueIndex"`
	HostIP          string                 `json:"HostIP" gorm:"not null"`
	Name            string                 `json:"Name" gorm:"unique;not null" validate:"required"`
	Path            string                 `json:"Path" gorm:"not null" validate:"required"`
	Pseudo          string                 `json:"Pseudo"`
	UserPermissions []model.UserPermission `json:"Users" gorm:"foreignKey:SambaShareID"`
	Status          string                 `json:"Status" gorm:"default:init"`
	CreatedAt       time.Time              `json:"CreatedAt"`
	UpdatedAt       time.Time              `json:"UpdatedAt"`
}

// UpdateSambaShareRequest defines the request payload for updating a Samba share
type UpdateSambaShareRequest struct {
	ID   string `json:"ID" validate:"required"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// UpdateSambaShareResponse defines the response payload for updating a Samba share
type UpdateSambaShareResponse struct {
	ID string `json:"ID"`
}

// DeleteSambaShareRequest defines the request payload for deleting a Samba share
type DeleteSambaShareRequest struct {
	ID uint `json:"ID" validate:"required"`
}

// DeleteSambaShareResponse defines the response payload for deleting a Samba share
type DeleteSambaShareResponse struct {
	ID uint `json:"ID"`
}

func (s *SambaShareServer) CreateShare(w *apiserver.Response, r *apiserver.Request) {
	in := &CreateSambaShareRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	bs, _ := json.Marshal(in.UserPermissions)

	if !strings.HasPrefix(in.Pseudo, "/") {
		in.Pseudo = "/" + in.Pseudo
	}

	share := model.SambaShare{
		HostIP:          in.HostIP,
		Name:            in.Name,
		Path:            in.Path,
		Pseudo:          in.Pseudo,
		UserPermissions: model.UserPermissionString(bs),
		Status:          model.SambaShareStatus_Init,
	}

	// Use actual DB instance
	if err := db.Instance().Create(&share).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := CreateSambaShareResponse{
		ID:   share.ID,
		Name: share.Name,
		Path: share.Path,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaShareServer) ListShares(w *apiserver.Response, r *apiserver.Request) {
	var shares []model.SambaShare

	// Use actual DB instance
	if err := db.Instance().Find(&shares).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	results := []SambaShare{}
	for _, share := range shares {
		results = append(results, SambaShare{
			ID:              share.ID,
			Name:            share.Name,
			Path:            share.Path,
			Pseudo:          share.Pseudo,
			UserPermissions: share.UserPermissions.Get(),
			Status:          share.Status,
			HostIP:          share.HostIP,
			CreatedAt:       share.CreatedAt,
			UpdatedAt:       share.UpdatedAt,
		})
	}

	out := ListSambaSharesResponse{
		Shares: results,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaShareServer) UpdateShare(w *apiserver.Response, r *apiserver.Request) {
	in := &UpdateSambaShareRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	var share model.SambaShare
	// Use actual DB instance
	if err := db.Instance().First(&share, in.ID).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	if in.Name != "" {
		share.Name = in.Name
	}
	if in.Path != "" {
		share.Path = in.Path
	}

	// Use actual DB instance
	if err := db.Instance().Save(&share).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := UpdateSambaShareResponse{
		ID: in.ID,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaShareServer) DeleteShare(w *apiserver.Response, r *apiserver.Request) {
	in := &DeleteSambaShareRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	var share model.SambaShare
	// Use actual DB instance
	if err := db.Instance().First(&share, in.ID).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// Use actual DB instance
	if err := db.Instance().Delete(&share).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := DeleteSambaShareResponse{
		ID: in.ID,
	}
	w.Write(retcode.StatusOK(out))
}
