package api

import (
	v1 "flutelake/fluteNAS/pkg/api/v1"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/server/terminal"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
)

func RegisterHandlersV1(
	as *apiserver.Apiserver,
	privateKey *util.LinkedRune,
	publicKey *util.LinkedRune,
	c cache.TinyCache,
	terms *terminal.WebTerminal,
) {
	const prefix string = "/v1"

	authApi := v1.NewAuthApi(privateKey, publicKey, c)
	termApi := v1.NewTerminalAPI(terms)
	wallpaperApi := v1.NewWallpapaerAPI(c)

	// check login status api
	as.Register(as.NewRoute().Prefix(prefix).Path("/hello").Handler(HelloFluteNAS))
	// =================================== public apis ===================================== //
	as.Register(as.NewRoute().Prefix(prefix).Path("/login").Handler(authApi.Login).AllowAnonymous(true))
	as.Register(as.NewRoute().Prefix(prefix).Path("/key").Handler(authApi.GetKey).AllowAnonymous(true))
	as.Register(as.NewRoute().Prefix(prefix).Path("/wallpaper").Handler(wallpaperApi.GetWallpaper).AllowAnonymous(true))

	//==================================== private apis ==================================== //
	as.Register(as.NewRoute().Prefix(prefix).Path("/logout").Handler(authApi.Logout))

	// web terminal
	as.Register(as.NewRoute().Prefix(prefix).Path("/terminal").Handler(termApi.CreateTerminal))

	// file download server
	fserver := v1.NewFileServer(c, "/mnt")
	as.HandleFunc("/files/download", fserver.ServerHttp)
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/listdir").Handler(v1.ListDir))
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/readdir").Handler(v1.ReadDir))
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/createdir").Handler(v1.CreateDir))
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/remove").Handler(v1.RemoveFile))
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/upload").Handler(v1.UploadFiles))
	as.Register(as.NewRoute().Prefix(prefix).Path("/files/download").Handler(fserver.DownloadFiles))

	// disk device
	as.Register(as.NewRoute().Prefix(prefix).Path("/disk/list").Handler(v1.ListDiskDevices))
	as.Register(as.NewRoute().Prefix(prefix).Path("/disk/set-mountpoint").Handler(v1.SetMountPoint))

	// samba users
	sambaUserServer := v1.SambaUserServer{}
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-user/create").Handler(sambaUserServer.CreateUser))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-user/list").Handler(sambaUserServer.ListUsers))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-user/update").Handler(sambaUserServer.UpdateUser))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-user/delete").Handler(sambaUserServer.DeleteUser))

	// samba shares
	sambaShareServer := v1.SambaShareServer{}
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-share/create").Handler(sambaShareServer.CreateShare))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-share/list").Handler(sambaShareServer.ListShares))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-share/update").Handler(sambaShareServer.UpdateShare))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-share/delete").Handler(sambaShareServer.DeleteShare))
	as.Register(as.NewRoute().Prefix(prefix).Path("/samba-share/status").Handler(sambaShareServer.SambaStatus))

	// nfs shares
	nfsShareServer := v1.NFSShareServer{}
	as.Register(as.NewRoute().Prefix(prefix).Path("/nfs-share/create").Handler(nfsShareServer.CreateNFSExport))
	as.Register(as.NewRoute().Prefix(prefix).Path("/nfs-share/list").Handler(nfsShareServer.ListNFSExports))
	as.Register(as.NewRoute().Prefix(prefix).Path("/nfs-share/update").Handler(nfsShareServer.UpdateNFSExport))
	as.Register(as.NewRoute().Prefix(prefix).Path("/nfs-share/delete").Handler(nfsShareServer.DeleteNFSExport))
	as.Register(as.NewRoute().Prefix(prefix).Path("/nfs-share/status").Handler(nfsShareServer.NFSStatus))

	// hosts
	as.Register(as.NewRoute().Prefix(prefix).Path("/host/list").Handler(v1.ListHosts))
}

func HelloFluteNAS(w *apiserver.Response, r *apiserver.Request) {
	// call this api:
	// curl -X POST http://10.0.1.106:8088/v1/hello -d '{"f1": "Hello"}' -H "Content-Type: application/json"
	// param, ok := r.Fields.(*model.HelloRequest)
	// if !ok {
	// 	w.WriteError(errors.New("type convert failed"), nil)
	// 	return
	// }
	param := &model.HelloRequest{}
	err := r.Unmarshal(param)
	if err != nil {
		w.WriteError(err, nil)
	}
	w.Write([]byte(fmt.Sprintf("Welcome to fluteNAS, %s", param.F1)))
}
