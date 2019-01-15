package worker

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"fmt"
	"context"
	"github.com/luckylgit/dscrond/common"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client //etcd 客户端api
	kv clientv3.KV          //etcd kv操作api
	lease clientv3.Lease    //etcd 租约操作api
	watcher clientv3.Watcher //etcd 监听watch
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
		watcher clientv3.Watcher
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
	 watcher = clientv3.NewWatcher(client)

     //填充单例
     G_jobMgr = &JobMgr{
     	client:client,
     	kv:kv,
     	lease:lease,
     	watcher:watcher,
	 }

	 //启动watchJobs
	 G_jobMgr.watchJobs()
	 return
}

//监听etcd任务变化
func (jmg *JobMgr) watchJobs()(err error){
	var (
		getResp *clientv3.GetResponse
		kvpair *mvccpb.KeyValue
		job *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
		jobEvent *common.JobEvent
	)
	//获取所有任务，获取当前集群的revsion
	if getResp,err = jmg.kv.Get(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithPrefix());err != nil {
		return
	}
	for _,kvpair =range getResp.Kvs{
		if job,err  = common.UnpackJob(kvpair.Value);err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVNET_SAVE,job)
			G_scheduler.PushSchedulerEvent(jobEvent) //推送event
		}
		continue
		 //if err = json.Unmarshal(kvpair.Value,&job);err == nil {
          //   //推送当前任务到scheduler
			// jobEvent = common.BuildJobEvent(common.JOB_EVNET_SAVE,job)
			// G_scheduler.PushSchedulerEvent(jobEvent) //推送event
		 //}
	}
	//从该revision向后监听变化
	go func() {
		//监听协程,从版本加1开始
		watchStartRevision = getResp.Header.Revision +1
        //启动监听,监听任务/cron/jobs/目录的后续变化
		watchChan = jmg.watcher.Watch(context.TODO(),
			common.JOB_SAVE_DIR,clientv3.WithRev(watchStartRevision),
				clientv3.WithPrefix())

		for watchResp = range watchChan {
			//返回的watchResp为多个返回事件
            for _,watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //保存任务
				     if job,err = common.UnpackJob(watchEvent.Kv.Value);err != nil {
				     	//反解任务失败忽略
				     	continue
					 }
					 //构造一个event，推送给调度协程
					jobEvent = common.BuildJobEvent(common.JOB_EVNET_SAVE,job)
					G_scheduler.PushSchedulerEvent(jobEvent)  //推送event
					//反序列化job,推送给调度协程update任务

				case mvccpb.DELETE: //删除任务
				    //让调度协程停止任务,删除事件delete任务
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))

					job = &common.Job{
						Name:jobName,
					}
					//构造一个删除事件event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE,job)
                    //推送到调度器scheduler
					G_scheduler.PushSchedulerEvent(jobEvent)  //推送event

				}
			}
		}
	}()
	return
}

//创建任务执行锁
func (jmg *JobMgr) CreateLock(jobName string)(jobLock *JobLock) {
    //返回锁
    jobLock = InitJobLock(jobName,jmg.kv,jmg.lease)
    return
}