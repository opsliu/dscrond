package worker

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"fmt"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client //etcd 客户端api
	kv clientv3.KV          //etcd kv操作api
	lease clientv3.Lease    //etcd 租约操作api
}

//单例模式
var (
	G_jobMgr *JobMgr
)

func InitJobMgr()(err error){
	//初始化配置建立连接
	var (
		etcdConf clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	//初始化配置
	etcdConf = clientv3.Config{
		Endpoints:G_config.EtcdHosts,
		DialTimeout:time.Duration(G_config.EtcdTimeout)*time.Millisecond,
        }
        //建立连接
	if client,err = clientv3.New(etcdConf);err != nil {
		fmt.Println("Etcd异常:",err)
		return
	}

	//得到kv和lease的api子集
     kv = clientv3.NewKV(client)
     lease = clientv3.NewLease(client)

     //填充单例
     G_jobMgr = &JobMgr{
     	client:client,
     	kv:kv,
     	lease:lease,
	 }
	 return
}