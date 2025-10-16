package main

import (
	"context"
	"errors"
	flutenasf "flutelake/fluteNAS/frontend/flute-nas"
	"flutelake/fluteNAS/pkg/api"
	"flutelake/fluteNAS/pkg/controller"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/server/terminal"
	"flutelake/fluteNAS/pkg/util"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func init() {
	// init logger settings
	flog.NewLogger(1000)

	// init api router
	dataPath, err := initDataDir()
	if err != nil {
		flog.Fatal(err)
	}

	// init database
	err = initDB(dataPath)
	if err != nil {
		flog.Fatal(err)
	}
}

// type FluteNAS struct {
// 	// http api server
// 	server *apiserver.Apiserver
// 	// ssl cert
// 	publicKey  *util.LinkedRune
// 	privateKey *util.LinkedRune
// }

func main() {

	c := cache.NewMemoryCache()

	server := apiserver.NewApiserver(c)
	// 前端文件
	server.SetFrontendFS(flutenasf.FrontendFiles)
	privateKey, publicKey, err := util.GenerateRSAKeyPair(512)
	if err != nil {
		flog.Fatal(err)
	}
	// nas := &FluteNAS{
	// 	server:     server,
	// 	publicKey:  puclicKey,
	// 	privateKey: privateKey,
	// }

	ctx, cancel := context.WithCancel(context.Background())
	terms := terminal.NewWebTerminal(600)

	// init db host table data
	initSelfHost()

	// create flute user and group
	node.CreateFluteUserAndGroup()

	// register apis
	api.RegisterHandlersV1(server, privateKey, publicKey, c, terms)

	// start terminal service
	go terms.Start(ctx.Done())
	server.HandleFunc("/ws/v1/terminal", terms.WebSocketHandler)

	// start controller manager
	cron := controller.NewCronJob()
	err = initController(cron)
	if err != nil {
		flog.Fatal(err)
	}
	go cron.Start()

	if err := server.Run(ctx); err != nil {
		cancel()
		flog.Fatal(err)
	}
}

func initDataDir() (string, error) {
	r, e := os.Getwd()
	if e != nil {
		return "", e
	}
	p := filepath.Join(r, ".flute")
	_, e = os.Stat(p)
	if e != nil {
		if os.IsNotExist(e) {
			return p, os.Mkdir(p, 0o644)
		} else {
			return p, e
		}
	}
	return p, nil
}

func initDB(pStr string) error {
	db.InitDB(pStr)

	// Migrate the table schema
	err := db.Instance().AutoMigrate(
		&model.MountPoint{},
		&model.Host{},
		&model.SambaUser{},
		&model.SambaShare{},
	// &Network{},
	// &Host{},
	// &Operation{},
	)
	if err != nil {
		return err
	}

	return nil
}

func initController(cron *controller.CronJob) error {
	// 15s 检查一次挂载点
	err := cron.AddJob("checkMountPoint", "@every 15s", controller.NewStorageDeviceController().MountPoint)
	if err != nil {
		return err
	}

	// 15s 检查一次samba user
	err = cron.AddJob("sambaUser", "@every 15s", controller.NewSambaUsereController().Do)
	if err != nil {
		return err
	}

	// 15s 检查一次samba share
	err = cron.AddJob("sambaShare", "@every 15s", controller.NewSambaShareController().Do)
	if err != nil {
		return err
	}

	return nil
}

// initSelfHost 初始化本机信息
func initSelfHost() {
	var localhost model.Host
	err := db.Instance().First(&localhost, "host_ip = ?", model.LocalHost).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		flog.Fatalf("Error query localhost: %v", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 收集本机信息
		osRelease, version := node.GetOS(model.LocalHost)
		kernelVersion := node.GetKernelVersion(model.LocalHost)
		arch := node.GetArch(model.LocalHost)
		hostname := node.GetHostname(model.LocalHost)
		sshPort, err := node.GetLocalHostSshPort()
		if err != nil {
			flog.Fatalf("Error get ssh port: %v", err)
		}
		localhost = model.Host{
			ID:        model.LocalHost,
			HostIP:    model.LocalHost,
			OS:        osRelease,
			OSVersion: version,
			Arch:      arch,
			Kernel:    kernelVersion,
			Hostname:  hostname,
			SSHPort:   sshPort,
		}
		if err := db.Instance().Create(&localhost).Error; err != nil {
			flog.Fatalf("Error creating localhost record: %v", err)
		}
	}
}
