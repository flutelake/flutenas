package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"path/filepath"
	"strings"
)

func ListDiskDevices(w *apiserver.Response, r *apiserver.Request) {
	in := &model.ListDiskDevicesRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	disks, err := node.DescribeDisk(in.HostIP)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 处理预期的挂载点和实际的挂载点不一致的问题
	var mountPoints []model.MountPoint
	db.Instance().Model(&model.MountPoint{}).Where("host_ip = ?", in.HostIP).Find(&mountPoints)
	mpMap := make(map[string]string, 0)
	for _, mp := range mountPoints {
		mpMap[mp.UUID] = mp.Path
	}
	for i, disk := range disks {
		if mp, ok := mpMap[disk.UUID]; ok {
			disks[i].SpecMountPoint = mp
		}
		// 接口返回的挂载点 不暴露前缀路径
		disks[i].MountPoint = strings.TrimPrefix(disks[i].MountPoint, "/mnt")
		disks[i].SpecMountPoint = strings.TrimPrefix(disks[i].SpecMountPoint, "/mnt")
	}

	out := &model.ListDiskDevicesResponse{
		Devices: disks,
	}
	w.Write(retcode.StatusOK(out))
}

func SetMountPoint(w *apiserver.Response, r *apiserver.Request) {
	in := &model.SetMountPointRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusParamInvalid(nil))
		return
	}

	var mountPoints []model.MountPoint
	result := db.Instance().Model(&model.MountPoint{}).Where("uuid = ? AND host_ip = ?", in.UUID, in.HostIP).Find(&mountPoints)
	if result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
	}
	if len(mountPoints) > 1 {
		w.WriteError(
			fmt.Errorf("there are many same uuid: %s mountpoint disk record on the host: %s", in.UUID, in.HostIP),
			retcode.StatusParamInvalid(nil),
		)
		return
	}
	// in.Path 为空 表示取消挂载，且无挂载记录 直接返回成功
	if len(mountPoints) == 0 && in.Path == "" {
		w.Write(retcode.StatusOK(&model.SetMountPointResponse{}))
	}

	if in.Path != "" {
		// if len(mountPoints) == 0 {

		// } else {
		// 	setMountPoint(w, r, in, &mountPoints[0])
		// }
		setMountPoint(w, r, in)

	} else {
		cancelMountPoint(w, r, in, &mountPoints[0])
	}
}

func setMountPoint(w *apiserver.Response, r *apiserver.Request, in *model.SetMountPointRequest) {
	host, err := GetHostInfo(w, in.HostIP)
	if err != nil {
		return
	}

	// cmd := node.NewExec().SetHost(host.HostIP)
	// todo 挂载前缀改成使用配置
	p := filepath.Join("/mnt", util.Trim(in.Path))

	// 检查是否已经挂载
	// mounted := false
	// mountedOther := ""
	// points, err := node.DescribeMountedPoint(host.HostIP)
	// if err != nil {
	// 	w.WriteError(err, retcode.StatusError(nil))
	// 	return
	// }
	// for _, item := range points {
	// 	if item.Device == in.Device {
	// 		if item.Point == p {
	// 			mounted = true
	// 		} else {
	// 			mountedOther = item.Point
	// 		}
	// 	}
	// }

	// if mountedOther != "" {
	// 	// 切换挂载点时接口中不要去解绑
	// 	// 已经挂载，但是路径不一致，需要先解挂载
	// 	// err := cmd.UmountDir(mountedOther)
	// 	// if err != nil {
	// 	// 	// 解卦失败的问题，暂时不返回错误，等控制器来解挂
	// 	// 	flog.Errorf("umount  %s error: %v", mountedOther, err)
	// 	// }
	// }

	// 挂载掉 要去掉 /dev/sdb记录，且UUID为fs uuid可能存在变化，挂载的时候需要根据实际情况判断

	result := db.Instance().
		Where(&model.MountPoint{UUID: in.UUID, HostIP: host.HostIP}).
		Assign(&model.MountPoint{Path: p}).
		FirstOrCreate(&model.MountPoint{
			UUID:   in.UUID,
			HostID: host.ID,
			HostIP: in.HostIP,
			Path:   p,
			Device: in.Device,
		})
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

func cancelMountPoint(w *apiserver.Response, r *apiserver.Request, in *model.SetMountPointRequest, record *model.MountPoint) {
	p := record.Path
	host, err := GetHostInfo(w, in.HostIP)
	if err != nil {
		return
	}
	cmd := node.NewExec().SetHost(host.HostIP)
	// 检查是否已经挂载
	mounted := false
	points, err := node.DescribeMountedPoint(host.HostIP)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	for _, item := range points {
		if item.Device == in.Device {
			if item.Point == p {
				mounted = true
			}
		}
	}
	if mounted {
		// 已经挂载， 且与取消前的路径一致， 执行解挂操作
		err := cmd.UmountDir(p)
		if err != nil {
			// flog.Errorf("umount  %s failed: %v", mountedOther, err)
			w.WriteError(err, retcode.StatusUmountDiskFailed(p))
			return
		}
	}

	record.Path = ""
	result := db.Instance().Save(record)
	if result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}
	if result.RowsAffected == 1 {
		flog.Debugf("create cancel mount-point record success, UUID: %s, HostIP: %s", in.UUID, in.HostIP)
	} else {
		flog.Debugf("cancel mount-point success, UUID: %s, HostIP: %s", in.UUID, in.HostIP)
	}

	// 看看是否存在挂载的情况
	w.Write(retcode.StatusOK(&model.SetMountPointResponse{}))
}
