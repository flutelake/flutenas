package controller

import (
	"bytes"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/scylladb/go-set"
)

var sambaShareLock sync.Mutex

type SambaShareController struct {
}

func NewSambaShareController() *SambaShareController {
	return &SambaShareController{}
}

func (s *SambaShareController) Do() {
	if !sambaShareLock.TryLock() {
		return
	}
	defer sambaShareLock.Unlock()
	hosts := []model.Host{}
	queryRes := db.Instance().Find(&hosts)
	if queryRes.Error != nil {
		flog.Errorf("cannot query hosts from db, error: %v", queryRes.Error)
		return
	}
	for _, h := range hosts {
		s.DoOnHost(h)
	}

}

func (s *SambaShareController) DoOnHost(host model.Host) {
	smbShares := []model.SambaShare{}
	// 找出所有的samba 用户
	queryRes := db.Instance().Where("host_ip = ?", host.HostIP).Find(&smbShares)
	if queryRes.Error != nil {
		flog.Errorf("cannot query samba users from db, error: %v", queryRes.Error)
		return
	}

	updateIDs := []uint{}
	deleteIDs := []uint{}
	change := false
	exports := []SambaExport{}
	for _, s := range smbShares {
		switch s.Status {
		case model.SambaShareStatus_Init, model.SambaShareStatus_Updating:
			change = true
			updateIDs = append(updateIDs, s.ID)
		case model.SambaShareStatus_Deleting:
			change = true
			deleteIDs = append(deleteIDs, s.ID)
			continue
		}
		perms := s.UserPermissions.Get()
		vaildUsers := set.NewStringSet()
		writeUsers := set.NewStringSet()
		everyone := false
		everyoneWriteAble := false
		for _, acl := range perms {
			vaildUsers.Add(acl.Username)
			if acl.Permission == model.SambaACL_ReadWrite {
				writeUsers.Add(acl.Username)
			}
		}
		if len(perms) == 1 && perms[0].Username == model.SambaUser_Anonymous {
			everyone = true
			if perms[0].Permission == model.SambaACL_ReadWrite {
				everyoneWriteAble = true
			}
		}
		ex := SambaExport{
			ShareID:           s.Pseudo, //fmt.Sprintf("%d", s.ID),
			Path:              s.Path,
			ValidUsers:        strings.Join(vaildUsers.List(), " "),
			WriteUsers:        strings.Join(writeUsers.List(), " "),
			Everyone:          everyone,
			EveryoneWriteAble: everyoneWriteAble,
		}
		exports = append(exports, ex)
	}

	if !change {
		return
	}

	buf, err := BuildSambaExports(exports)
	if err != nil {
		flog.Errorf("build smb.conf failed, error: %v", err)
		return
	}
	content := buf.String()

	cmd := node.NewExec().SetHost(host.HostIP)
	err = cmd.WriteFile("/etc/samba/smb.conf", []byte(content), 0644)
	if err != nil {
		flog.Errorf("write smb.conf into host: %s, failed, error: %v", host.HostIP, err)
		return
	}

	_, err = cmd.Command("smbcontrol smbd reload-config")
	if err != nil {
		flog.Errorf("reload smb.conf on host: %s, failed, error: %v", host.HostIP, err)
		return
	}

	if len(updateIDs) > 0 {
		result := db.Instance().Model(&model.SambaShare{}).Where("ID IN ?", updateIDs).Updates(map[string]interface{}{
			"status":     model.SambaShareStatus_Active,
			"updated_at": time.Now(),
		})
		if result.Error != nil {
			flog.Errorf("update samba shares status failed, error: %v", err)
			return
		}
	}
	if len(deleteIDs) > 0 {
		result := db.Instance().Where("ID IN ?", updateIDs).Delete(&model.SambaShare{})
		if result.Error != nil {
			flog.Errorf("delete samba shares failed, error: %v", err)
			return
		}
	}

}

func BuildSambaExports(exports []SambaExport) (*bytes.Buffer, error) {
	var buff bytes.Buffer
	/*
		1. workgroup = SAMBA
		定义工作组名称为"SAMBA"
		这是 Windows 网络中的工作组名称，连接的 Windows 客户端需要在同一个工作组才能看到这个共享

		2. security = user
		设置安全级别为"用户级"验证
		表示访问共享时需要提供用户名和密码
		这是目前 Samba 推荐的安全模式

		3. passdb backend = tdbsam
		指定密码数据库后端为 tdbsam
		tdbsam 是一个轻量级的数据库，用于存储用户账号信息
		比早期的 smbpasswd 文件更安全和高效

		4. kernel share modes = no
		禁用内核共享模式
		这可以提高性能，因为它避免了不必要的文件锁定检查

		5. posix locking = no
		禁用 POSIX 锁定机制
		在某些情况下可以提高性能，特别是当多个客户端访问相同文件时

		6. kernel oplocks = yes
		启用内核级机会锁（opportunistic locks）
		机会锁可以提高性能，允许客户端在本地缓存文件
		当其他客户端需要访问文件时，服务器会通知持有机会锁的客户端
	*/
	buff.WriteString(`[global]
    workgroup = SAMBA
    security = user

    passdb backend = tdbsam

    kernel share modes = no
    posix locking = no
    kernel oplocks = yes

`)

	tmpl, err := template.New("smb_export").Parse(_sambaShareTemplate)
	if err != nil {
		return nil, err
	}

	for _, v := range exports {
		err = tmpl.Execute(&buff, v)
		if err != nil {
			return nil, err
		}
	}

	return &buff, nil
}

type SambaExport struct {
	ShareID           string
	Path              string
	ValidUsers        string
	WriteUsers        string
	Everyone          bool
	EveryoneWriteAble bool
}

const _sambaShareTemplate = `
[{{.ShareID}}]
    path = {{.Path}}
    writeable = yes
	{{- if .EveryoneWriteAble}}
    read only = yes
	{{- else}}
    read only = no
	{{- end}}
	{{- if .Everyone}}
    guest ok = yes
    browseable = yes
    public = yes
	{{- else}}
    guest ok = no
    valid users = {{.ValidUsers}}
    write list = {{.WriteUsers}}
	{{- end}}
    create mask = 0644
    directory mask = 0755
`

// force user = root
// force group = root
