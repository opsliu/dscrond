package worker

import (
	"go.etcd.io/etcd/clientv3"
	"context"
	"github.com/luckylgit/dscrond/common"
)

type JobLock struct {
	kv clientv3.KV
	lease clientv3.Lease
	jobName string //任务名称
	cancelFunc context.CancelFunc //取消机制，用于终止自动续租
	leaseId clientv3.LeaseID //租约id
	isLocked bool  //是否上锁成功
}

var (
	G_jobLock *JobLock
)


//初始化一把锁
func InitJobLock(jobName string,kv clientv3.KV,lease clientv3.Lease) (jobLock *JobLock){
	G_jobLock =  &JobLock{
		kv:kv,
		lease:lease,
		jobName:jobName,
	}
	return &JobLock{
		kv:kv,
		lease:lease,
		jobName:jobName,
	}
}

//乐观尝试上锁
func (jbl *JobLock) TryLock()(err error){
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelFunc context.CancelFunc
		cancelCtx context.Context
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		txn  clientv3.Txn
		lockKey string
		txnResp *clientv3.TxnResponse
	)
	//创建租约5秒
	if leaseGrantResp,err = jbl.lease.Grant(context.TODO(),5);err != nil {
		return
	}
	//自动续租
	cancelCtx,cancelFunc =context.WithCancel(context.TODO()) //取消自动续租
	if keepRespChan,err =  jbl.lease.KeepAlive(cancelCtx,leaseGrantResp.ID);err != nil {
		goto  FAIL
	}
	//处理续租应答协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
			case keepResp = <- keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
		END:
	}()
	//创建事务
	txn = jbl.kv.Txn(context.TODO())
	//锁路径
	lockKey = common.JOB_LOCK_DIR + jbl.jobName
	//事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey),"=",0)).
		Then(clientv3.OpPut(lockKey,"",clientv3.WithLease(leaseGrantResp.ID))).
		Else(clientv3.OpGet(lockKey)) //判断事务，如果等于0 则创建key，否则获取key
     if txnResp, err = txn.Commit();err != nil { //如果抢锁失败跳出
     	goto FAIL
	 }

	//成功返回，失败释放租约
	if !txnResp.Succeeded {
		err = common.ERROR_LOCK_ALREADY_EXIST
		goto FAIL //失败，释放租约
	}

	//记录租约id
	jbl.leaseId = leaseGrantResp.ID
	jbl.cancelFunc = cancelFunc
	jbl.isLocked = true
	return
FAIL:
	cancelFunc()//取消自动续租
	jbl.lease.Revoke(context.TODO(),leaseGrantResp.ID) //释放掉租约id
	return
}

func (jbl *JobLock) UnLock()(){
	if jbl.isLocked{
		jbl.cancelFunc()//取消协程
		jbl.lease.Revoke(context.TODO(),jbl.leaseId)
	}

}
