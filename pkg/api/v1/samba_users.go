package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/db"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
)

type SambaUserServer struct{}

func (s *SambaUserServer) CreateUser(w *apiserver.Response, r *apiserver.Request) {
	in := &model.SambaUser{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// todo 检查用户名是否已存在

	in.Status = model.SambaUserStatus_Init
	// Use actual DB instance
	if err := db.Instance().Create(&in).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := model.CreateSambaUserResponse{
		ID:       in.ID,
		Username: in.Username,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaUserServer) ListUsers(w *apiserver.Response, r *apiserver.Request) {
	var users []model.SambaUser

	// Use actual DB instance
	if err := db.Instance().Find(&users).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := model.ListSambaUsersResponse{
		Users: users,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaUserServer) UpdateUser(w *apiserver.Response, r *apiserver.Request) {
	in := &model.UpdateSambaUserRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	var user model.SambaUser
	// Use actual DB instance
	if err := db.Instance().First(&user, in.ID).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	user.Password = in.Password
	user.Status = model.SambaUserStatus_ChangingPWD

	// Use actual DB instance
	if err := db.Instance().Save(&user).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := model.UpdateSambaUserResponse{
		ID: in.ID,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *SambaUserServer) DeleteUser(w *apiserver.Response, r *apiserver.Request) {
	in := &model.DeleteSambaUserRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	var user model.SambaUser
	// Use actual DB instance
	if err := db.Instance().First(&user, in.ID).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	// Use actual DB instance
	user.Status = model.SambaUserStatus_Deleting
	if err := db.Instance().Save(&user).Error; err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	// if err := db.Instance().Delete(&user).Error; err != nil {
	// 	w.WriteError(err, retcode.StatusError(nil))
	// 	return
	// }

	out := model.DeleteSambaUserResponse{
		ID: in.ID,
	}
	w.Write(retcode.StatusOK(out))
}
