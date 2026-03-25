package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SendInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/update" {
			http.NotFound(w, r)
			return
		}

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	agent := NewAgent(ts.URL, 10, 2)

	info := map[string]float64{
		"info": 0.0,
	}

	err := agent.sendInfo(info, 1)
	require.NoError(t, err)
}
