package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"
)

type OpenNFSServerRequest struct {
	Open string `json:"Open" validate:"required,oneof=open close"`
}

type NFSExport struct {
	gorm.Model
	ID          uint       `json:"ID" gorm:"primaryKey,autoIncrement"`
	HostIP      string     `json:"HostIP" gorm:"not null;index"`
	Name        string     `json:"Name" gorm:"not null"`
	Path        string     `json:"Path" gorm:"not null"`
	Pseudo      string     `json:"Pseudo" gorm:"not null;uniqueIndex:idx_host_pseudo"`
	DefaultACL  string     `json:"DefaultACL" gorm:"not null;default:'None'"`
	Acls        string     `json:"Acls" gorm:"not null;default:'[]'"`
	Protocols   string     `json:"Protocols" gorm:"not null;default:'3,4'"`
	Status      string     `json:"Status" gorm:"not null;default:'enabled';index"`
	LastApplied *time.Time `json:"LastApplied" gorm:"nullable"`
	TestResult  string     `json:"TestResult" gorm:"nullable"`
	CreatedAt   time.Time  `json:"CreatedAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"UpdatedAt" gorm:"autoUpdateTime"`
	// AclsMapped  []NFSAcl   `json:"AclsMapped" gorm:"-"`
}

type NFSExportMapped struct {
	NFSExport
	AclsMapped []NFSAcl `json:"AclsMapped"`
}

// TableName 指定表名
func (NFSExport) TableName() string {
	return "nfs_exports"
}

// GetAcls 解析ACL JSON
func (e *NFSExport) GetAcls() ([]NFSAcl, error) {
	var acls []NFSAcl
	if err := json.Unmarshal([]byte(e.Acls), &acls); err != nil {
		return nil, err
	}
	return acls, nil
}

// SetAcls 设置ACL JSON
func (e *NFSExport) SetAcls(acls []NFSAcl) error {
	data, err := json.Marshal(acls)
	if err != nil {
		return err
	}
	e.Acls = string(data)
	return nil
}

// MarshalJSON customizes JSON marshaling for NFSExport
func (e NFSExport) MarshalJSON() ([]byte, error) {
	type Alias NFSExport
	acls, err := e.GetAcls()
	if err != nil {
		// If parsing fails, return empty array
		acls = []NFSAcl{}
	}

	return json.Marshal(&struct {
		Acls []NFSAcl `json:"Acls"`
		*Alias
	}{
		Acls:  acls,
		Alias: (*Alias)(&e),
	})
}

// UnmarshalJSON customizes JSON unmarshaling for NFSExport
func (e *NFSExport) UnmarshalJSON(data []byte) error {
	type Alias NFSExport
	aux := &struct {
		Acls []NFSAcl `json:"Acls"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert the ACL slice back to JSON string for storage
	if aux.Acls != nil {
		jsonData, err := json.Marshal(aux.Acls)
		if err != nil {
			return err
		}
		e.Acls = string(jsonData)
	} else {
		e.Acls = "[]"
	}

	// Copy other fields
	*e = NFSExport(*aux.Alias)

	return nil
}

type NFSAcl struct {
	IPRange    string `json:"IPRange"`
	Permission string `json:"Permission"`
}

// Validate 验证ACL规则
func (a *NFSAcl) Validate() error {
	if a.IPRange == "" {
		return errors.New("IPRange不能为空")
	}
	if a.Permission != "RO" && a.Permission != "RW" {
		return errors.New("Permission必须是RO或RW")
	}
	// 验证IPRange格式（简化验证）
	if !strings.Contains(a.IPRange, "/") && net.ParseIP(a.IPRange) == nil {
		return errors.New("IPRange格式无效")
	}
	return nil
}

// UpdateStatus 更新导出规则状态
func UpdateStatus(db *gorm.DB, id uint, status string) error {
	if status != "enabled" && status != "disabled" {
		return fmt.Errorf("无效的状态: %s, 必须是enabled或disabled", status)
	}

	result := db.Model(&NFSExport{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("导出规则不存在")
	}
	return nil
}

// UpdateLastApplied 更新最后同步时间
func UpdateLastApplied(db *gorm.DB, id uint, t time.Time) error {
	result := db.Model(&NFSExport{}).Where("id = ?", id).Update("last_applied", t)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("导出规则不存在")
	}
	return nil
}

// UpdateTestResult 更新测试结果
func UpdateTestResult(db *gorm.DB, id uint, result string) error {
	resultDB := db.Model(&NFSExport{}).Where("id = ?", id).Update("test_result", result)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	if resultDB.RowsAffected == 0 {
		return errors.New("导出规则不存在")
	}
	return nil
}

// GetEnabledByHostIP 获取启用的导出规则（控制器同步使用）
func GetEnabledByHostIP(db *gorm.DB, hostIP string) ([]NFSExport, error) {
	var exports []NFSExport
	err := db.Where("host_ip = ? AND status = ?", hostIP, "enabled").Find(&exports).Error
	if err != nil {
		return nil, err
	}
	return exports, nil
}
