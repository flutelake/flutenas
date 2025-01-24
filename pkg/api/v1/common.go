package v1

import (
	"errors"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"

	"gorm.io/gorm"
)

func GetHostInfo(w *apiserver.Response, hostIP string) (*model.Host, error) {
	// 这段逻辑看看能不能移到公共的地方
	var host model.Host
	err := db.Instance().First(&host, "host_ip = ?", hostIP).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteError(err, retcode.StatusParamInvalid("HostID"))
		} else {
			flog.Errorf("Error query localhost: %v", err)
			w.WriteError(err, retcode.StatusError(nil))
		}
		return nil, err
	}
	return &host, nil
}
