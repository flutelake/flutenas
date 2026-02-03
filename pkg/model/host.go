package model

import (
	"time"
	"gorm.io/gorm"
)

type Host struct {
	gorm.Model
	ID        string `json:"ID"` // 正常情况下 ID和IP是相同的内容
	HostIP    string `json:"HostIP"`
	Hostname  string `json:"Hostname"`
	AliasName string `json:"AliasName"`
	OS        string `json:"OS"`
	OSVersion string `json:"OSVersion"`
	Arch      string `json:"Arch"`
	Kernel    string `json:"Kernel"`
	SSHPort   string `json:"SSHPort"`
	DistroID         string    `json:"DistroID"`         // 发行版ID (ubuntu, debian, centos等)
	DistroVersion    string    `json:"DistroVersion"`    // 发行版版本
	DistroIDLike     string    `json:"DistroIDLike"`     // 发行版家族 (debian, rhel等)
	PackageManager   string    `json:"PackageManager"`   // 包管理器 (apt, yum, dnf等)
	NFSInstalled     bool      `json:"NFSInstalled"`     // NFS-Ganesha是否已安装
	NFSVersion       string    `json:"NFSVersion"`       // NFS-Ganesha版本
	NFSServiceStatus string    `json:"NFSServiceStatus"` // NFS服务状态 (running, stopped等)
	LastChecked      time.Time `json:"LastChecked"`      // 最后检查时间
}

type ListHostsRequest struct {
}

type ListHostsResponse struct {
	Hosts []Host
}
