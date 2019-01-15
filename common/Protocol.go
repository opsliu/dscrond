package common

import (
	"encoding/json"
	"strings"
	"github.com/gorhill/cronexpr"
	"time"
)

//任务
type Job struct {
	Name string `json:"name"` //定时任务的名字
	Command string `json:"command"` //定时任务的命令
	CronExpr string `json:"cronExpr"` //定时任务的crontab表达式
}

//http接口返回应答
type Response struct {
	Errno int `json:"error"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

type JobEvent struct {
	EventType int //save delete
	Job       *Job
}

type JobSchedulerPlan struct {
	Job *Job                      //job信息
	CronExpr *cronexpr.Expression //crontab表达式
	NextTime time.Time           //下次的执行时间
}

//任务执行状态
type JobExcuteInfo struct {
   Job *Job
   PlanTime time.Time //计划开始时间
   RealTime time.Time //实际开始时间
}

//任务执行结果
type JobExcuteResult struct {
	ExcuteInfo *JobExcuteInfo  //任务本身相关信息
	Output []byte  //正常结果的输出
	Err error //脚本错误的输出
	StartTime time.Time
	EndTime time.Time
}


func BuildResp(errno int,msg string,data interface{})(resp []byte,err error ){
    var (
    	response Response
	)
    response.Errno = errno
    response.Msg = msg
    response.Data = data

    //序列化
    resp ,err = json.Marshal(response)
    return
}

//反序列化job
func UnpackJob(v []byte)(ret *Job,err error){
     var job *Job
     job = &Job{}
	if err = json.Unmarshal(v,&job);err != nil {
		return
	}
	ret = job
	return

}

//从etcd的/cron/jobs/jobname获取jobname
func ExtractJobName(jobKey string) (string){
	return strings.TrimPrefix(jobKey,JOB_SAVE_DIR)
}

//定义event任务变化事件update delete
func BuildJobEvent(eventType int,job *Job)(jobEvent *JobEvent){
	return &JobEvent{
		EventType:eventType,
		Job:job,
	}
}

func BuildJobSchedulerPlan(job *Job) (jobSchedulerPlan *JobSchedulerPlan,err error){
	var (
		expr *cronexpr.Expression
	)

	//解析表达式
	if expr,err = cronexpr.Parse(job.CronExpr);err != nil {
		return
	}


	//构造下次执行计划
	jobSchedulerPlan = &JobSchedulerPlan{
		Job:job,
		CronExpr:expr,
		NextTime:expr.Next(time.Now()),
	}
    return
}

func BuildJobExcuteInfo(jobSchedulerPlan *JobSchedulerPlan) (jobExcuteInfo *JobExcuteInfo) {
	return &JobExcuteInfo{
		Job: jobSchedulerPlan.Job,
		PlanTime:jobSchedulerPlan.NextTime,
		RealTime: time.Now(),//真实执行的时间
	}
}


func BuildJobExecuteResult(jobExcuteInfo *JobExcuteInfo,cmdOutput []byte,err error)(jobExcuteResult *JobExcuteResult){
	return &JobExcuteResult{
		ExcuteInfo:jobExcuteInfo,
		Output:cmdOutput,
		Err:err,
	}
}