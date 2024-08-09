package apiserver

import (
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/util"
	"net/http"
)

type Response struct {
	ResponseWriter http.ResponseWriter
	// todo retcode
	fields any

	cookie *Session
}

func (r *Response) Write(data any) {
	r.fields = data
}

func (r *Response) WriteError(err error, data any) {
	flog.Errorf("request error: %v", err)
	r.fields = data
}

func (r *Response) SetCookie(userInfo any) {
	r.cookie = &Session{
		SessionID: util.RandStringRunes(32),
		UserInfo:  userInfo,
	}
}
