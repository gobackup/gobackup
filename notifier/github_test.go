package notifier

import (
	"fmt"
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_parseGitHubIssue(t *testing.T) {
	issue, err := parseGitHubIssue("https://github.com/huacnlee/gobackup/pull/111")
	assert.NoError(t, err)
	assert.Equal(t, "huacnlee", issue.group)
	assert.Equal(t, "gobackup", issue.repo)
	assert.Equal(t, "111", issue.issueId)
	assert.Equal(t, "https://api.github.com/repos/huacnlee/gobackup/issues/111/comments", issue.commentURL())

	issue, err = parseGitHubIssue("https://github.com/huacnlee/gobackup/issues/111")
	assert.NoError(t, err)
	assert.Equal(t, "huacnlee", issue.group)
	assert.Equal(t, "gobackup", issue.repo)
	assert.Equal(t, "111", issue.issueId)
	assert.Equal(t, "https://api.github.com/repos/huacnlee/gobackup/issues/111/comments", issue.commentURL())

	issue, err = parseGitHubIssue("http://github.com/huacnlee/gobackup/issues/111/foo/bar?foo=1")
	assert.NoError(t, err)
	assert.Equal(t, "huacnlee", issue.group)
	assert.Equal(t, "gobackup", issue.repo)
	assert.Equal(t, "111", issue.issueId)

	_, err = parseGitHubIssue("https://github.com/huacnlee")
	assert.EqualError(t, err, "invalid GitHub issue URL: https://github.com/huacnlee")
}

func Test_NewGitHub(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	s := NewGitHub(base)
	s.viper.Set("token", "this-is-github-access-token")
	s.viper.Set("url", "https://github.com/huacnlee/gobackup/issues/111")

	assert.Equal(t, "GitHub Comment", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"body":"This is title\n\nThis is body"}`, string(body))

	url, err := s.buildWebhookURL("https://github.com/huacnlee")
	assert.EqualError(t, err, "invalid GitHub issue URL: https://github.com/huacnlee")
	assert.Equal(t, "https://github.com/huacnlee", url)

	url, err = s.buildWebhookURL("https://github.com/huacnlee/gobackup/issues/111")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.github.com/repos/huacnlee/gobackup/issues/111/comments", url)

	headers := s.buildHeaders()
	assert.Equal(t, "Bearer this-is-github-access-token", headers["Authorization"])

	respBody := `{"message":"Must have admin rights to Repository.","documentation_url":"https://docs.github.com/rest"}`
	err = s.checkResult(404, []byte(respBody))
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", 404, respBody))

	respBody = `{}`
	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	respBody = `{}`
	err = s.checkResult(201, []byte(respBody))
	assert.NoError(t, err)
}
