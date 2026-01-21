package model

type ListDirRequest struct {
	Path string `doc:"路径" `
}

type ListDirResponse struct {
	Dirs []string `doc:"文件夹列表"`
}

type ReadDirRequest struct {
	Path string `doc:"路径"`
}

type ReadDirResponse struct {
	Entries []FileEntry `doc:"文件列表"`
}

type FileEntry struct {
	Name    string
	IsDir   bool
	Size    int64
	LastMod int64
	Kind    string // MIME Type
}

type CreateDirRequest struct {
	Path string `doc:"文件夹名称" validate:"required"`
}

type CreateDirResponse struct {
	Entities string `doc:"文件列表"`
}

type UploadFilesRequest struct {
}

type UploadFilesResponse struct {
	Names []string `doc:"上传成功的文件名"`
}

type RemoveFileRequest struct {
	Path  string   `doc:"文件路径" validate:"required"`
	Paths []string `doc:"文件路径多个" validate:"required"`
}

type RemoveFileResponse struct {
}

type DownloadFilesRequest struct {
	Path string `doc:"文件路径" validate:"required"`
}

type DownloadFilesResponse struct {
	Token    string
	Location string
}
