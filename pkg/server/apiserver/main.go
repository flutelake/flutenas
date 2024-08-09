package apiserver

import (
	"context"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/flog"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Apiserver struct {
	DataPath string
	// Routes   map[string]*Route
	// Server   *http.Server
	serveMux *http.ServeMux
	cache    cache.TinyCache
	address  string
	certKey  string
	certPerm string
}

func NewApiserver(c cache.TinyCache) *Apiserver {
	return &Apiserver{
		// Routes: make(map[string]*Route),
		// Server: &http.Server{
		// 	Addr: ":8088",
		// },
		serveMux: http.NewServeMux(),
		cache:    c,
		address:  ":8088",
		certKey:  "",
		certPerm: "",
	}
}

func (a *Apiserver) NewRoute() *Route {
	return &Route{}
}

func (a *Apiserver) Register(route *Route) {
	// if _, ok := a.Routes[route.path]; ok {
	// 	flog.Fatalf("Route %s already exists", route.path)
	// }

	flog.Infof("Register route: %s", route.GetPath())
	route.cache = a.cache
	a.serveMux.Handle(route.GetPath(), route)
}

func (s *Apiserver) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	flog.Infof("Register route: %s", pattern)
	s.serveMux.HandleFunc(pattern, handler)
}

func (s *Apiserver) Run(ctx context.Context) (err error) {
	stopChan := ctx.Done()
	go func() {
		<-stopChan
		fmt.Println("shutdown server...")
		// _ = s.Server.Shutdown(context.Background())
	}()

	// develop for frontend route
	targetURL, err := url.Parse("http://10.0.1.106:5173")
	if err != nil {
		flog.Fatalf("Invalid frontend proxy target URL, %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	s.serveMux.Handle("/", proxy)
	// develop for frontend route --------- end

	flog.Infof("Starting server on %s", s.address)
	if s.certKey != "" {
		err = http.ListenAndServeTLS(s.address, s.certKey, s.certPerm, s.serveMux)
	} else {
		err = http.ListenAndServe(s.address, s.serveMux)
	}

	return err
}
