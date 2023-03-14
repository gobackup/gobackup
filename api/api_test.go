package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
)

var (
	testAPIToken = "foo-bar-dar"
	headers      = map[string]string{
		"Authorization": testAPIToken,
	}
)

func init() {
	config.Init("../gobackup_test.yml")
}

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

	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json")
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
	code, body := invokeHttp("GET", "/api/config", headers, nil)

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"models": []string{"base_test", "demo", "expand_env", "normal_files", "test_model"}}, body)
}

func TestAPIPostPeform(t *testing.T) {
	code, body := invokeHttp("POST", "/api/perform", headers, gin.H{"model": "test_model"})

	assert.Equal(t, 200, code)
	assertMatchJSON(t, gin.H{"message": "Backup: test_model performed in background."}, body)
}
