package master

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"fmt"
	"github.com/luckylgit/dscrond/common"
	"context"
	"encoding/json"
	"go.etcd.io/etcd/mvcc/mvccpb"
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

//保存
func (jmg *JobMgr) SaveJob(job *common.Job)(oldjob *common.Job ,err error){
    //
    var (
    	jobKey string
    	jobValue []byte
    	putResp *clientv3.PutResponse
    	oldJobObj common.Job
	)
    jobKey = common.JOB_SAVE_DIR+job.Name
    if jobValue,err = json.Marshal(*job);err != nil {
    	return
	}
	//保存etcd
	if putResp,err = jmg.kv.Put(context.TODO(),jobKey,string(jobValue),clientv3.WithPrevKV());err != nil {
		return
	}
    //如果是更新返回旧值
    if putResp.PrevKv != nil {
    	if err = json.Unmarshal(putResp.PrevKv.Value,&oldJobObj);err != nil {
    		err = nil
			return
		}
		oldjob = &oldJobObj
	}
	return
}

//删除job
func (jmg *JobMgr) DeleteJob(name string)(oldJob *common.Job,err error){
	var (
		jobKey string
		delResp *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	jobKey = common.JOB_SAVE_DIR+name
	if delResp,err = jmg.kv.Delete(context.TODO(),jobKey,clientv3.WithPrevKV());err != nil {
		return
	}
	//删除之后返回旧值
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value,&oldJobObj);err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
//列出所有
func (jmg *JobMgr) ListJobs()(jobList []*common.Job,err error){
    var (
    	dirKey string
		getListResp *clientv3.GetResponse
		keyValuePair *mvccpb.KeyValue
		job *common.Job
	)
    dirKey = common.JOB_SAVE_DIR
    if getListResp,err = jmg.kv.Get(context.TODO(),dirKey,clientv3.WithPrefix());err != nil {
		return
	}

	fmt.Println(getListResp)
	jobList = make([]*common.Job,0)
	for _,keyValuePair =range  getListResp.Kvs{
		job = &common.Job{}
		if err = json.Unmarshal(keyValuePair.Value,job);err != nil {
			fmt.Println(err)
         	err = nil
         	continue
        }
		jobList = append(jobList,job)
	}
	return
}

//杀死任务
//向/cron/killer/添加杀死的任务
func (jmg *JobMgr) KillJob(name string)(err error){
    var (
    	//KillResp *clientv3.PutResponse
    	jobKey string
    	jobVal string
    	lease *clientv3.LeaseGrantResponse
	)
    jobKey = common.JOB_KILL_DIR + name
	jobVal = name
	if lease,err = jmg.lease.Grant(context.TODO(),1);err != nil {
		return
	}

    if _,err = jmg.kv.Put(context.TODO(),jobKey,jobVal,clientv3.WithLease(lease.ID));err != nil {
    	return
	}
	return
}