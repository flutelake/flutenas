package v1

import (
	"encoding/json"
	"errors"
	"fmt"

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
	ID         uint   `json:"ID" validate:"required"`
	HostIP     string `json:"HostIP"`
	Name       string `json:"Name"`
	Path       string `json:"Path"`
	Pseudo     string `json:"Pseudo"`
	DefaultACL string `json:"DefaultACL"`
	Acls       string `json:"Acls"`
	Protocols  string `json:"Protocols"`
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
		ID:         in.ID,
		HostIP:     in.HostIP,
		Name:       in.Name,
		Path:       in.Path,
		Pseudo:     in.Pseudo,
		DefaultACL: in.DefaultACL,
		Acls:       in.Acls,
		Protocols:  in.Protocols,
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
		// 安装NFS-Ganesha，带apt锁检测和等待机制
		installScript := `
            # 检测操作系统类型
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

// StartNFSServerRequest 启动NFS服务请求
type StartNFSServerRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

// StartNFSServerResponse 启动NFS服务响应
type StartNFSServerResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

// StopNFSServerRequest 停止NFS服务请求
type StopNFSServerRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

// StopNFSServerResponse 停止NFS服务响应
type StopNFSServerResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

// GetNFSServerStatusRequest 获取NFS服务状态请求
type GetNFSServerStatusRequest struct {
	HostIP string `json:"HostIP" validate:"required"`
}

// GetNFSServerStatusResponse 获取NFS服务状态响应
type GetNFSServerStatusResponse struct {
	Status  string `json:"Status"`
	Uptime  string `json:"Uptime"`
	Message string `json:"Message"`
}

// StartNFSServer 启动NFS服务
func (s *NFSShareServer) StartNFSServer(w *apiserver.Response, r *apiserver.Request) {
	in := &StartNFSServerRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 记录审计日志
	flog.Infof("[AUDIT][NFS] User: %s, Action: start-service, Host: %s",
		getCurrentUser(r), in.HostIP)

	// 启动服务
	if err := node.StartNFSServerControl(in.HostIP); err != nil {
		flog.Errorf("start NFS server on host: %s, error: %v", in.HostIP, err)
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &StartNFSServerResponse{
		Status:  "running",
		Message: "NFS服务已启动",
	}
	w.Write(retcode.StatusOK(out))
}

// StopNFSServer 停止NFS服务
func (s *NFSShareServer) StopNFSServer(w *apiserver.Response, r *apiserver.Request) {
	in := &StopNFSServerRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 记录审计日志
	flog.Infof("[AUDIT][NFS] User: %s, Action: stop-service, Host: %s",
		getCurrentUser(r), in.HostIP)

	// 停止服务
	if err := node.StopNFSServerControl(in.HostIP); err != nil {
		flog.Errorf("stop NFS server on host: %s, error: %v", in.HostIP, err)
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &StopNFSServerResponse{
		Status:  "stopped",
		Message: "NFS服务已停止",
	}
	w.Write(retcode.StatusOK(out))
}

// GetNFSServerStatus 获取NFS服务状态
func (s *NFSShareServer) GetNFSServerStatus(w *apiserver.Response, r *apiserver.Request) {
	in := &GetNFSServerStatusRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 获取服务状态
	status, uptime, err := node.GetNFSServerStatusControl(in.HostIP)
	if err != nil {
		flog.Errorf("get NFS server status on host: %s, error: %v", in.HostIP, err)
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &GetNFSServerStatusResponse{
		Status:  status,
		Uptime:  uptime,
		Message: fmt.Sprintf("NFS服务状态: %s", status),
	}
	w.Write(retcode.StatusOK(out))
}

// ValidateNFSConfigRequest 验证NFS配置请求
type ValidateNFSConfigRequest struct {
	ConfigPath string `json:"ConfigPath"`
}

// ValidateNFSConfigResponse 验证NFS配置响应
type ValidateNFSConfigResponse struct {
	Valid   bool     `json:"Valid"`
	Errors  []string `json:"Errors"`
	Message string   `json:"Message"`
}

// ValidateNFSConfig 验证NFS配置文件语法
func (s *NFSShareServer) ValidateNFSConfig(w *apiserver.Response, r *apiserver.Request) {
	in := &ValidateNFSConfigRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 如果未指定配置路径，使用默认路径
	configPath := in.ConfigPath
	if configPath == "" {
		configPath = "/etc/ganesha/ganesha.conf"
	}

	// 验证配置
	if err := node.ValidateNFSConfigFile(configPath); err != nil {
		out := &ValidateNFSConfigResponse{
			Valid:   false,
			Errors:  []string{err.Error()},
			Message: "配置验证失败",
		}
		w.Write(retcode.StatusOK(out))
		return
	}

	out := &ValidateNFSConfigResponse{
		Valid:   true,
		Errors:  []string{},
		Message: "配置验证通过",
	}
	w.Write(retcode.StatusOK(out))
}

// UpdateExportStatusRequest 更新导出规则状态请求
type UpdateExportStatusRequest struct {
	ID     uint   `json:"ID" validate:"required"`
	Status string `json:"Status" validate:"required,oneof=enabled disabled"`
}

// UpdateExportStatusResponse 更新导出规则状态响应
type UpdateExportStatusResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

// UpdateExportStatus 更新导出规则状态（启用/禁用）
func (s *NFSShareServer) UpdateExportStatus(w *apiserver.Response, r *apiserver.Request) {
	in := &UpdateExportStatusRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 验证状态值
	if in.Status != "enabled" && in.Status != "disabled" {
		w.WriteError(errors.New("状态必须是enabled或disabled"), retcode.StatusError(nil))
		return
	}

	// 记录审计日志
	flog.Infof("[AUDIT][NFS] User: %s, Action: update-status, ID: %d, Status: %s",
		getCurrentUser(r), in.ID, in.Status)

	// 更新状态
	dbInstance := db.Instance()
	if err := model.UpdateStatus(dbInstance, in.ID, in.Status); err != nil {
		flog.Errorf("update NFS export status failed, ID: %d, error: %v", in.ID, err)
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 获取更新后的导出规则
	var updatedExport model.NFSExport
	if result := dbInstance.First(&updatedExport, in.ID); result.Error != nil {
		flog.Errorf("get NFS export after status update failed, ID: %d, error: %v", in.ID, result.Error)
		w.WriteError(result.Error, retcode.StatusError(nil))
		return
	}

	// 尝试触发配置同步（异步进行，不阻塞响应）
	go func() {
		if err := triggerNFSConfigSync(updatedExport.HostIP); err != nil {
			flog.Errorf("trigger NFS config sync after status update failed, HostIP: %s, error: %v",
				updatedExport.HostIP, err)
		}
	}()

	out := &UpdateExportStatusResponse{
		Status:  in.Status,
		Message: fmt.Sprintf("导出规则状态已更新为%s", in.Status),
	}
	w.Write(retcode.StatusOK(out))
}

// TestExportConfigRequest 测试导出规则配置请求
type TestExportConfigRequest struct {
	Exports []model.NFSExport `json:"Exports" validate:"required"`
}

// TestExportConfigResponse 测试导出规则配置响应
type TestExportConfigResponse struct {
	Valid   bool     `json:"Valid"`
	Errors  []string `json:"Errors"`
	Message string   `json:"Message"`
}

// TestExportConfig 测试导出规则配置（模拟NFS-Ganesha配置加载）
func (s *NFSShareServer) TestExportConfig(w *apiserver.Response, r *apiserver.Request) {
	in := &TestExportConfigRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// 记录审计日志
	flog.Infof("[AUDIT][NFS] User: %s, Action: test-export-config, ExportCount: %d",
		getCurrentUser(r), len(in.Exports))

	// 过滤出启用的导出规则
	var enabledExports []model.NFSExport
	for _, export := range in.Exports {
		if export.Status == "enabled" {
			enabledExports = append(enabledExports, export)
		}
	}

	// 生成临时配置并验证
	testResult, err := node.TestNFSExportConfig(enabledExports)
	if err != nil {
		out := &TestExportConfigResponse{
			Valid:   false,
			Errors:  []string{err.Error()},
			Message: "导出规则配置测试失败",
		}
		w.Write(retcode.StatusOK(out))
		return
	}

	out := &TestExportConfigResponse{
		Valid:   testResult.Valid,
		Errors:  testResult.Errors,
		Message: testResult.Message,
	}
	w.Write(retcode.StatusOK(out))
}

// triggerNFSConfigSync 触发NFS配置同步
func triggerNFSConfigSync(hostIP string) error {
	// 获取所有启用的导出规则
	exports, err := model.GetEnabledByHostIP(db.Instance(), hostIP)
	if err != nil {
		return fmt.Errorf("get enabled exports for host %s failed: %w", hostIP, err)
	}

	// 使用exec来生成并应用配置
	cmd := node.NewExec().SetHost(hostIP)
	if err := cmd.RefreshNFSGaneshaConfig(exports); err != nil {
		return fmt.Errorf("refresh NFS config for host %s failed: %w", hostIP, err)
	}

	return nil
}

// getCurrentUser 获取当前用户（简化实现）
func getCurrentUser(r *apiserver.Request) string {
	// TODO: 从session或token中获取实际用户名
	return "admin"
}
