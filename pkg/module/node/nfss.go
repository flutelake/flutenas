package node

import (
	"crypto/md5"
	"encoding/hex"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/flog"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flosch/pongo2"
)

func (x *Exec) StartNFSGanesha(nfs []model.NFSExport) error {
	// 安装应用
	bs, err := x.Command("apt install nfs-ganesha nfs-ganesha-vfs")
	if err != nil {
		flog.Errorf("install nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}

	return x.generatehNFSGaneshaConfigAndStart(nfs)
}

func (x *Exec) StopNFSGanesha() error {
	bs, err := x.Command("systemctl disable nfs-ganesha && systemctl stop nfs-ganesha")
	if err != nil {
		flog.Errorf("start nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}
	return nil
}
func (x *Exec) RefreshNFSGaneshaConfig(nfs []model.NFSExport) error {
	// Filter out disabled exports - only generate config for enabled exports
	enabledExports := make([]model.NFSExport, 0)
	for _, export := range nfs {
		if export.Status == "enabled" {
			enabledExports = append(enabledExports, export)
		}
	}

	return x.generatehNFSGaneshaConfigAndStart(enabledExports)
}

func (x *Exec) generatehNFSGaneshaConfigAndStart(nfs []model.NFSExport) error {
	config := `NFS_CORE_PARAM {
        mount_path_pseudo = true;
        Protocols = 3,4,9P;
}

EXPORT_DEFAULTS {
        Access_Type = RW;
        Squash = all_squash;
		## todo replace uid gid
        Anonymous_Uid = 1000;
        Anonymous_Gid = 1000;
}

MDCACHE {
        Entries_HWMark = 100000;
}

LOG {

	Default_Log_Level = WARN;
	Components {
			FSAL = INFO;
			NFS4 = EVENT;
	}

	Facility {
			name = FILE;
			destination = "/var/log/ganesha.log";
			enable = active;
	}
}
{% for e in exports %}
EXPORT
{		
		# Export {{ e.Name }}
        Export_Id = {{ e.Id }};
        Path = {{ e.Path }};
        Pseudo = {{ e.Pseudo }};
        Protocols = 3,4;
        Access_Type = RW;
        FSAL {
                Name = VFS;
        }
        Clients = {{ e.IPWhiteRange }};
}
{% endfor %}	`

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromString(config)
	if err != nil {
		flog.Errorf("Error compiling nfs export template: %v", err)
		return err
	}
	// Now you can render the template with the given
	// pongo2.Context how often you want to.
	out, err := tpl.Execute(pongo2.Context{"exports": nfs})
	if err != nil {
		panic(err)
	}
	// fmt.Println(out)

	// 写入配置
	err = os.WriteFile("/etc/ganesha/ganesha.conf", []byte(out), 0644)
	if err != nil {
		flog.Errorf("Error writing nfs export config: %v", err)
		return err
	}

	// 启动应用
	bs, err := x.Command("systemctl enable nfs-ganesha && systemctl restart nfs-ganesha")
	if err != nil {
		flog.Errorf("start nfs-ganesha error, output: %s, err: %s", string(bs), err)
		return err
	}

	return nil
}

// StartNFSServerControl 启动NFS服务（独立函数）
func StartNFSServerControl(hostIP string) error {
	cmd := NewExec().SetHost(hostIP)

	// 验证配置文件语法
	if err := ValidateNFSConfigFile("/etc/ganesha/ganesha.conf"); err != nil {
		flog.Warnf("NFS configuration validation failed: %v", err)
		// 不阻止启动，继续尝试
	}

	// 启动服务
	bs, err := cmd.Command("systemctl start nfs-ganesha")
	if err != nil {
		return fmt.Errorf("start nfs-ganesha failed: %w, output: %s", err, string(bs))
	}

	flog.Infof("NFS service started successfully on host: %s", hostIP)
	return nil
}

// StopNFSServerControl 停止NFS服务（独立函数）
func StopNFSServerControl(hostIP string) error {
	cmd := NewExec().SetHost(hostIP)

	// 停止服务
	bs, err := cmd.Command("systemctl stop nfs-ganesha")
	if err != nil {
		return fmt.Errorf("stop nfs-ganesha failed: %w, output: %s", err, string(bs))
	}

	flog.Infof("NFS service stopped successfully on host: %s", hostIP)
	return nil
}

// GetNFSServerStatusControl 获取NFS服务状态（独立函数）
func GetNFSServerStatusControl(hostIP string) (status string, uptime string, err error) {
	cmd := NewExec().SetHost(hostIP)

	// 服务存在，检查其活动状态
	// Using CommandWithExitCode to capture output even when systemctl returns non-zero exit code
	bs, err := cmd.CommandWithoutExitCode("systemctl is-active nfs-ganesha")
	if err != nil {
		return "unknown", "", fmt.Errorf("get nfs-ganesha status failed: %w", err)
	}

	statusStr := strings.TrimSpace(string(bs))
	switch statusStr {
	case "active":
		status = "running"
	case "inactive":
		status = "stopped"
	case "activating":
		status = "starting"
	case "deactivating":
		status = "stopping"
	default:
		status = "unknown"
	}

	// 获取运行时间（如果服务正在运行）
	if status == "running" {
		uptimeBs, _ := cmd.Command("systemctl show nfs-ganesha --property=ActiveEnterTimestamp --value")
		uptime = strings.TrimSpace(string(uptimeBs))
	} else {
		uptime = ""
	}

	return status, uptime, nil
}

// ValidateNFSConfigFile 验证NFS配置文件语法（独立函数）
func ValidateNFSConfigFile(configPath string) error {
	if configPath == "" {
		configPath = "/etc/ganesha/ganesha.conf"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// 使用ganesha.nfsd验证语法
	cmd := NewExec()
	bs, err := cmd.Command(fmt.Sprintf("ganesha.nfsd -f %s -t", configPath))
	if err != nil {
		return fmt.Errorf("config validation failed: %w, output: %s", err, string(bs))
	}

	return nil
}

// CheckNFSGaneshaInstallation 检查NFS-Ganesha安装状态
// 返回值:
//   - installed: 是否已安装
//   - version: 版本号（如果已安装）
//   - serviceStatus: 服务状态 (running, stopped, not_installed)
//   - err: 错误信息
func CheckNFSGaneshaInstallation(host string) (installed bool, version string, serviceStatus string, err error) {
	cmd := NewExec().SetHost(host)
	defer cmd.Close()

	// 默认返回值
	installed = false
	version = ""
	serviceStatus = "not_installed"
	err = nil

	// 步骤1: 检测发行版类型，确定包管理器
	distroInfo, err := DetectDistro(host)
	if err != nil {
		return false, "", "unknown", fmt.Errorf("failed to detect distro: %v", err)
	}

	// 步骤2: 检查NFS-Ganesha是否已安装
	switch distroInfo.PackageManager {
	case "apt":
		// Debian/Ubuntu系统使用dpkg，检查状态为ii (installed)的包，（如果包被remove但没有purge， dpkg -l仍然会列出来）
		output, err := cmd.Command("dpkg -l nfs-ganesha 2>/dev/null | grep '^ii'")
		if err != nil || len(output) == 0 {
			// 未安装或查询失败
			return false, "", "not_installed", nil
		}
		installed = true

		// 解析版本号 (输出格式: ii  nfs-ganesha 3.5-1 amd64)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			fields := strings.Fields(lines[0])
			if len(fields) >= 3 {
				version = fields[2] // 版本号在第三列
			}
		}

	case "yum", "dnf":
		// RHEL/CentOS系统使用rpm，检查确切的包名
		output, err := cmd.Command("rpm -q nfs-ganesha 2>/dev/null")
		if err != nil || len(output) == 0 {
			// 未安装或查询失败
			return false, "", "not_installed", nil
		}
		installed = true

		// 解析版本号 (输出格式: nfs-ganesha-3.5-1.el8.x86_64)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			// 提取版本号 (nfs-ganesha-3.5-1.el8.x86_64 -> 3.5)
			pkgName := lines[0]
			parts := strings.Split(pkgName, "-")
			if len(parts) >= 3 {
				version = parts[2] // 版本号通常是第三部分
			}
		}

	default:
		// 未知或不支持的包管理器
		return false, "", "unknown", fmt.Errorf("unsupported package manager: %s", distroInfo.PackageManager)
	}

	// 步骤3: 如果已安装，检查服务状态
	if installed {
		status, _, err := GetNFSServerStatusControl(host)
		if err != nil {
			// 获取状态失败，标记为unknown
			serviceStatus = "unknown"
		} else {
			serviceStatus = status
		}
	}

	return installed, version, serviceStatus, nil
}

func CheckAndMaintainNFSService() error {

	cmd := NewExec().SetHost("127.0.0.1")
	defer cmd.Close()
	output, err := cmd.CommandWithoutExitCode("systemctl is-active nfs-ganesha")
	if err != nil {
		return fmt.Errorf("failed to check NFS-Ganesha enable status: %w", err)
	}
	if strings.TrimSpace(string(output)) != "active" {
		cmd.CommandWithoutExitCode("systemctl start nfs-ganesha")
	}

	output, err = cmd.CommandWithoutExitCode("systemctl is-enabled nfs-ganesha")
	if err != nil {
		return fmt.Errorf("failed to check NFS-Ganesha enable status: %w", err)
	}

	enabledState := strings.TrimSpace(string(output))
	if enabledState != "enabled" {
		flog.Warnf("NFS-Ganesha service is not enabled at boot: %s", enabledState)
		cmd.CommandWithoutExitCode("systemctl enable nfs-ganesha")
	}

	return nil
}

// TestResult represents the result of a configuration test
type TestResult struct {
	Valid   bool
	Errors  []string
	Message string
}

// GenerateNFSConfig generates NFS-Ganesha configuration from database exports
func GenerateNFSConfig(exports []model.NFSExport) (string, error) {
	configTemplate := `NFS_CORE_PARAM {
        mount_path_pseudo = true;
        Protocols = 3,4,9P;
		NFS_Port_Check = false;
}

EXPORT_DEFAULTS {
        Access_Type = None;
        Squash = all_squash;
        Anonymous_Uid = {{ uid }};
        Anonymous_Gid = {{ gid}};
}

MDCACHE {
        Entries_HWMark = 100000;
}

LOG {

	Default_Log_Level = WARN;
	Components {
			FSAL = INFO;
			NFS4 = EVENT;
	}

	Facility {
			name = FILE;
			destination = "/var/log/ganesha.log";
			enable = active;
	}
}
{% for e in exports %}
EXPORT
{		
        # Export {{ e.Name }}
        Export_Id = {{ e.ID }};
        Path = {{ e.Path }};
        Pseudo = {{ e.Pseudo }};
        Protocols = 3,4;
        Access_Type = {{ e.DefaultACL }};

        FSAL {
                Name = VFS;
        }
	{% for acl in e.AclsMapped %}
        CLIENT {
                Clients = {{ acl.IPRange }};
                Access_Type = {{ acl.Permission }};
        }
	{% endfor %}
}
{% endfor %}`

	uid, gid := GetFluteUIDGID()

	nfss := make([]model.NFSExportMapped, len(exports))
	for i, export := range exports {
		acls, err := export.GetAcls()
		if err != nil {
			// If parsing fails, use empty array
			acls = []model.NFSAcl{}
		}
		export.Path = filepath.Join("/mnt", export.Path)
		nfss[i] = model.NFSExportMapped{
			NFSExport:  export,
			AclsMapped: acls,
		}
	}

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromString(configTemplate)
	if err != nil {
		return "", fmt.Errorf("error compiling NFS export template: %v", err)
	}

	// Now you can render the template with the given
	// pongo2.Context how often you want to.
	out, err := tpl.Execute(pongo2.Context{"uid": uid, "gid": gid, "exports": nfss})
	if err != nil {
		return "", fmt.Errorf("error executing NFS export template: %v", err)
	}

	return out, nil
}

// WriteFile writes content to a file on the specified host
func WriteFile(host string, filePath string, content []byte, perm os.FileMode) error {
	cmd := NewExec().SetHost(host)
	defer cmd.Close()

	return cmd.WriteFile(filePath, content, perm)
}

// MoveFile moves/renames a file on the specified host
func MoveFile(host string, srcPath string, dstPath string) error {
	cmd := NewExec().SetHost(host)
	defer cmd.Close()

	mvCmd := fmt.Sprintf("mv '%s' '%s'", srcPath, dstPath)
	output, err := cmd.Command(mvCmd)
	if err != nil {
		return fmt.Errorf("move file failed: %w, output: %s", err, string(output))
	}

	return nil
}

// BackupFile creates a backup of a file on the specified host
func BackupFile(host string, srcPath string, backupPath string) error {
	cmd := NewExec().SetHost(host)
	defer cmd.Close()

	cpCmd := fmt.Sprintf("cp '%s' '%s'", srcPath, backupPath)
	output, err := cmd.Command(cpCmd)
	if err != nil {
		return fmt.Errorf("backup file failed: %w, output: %s", err, string(output))
	}

	return nil
}

// RemoveFile removes a file on the specified host
func RemoveFile(host string, filePath string) error {
	cmd := NewExec().SetHost(host)
	defer cmd.Close()

	rmCmd := fmt.Sprintf("rm -f '%s'", filePath)
	output, err := cmd.Command(rmCmd)
	if err != nil {
		return fmt.Errorf("remove file failed: %w, output: %s", err, string(output))
	}

	return nil
}

// ReloadNFSConfig sends SIGHUP signal to trigger NFS-Ganesha to reload its configuration
func ReloadNFSConfig(host string) error {
	cmd := NewExec().SetHost(host)

	// Send SIGHUP signal to trigger NFS-Ganesha to reload configuration
	reloadCmd := "pid=$(pgrep ganesha.nfsd) && if [ -n \"$pid\" ]; then kill -HUP $pid; echo 'reload-success'; else echo 'process-not-found'; fi"

	output, err := cmd.Command(reloadCmd)
	if err != nil {
		return fmt.Errorf("reload NFS config failed: %w, output: %s", err, string(output))
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "process-not-found" {
		return fmt.Errorf("NFS-Ganesha process not found, cannot reload configuration")
	}

	if outputStr != "reload-success" {
		return fmt.Errorf("unexpected output during config reload: %s", outputStr)
	}

	return nil
}

// TestNFSExportConfig tests the NFS export configuration by generating a temporary config and validating it
func TestNFSExportConfig(exports []model.NFSExport) (*TestResult, error) {
	// Generate a temporary configuration with the provided exports
	config := `NFS_CORE_PARAM {
        mount_path_pseudo = true;
        Protocols = 3,4,9P;
}

EXPORT_DEFAULTS {
        Access_Type = RW;
        Squash = all_squash;
		## todo replace uid gid
        Anonymous_Uid = 1000;
        Anonymous_Gid = 1000;
}

MDCACHE {
        Entries_HWMark = 100000;
}

LOG {

	Default_Log_Level = WARN;
	Components {
			FSAL = INFO;
			NFS4 = EVENT;
	}

	Facility {
			name = FILE;
			destination = "/var/log/ganesha.log";
			enable = active;
	}
}
{% for e in exports %}
EXPORT
{		
		# Export {{ e.Name }}
        Export_Id = {{ e.Id }};
        Path = {{ e.Path }};
        Pseudo = {{ e.Pseudo }};
        Protocols = 3,4;
        Access_Type = RW;
        FSAL {
                Name = VFS;
        }
        Clients = {{ e.IPWhiteRange }};
}
{% endfor %}	`

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromString(config)
	if err != nil {
		return &TestResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Error compiling NFS export template: %v", err)},
			Message: "Template compilation failed",
		}, nil
	}

	// Now you can render the template with the given
	// pongo2.Context how often you want to.
	out, err := tpl.Execute(pongo2.Context{"exports": exports})
	if err != nil {
		return &TestResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Error executing NFS export template: %v", err)},
			Message: "Template execution failed",
		}, nil
	}

	// Write to a temporary file for validation
	tmpFile := "/tmp/ganesha_test_" + fmt.Sprintf("%d", time.Now().Unix()) + ".conf"
	if err := os.WriteFile(tmpFile, []byte(out), 0644); err != nil {
		return &TestResult{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Error writing temporary config file: %v", err)},
			Message: "Failed to create temporary config file",
		}, nil
	}
	defer os.Remove(tmpFile) // Clean up the temp file

	// Use ganesha.nfsd to validate the configuration syntax
	cmd := NewExec()
	validationCmd := fmt.Sprintf("ganesha.nfsd -f %s -t", tmpFile)
	bs, err := cmd.Command(validationCmd)

	if err != nil {
		// The command might fail for various reasons, including validation errors
		// Return the validation result based on the output
		errorMsg := fmt.Sprintf("Config validation failed: %s", string(bs))
		if bs == nil || len(bs) == 0 {
			errorMsg = fmt.Sprintf("Config validation failed: %v", err)
		}

		return &TestResult{
			Valid:   false,
			Errors:  []string{errorMsg},
			Message: "Configuration validation failed",
		}, nil
	}

	// If we reach here, the validation was successful
	return &TestResult{
		Valid:   true,
		Errors:  []string{},
		Message: "Configuration validation passed",
	}, nil
}

// CompareAndReplaceNFSConfig compares the current NFS config with the new config using MD5 checksum
// and only replaces if the content is different
func CompareAndReplaceNFSConfig(hostIP string, newConfig string) error {
	configPath := "/etc/ganesha/ganesha.conf"

	// Calculate MD5 of the new config using Go's crypto/md5
	newConfigHash := md5.Sum([]byte(newConfig))
	newConfigHashStr := hex.EncodeToString(newConfigHash[:])

	// Read current config file and calculate its MD5 using Go's crypto/md5
	cmd := NewExec().SetHost(hostIP)
	defer cmd.Close()

	// Check if config file exists
	checkCmd := fmt.Sprintf("test -f '%s' && echo 'exists' || echo 'not_found'", configPath)
	existsOutput, err := cmd.Command(checkCmd)
	if err != nil {
		return fmt.Errorf("failed to check if config file exists: %w", err)
	}

	// If config file doesn't exist, we need to create it
	if strings.TrimSpace(string(existsOutput)) == "not_found" {
		flog.Infof("NFS config file does not exist on host %s, will create new config", hostIP)
		return nil // File doesn't exist, so we should proceed with writing
	}

	// Read the current config file content
	readCmd := fmt.Sprintf("cat '%s'", configPath)
	currentConfigContent, err := cmd.Command(readCmd)
	if err != nil {
		return fmt.Errorf("failed to read current config file: %w", err)
	}

	// Calculate MD5 of current config content using Go's crypto/md5
	currentConfigHash := md5.Sum(currentConfigContent)
	currentConfigHashStr := hex.EncodeToString(currentConfigHash[:])

	// Compare MD5 hashes
	if currentConfigHashStr == newConfigHashStr {
		flog.Infof("NFS config on host %s is already up to date (MD5: %s), skipping update", hostIP, currentConfigHashStr)
		return fmt.Errorf("config unchanged: current and new configs are identical")
	}

	flog.Infof("NFS config on host %s has changed (current MD5: %s, new MD5: %s), will update",
		hostIP, currentConfigHashStr, newConfigHashStr)
	return nil // Config is different, proceed with update
}
