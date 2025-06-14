package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func DescribeDisk(hostIP string) ([]model.DiskDevice, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("lsblk -npbP -oNAME,SIZE,SERIAL,TYPE,WWN,VENDOR,MOUNTPOINT,HOTPLUG,ROTA,FSTYPE,PKNAME,MODEL,UUID,PARTUUID")
	if err != nil {
		return nil, fmt.Errorf("exec error: %s", err)
	}

	blocks := make([]string, 0, 10)
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		blocks = append(blocks, sc.Text())
	}

	systemDisk := ""

	disks := make([]model.DiskDevice, 0)
	for _, v := range blocks {
		fields := strings.Fields(strings.ReplaceAll(v, `"`, ``))
		// if len(fields) != 3 {
		// 	continue
		// }
		disk := model.DiskDevice{}
		pkname := ""
		for _, f := range fields {
			strs := strings.Split(f, "=")
			if len(strs) != 2 {
				continue
			}
			switch strs[0] {
			case "NAME":
				disk.Name = strs[1]
			case "SERIAL":
				disk.Serial = strs[1]
			case "SIZE":
				size, _ := strconv.ParseUint(strs[1], 10, 64)
				disk.Size = util.FormatStorageSize(size)
			case "TYPE":
				disk.Type = strs[1]
			case "WWN":
				disk.WWN = strs[1]
			case "VENDOR":
				disk.Vendor = strs[1]
			case "MOUNTPOINT":
				disk.MountPoint = strs[1]
			case "HOTPLUG":
				disk.HotPlug = util.StringToBool(strs[1])
			case "ROTA":
				disk.Rota = util.StringToBool(strs[1])
			case "MODEL":
				disk.Model = strs[1]
			case "FSTYPE":
				disk.FsType = strs[1]
			case "PKNAME":
				pkname = strs[1]
			case "UUID":
				disk.UUID = strs[1]
			case "PARTUUID":
				disk.PartUUID = strs[1]
			default:
				continue
			}
		}
		if disk.Type == "part" {
			if disk.MountPoint == "/" {
				systemDisk = pkname
			}
			continue
		}
		if disk.Type != "disk" {
			continue
		}
		disks = append(disks, disk)
	}

	for i, d := range disks {
		if d.Name == systemDisk {
			disks[i].IsSystemDisk = true
		}
	}

	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Serial < disks[j].Serial
	})
	return disks, nil
}

func DescribeMountedPoint(hostIP string) ([]model.MountedPoint, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("mount -l")
	if err != nil {
		return nil, fmt.Errorf("exec error: %s", err)
	}

	// lines := make([]string, 0, 10)
	result := make([]model.MountedPoint, 0)
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		// lines = append(lines, sc.Text())
		line := sc.Text()
		fields := strings.Fields(line)

		point := model.MountedPoint{}
		if len(fields) >= 3 {
			if fields[1] == "on" {
				point.Device = fields[0]
				point.Point = fields[2]
			}
		} else {
			continue
		}

		if len(fields) >= 5 {
			if fields[3] == "type" {
				point.FsType = fields[4]
			}
		}

		if len(fields) >= 6 {
			opt := fields[5]
			opt, _ = strings.CutPrefix(opt, "(")
			opt, _ = strings.CutSuffix(opt, ")")
			point.Options = opt
		}
		result = append(result, point)
	}

	return result, nil
}
