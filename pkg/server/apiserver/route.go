package apiserver

import (
	"encoding/json"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/flog"
	"fmt"
	"net/http"
	"time"
)

type Route struct {
	prefix             string
	path               string
	function           func(*Response, *Request)
	allowAnonymous     bool
	permissionRequired string
	cache              cache.TinyCache
}

// allowAnonymous bool, permissionRequired string, preFilters ...*Filter
// func NewRoute(path string, function func(Response, *Request)) *Route {
// 	return &Route{
// 		path:               path,
// 		function:           function,
// 		allowAnonymous:     false,
// 		permissionRequired: "",
// 		preFilters:         []*Filter{},
// 	}
// }

func (h *Route) GetPath() string { return h.prefix + h.path }

func (h *Route) Path(path string) *Route { h.path = path; return h }

func (h *Route) Prefix(prefix string) *Route {
	h.prefix = prefix
	return h
}

func (h *Route) Handler(function func(*Response, *Request)) *Route { h.function = function; return h }

func (h *Route) AllowAnonymous(allow bool) *Route {
	h.allowAnonymous = allow
	return h
}

func (h *Route) Permission(permission string) *Route { h.permissionRequired = permission; return h }

func (h *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// todo local develop test
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	resp := &Response{ResponseWriter: w}
	req := &Request{Request: r}

	// pre filter
	status := filterAuth(h.cache, resp, req)
	switch status {
	case http.StatusUnauthorized:
		if h.allowAnonymous {
			goto handle
		}
		w.WriteHeader(status)
		return
	case http.StatusOK:
		goto handle
	default:
		w.WriteHeader(status)
		return
	}

handle:
	// handler
	h.function(resp, req)

	if resp.cookie != nil {
		// save into cache
		h.cache.SetExpired(fmt.Sprintf("Session:%s", resp.cookie.SessionID), resp.cookie, time.Hour*9)
	}

	if resp.fields == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		bs, err := json.Marshal(resp.fields)
		if err != nil {
			flog.Errorf("marshal response body data failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.ResponseWriter.Write(bs)
	}
}
