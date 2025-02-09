package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/server/terminal"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
)

type TerminalAPI struct {
	terms *terminal.WebTerminal
}

func NewTerminalAPI(terms *terminal.WebTerminal) *TerminalAPI {
	return &TerminalAPI{
		terms: terms,
	}
}

func (a *TerminalAPI) CreateTerminal(w *apiserver.Response, r *apiserver.Request) {
	in := &model.CreateTerminalRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError)
		return
	}
	// get username and password from session
	userinfo, ok := r.Session.UserInfo.(model.SessionUserInfo)
	if !ok {
		w.WriteError(fmt.Errorf("format session error"), retcode.StatusError)
		return
	}
	flog.Infof("username: %s, password: %s", userinfo.Username, userinfo.Password.String())
	// if host_ip not eq localhost, get host ip from db
	hostInfo, err := GetHostInfo(w, in.HostIP)
	if err != nil {
		return
	}

	ips := util.SourceIPs(r.Request)
	srcIP := ""
	if len(ips) > 0 {
		srcIP = ips[0].String()
	}

	token, err := a.terms.CreateTerminal(terminal.CreateTerminalParam{
		Hostname:           hostInfo.Hostname,
		BrowserFinderPrint: "",
		SourceIP:           srcIP,
		User:               userinfo.Username,
		TerminalName:       in.TerminalName,
		Host: terminal.Host{
			Hostname: hostInfo.Hostname,
			Host:     in.HostIP,
			Port:     hostInfo.SSHPort,
			Username: userinfo.Username,
			Password: userinfo.Password.String(),
		},
	})
	if err != nil {
		w.WriteError(err, nil)
	}
	out := model.CreateTerminalResponse{
		Token: token,
	}
	w.Write(retcode.StatusOK(out))
}
