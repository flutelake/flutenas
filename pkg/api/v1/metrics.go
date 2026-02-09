package v1

import (
	"encoding/json"
	"flutelake/fluteNAS/pkg/module/retcode"
	"flutelake/fluteNAS/pkg/server/apiserver"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type VictoriaQueryRangeRequest struct {
	Query string `json:"Query"`
	Start int64  `json:"Start"`
	End   int64  `json:"End"`
	Step  int64  `json:"Step"`
}

func QueryVictoriaMetricsRange(w *apiserver.Response, r *apiserver.Request) {
	// baseURL := os.Getenv("VICTORIA_METRICS_QUERY_URL")
	// if baseURL == "" {
	// 	w.WriteError(fmt.Errorf("victoria metrics query url not configured"), retcode.StatusError(nil))
	// 	return
	// }
	baseURL := "http://127.0.0.1:8086"

	in := &VictoriaQueryRangeRequest{}
	if err := r.Unmarshal(in); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	if in.Query == "" {
		w.WriteError(fmt.Errorf("query is required"), retcode.StatusError(nil))
		return
	}

	if in.End == 0 {
		in.End = time.Now().Unix()
	}
	if in.Start == 0 {
		in.Start = in.End - 300
	}
	if in.Step <= 0 {
		in.Step = 10
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", in.Query)
	q.Set("start", fmt.Sprintf("%d", in.Start))
	q.Set("end", fmt.Sprintf("%d", in.End))
	q.Set("step", fmt.Sprintf("%d", in.Step))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(r.Request.Context(), http.MethodGet, u.String(), nil)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteError(fmt.Errorf("victoria metrics returned status %d", resp.StatusCode), retcode.StatusError(nil))
		return
	}

	var out any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		w.WriteError(err, retcode.StatusError(nil))
		return
	}

	w.Write(retcode.StatusOK(out))
}
