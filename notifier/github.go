package notifier

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type githubCommentPayload struct {
	Body string `json:"body"`
}

type githubIssue struct {
	group   string
	repo    string
	issueId string
}

var (
	githubIssueRe = regexp.MustCompile(`^http[s]?:\/\/github\.com\/([^\/]+)\/([^\/]+)\/(issues|pull)\/([^\/]+)`)
)

func parseGitHubIssue(url string) (*githubIssue, error) {
	if !githubIssueRe.MatchString(url) {
		return nil, fmt.Errorf("invalid GitHub issue URL: %s", url)
	}

	matches := githubIssueRe.FindAllStringSubmatch(url, -1)

	return &githubIssue{
		group:   matches[0][1],
		repo:    matches[0][2],
		issueId: matches[0][4],
	}, nil
}

func (issue githubIssue) commentURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s/comments", issue.group, issue.repo, issue.issueId)
}

// type: github
// url: https://github.com/gobackup/gobackup/issues/111
// access_token: xxxxxxxxxx
func NewGitHub(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "GitHub Comment",
		method:      "POST",
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			payload := githubCommentPayload{
				Body: fmt.Sprintf("%s\n\n%s", title, message),
			}

			return json.Marshal(payload)
		},
		buildWebhookURL: func(url string) (string, error) {
			issue, err := parseGitHubIssue(url)
			if err != nil {
				return url, err
			}

			return issue.commentURL(), nil
		},
		buildHeaders: func() map[string]string {
			return map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", base.viper.GetString("token")),
			}
		},
		checkResult: func(status int, body []byte) error {
			if status == 200 || status == 201 {
				return nil
			}

			return fmt.Errorf(`status: %d, body: %s`, status, string(body))
		},
	}
}
