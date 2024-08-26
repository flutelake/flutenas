package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"path/filepath"
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

	// 处理预期的挂载点和实际的挂载点不一致的问题
	var mountPoints []model.MountPoint
	db.Instance().Model(&model.MountPoint{}).Find(&mountPoints)
	mpMap := make(map[string]string, 0)
	for _, mp := range mountPoints {
		mpMap[mp.UUID] = mp.Path
	}
	for i, disk := range disks {
		if mp, ok := mpMap[disk.UUID]; ok && mp != disk.MountPoint {
			disks[i].SpecMountPoint = mp
		}
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
	// todo 挂载前缀改成使用配置
	p := filepath.Join("/mnt", in.Path)
	// todo 检查是否已经存在挂载点
	// todo 检查是否已经挂载，如果已经挂载需要先解挂载（解挂可能会遇到device busy的问题），再重新挂载

	result := db.Instance().FirstOrCreate(&model.MountPoint{
		UUID:   in.UUID,
		Node:   in.Node,
		Path:   p,
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
