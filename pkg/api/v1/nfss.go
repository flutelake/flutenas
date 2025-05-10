package v1

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/util"
)

type NFSShareServer struct{}

// OpenNFSServerRequestResponse 定义开启/关闭NFS服务器的响应
type OpenNFSServerRequestResponse struct {
	Status string `json:"Status"`
}

// NFSExportResponse 定义NFS导出响应
type NFSExportResponse struct {
	Export model.NFSExport `json:"Export"`
}

// NFSExportsResponse 定义NFS导出列表响应
type NFSExportsResponse struct {
	Exports []model.NFSExport `json:"Exports"`
}

type DeleteNFSExportRequest struct {
	ID uint `json:"ID" validate:"required"`
}

// OpenNFSServer 开启或关闭NFS服务器
// func (s *NFSShareServer) OpenNFSServer(w *apiserver.Response, r *apiserver.Request) {
// 	in := &model.OpenNFSServerRequest{}
// 	if err := r.Unmarshal(in); err != nil {
// 		w.WriteError(err, retcode.StatusError(nil))
// 		return
// 	}

// 	// TODO: 实现开启或关闭NFS服务器的逻辑

// 	out := &OpenNFSServerRequestResponse{
// 		Status: "success",
// 	}
// 	w.Write(retcode.StatusOK(out))
// }

// CreateNFSExport 创建NFS共享
func (s *NFSShareServer) CreateNFSExport(w *apiserver.Response, r *apiserver.Request) {
	// 解析请求
	export := &model.NFSExport{}
	if err := r.Unmarshal(export); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 验证ACLs格式
	if export.Acls != "" {
		var acls []model.NFSAcl
		if err := json.Unmarshal([]byte(export.Acls), &acls); err != nil {
			w.WriteError(errors.New("ACLs格式无效，应为有效的JSON数组"), retcode.StatusError(nil))
			return
		}
	}

	// 创建记录
	if result := db.Instance().Create(export); result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}

	// 返回响应
	response := &NFSExportResponse{
		Export: *export,
	}
	w.Write(retcode.StatusOK(response))
}

// DeleteNFSExport 删除NFS共享
func (s *NFSShareServer) DeleteNFSExport(w *apiserver.Response, r *apiserver.Request) {
	// 获取ID参数
	in := &DeleteNFSExportRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 查找记录
	var export model.NFSExport
	if result := db.Instance().First(&export, in.ID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			w.WriteError(errors.New("未找到NFS共享"), retcode.StatusError(nil))
		} else {
			w.WriteError(result.Error, retcode.StatusError(nil))
		}
		return
	}

	// 删除记录
	if result := db.Instance().Delete(&export); result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}

	// 返回成功响应
	w.Write(retcode.StatusOK(map[string]string{"message": "NFS共享已成功删除"}))
}

// 添加UpdateNFSExportRequest结构体
type UpdateNFSExportRequest struct {
	ID        uint   `json:"ID" validate:"required"`
	HostIP    string `json:"HostIP"`
	Name      string `json:"Name"`
	Path      string `json:"Path"`
	Pseudo    string `json:"Pseudo"`
	Acls      string `json:"Acls"`
	Protocols string `json:"Protocols"`
}

// UpdateNFSExport 更新NFS共享
func (s *NFSShareServer) UpdateNFSExport(w *apiserver.Response, r *apiserver.Request) {
	// 解析更新请求
	in := &UpdateNFSExportRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 查找现有记录
	var existingExport model.NFSExport
	if result := db.Instance().First(&existingExport, in.ID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			w.WriteError(errors.New("未找到NFS共享"), retcode.StatusError(nil))
		} else {
			w.WriteError(result.Error, retcode.StatusError(nil))
		}
		return
	}

	// 验证ACLs格式
	if in.Acls != "" {
		var acls []model.NFSAcl
		if err := json.Unmarshal([]byte(in.Acls), &acls); err != nil {
			w.WriteError(errors.New("ACLs格式无效，应为有效的JSON数组"), retcode.StatusError(nil))
			return
		}
	}

	// 更新记录
	updatedExport := model.NFSExport{
		ID:        in.ID,
		HostIP:    in.HostIP,
		Name:      in.Name,
		Path:      in.Path,
		Pseudo:    in.Pseudo,
		Acls:      in.Acls,
		Protocols: in.Protocols,
	}

	// 保留创建时间
	updatedExport.CreatedAt = existingExport.CreatedAt

	// 保存更新
	if result := db.Instance().Save(&updatedExport); result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}

	// 返回响应
	response := &NFSExportResponse{
		Export: updatedExport,
	}
	w.Write(retcode.StatusOK(response))
}

// NFSExportRequest 定义NFS共享请求参数
type NFSExportRequest struct {
	HostIP string `json:"HostIP"`
}

// ListNFSExports 列出指定主机IP的NFS共享
func (s *NFSShareServer) ListNFSExports(w *apiserver.Response, r *apiserver.Request) {
	// 从请求体中获取参数
	var in NFSExportRequest
	if err := r.Unmarshal(&in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 查询记录
	var exports []model.NFSExport
	query := db.Instance()
	if in.HostIP != "" {
		query = query.Where("host_ip = ?", in.HostIP)
	}

	if result := query.Find(&exports); result.Error != nil {
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}

	// 返回响应
	response := &NFSExportsResponse{
		Exports: exports,
	}
	w.Write(retcode.StatusOK(response))
}

type NFSStatusRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

type NFSStatusResponse struct {
	Installed bool `json:"Installed"`
	Actived   bool `json:"Actived"`
}

func (s *NFSShareServer) NFSStatus(w *apiserver.Response, r *apiserver.Request) {
	in := &NFSStatusRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	cmd := node.NewExec().SetHost(in.HostIP)

	checkInstalled := `
        if command -v ganesha.nfsd >/dev/null 2>&1; then
            echo "installed"
        else
            echo "not_installed"
        fi`

	resBs, err := cmd.Command(checkInstalled)
	if err != nil {
		flog.Errorf("check NFS-Ganesha is installed on host: %s, error: %v, stdout: %s", in.HostIP, err, string(resBs))
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	if util.Trim(string(resBs)) == "not_installed" {
		installScript := `
            if command -v apt-get >/dev/null 2>&1; then
                # Debian/Ubuntu
                apt-get update && apt-get install -y nfs-ganesha
                systemctl enable nfs-ganesha
                systemctl start nfs-ganesha
            elif command -v yum >/dev/null 2>&1; then
                # CentOS/RHEL
                yum -y install nfs-ganesha
                systemctl enable nfs-ganesha
                systemctl start nfs-ganesha
            elif command -v dnf >/dev/null 2>&1; then
                # Fedora
                dnf -y install nfs-ganesha
                systemctl enable nfs-ganesha
                systemctl start nfs-ganesha
            else
                echo "Unsupported operating system"
                exit 1
            fi`

		resBs, err = cmd.Command(installScript)
		if err != nil {
			flog.Errorf("try to install NFS-Ganesha service on host: %s, error: %v, stdout: %s", in.HostIP, err, string(resBs))
			w.WriteError(err, retcode.StatusError(nil))
			return
		}
	}
	// 检查 NFS-Ganesha 服务状态
	checkActive := `systemctl is-active nfs-ganesha`
	activeResult, err := cmd.Command(checkActive)
	if err != nil {
		flog.Errorf("get NFS-Ganesha service status on host: %s, error: %v, stdout: %s", in.HostIP, err, string(resBs))
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := NFSStatusResponse{
		Installed: util.Trim(string(resBs)) == "installed",
		Actived:   util.Trim(string(activeResult)) == "active",
	}

	w.Write(retcode.StatusOK(out))
}
