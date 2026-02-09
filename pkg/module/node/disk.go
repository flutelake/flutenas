package node

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
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

func EnsureDiskEmptyForMkfs(hostIP string, device string) error {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	if !strings.HasPrefix(device, "/dev/") || strings.ContainsAny(device, " \t\r\n") {
		return fmt.Errorf("invalid device: %s", device)
	}

	out, err := exec.CommandWithoutExitCode(fmt.Sprintf("lsblk -n -o TYPE,FSTYPE,MOUNTPOINT %s", device))
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return errors.New("disk not found")
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		t := fields[0]
		fstype := ""
		mountpoint := ""
		if len(fields) >= 2 {
			fstype = fields[1]
		}
		if len(fields) >= 3 {
			mountpoint = fields[2]
		}

		if i == 0 {
			if t != "disk" {
				return fmt.Errorf("device is not a disk: %s", device)
			}
			if fstype != "" || mountpoint != "" {
				return fmt.Errorf("disk is not empty: %s", device)
			}
			continue
		}

		return fmt.Errorf("disk has partitions or children: %s", device)
	}

	return nil
}

func MkfsDisk(hostIP string, device string, fsType string) error {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	if !strings.HasPrefix(device, "/dev/") || strings.ContainsAny(device, " \t\r\n") {
		return fmt.Errorf("invalid device: %s", device)
	}

	fsType = strings.ToLower(strings.TrimSpace(fsType))
	var mkfsCmd string
	switch fsType {
	case "ext4":
		label, err := genUniqueDiskLabel(exec, fsType)
		if err != nil {
			return err
		}
		labelArg := fmt.Sprintf("-L '%s'", escapeSingleQuotes(label))
		mkfsCmd = fmt.Sprintf("mkfs.ext4 -F %s %s", labelArg, device)
	case "xfs":
		label, err := genUniqueDiskLabel(exec, fsType)
		if err != nil {
			return err
		}
		labelArg := fmt.Sprintf("-L '%s'", escapeSingleQuotes(label))
		mkfsCmd = fmt.Sprintf("mkfs.xfs -f %s %s", labelArg, device)
	case "btrfs":
		label, err := genUniqueDiskLabel(exec, fsType)
		if err != nil {
			return err
		}
		labelArg := fmt.Sprintf("-L '%s'", escapeSingleQuotes(label))
		mkfsCmd = fmt.Sprintf("mkfs.btrfs -f %s %s", labelArg, device)
	default:
		if !isSafeFsType(fsType) {
			return fmt.Errorf("unsupported filesystem: %s", fsType)
		}
		mkfsCmd = fmt.Sprintf("mkfs.%s %s", fsType, device)
	}

	supported, err := listSupportedMkfsFilesystems(exec)
	if err != nil {
		return err
	}
	supportedSet := map[string]struct{}{}
	for _, t := range supported {
		supportedSet[t] = struct{}{}
	}
	if _, ok := supportedSet[fsType]; !ok {
		return fmt.Errorf("unsupported filesystem: %s", fsType)
	}

	cmd := fmt.Sprintf("if command -v wipefs >/dev/null 2>&1; then wipefs -a %s; fi; %s; sync; udevadm settle >/dev/null 2>&1 || true", device, mkfsCmd)
	bs, err := exec.Command(cmd)
	if err != nil {
		return fmt.Errorf("mkfs failed: %w, output: %s", err, string(bs))
	}

	return nil
}

func ListSupportedMkfsFilesystems(hostIP string) ([]string, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	return listSupportedMkfsFilesystems(exec)
}

func listSupportedMkfsFilesystems(exec *Exec) ([]string, error) {
	cmd := "while read -r f1 f2; do if [ \"$f1\" = \"nodev\" ]; then continue; fi; t=\"$f1\"; if [ -n \"$f2\" ]; then t=\"$f2\"; fi; if command -v \"mkfs.$t\" >/dev/null 2>&1; then echo \"$t\"; fi; done < /proc/filesystems"
	out, err := exec.CommandWithoutExitCode(cmd)
	if err != nil {
		return nil, err
	}

	fsTypes := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fsTypes[line] = true
	}

	supportedFsTypes := []string{}
	for _, item := range []string{"ext4", "xfs", "btrfs"} {
		if _, ok := fsTypes[item]; !ok {
			continue
		}
		supportedFsTypes = append(supportedFsTypes, item)
	}

	return supportedFsTypes, nil
}

func genUniqueDiskLabel(exec *Exec, fsType string) (string, error) {
	const prefix = "flutedisk"

	maxLen := 16
	switch fsType {
	case "xfs":
		maxLen = 12
	case "btrfs":
		maxLen = 256
	}

	suffixLen := maxLen - len(prefix)
	if suffixLen <= 0 {
		return "", fmt.Errorf("label prefix too long for filesystem: %s", fsType)
	}
	if fsType != "btrfs" && suffixLen > 7 {
		suffixLen = 7
	}
	if fsType == "xfs" && suffixLen > 3 {
		suffixLen = 3
	}

	existing, err := listExistingDiskLabels(exec)
	if err != nil {
		return "", err
	}

	for i := 0; i < 32; i++ {
		suffix, err := randLabelSuffix(suffixLen)
		if err != nil {
			return "", err
		}
		label := prefix + suffix
		if _, ok := existing[label]; ok {
			continue
		}
		if !regexpLikeLabel(label) {
			continue
		}
		return label, nil
	}

	return "", fmt.Errorf("failed to generate unique label")
}

func listExistingDiskLabels(exec *Exec) (map[string]struct{}, error) {
	out, err := exec.CommandWithoutExitCode("if [ -d /dev/disk/by-label ]; then ls -1 /dev/disk/by-label; fi")
	if err != nil {
		return nil, err
	}

	set := map[string]struct{}{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		set[line] = struct{}{}
	}
	return set, nil
}

func randLabelSuffix(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	if n <= 0 {
		return "", nil
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	out := make([]byte, n)
	for i := range b {
		out[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(out), nil
}

func escapeSingleQuotes(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `'\''`)
	return s
}

func isSafeFsType(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' || r == '+' {
			continue
		}
		return false
	}
	return true
}

func regexpLikeLabel(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
