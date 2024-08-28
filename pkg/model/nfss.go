package model

import "gorm.io/gorm"

type OpenNFSServerRequest struct {
	Open string `json:"Open" validate:"required,oneof=open close"`
}

type NFSExport struct {
	gorm.Model
	Id           uint   `json:"UUID" gorm:"uniqueIndex,primaryKey"`
	Name         string `json:"Name"`
	Path         string `json:"Path"`
	PseudoPath   string `json:"PseudoPath" gorm:"unique"`
	IPWhiteRange string `json:"IPWhiteRange"`
}

type OpenNFSServerRequestResponse struct {
}

type CreateNFSExportRequest struct {
}

type CreateNFSExportRequestResponse struct {
}

type DeleteNFSExportRequest struct {
}

type DeleteNFSExportRequestResponse struct {
}

type UpdateNFSExportRequest struct {
}

type UpdateNFSExportRequestResponse struct {
}

type ListNFSExportsRequest struct {
}

type ListNFSExportsResponse struct {
}
