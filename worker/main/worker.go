package main
import (
	"runtime"
	"fmt"
	"flag"
	"time"
	"github.com/luckylgit/dscrond/worker"
)

var (
	confFielPath string
)

func initEnv(){
	runtime.GOMAXPROCS(runtime.NumCPU()) //根据系统环境初始化线程
}

func initArgs(){
	//worker配置文件初始话
	flag.StringVar(&confFielPath,"config","./config/worker.json","请指定配置文件的路径")
	flag.Parse()
}
func main(){
	var (
		err error
	)
	//读取配置
	initArgs()
	if err = worker.InitConfig(confFielPath);err != nil {
		goto ERR
	}
	//开始初始化线程
	initEnv()

	//从etcd获取任务


	for {
		time.Sleep(1 *time.Second)
	}
	return //正常退出
ERR:
	fmt.Println("启动异常:",err) //异常退出打印错误信息
}