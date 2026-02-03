package node

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// DistroInfo 包含Linux发行版信息
type DistroInfo struct {
	ID             string   // 发行版ID (ubuntu, debian, centos等)
	Version        string   // 发行版版本
	IDLike         []string // 发行版家族 (debian, rhel等)
	PrettyName     string   // 完整发行版名称
	PackageManager string   // 包管理器 (apt, yum, dnf等)
}

// DetectDistro 检测远程主机的Linux发行版信息
func DetectDistro(host string) (DistroInfo, error) {
	exec := NewExec().SetHost(host)
	defer exec.Close()

	info := DistroInfo{
		ID:             "unknown",
		Version:        "unknown",
		IDLike:         []string{},
		PrettyName:     "Unknown Linux Distribution",
		PackageManager: "unknown",
	}

	// 读取 /etc/os-release 文件
	output, err := exec.Command("cat /etc/os-release")
	if err != nil {
		return info, fmt.Errorf("failed to read /etc/os-release: %v", err)
	}

	// 解析文件内容
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "ID=") {
			info.ID = strings.ToLower(strings.Trim(line[len("ID="):], `"`))
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			info.Version = strings.Trim(line[len("VERSION_ID="):], `"`)
		} else if strings.HasPrefix(line, "ID_LIKE=") {
			idLikeStr := strings.Trim(line[len("ID_LIKE="):], `"`)
			info.IDLike = strings.Fields(idLikeStr)
		} else if strings.HasPrefix(line, "PRETTY_NAME=") {
			info.PrettyName = strings.Trim(line[len("PRETTY_NAME="):], `"`)
		}
	}

	if err := scanner.Err(); err != nil {
		return info, fmt.Errorf("failed to parse /etc/os-release: %v", err)
	}

	// 确定包管理器
	info.PackageManager = GetPackageManager(info.ID, info.IDLike)

	return info, nil
}

// GetPackageManager 根据发行版ID和家族确定包管理器
func GetPackageManager(distroID string, idLike []string) string {
	distroID = strings.ToLower(distroID)

	// 检查主要发行版ID
	switch distroID {
	case "ubuntu", "debian", "raspbian", "linuxmint", "pop", "elementary":
		return "apt"
	case "centos", "rhel", "redhat", "fedora", "rocky", "almalinux", "oracle":
		// 对于RHEL系，需要区分版本
		return "yum"
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
		return "zypper"
	case "arch", "manjaro", "endeavouros":
		return "pacman"
	case "alpine":
		return "apk"
	}

	// 如果ID不匹配，检查ID_LIKE家族
	for _, like := range idLike {
		like = strings.ToLower(like)
		switch like {
		case "debian":
			return "apt"
		case "rhel", "redhat", "fedora":
			return "yum"
		case "suse":
			return "zypper"
		case "arch":
			return "pacman"
		}
	}

	// 默认返回unknown
	return "unknown"
}

// GetInstallCommands 根据发行版、包管理器和当前用户生成NFS-Ganesha安装命令
func GetInstallCommands(distroInfo DistroInfo, currentUser string) []string {
	commands := []string{}

	// 判断是否需要sudo：如果当前用户不是root，则需要sudo
	useSudo := currentUser != "root"

	switch distroInfo.PackageManager {
	case "apt":
		if useSudo {
			commands = []string{
				"sudo apt update",
				"sudo apt install -y nfs-ganesha nfs-ganesha-vfs",
			}
		} else {
			commands = []string{
				"apt update",
				"apt install -y nfs-ganesha nfs-ganesha-vfs",
			}
		}
	case "yum":
		if useSudo {
			commands = []string{
				"sudo yum install -y nfs-ganesha nfs-ganesha-vfs",
			}
		} else {
			commands = []string{
				"yum install -y nfs-ganesha nfs-ganesha-vfs",
			}
		}
	case "dnf":
		if useSudo {
			commands = []string{
				"sudo dnf install -y nfs-ganesha nfs-ganesha-vfs",
			}
		} else {
			commands = []string{
				"dnf install -y nfs-ganesha nfs-ganesha-vfs",
			}
		}
	case "zypper":
		if useSudo {
			commands = []string{
				"sudo zypper install -y nfs-ganesha nfs-ganesha-vfs",
			}
		} else {
			commands = []string{
				"zypper install -y nfs-ganesha nfs-ganesha-vfs",
			}
		}
	default:
		// 对于未知的发行版，返回通用的构建安装命令
		commands = []string{
			"# 请根据您的发行版手动安装NFS-Ganesha",
			"# 源码构建: https://github.com/nfs-ganesha/nfs-ganesha",
		}
	}

	return commands
}
