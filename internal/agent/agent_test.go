package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SendInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		pathsParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathsParts) != 4 || pathsParts[0] != "update" {
			http.NotFound(w, r)
			return
		}
	}))
	defer ts.Close()

	agent := NewAgent(ts.URL, 10, 2)

	info := map[string]float64{
		"info": 0.0,
	}

	err := agent.sendInfo(info, 1)
	require.NoError(t, err)
}
