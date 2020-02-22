package master

import (
	"encoding/json"
	"io/ioutil"
)

// Config struct
type Config struct {
	APIPort         int      `json:"apiPort`
	APIReadTimeout  int      `json:"apiReadTimeout`
	APIWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
}

var (
	// 单例
	G_config *Config
)

// InitConfig func
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	G_config = &conf

	// fmt.Println(conf)

	return
}
