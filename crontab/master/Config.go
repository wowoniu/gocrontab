package master

import (
	"encoding/json"
	"io/ioutil"
)

//HTTP SERVER 配置
type Config struct {
	ApiPort         int      `json:"api_port"`          //API端口
	ReadTimeout     int      `json:"read_timeout"`      //API读超时时间 毫秒
	WriteTimeout    int      `json:"write_timeout"`     //API写超时时间 毫秒
	EtcdEndpoints   []string `json:"etcd_endpoints"`    //etcd 集群
	EtcdDialTimeout int      `json:"etcd_dial_timeout"` //etcd连接超时时间 毫秒
}

//全局单例配置
var (
	G_config *Config
)

//加载配置
func LoadConfig(configFileName string) (err error) {
	var (
		content []byte
		config  Config
	)
	if content, err = ioutil.ReadFile(configFileName); err != nil {
		return
	}

	//JSON反序列化
	if err = json.Unmarshal(content, &config); err != nil {
		return
	}

	G_config = &config

	return
}
