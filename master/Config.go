package master

import (
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	ApiPort int 			`json:"apiPort"`
	ApiHost string 			`json:"apiHost"`
	ApiReadTimeout int 		`json:"apiReadTimeout"`
	ApiWriteTimeout int 	`json:"apiWriteTimeout"`
	EtcdHosts []string 		`json:"etcdHosts"`
    EtcdTimeout int         `json:"etcdTimeout"`
    StaticDir string        `json:"staticDir"`
}

var (
	G_config *Config
)


//json解析配置文件
type ConfJsonStruct struct {
}

func NewConfJsonStruct() *ConfJsonStruct {
	return &ConfJsonStruct{}
}

//加载配置文件
func (cjst *ConfJsonStruct) Load(filename string,v interface{})(err error){
	var (
		data []byte
	)
	if data,err = ioutil.ReadFile(filename);err != nil {
		//fmt.Println("配置文件读取失败:",err)
		return
	}

	if err = json.Unmarshal(data,v); err != nil {
		//fmt.Println("配置文件json序列反序列化失败:",err)
		return
	}
    return
}

//初始化读取配置文件入口
func InitConfig(path string)(err error){
	  var (
		conf Config
		confPrase  *ConfJsonStruct
	  )
	  confPrase = NewConfJsonStruct()
	  conf = Config{}

	  if err = confPrase.Load(path,&conf);err != nil {
	  	return
	  }
	  G_config = &conf
	  return
}
