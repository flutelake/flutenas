package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
)

func ListDiskDevices(w *apiserver.Response, r *apiserver.Request) {
	in := &model.ListDiskDevicesRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	disks, err := node.DescribeDisk()
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &model.ListDiskDevicesResponse{
		Devices: disks,
	}
	w.Write(retcode.StatusOK(out))
}

func SetMountPoint(w *apiserver.Response, r *apiserver.Request) {
	in := &model.SetMountPointRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	result := db.Instance().FirstOrCreate(&model.MountPoint{
		UUID:   in.UUID,
		Node:   in.Node,
		Path:   in.Path,
		Device: in.Device,
	}, "UUID = ?", in.UUID)
	if result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}
	if result.RowsAffected == 1 {
		flog.Debugf("new mount-point record create, UUID: %s", in.UUID)
	} else {
		flog.Debugf("mount-point record updated, UUID: %s", in.UUID)
	}

	w.Write(retcode.StatusOK(&model.SetMountPointResponse{}))
}
