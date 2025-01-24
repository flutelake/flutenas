package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/module/flog"
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
