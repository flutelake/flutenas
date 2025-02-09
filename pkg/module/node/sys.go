package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/module/flog"
	"os"
	"strings"
)

/*
# cat /etc/os-release
PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
NAME="Debian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"
*/
func GetOS(host string) (osRelease string, version string) {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	osRelease, version = "Unknown", "Unknown"
	file, err := exec.Command("cat /etc/os-release")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			osRelease = strings.ToLower(strings.Trim(line[len("ID="):], `"`))
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(line[len("VERSION_ID="):], `"`)
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}

	return
}

func GetKernelVersion(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("uname -r")
	if err != nil {
		flog.Fatalf("Error getting kernel version: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetArch(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("uname -m")
	if err != nil {
		flog.Fatalf("Error getting architecture: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetHostname(host string) string {
	exec := NewExec().SetHost(host)
	defer exec.Close()
	output, err := exec.Command("hostname")
	if err != nil {
		flog.Fatalf("Error getting hostname: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func GetLocalHostSshPort() (string, error) {
	// 打开 sshd 配置文件
	file, err := os.Open("/etc/ssh/sshd_config")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 默认端口是 22
	port := "22"

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// 忽略注释和空行
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// 查找 "Port" 配置项
		if strings.HasPrefix(line, "Port") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				// 解析端口号
				port = parts[1]
			}
			break // 找到第一个 Port 配置后退出
		}
	}

	// 检查是否有读取错误
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return port, nil
}
