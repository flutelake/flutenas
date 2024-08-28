package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
)

func OpenNFSServer(w *apiserver.Response, r *apiserver.Request) {
	in := &model.OpenNFSServerRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	// TODO:

	out := &model.OpenNFSServerRequestResponse{}
	w.Write(retcode.StatusOK(out))
}

// CreateNFSExport 增加分享路径
func CreateNFSExport(w *apiserver.Response, r *apiserver.Request) {
}

// DeleteNFSExport 删除分享路径
func DeleteNFSExport(w *apiserver.Response, r *apiserver.Request) {
}

// UpdateNFSExport 更新分享路径
func UpdateNFSExport(w *apiserver.Response, r *apiserver.Request) {
}

// ListNFSExports list all nfs exports path
func ListNFSExports(w *apiserver.Response, r *apiserver.Request) {}
