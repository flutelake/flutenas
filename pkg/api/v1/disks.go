package v1

import (
	"flutelake/fluteNAS/pkg/model"
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
