package controller

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"fmt"
	"time"
)

type Controller struct {
}

func (c *Controller) Run() {
	// go 	// TODO: implement
}

func (c *Controller) Stop() {
	// TODO: implement
}

func (c *Controller) MountPoint() {
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		// 检查挂载点
		var mountPoints []model.MountPoint
		result := db.Instance().Find(&mountPoints)
		if result.Error != nil {
			flog.Errorf("Error query db mount points: %v", result.Error)
			continue
		}
		if result.RowsAffected == 0 {
			continue
		}
		mpMap := make(map[string][]model.MountPoint)
		for _, mountPoint := range mountPoints {
			mpMap[mountPoint.Node] = append(mpMap[mountPoint.Node], mountPoint)
		}
		for n, mps := range mpMap {
			// todo 查询节点信息 mountPoint.Node
			exec := node.NewExec().SetHost(n)
			points, err := node.DescribeMountedPoint()
			if err != nil {
				flog.Errorf("Error describe mount point: %v", err)
				continue
			}

			devicePonintMap := make(map[string]model.MountedPoint)
			for _, p := range points {
				devicePonintMap[p.Device] = p
			}

			for _, mp := range mps {
				mounted, ok := devicePonintMap[mp.Device]
				if ok {
					if mounted.Point != mp.Path {
						// 解绑 todo
						_, err := exec.Command(fmt.Sprintf("umount -f %s", mounted.Point))
						if err != nil {
							flog.Errorf("Error umount(device: %s, point: %s): %v", mounted.Device, mounted.Point, err)
							continue
						}
					} else {
						// 已正确挂载
						continue
					}
				}

				if _, err := exec.Command(fmt.Sprintf("mount %s %s", mp.Device, mp.Path)); err != nil {
					flog.Errorf("Error mount point: %v", err)
					continue
				}
			}
		}
	}
}
