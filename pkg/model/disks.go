package model

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
	HotPlug      bool
	Rota         bool
	IsSystemDisk bool
}
