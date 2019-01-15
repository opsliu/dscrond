package worker

import (
	"github.com/luckylgit/dscrond/common"
	"time"
	"fmt"
)

//调度任务
type Scheduler struct {
	JobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulerPlan //任务名称以及任务下次执行的相关信息记录
	jobExcutingTable map[string]*common.JobExcuteInfo
	jobExcuteResultChan chan *common.JobExcuteResult
}

var (
	G_scheduler *Scheduler
)

//遍历schedulerTable 的map 统计最近要过期的任务时间，如果已经过期或者即将过期则重新设置table表中的任务执行和时间
//并统计最近要过期的一个时间，让监听睡眠指定的时间再次遍历计划表
//如果表为空则设置睡眠1秒
func (sch *Scheduler) TryScheduler()(schedulerAfter time.Duration) {
	var (
		jobSchrPlan *common.JobSchedulerPlan
		now time.Time
		nearTime *time.Time //因为默认是值传递，程序需要修改此值，所以需要传入指针类型
	)

	 //获取当前时间
	 now = time.Now()

	//如果计划表为空设置默认睡眠时间
	if len(sch.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second //睡眠1秒
		return
	}
      for _,jobSchrPlan = range sch.jobPlanTable {
      	  if jobSchrPlan.NextTime.Equal(now) ||jobSchrPlan.NextTime.Before(now){
      	  	//TODO:尝试执行任务
      	  	  sch.TryStartJob(jobSchrPlan)
			  jobSchrPlan.NextTime = jobSchrPlan.CronExpr.Next(now)
		  }

		  //统计最近要过期的任务时间
		  if nearTime == nil || jobSchrPlan.NextTime.Before(*nearTime){
		  	nearTime = &jobSchrPlan.NextTime
		  }

		  //设置下次间隔调度时间
		  schedulerAfter = (*nearTime).Sub(now)
	  }
	  return
}

//调度协程
func (sch *Scheduler) SchedulerLoop(){
	var (
		jobEvent *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobExcuteRes *common.JobExcuteResult
	)

	//启动主循环，初始化一次 计算下一次调度任务执行时间
	schedulerAfter = G_scheduler.TryScheduler()
	schedulerTimer = time.NewTimer(schedulerAfter)


    for {
		select {
    	case jobEvent = <-sch.JobEventChan: //监听任务变化事件

			//操作维护处理的任务增删改查
			  sch.handlerJobEvent(jobEvent)
			  schedulerAfter = sch.TryScheduler()

    	case <- schedulerTimer.C: //最近要执行时间到期，遍历计划table
    	case jobExcuteRes = <- sch.jobExcuteResultChan: //从结果chan中取出结果
    	     sch.handlerJobResult(jobExcuteRes)
		}
		schedulerAfter = sch.TryScheduler()
		schedulerTimer.Reset(schedulerAfter) //重置定时器
	}
}


//处理任务事件
func (sch *Scheduler) handlerJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulerPlan *common.JobSchedulerPlan
		err error
        jobExist bool
	)
	switch jobEvent.EventType {
	case common.JOB_EVNET_SAVE:
		if jobSchedulerPlan,err = common.BuildJobSchedulerPlan(jobEvent.Job);err != nil {
			return
		}
		sch.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulerPlan,jobExist = sch.jobPlanTable[jobEvent.Job.Name];jobExist{
			delete(sch.jobPlanTable,jobEvent.Job.Name) //如果存在就删除掉
		}
	}
}
//推送任务事件
func (sch *Scheduler) PushSchedulerEvent(jobEvent *common.JobEvent){
	sch.JobEventChan <- jobEvent
}

//处理回传的结果
func (sch *Scheduler) handlerJobResult(jobExcuteRes *common.JobExcuteResult){
    //删除执行状态
    delete(sch.jobExcutingTable,jobExcuteRes.ExcuteInfo.Job.Name) //从任务执行表中删除任务因为赢执行完成
    fmt.Println("任务执行完成:",string(jobExcuteRes.ExcuteInfo.Job.Name),string(jobExcuteRes.Output),jobExcuteRes.Err)
}
//推送执行结果到结果chan
func (sch *Scheduler) PushJobExcuteResult(jobRes *common.JobExcuteResult) {
   sch.jobExcuteResultChan <- jobRes
}
//初始化调度器
func InitScheduler()(err error){
	G_scheduler = &Scheduler{
		JobEventChan:make(chan *common.JobEvent,1000),
		jobPlanTable:make(map[string]*common.JobSchedulerPlan),
		jobExcutingTable:make(map[string]*common.JobExcuteInfo),
		jobExcuteResultChan:make(chan *common.JobExcuteResult,1000),
	}
	go G_scheduler.SchedulerLoop()
	return
}

//尝试执行任务
func (sch *Scheduler) TryStartJob(jobPlan *common.JobSchedulerPlan) {
	var (
		jobExcuting bool
		jobExcuteInfo  *common.JobExcuteInfo
	)

   //调度 +执行：调度是检查任务是否过期，执行表示开始处理任务，任务有可能执行很久，所以1分钟只能执行一次 ,防止并发

   //如果任务正在执行 跳过本次执行
   if jobExcuteInfo,jobExcuting = sch.jobExcutingTable[jobPlan.Job.Name];jobExcuting {
   	   fmt.Println("此任务上次调度正在执行,跳过本次执行:",jobExcuteInfo.Job.Name)
       return
   }

   //构建执行状态
	jobExcuteInfo = common.BuildJobExcuteInfo(jobPlan)

	//保存任务执行状态
	sch.jobExcutingTable[jobPlan.Job.Name] = jobExcuteInfo

	//执行任务
	//TODO: 执行任务
	fmt.Println("执行任务:",jobExcuteInfo.Job.Name,jobExcuteInfo.RealTime,jobExcuteInfo.PlanTime)
	G_excutor.ExcuteJob(jobExcuteInfo)

}