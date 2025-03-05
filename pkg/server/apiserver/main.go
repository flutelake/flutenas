package apiserver

import (
	"context"
	"embed"
	"errors"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type Apiserver struct {
	DataPath string
	// Routes   map[string]*Route
	// Server   *http.Server
	serveMux   *http.ServeMux
	cache      cache.TinyCache
	address    string
	certKey    string
	certPerm   string
	frontendFS embed.FS
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

	env := os.Getenv("ENV")
	if env == "prod" {
		// 生产环境由后端提供前端文件路由服务
		index, err := s.frontendFS.ReadFile("build/index.html")
		if err != nil {
			panic(err)
		}
		s.serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			clientIP := util.GetClientIP(r)
			defer flog.Infof("WebServer [%s %s] [client: %s]", r.Method, r.RequestURI, clientIP.String())

			// 如果请求的是根路径，直接返回index.html
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html")
				w.Write(index)
				return
			}

			// r.RequestURI 自带'/'前缀
			absolute := "build" + r.RequestURI
			ext := filepath.Ext(absolute)
			bs, err := s.frontendFS.Open(absolute)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					w.WriteHeader(http.StatusNotFound)
					w.Header().Set("Content-Type", "text/html")
					w.Write(index)
					return
				}

				flog.Errorf("WebServer [%s %s], [error] %s", r.Method, r.RequestURI, err.Error())
			}
			defer bs.Close()
			contentType := mime.TypeByExtension(ext)
			w.Header().Set("Content-Type", contentType)
			io.Copy(w, bs)
		})
	} else {
		// 开发环境由前端Webserver提供服务
		// develop for frontend route
		targetURL, err := url.Parse("http://10.0.1.106:5173")
		if err != nil {
			flog.Fatalf("Invalid frontend proxy target URL, %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		s.serveMux.Handle("/", proxy)
		// develop for frontend route --------- end
	}

	flog.Infof("Starting server on %s", s.address)
	if s.certKey != "" {
		err = http.ListenAndServeTLS(s.address, s.certKey, s.certPerm, s.serveMux)
	} else {
		err = http.ListenAndServe(s.address, s.serveMux)
	}

	return err
}

func (s *Apiserver) SetFrontendFS(fs embed.FS) {
	s.frontendFS = fs
}
