package model

import "gorm.io/gorm"

type ListDiskDevicesRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

type ListDiskDevicesResponse struct {
	Devices []DiskDevice
}

type DiskDevice struct {
	Name           string
	Type           string
	Size           string
	Vendor         string
	Model          string
	Serial         string
	WWN            string
	MountPoint     string
	SpecMountPoint string
	FsType         string // filesytem type, empty if not formatted
	UUID           string // filesystem UUID, empty if not formatted
	PartUUID       string // Partition UUID, emtpty if not partitioned
	HotPlug        bool
	Rota           bool
	IsSystemDisk   bool
}

type MountPoint struct {
	gorm.Model
	UUID   string `json:"UUID" gorm:"uniqueIndex"`
	HostID string `json:"HostID"`
	HostIP string `json:"HostIP"`
	Device string `json:"Device"`
	Path   string `json:"PATH"`
}

// result of mount -l on node
type MountedPoint struct {
	Device  string `json:"Device"`
	Point   string `json:"Point"`
	FsType  string `json:"FsType"`
	Options string `json:"Options"`
}

type SetMountPointRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
	Device string `json:"Device" validate:"required"`
	UUID   string `json:"UUID" validate:"required"`
	Path   string `json:"Path"`
}

type SetMountPointResponse struct {
}

type MkfsDiskRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
	Device string `json:"Device" validate:"required"`
	FsType string `json:"FsType" validate:"required"`
}

type MkfsDiskResponse struct {
	Device DiskDevice `json:"Device"`
}

type ListSupportedMkfsFilesystemsRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

type ListSupportedMkfsFilesystemsResponse struct {
	FsTypes []string `json:"FsTypes"`
}
