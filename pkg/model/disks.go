package model

import "gorm.io/gorm"

type ListDiskDevicesRequest struct {
}

type ListDiskDevicesResponse struct {
	Devices []DiskDevice
}

type DiskDevice struct {
	Name         string
	Type         string
	Size         string
	Vendor       string
	Model        string
	Serial       string
	WWN          string
	MountPoint   string
	FsType       string // filesytem type, empty if not formatted
	UUID         string // filesystem UUID, empty if not formatted
	PartUUID     string // Partition UUID, emtpty if not partitioned
	HotPlug      bool
	Rota         bool
	IsSystemDisk bool
}

type MountPoint struct {
	gorm.Model
	UUID   string `json:"UUID" gorm:"uniqueIndex"`
	Node   string `json:"Node"`
	Device string `json:"Device"`
	Path   string `json:"PATH"`
}
