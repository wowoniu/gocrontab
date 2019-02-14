package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
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
