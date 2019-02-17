package worker

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"
)

type WebExecutor struct {
	Url            string //web触发地址
	ConnectTimeout int64  //HTTP连接超时时间
}

var G_webExecutor *WebExecutor

func init() {
	G_webExecutor = &WebExecutor{}
}

//执行命令
func (this *WebExecutor) Exec(ctx context.Context, url string) (output []byte, err error) {
	var (
		request  *http.Request
		client   *http.Client
		response *http.Response
	)
	if request, err = http.NewRequest("POST", url, nil); err != nil {
		return
	}
	request.WithContext(ctx)
	client = &http.Client{
		Timeout: time.Duration(G_config.HttpTimeout) * time.Microsecond,
	}
	if response, err = client.Do(request); err != nil {
		return
	}
	output, err = ioutil.ReadAll(response.Body)
	return
}
