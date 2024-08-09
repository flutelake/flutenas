package apiserver

import (
	"context"
	"time"
)

type HttpContext struct {
	UserInfo        string
	AnonymousAccess bool
	UUID            string
}

func (HttpContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (HttpContext) Done() <-chan struct{} {
	return nil
}

func (HttpContext) Err() error {
	return nil
}

func (HttpContext) Value(key any) any {
	return nil
}

type ctxMarker struct{}

var (
	ctxMarkerKey = &ctxMarker{}
)

func Extract(ctx context.Context) string {
	u, ok := ctx.Value(ctxMarkerKey).(string)
	if !ok {
		// return anonymous user
		return ""
		// 	return User{
		// 		anonymous: true,
		// 	}
	}
	return u
}

func SetInContext(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, user)
}
