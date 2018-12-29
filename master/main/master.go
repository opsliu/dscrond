package main

import (
	"runtime"
	"github.com/luckylgit/dscrond/master"
	"fmt"
	"flag"
)

var (
	confFielPath string
)

func initEnv(){
	runtime.GOMAXPROCS(runtime.NumCPU()) //根据系统环境初始化线程
}

func initArgs(){
	flag.StringVar(&confFielPath,"config","./config/master.json","请指定配置文件的路径")
	flag.Parse()
}
func main(){
	var (
		err error
	)
	//读取配置
	initArgs()
	if err = master.InitConfig(confFielPath);err != nil {
		goto ERR
	}
	//开始初始化线程
	initEnv()

	//任务管理器启动
	if err = master.InitJobMgr();err != nil {
		goto ERR
	}
	//启动apiHttp服务
	if err = master.InitApiServer();err !=nil{
		goto ERR
	}
	return //正常退出
ERR:
	fmt.Println("启动异常:",err) //异常退出打印错误信息
}