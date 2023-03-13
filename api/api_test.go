package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/longbridgeapp/assert"
)

func assertResponseJSON(t *testing.T, w *httptest.ResponseRecorder, status int, data map[string]any) {
	t.Helper()

	expectedJSON, err := json.Marshal(data)
	assert.NoError(t, err)

	assert.Equal(t, w.Code, status)

	assert.Equal(t, w.Body.String(), string(expectedJSON))
}

func TestAPIStatus(t *testing.T) {
	r := setupRouter("master", "foo-bar-dar")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/status", nil)
	r.ServeHTTP(w, req)

	assertResponseJSON(t, w, 403, gin.H{"error": "Access denied"})

	w = httptest.NewRecorder()
	req.Header.Add("Authorization", "foo-bar-dar")
	r.ServeHTTP(w, req)

	assertResponseJSON(t, w, 200, gin.H{"models": 0})
}
