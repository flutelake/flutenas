package model

import "gorm.io/gorm"

type OpenNFSServerRequest struct {
	Open string `json:"Open" validate:"required,oneof=open close"`
}

type NFSExport struct {
	gorm.Model
	ID        uint   `json:"ID" gorm:"uniqueIndex,primaryKey"`
	HostIP    string `json:"HostIP"`
	Name      string `json:"Name"`
	Path      string `json:"Path"`
	Pseudo    string `json:"Pseudo" gorm:"unique"`
	Acls      string `json:"Acls"`
	Protocols string `json:"Protocols"`
	Status    string `json:"Status" gorm:"default:init"`
}

type NFSAcl struct {
	IPRange    string `json:"IPRange"`
	Permission string `json:"Permission"`
}
