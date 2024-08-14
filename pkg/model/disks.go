package model

type ListDiskDevicesRequest struct {
}

type ListDiskDevicesResponse struct {
	Devices []DiskDevice
}

type DiskDevice struct {
	Name   string
	Type   string
	Path   string
	Size   uint64
	Vendor string
	Model  string
	Serial string
	WWN    string
}
