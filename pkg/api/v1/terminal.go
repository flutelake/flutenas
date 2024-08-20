package v1

import (
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"flutelake/fluteNAS/pkg/server/terminal"
	"flutelake/fluteNAS/pkg/util"
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

	ips := util.SourceIPs(r.Request)
	srcIP := ""
	if len(ips) > 0 {
		srcIP = ips[0].String()
	}

	token, err := a.terms.CreateTerminal(terminal.CreateTerminalParam{
		Hostname:           in.Hostname,
		BrowserFinderPrint: "",
		SourceIP:           srcIP,
		User:               "root",
		TerminalName:       in.TerminalName,
		Host: terminal.Host{
			Hostname: in.Hostname,
			Host:     "127.0.0.1",
			Port:     "22",
			Username: "root",
			Password: "",
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
