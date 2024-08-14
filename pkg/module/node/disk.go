package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/model"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func DescribeDisk(host string) ([]model.DiskDevice, error) {
	// 查看所有的设备
	// 类型 FSTYPE 数据在某些情况下不准确和 blkid 显示 Type 数据不一致，不能有效判断是否拥有 LV，即便所在节点重启后数据恢复正常
	cmd := exec.Command("sh", "-c", "lsblk", "-ndpbP", "-oNAME,SIZE,SERIAL,TYPE,WWN,VENDOR")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("[exec.Command] error: %s", err)
	}

	blocks := make([]string, 0, 10)
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		blocks = append(blocks, sc.Text())
	}

	disks := make([]model.DiskDevice, 0, 3)
	for _, v := range blocks {
		fields := strings.Fields(strings.ReplaceAll(v, `"`, ``))
		// if len(fields) != 3 {
		// 	continue
		// }
		disk := model.DiskDevice{}
		for _, f := range fields {
			strs := strings.Split(f, "=")
			if len(strs) != 2 {
				continue
			}
			switch strs[0] {
			case "SN":
				disk.Serial = strs[1]
			case "NAME":
				disk.Name = strs[1]
			case "SIZE":
				size, _ := strconv.ParseUint(strs[1], 10, 64)
				disk.Size = size
			case "TYPE":
				disk.Type = strs[1]
			case "WWN":
				disk.WWN = strs[1]
			case "VENDOR":
				disk.Vendor = strs[1]
			// case "MODEL":
			// 	disk.Model = strs[1]
			default:
				continue
			}
		}
		disks = append(disks, disk)
	}

	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Serial < disks[j].Serial
	})
	return disks, nil
}
