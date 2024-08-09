package apiserver

import (
	"flutelake/fluteNAS/pkg/module/cache"
	"net/http"
)

// auth
func filterAuth(c cache.TinyCache, resp *Response, req *Request) int {
	cookie, err := req.Request.Cookie("sid")
	if err != nil {
		// Cookie 不存在或读取失败
		return http.StatusUnauthorized
	}
	if c == nil {
		return http.StatusInternalServerError
	}

	userIntf, ok := c.Get("Session:" + cookie.Value)
	if !ok {
		// Cookie 有效
		return http.StatusUnauthorized
	}

	sess, ok := userIntf.(*Session)
	if !ok {
		return http.StatusInternalServerError
	}
	req.Session = sess
	// save session info into ctx
	// req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), "session", sess))

	return http.StatusOK
}
