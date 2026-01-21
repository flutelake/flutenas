package v1

import (
	"archive/zip"
	"errors"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/cache"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/node"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileServer struct {
	cache    cache.TinyCache
	rootPath string
}

func NewFileServer(cache cache.TinyCache, rootPath string) *FileServer {
	return &FileServer{
		cache:    cache,
		rootPath: rootPath,
	}
}

func ListDir(w *apiserver.Response, r *apiserver.Request) {
	in := &model.ListDirRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	p := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+in.Path))
	entities, err := os.ReadDir(p)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	dirs := make([]string, 0)
	for _, item := range entities {
		if item.IsDir() {
			dirs = append(dirs, item.Name())
		}
	}

	out := model.ListDirResponse{
		Dirs: dirs,
	}
	w.Write(retcode.StatusOK(out))
}

func ReadDir(w *apiserver.Response, r *apiserver.Request) {
	in := &model.ReadDirRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	p := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+in.Path))
	entities, err := os.ReadDir(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteError(err, retcode.StatusDirNotExist(nil))
			return
		}
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	enDirs := make([]model.FileEntry, 0)
	ens := make([]model.FileEntry, 0)
	for _, item := range entities {
		en := model.FileEntry{
			Name:  item.Name(),
			IsDir: item.IsDir(),
		}
		if item.IsDir() {
			enDirs = append(enDirs, en)
			continue
		}
		fileInfo, err := item.Info()
		if err != nil {
			continue
		}
		en.LastMod = fileInfo.ModTime().UnixMilli()
		if !item.IsDir() {
			en.Size = int64(fileInfo.Size())
			en.Kind = mime.TypeByExtension(filepath.Ext(item.Name()))
		}
		ens = append(ens, en)
	}

	out := model.ReadDirResponse{
		Entries: append(enDirs, ens...),
	}
	w.Write(retcode.StatusOK(out))
}

func CreateDir(w *apiserver.Response, r *apiserver.Request) {
	in := &model.CreateDirRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	p := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+in.Path))
	err := os.Mkdir(p, 0o644)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	uid, gid := node.GetFluteUIDGID()
	err = os.Chown(p, uid, gid)
	if err != nil {
		defer os.Remove(p)
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	out := model.CreateDirResponse{}
	w.Write(retcode.StatusOK(out))
}

func RemoveFile(w *apiserver.Response, r *apiserver.Request) {
	in := &model.RemoveFileRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	ps := []string{}
	if in.Path != "" {
		p := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+in.Path))
		ps = append(ps, p)
	}
	if in.Paths != nil {
		for _, item := range in.Paths {
			p := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+item))
			ps = append(ps, p)
		}
	}

	for _, p := range ps {
		_, err := os.Stat(p)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				w.Write(retcode.StatusOK)
				return
			}
			w.WriteError(err, retcode.StatusError(nil))
			return
		}

		err = os.RemoveAll(p)
		if err != nil {
			w.WriteError(err, retcode.StatusError(nil))
			return
		}
	}
	out := model.RemoveFileResponse{}
	w.Write(retcode.StatusOK(out))
}

func UploadFiles(w *apiserver.Response, r *apiserver.Request) {
	dir := r.Request.URL.Query().Get("FilePath")
	if strings.HasSuffix(dir, string(filepath.Separator)) {
		dir = strings.TrimRight(dir, string(filepath.Separator))
	}
	path := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+dir))
	if _, err := os.Stat(path); err != nil {
		// donot care dir existed or not exist
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	// newName := q.Get("NewName")

	reader, err := r.Request.MultipartReader()
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	filenames := make([]string, 0, 4)
	var file *os.File
	for {
		part, err := reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			w.WriteError(err, retcode.StatusError(nil))
			return
		}

		if part.FileName() == "" {
			continue
		} else {
			name := part.FileName()
			if name == "" {
				w.WriteError(fmt.Errorf("multipart name is empty"), retcode.StatusError(nil))
				return
			}
			p := filepath.Join(path, name)
			file, err = os.Create(p)
			if err != nil {
				w.WriteError(err, retcode.StatusError(nil))
				return
			}
			defer file.Close()
			io.Copy(file, part)
			filenames = append(filenames, part.FileName())
			node.Belong2Flute(p)
		}
	}

	out := model.UploadFilesResponse{
		Names: filenames,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *FileServer) DownloadFiles(w *apiserver.Response, r *apiserver.Request) {
	in := &model.DownloadFilesRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	if in.Path == string(filepath.Separator) {
		w.WriteError(fmt.Errorf("file path is empty"), retcode.StatusDirEmpty(nil))
		return
	}
	path := filepath.Join("/mnt", filepath.Clean(string(filepath.Separator)+in.Path))
	token := util.RandStringRunes(16)
	_, err := os.Stat(path)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	s.cache.SetExpired(fmt.Sprintf("fsdownload:%s", token), path, time.Second*60)
	out := model.DownloadFilesResponse{
		Token:    token,
		Location: "/files/download?Token=" + token,
	}
	w.Write(retcode.StatusOK(out))
}

func (s *FileServer) ServerHttp(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("Token")
	if t == "" {
		http.Error(w, "Token is empty", http.StatusUnauthorized)
	}

	path, ok := s.cache.BurnAfterGet(fmt.Sprintf("fsdownload:%s", t))
	if !ok || path == "" {
		http.Error(w, "Token is invalid", http.StatusUnauthorized)
	}
	p, ok := path.(string)
	if !ok {
		http.Error(w, "File is invalid", http.StatusUnauthorized)
	}
	// 如果不在完全范围内的文件需要加日志记录
	if !strings.HasPrefix(p, s.rootPath) {
		flog.Warnf("file: %s been downloaded", path)
	}

	entry, err := os.Stat(p)
	if err != nil {
		http.Error(w, "File is invalid", http.StatusUnauthorized)
	}
	if entry.IsDir() {
		// 边打包压缩边传输
		serverDir(w, p, entry)
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", entry.Name()))
		http.ServeFile(w, r, p)
	}
}

func serverDir(w http.ResponseWriter, p string, entry fs.FileInfo) {
	// 创建一个压缩文件
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", entry.Name()))

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// 遍历文件夹中的所有文件和子文件夹
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == p {
			return nil
		}

		// 创建压缩文件中的文件
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置压缩文件中的文件路径
		header.Name, err = filepath.Rel(p, path)
		if err != nil {
			return err
		}

		// 如果是文件夹，则添加一个目录条目
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，则将文件内容写入压缩文件
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
