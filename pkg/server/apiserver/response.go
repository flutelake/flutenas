package apiserver

import (
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/util"
	"net/http"
	"time"
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
	http.SetCookie(r.ResponseWriter, &http.Cookie{
		Name:  "sid",
		Value: r.cookie.SessionID,
		Path:  "/",
		// httpOnly 阻止在浏览器控制台中通过document.cookie获取cookie
		HttpOnly: true,
		// Secure:   true,
		Expires: time.Now().Add(time.Hour * 9),
	})
}

func (r *Response) NullCookie() {
	r.cookie = nil
	http.SetCookie(r.ResponseWriter, &http.Cookie{
		Name:  "sid",
		Value: "",
		Path:  "/",
		// httpOnly 阻止在浏览器控制台中通过document.cookie获取cookie
		HttpOnly: true,
		// Secure:   true,
		Expires: time.Now().Add(time.Minute * 1),
	})
}
