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
    	postJob string
    	job common.Job
		oldJob *common.Job
		bytes []byte

	)

    if err = req.ParseForm();err != nil{
    	goto ERR
	}
	postJob = req.PostForm.Get("job")
    if err = json.Unmarshal([]byte(postJob),&job);err != nil {
    	goto ERR
	}

    //保存到job
    if  oldJob,err = G_jobMgr.SaveJob(&job);err != nil {
		goto ERR
	}
    //正常的应答返回({"errno:"0","msg";"","data":"json"})
	if bytes,err = common.BuildResp(0,"sucess",oldJob);err == nil {
		rsp.Write(bytes)
	}

    return
	ERR:

	//异常
	if bytes,err = common.BuildResp(-1,err.Error(),nil);err == nil {
			rsp.Write(bytes)
	}
}


//删除任务(post:{name:job}
func handleJobDelete(rsp http.ResponseWriter,req *http.Request){
    var (
    	err error
    	oldJob *common.Job
    	name string
		bytes []byte
	)
    if err = req.ParseForm();err != nil {
    	goto ERR
	}

	name = req.PostForm.Get("name")
    if oldJob,err = G_jobMgr.DeleteJob(name);err != nil {
    	goto ERR
	}
	//正常的删除应答返回({"errno:"0","msg";"","data":"json"})
	if bytes,err = common.BuildResp(0,"sucess",oldJob);err == nil {
		rsp.Write(bytes)
	}

	return
	ERR:

	//异常
	if bytes,err = common.BuildResp(-1,err.Error(),nil);err == nil {
		rsp.Write(bytes)
	}
}

//列出所有任务
func handleJobList(rsp http.ResponseWriter,req *http.Request){
   var (
   	err error
   	jobLists []*common.Job
   	bytes []byte
   )

   if jobLists,err = G_jobMgr.ListJobs();err != nil {
   	  goto ERR
   }
	//正常的查询任务列表应答返回({"errno:"0","msg";"","data":"json"})
	if bytes,err = common.BuildResp(0,"sucess",jobLists);err == nil {
		rsp.Write(bytes)
	}
   return

   ERR:
   //异常
	if bytes,err = common.BuildResp(-1,err.Error(),nil);err == nil {
		   rsp.Write(bytes)
	 }

}

//添加杀死任务
func handleJobKill(rsp http.ResponseWriter,req *http.Request){
	var (
		jobName string
		err error
		bytes []byte
	)
	if err = req.ParseForm();err != nil {
		goto ERR
	}
	jobName = req.PostForm.Get("name")
	if err = G_jobMgr.KillJob(jobName);err != nil {
		goto ERR
	}
	//正常kill
	if bytes,err = common.BuildResp(0,"sucess","");err == nil {
		rsp.Write(bytes)
	}
	return
	ERR:

	//添加杀死任务失败
	if bytes,err = common.BuildResp(-1,err.Error(),"");err == nil {
		rsp.Write(bytes)
	}
	return
}
//初始化api服务
func InitApiServer()(err error){
	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		address string
		staticDir http.Dir
		staticHandler http.Handler
	)
	//读取配置文件初始化接口
	//if conf,err = InitConfig();err != nil {
	//	//fmt.Println("配置文件读取失败:",err)
	//	return
	//}
	//定义接口路由
	mux = http.NewServeMux()
	mux.HandleFunc("/jobs/save",handleJobSave)
	mux.HandleFunc("/jobs/delete",handleJobDelete)
	mux.HandleFunc("/jobs/list",handleJobList)
	mux.HandleFunc("/jobs/kill",handleJobKill)

    //静态页面
    staticDir = http.Dir(G_config.StaticDir)
	staticHandler =http.FileServer(staticDir)
    mux.Handle("/",http.StripPrefix("/",staticHandler))


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