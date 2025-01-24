package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
)

func ListHosts(w *apiserver.Response, r *apiserver.Request) {
	var hosts []model.Host
	err := db.Instance().Find(&hosts).Error
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := &model.ListHostsResponse{
		Hosts: hosts,
	}
	w.Write(retcode.StatusOK(out))
}
