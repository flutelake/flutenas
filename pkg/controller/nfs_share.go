package controller

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
)

var nfsExporterLock sync.Mutex

// NFSShareController NFS分享控制器
type NFSShareController struct {
}

// NewNFSShareController 创建NFS分享控制器
func NewNFSShareController() *NFSShareController {
	return &NFSShareController{}
}

// syncNFSConfig 同步NFS配置
func (c *NFSShareController) Do() {
	if !nfsExporterLock.TryLock() {
		return
	}

	defer nfsExporterLock.Unlock()
	startTime := time.Now()
	flog.Debugf("Starting NFS config sync...")

	// 获取所有启用的NFS导出规则
	var exports []model.NFSExport
	err := db.Instance().Where("status = ?", "enabled").Find(&exports).Error
	if err != nil {
		flog.Errorf("Failed to fetch enabled NFS exports: %v", err)
		return
	}

	if len(exports) == 0 {
		flog.Debugf("No enabled NFS exports found, skipping sync")
		return
	}

	// 检查nfs-ganesha服务是否运行，未运行则启动
	// 检查nfs-ganesha服务是否设置开机自启，未设置则设置
	if err := node.CheckAndMaintainNFSService(); err != nil {
		flog.Warnf("nfs-ganesha service checking, error: %v", err)
	}

	// 按HostIP分组处理
	hostExports := make(map[string][]model.NFSExport)
	for _, export := range exports {
		hostIP := export.HostIP
		if hostIP == "" {
			hostIP = "127.0.0.1" // 默认本地
		}
		hostExports[hostIP] = append(hostExports[hostIP], export)
	}

	// 为每个主机生成配置并热重载
	for hostIP, hostExports := range hostExports {
		if err := c.syncHostNFSConfig(hostIP, hostExports); err != nil {
			flog.Errorf("Failed to sync NFS config for host %s: %v", hostIP, err)
		}
	}

	flog.Debugf("NFS config sync completed in %v", time.Since(startTime))
}

// syncHostNFSConfig 同步指定主机的NFS配置
func (c *NFSShareController) syncHostNFSConfig(hostIP string, exports []model.NFSExport) error {
	if len(exports) == 0 {
		return nil
	}

	flog.Infof("Syncing NFS config for host %s, %d exports", hostIP, len(exports))

	// 步骤1: 生成NFS配置（使用标准化函数）
	config, err := node.GenerateNFSConfig(exports)
	if err != nil {
		return fmt.Errorf("failed to generate NFS config: %w", err)
	}

	// 步骤2: 检查配置文件是否发生变化
	if err := node.CompareAndReplaceNFSConfig(hostIP, config); err != nil {
		// If configs are identical, skip the update process
		if strings.Contains(err.Error(), "config unchanged") {
			flog.Infof("NFS config for host %s is already up to date, skipping update", hostIP)
			return nil
		}
		return fmt.Errorf("NFS config comparison failed: %w", err)
	}

	// 步骤3: 备份当前配置
	if err := c.BackupNFSConfig(hostIP); err != nil {
		flog.Warnf("Failed to backup current NFS config before update: %v", err)
		// Continue even if backup fails
	}

	// 步骤4: 写入配置文件
	configPath := "/etc/ganesha/ganesha.conf"
	if err := node.WriteFile(hostIP, configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write NFS config: %w", err)
	}

	// 步骤5: 热重载NFS服务（使用 standardized function）
	if err := node.ReloadNFSConfig(hostIP); err != nil {
		// 热重载失败，尝试回滚
		flog.Errorf("NFS hot reload failed, attempting rollback: %v", err)
		if rollbackErr := c.RollbackNFSConfig(hostIP); rollbackErr != nil {
			flog.Errorf("Failed to rollback NFS config: %v", rollbackErr)
			return fmt.Errorf("NFS reload failed: %w, rollback also failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("NFS reload failed: %w, config has been rolled back", err)
	}

	// 步骤6: 更新最后同步时间
	now := time.Now()
	for _, export := range exports {
		if err := model.UpdateLastApplied(db.Instance(), export.ID, now); err != nil {
			flog.Warnf("Failed to update last_applied for export %d: %v", export.ID, err)
		}
	}

	flog.Infof("Successfully synced NFS config for host %s", hostIP)
	return nil
}

// BackupNFSConfig 备份当前NFS配置
func (c *NFSShareController) BackupNFSConfig(hostIP string) error {
	configPath := "/etc/ganesha/ganesha.conf"
	backupPath := fmt.Sprintf("/etc/ganesha/ganesha.conf.backup.%d", time.Now().Unix())

	if err := node.BackupFile(hostIP, configPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup NFS config: %w", err)
	}

	flog.Infof("NFS config backed up to %s for host %s", backupPath, hostIP)
	return nil
}

// RollbackNFSConfig 回滚NFS配置
func (c *NFSShareController) RollbackNFSConfig(hostIP string) error {
	configPath := "/etc/ganesha/ganesha.conf"

	// Find the most recent backup file
	cmd := node.NewExec().SetHost(hostIP)
	defer cmd.Close()

	findCmd := "ls -t /etc/ganesha/ganesha.conf.backup.* 2>/dev/null | head -n1"
	backupPathOutput, err := cmd.Command(findCmd)
	if err != nil || len(backupPathOutput) == 0 {
		return fmt.Errorf("no backup file found for rollback: %w", err)
	}

	backupPath := strings.TrimSpace(string(backupPathOutput)) // remove newline and whitespace

	if backupPath == "" {
		return fmt.Errorf("no backup file found for rollback")
	}

	// Move backup to current config
	if err := node.MoveFile(hostIP, backupPath, configPath); err != nil {
		return fmt.Errorf("failed to rollback NFS config: %w", err)
	}

	// Reload the rolled-back config
	if reloadErr := node.ReloadNFSConfig(hostIP); reloadErr != nil {
		return fmt.Errorf("rollback successful but reload failed: %w", reloadErr)
	}

	flog.Infof("NFS config rolled back from %s for host %s", backupPath, hostIP)
	return nil
}
