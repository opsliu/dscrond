package master

import (
	"net/http"
	"net"
	"time"
	"fmt"
	"strconv"
	"github.com/luckylgit/dscrond/common"
	"encoding/json"
)


type ApiServer struct {
	httpServer *http.Server
}

var (
	//单例模式
	G_apiServer *ApiServer
)
//保存任务接口
func handleJobSave(rsp http.ResponseWriter,req *http.Request){
    //保存任务etcd
    var(
    	err error
    	jobName string
    	job common.Job

	)
    if err = req.ParseForm();err != nil{
    	goto ERR
	}

	jobName = req.PostForm.Get("job")
    if err = json.Unmarshal([]byte(jobName),&job);err != nil {
    	goto ERR
	}

    //保存到job


	ERR:
}


//初始化api服务
func InitApiServer()(err error){
	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		address string
	)
	//读取配置文件初始化接口
	//if conf,err = InitConfig();err != nil {
	//	//fmt.Println("配置文件读取失败:",err)
	//	return
	//}
	//定义接口路由
	mux = http.NewServeMux()
	mux.HandleFunc("/jobs/save",handleJobSave)


	//启动监听
	address = G_config.ApiHost+":"+strconv.Itoa(G_config.ApiPort)
	if listener,err = net.Listen("tcp",address);err != nil {
		return
	}
    //创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:time.Duration(G_config.ApiReadTimeout)*time.Millisecond,
		WriteTimeout:time.Duration(G_config.ApiWriteTimeout)*time.Millisecond,
		Handler:mux,
	}

	//赋值单例模式
    G_apiServer = &ApiServer{
    	httpServer:httpServer,
	}

	//启动服务端
	go httpServer.Serve(listener)
    fmt.Println("Master Server已经启动:",address)
	return

}