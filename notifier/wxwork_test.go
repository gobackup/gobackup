package notifier

import (
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
	"testing"
)

func Test_WxWork(t *testing.T) {
	base := &Base{}

	s := NewWxWork(base)
	assert.Equal(t, "WxWork", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"msgtype":"text","text":{"content":"This is title\n\nThis is body"}}`, string(body))

	respBody := `{"errcode":0,"errmsg":"ok"}`
	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	respBody = `{"errcode":93000,"errmsg":"invalid webhook url, hint: [1726032284347804201022717], from ip: xx.xx.xx.xx, more info at https://open.work.weixin.qq.com/devtool/query?e=93000"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)

	err = s.checkResult(200, []byte(respBody))
	assert.EqualError(t, err, "status: 200, body: "+respBody)
}

func Test_WxWorkNotify(t *testing.T) {
	v := viper.New()
	v.Set("url", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxx-xxxx-xxxx-xxxx")

	base := &Base{
		viper: v,
	}

	s := NewWxWork(base)

	err := s.notify("hello", "world")

	assert.NoError(t, err)
}
