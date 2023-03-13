package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

var (
	testAPIToken = "foo-bar-dar"
	headers      = map[string]string{
		"Authorization": testAPIToken,
	}
)

func assertMatchJSON(t *testing.T, expected map[string]any, actual string) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedJSON), actual)
}

func invokeHttp(method string, path string, headers map[string]string, data map[string]any) (statusCode int, body string) {
	r := setupRouter("master", testAPIToken)
	w := httptest.NewRecorder()

	bodyBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	for key := range headers {
		req.Header.Add(key, headers[key])
	}

	r.ServeHTTP(w, req)

	return w.Code, w.Body.String()
}

func TestWithoutAPIToken(t *testing.T) {
	code, body := invokeHttp("GET", "/api/config", nil, nil)
	assert.Equal(t, 403, code)
	assertMatchJSON(t, gin.H{"error": "Access denied"}, body)
}

func TestAPIStatus(t *testing.T) {
	code, body := invokeHttp("GET", "/status", nil, nil)

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "GoBackup is running.", "version": "master"}, body)
}

func TestAPIGetModels(t *testing.T) {
	models := map[string]any{
		"foo": map[string]any{
			"archive": map[string]any{
				"excludes": []string{"/home/ubuntu/.ssh/known_hosts", "/etc/logrotate.d/syslog"},
			},
			"databases": map[string]any{
				"dummy_test": map[string]any{
					"type":     "mysql",
					"host":     "localhost",
					"port":     3306,
					"database": "dummy_test",
				},
			},
		},
	}

	viper.Set("models", models)

	code, body := invokeHttp("GET", "/api/config", headers, nil)

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"models": models}, body)
}
