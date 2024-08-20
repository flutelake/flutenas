package main

import (
	"context"
	"flutelake/fluteNAS/pkg/api"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/server/terminal"
	"flutelake/fluteNAS/pkg/util"
	"os"
	"path/filepath"
)

func init() {
	// init logger settings
	flog.NewLogger(1000)

	// init api router
	_, err := initDataDir()
	if err != nil {
		flog.Fatal(err)
	}

	// init database
	// err = initDB(dataPath)
	// if err != nil {
	// 	flog.Fatal(err)
	// }
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

	// register apis
	api.RegisteHandlersV1(server, privateKey, publicKey, c, terms)

	// start terminal service
	go terms.Start(ctx.Done())
	server.HandleFunc("/ws/v1/terminal", terms.WebSocketHandler)

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
	// &Network{},
	// &Host{},
	// &Operation{},
	)
	if err != nil {
		return err
	}

	return nil
}
