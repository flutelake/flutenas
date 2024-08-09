package apiserver

import "flutelake/fluteNAS/pkg/util"

type Session struct {
	SessionID string
	UserInfo  interface{}
}

func NewSession(userInfo interface{}) *Session {
	sessionID := util.RandStringRunes(64)
	return &Session{
		SessionID: sessionID,
		UserInfo:  userInfo,
	}
}
