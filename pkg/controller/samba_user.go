package controller

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"strings"
	"sync"
)

var sambaUserLock sync.Mutex

type SambaUserController struct {
}

func NewSambaUsereController() *SambaUserController {
	return &SambaUserController{}
}

func (s *SambaUserController) Do() {
	if !sambaUserLock.TryLock() {
		return
	}
	defer sambaUserLock.Unlock()
	// flog.Debugf("start to check samba users...")
	smbUsers := []model.SambaUser{}
	// 找出所有的samba 用户
	queryRes := db.Instance().Find(&smbUsers)
	if queryRes.Error != nil {
		flog.Errorf("cannot query samba users from db, error: %v", queryRes.Error)
		return
	}

	for _, u := range smbUsers {
		switch u.Status {
		case model.SambaUserStatus_Active:
			// 检查用户是否存在
			if err := s.checkSambaUser(&u); err != nil {
				flog.Errorf("failed to check samba user %s: %v", u.Username, err)
				continue
			}
		case model.SambaUserStatus_Init:
			if err := s.createSambaUser(&u); err != nil {
				flog.Errorf("failed to create samba user %s: %v", u.Username, err)
				continue
			}
		case model.SambaUserStatus_ChangingPWD:
			if err := s.updateSambaUserPassword(&u); err != nil {
				flog.Errorf("failed to update samba user password %s: %v", u.Username, err)
				continue
			}
		case model.SambaUserStatus_Deleting:
			if err := s.deleteSambaUser(&u); err != nil {
				flog.Errorf("failed to delete samba user %s: %v", u.Username, err)
				continue
			}
		}
	}
}

// createSambaUser 创建新的Samba用户
func (s *SambaUserController) createSambaUser(user *model.SambaUser) error {
	cmd := node.NewExec().SetHost(user.HostIP)
	// 1. 创建系统用户
	// useradd -M not create home directory
	// useradd -s 指定登录shell，如果不指定会默认使用/bin/bash
	if bs, err := cmd.Command(fmt.Sprintf("id %s || useradd -M -s /sbin/nologin %s", user.Username, user.Username)); err != nil {
		return fmt.Errorf("failed to create system user: %v, stdout: %s", err, string(bs))
	}

	// 2. 创建Samba用户并设置密码
	bs, err := cmd.Command(fmt.Sprintf("(echo %s; echo %s) | smbpasswd -a %s", user.Password, user.Password, user.Username))
	if err != nil {
		return fmt.Errorf("failed to create samba user: %v, stdout: %s", err, string(bs))
	}

	// 3. 更新用户状态为激活
	if err := db.Instance().Model(user).Update("status", model.SambaUserStatus_Active).Error; err != nil {
		return fmt.Errorf("failed to update user status: %v", err)
	}
	flog.Infof("create samba user: %s successed", user.Username)
	return nil
}

// updateSambaUserPassword 更新Samba用户密码
func (s *SambaUserController) updateSambaUserPassword(user *model.SambaUser) error {
	cmd := node.NewExec().SetHost(user.HostIP)
	// 1. 新Samba用户密码
	bs, err := cmd.Command(fmt.Sprintf("(echo %s; echo %s) | smbpasswd -a %s", user.Password, user.Password, user.Username))
	if err != nil {
		return fmt.Errorf("failed to update samba password: %v, stdout: %s", err, string(bs))
	}

	// 3. 更新用户状态为激活
	if err := db.Instance().Model(user).Update("status", model.SambaUserStatus_Active).Error; err != nil {
		return fmt.Errorf("failed to update user status: %v", err)
	}

	return nil
}

// deleteSambaUser 删除Samba用户
func (s *SambaUserController) deleteSambaUser(user *model.SambaUser) error {
	cmd := node.NewExec().SetHost(user.HostIP)
	// 1. 检查samba用户是否存在
	bs, err := cmd.Command(fmt.Sprintf("pdbedit --list | grep %s: | wc -l", user.Username))
	if err != nil {
		return fmt.Errorf("pdbedit failed to list samba user: %v, stdout: %s", err, string(bs))
	}
	if util.Trim(string(bs)) != "0" {
		// 2. 删除samba用户
		bs, err = cmd.Command(fmt.Sprintf("pdbedit --delete --user=%s", user.Username))
		if err != nil {
			return fmt.Errorf("pdbedit failed to delete samba user: %v, stdout: %s", err, string(bs))
		}
	}

	// 3. 检查系统用户是否存在
	bs, err = cmd.Command(fmt.Sprintf("cat /etc/passwd | grep %s: | wc -l", user.Username))
	if err != nil {
		return fmt.Errorf("cat os user failed: %v, stdout: %s", err, string(bs))
	}
	// 4. 删除系统用户
	if util.Trim(string(bs)) != "0" {
		bs, err := cmd.Command(fmt.Sprintf("userdel -r %s", user.Username))
		if err != nil {
			return fmt.Errorf("delete os user failed: %v, stdout: %s", err, string(bs))
		}
	}

	// 3. 从数据库中删除用户记录
	if err := db.Instance().Delete(user).Error; err != nil {
		return fmt.Errorf("failed to delete user from database: %v", err)
	}
	return nil
}

func (s *SambaUserController) checkSambaUser(user *model.SambaUser) error {
	cmd := node.NewExec().SetHost(user.HostIP)

	// 检查os samba user 是否存在
	bs, err := cmd.Command(fmt.Sprintf("id %s", user.Username))
	if err != nil {
		if strings.Contains(err.Error(), "no such user") {
			// 不存在 则去创建
			return s.createSambaUser(user)
		}
		flog.Errorf("check samba user on host: %s, exec output: %s, error: %v", user.HostIP, string(bs), err)
		return err
	}

	bs, err = cmd.Command(fmt.Sprintf("pdbedit -L %s", user.Username))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// 不存在 则去创建
			return s.createSambaUser(user)
		}
		flog.Errorf("pdbedit check samba user on host: %s, exec output: %s, error: %v", user.HostIP, string(bs), err)
		return err
	}

	return nil
}
