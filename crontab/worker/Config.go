package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	WorkerName      string   `json:"worker_name"`          //当前节点名称
	EtcdEndpoints   []string `json:"etcd_endpoints"`       //etcd 集群
	EtcdDialTimeout int      `json:"etcd_dial_timeout"`    //etcd连接超时时间 毫秒
	ExecuteShell    string   `json:"execute_shell"`        //解析命令的shell
	MongoHost       string   `json:"mongo_host"`           //mongodb配置
	HttpTimeout     int      `json:"http_connect_timeout"` //http连接超时时间
}

var G_config *Config

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
