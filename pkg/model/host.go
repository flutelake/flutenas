package model

import "gorm.io/gorm"

type Host struct {
	gorm.Model
	ID        uint   `json:"ID" gorm:"uniqueIndex"`
	HostIP    string `json:"HostIP"`
	Hostname  string `json:"Hostname"`
	AliasName string `json:"AliasName"`
	OS        string `json:"OS"`
	OSVersion string `json:"OSVersion"`
	Arch      string `json:"Arch"`
	Kernel    string `json:"Kernel"`
	SSHPort   string `json:"SSHPort"`
}

type ListHostsRequest struct {
}

type ListHostsResponse struct {
	Hosts []Host
}
