package controller

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
)

type StorageDeviceController struct {
}

func NewStorageDeviceController() *StorageDeviceController {
	return &StorageDeviceController{}
}

func (s *StorageDeviceController) MountPoint() {
	// ticker := time.NewTicker(10 * time.Second)
	// for range ticker.C {
	// }
	// 检查挂载点
	var mountPoints []model.MountPoint
	result := db.Instance().Find(&mountPoints)
	if result.Error != nil {
		flog.Errorf("Error query db mount points: %v", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		return
	}
	mpMap := make(map[string][]model.MountPoint)
	for _, mountPoint := range mountPoints {
		mpMap[mountPoint.HostID] = append(mpMap[mountPoint.HostID], mountPoint)
	}
	for n, mps := range mpMap {
		var host model.Host
		if err := db.Instance().First(&host, "ID = ?", n).Error; err != nil {
			flog.Errorf("Error get host info, id: %s, err: %v", n, err)
			continue
		}

		exec := node.NewExec().SetHost(host.HostIP)
		points, err := node.DescribeMountedPoint(host.HostIP)
		if err != nil {
			flog.Errorf("Error describe mount point: %v", err)
			continue
		}

		devicePointMap := make(map[string]model.MountedPoint)
		for _, p := range points {
			devicePointMap[p.Device] = p
		}

		for _, mp := range mps {
			if mp.Path == "" || util.Trim(mp.Path) == "/mnt" {
				continue
			}
			mounted, ok := devicePointMap[mp.Device]
			if ok {
				if mounted.Point != mp.Path {
					// 解绑
					err := exec.UmountDir(mounted.Point)
					if err != nil {
						flog.Errorf("Error umount(device: %s, point: %s): %v", mounted.Device, mounted.Point, err)
						continue
					}
				} else {
					// 已正确挂载
					continue
				}
			}
			// 检查mp.Path路径是否存在，不存在则创建
			if _, err := exec.Command(fmt.Sprintf("mkdir -p %s", mp.Path)); err != nil {
				flog.Errorf("Error creating mount point directory: %s, err: %v", mp.Path, err)
				continue
			}

			cmdstr := fmt.Sprintf("mount %s %s", mp.Device, mp.Path)
			if _, err := exec.Command(cmdstr); err != nil {
				flog.Errorf("Error mount point: %v, mount cmd: %s", err, cmdstr)
				continue
			}
		}
	}
}
