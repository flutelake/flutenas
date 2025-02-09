package model

type CreateTerminalRequest struct {
	HostIP       string `doc:"节点序列号" validate:"required"`
	FinderPrint  string `doc:"浏览器指纹" validate:"required"`
	TerminalName string `doc:"节点上面的终端会话名称，创建临时会话时传空"`
}

type CreateTerminalResponse struct {
	Token string `doc:"控制台会话token" validate:"required"`
}
