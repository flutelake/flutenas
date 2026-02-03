package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
)

func ListHosts(w *apiserver.Response, r *apiserver.Request) {
	var hosts []model.Host
	err := db.Instance().Find(&hosts).Error
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &model.ListHostsResponse{
		Hosts: hosts,
	}
	w.Write(retcode.StatusOK(out))
}

// HostSystemInfoResponse 主机系统信息响应
type HostSystemInfoResponse struct {
	HostIP           string   `json:"HostIP"`
	Hostname         string   `json:"Hostname"`
	DistroID         string   `json:"DistroID"`
	DistroVersion    string   `json:"DistroVersion"`
	DistroIDLike     string   `json:"DistroIDLike"`
	PackageManager   string   `json:"PackageManager"`
	Kernel           string   `json:"Kernel"`
	Arch             string   `json:"Arch"`
	NFSInstalled     bool     `json:"NFSInstalled"`
	NFSVersion       string   `json:"NFSVersion"`
	NFSServiceStatus string   `json:"NFSServiceStatus"`
	InstallCommands  []string `json:"InstallCommands"`
}

// GetHostSystemInfo 获取主机系统详细信息
func GetHostSystemInfo(w *apiserver.Response, r *apiserver.Request) {
	// 获取HostIP参数
	hostIP := r.Request.URL.Query().Get("HostIP")
	if hostIP == "" {
		hostIP = "127.0.0.1"
	}

	// 获取基础系统信息
	osRelease, osVersion := node.GetOS(hostIP)
	hostname := node.GetHostname(hostIP)
	kernel := node.GetKernelVersion(hostIP)
	arch := node.GetArch(hostIP)

	// 检测发行版信息
	distroInfo, err := node.DetectDistro(hostIP)
	if err != nil {
		flog.Errorf("Failed to detect distro for host %s: %v", hostIP, err)
		// 使用基础OS信息作为fallback
		distroInfo.ID = osRelease
		distroInfo.Version = osVersion
		distroInfo.PackageManager = node.GetPackageManager(osRelease, nil)
	}

	// 检查NFS-Ganesha安装状态
	installed, version, serviceStatus, err := node.CheckNFSGaneshaInstallation(hostIP)
	if err != nil {
		flog.Warnf("Failed to check NFS installation for host %s: %v", hostIP, err)
		// 继续返回其他信息
	}

	// 生成安装命令
	var installCommands []string
	if !installed {
		// 获取当前用户信息
		userinfo, ok := r.Session.UserInfo.(model.SessionUserInfo)
		currentUser := "admin" // 默认用户
		if ok {
			currentUser = userinfo.Username
		}
		installCommands = node.GetInstallCommands(distroInfo, currentUser)
	}

	// 构建响应
	response := HostSystemInfoResponse{
		HostIP:           hostIP,
		Hostname:         hostname,
		DistroID:         distroInfo.ID,
		DistroVersion:    distroInfo.Version,
		DistroIDLike:     "",
		PackageManager:   distroInfo.PackageManager,
		Kernel:           kernel,
		Arch:             arch,
		NFSInstalled:     installed,
		NFSVersion:       version,
		NFSServiceStatus: serviceStatus,
		InstallCommands:  installCommands,
	}

	// 将DistroIDLike数组转换为字符串
	if len(distroInfo.IDLike) > 0 {
		response.DistroIDLike = distroInfo.IDLike[0]
	}

	w.Write(retcode.StatusOK(response))
}
