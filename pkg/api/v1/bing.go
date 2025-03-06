package v1

// bing 每日一图

// handlers/wallpaper.go
import (
	"encoding/json"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"io"
	"net/http"
	"time"
)

type WallpaperAPI struct {
	cache cache.TinyCache
}

func NewWallpapaerAPI(c cache.TinyCache) *WallpaperAPI {
	return &WallpaperAPI{
		cache: c,
	}
}

type BingResponse struct {
	Images []struct {
		URL       string `json:"url"`
		Copyright string `json:"copyright"`
		Title     string `json:"title"`
	} `json:"images"`
}

type WallpaperResponse struct {
	URL       string `json:"url"`
	Copyright string `json:"copyright"`
	Title     string `json:"title"`
}

const BingWallpaperKey = "BingWallpaper"

func (s *WallpaperAPI) GetWallpaper(w *apiserver.Response, r *apiserver.Request) {
	// Get From Cache Firsh
	wp, ok := s.cache.Get(BingWallpaperKey)
	if ok {
		// if wallpaper, ok := resp.(WallpaperResponse); ok {
		w.Write(retcode.StatusOK(wp))
		// }
	}
	// 调用必应API
	resp, err := http.Get("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1")
	if err != nil {
		// 返回错误响应
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 返回错误响应
		return
	}

	var bingResp BingResponse
	if err := json.Unmarshal(body, &bingResp); err != nil {
		// 返回错误响应
		return
	}

	if len(bingResp.Images) == 0 {
		// 返回错误响应
		return
	}

	// 构造返回数据
	wallpaper := WallpaperResponse{
		URL:       "https://www.bing.com" + bingResp.Images[0].URL,
		Copyright: bingResp.Images[0].Copyright,
		Title:     bingResp.Images[0].Title,
	}

	s.cache.SetExpired(BingWallpaperKey, wallpaper, time.Hour*20)

	// 返回成功响应
	w.Write(retcode.StatusOK(wallpaper))
}
