package worker

import (
	"github.com/luckylgit/dscrond/common"
	"time"
)

//调度任务
type Scheduler struct {
	JobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulerPlan //任务名称以及任务下次执行的相关信息记录
}

var (
	G_scheduler *Scheduler
)

//遍历schedulerTable 的map 统计最近要过期的任务时间，如果已经过期或者即将过期则重新设置table表中的任务执行和时间
//并统计最近要过期的一个时间，让监听睡眠指定的时间再次遍历计划表
//如果表为空则设置睡眠1秒
func (sch *Scheduler) TryScheduler()(schedulerAfter time.Duration) {
	var (
		jobSchedulerPlan *common.JobSchedulerPlan
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
      for jobSchedulerPlan = range sch.jobPlanTable {
      	  if jobSchedulerPlan.NextTime.Equal(now) ||jobSchedulerPlan.NextTime.Before(now){
      	  	//TODO:尝试执行任务
      	  	jobSchedulerPlan.NextTime = jobSchedulerPlan.CronExpr.Next(now)
		  }

		  //统计最近要过期的任务时间
		  if nearTime == nil || jobSchedulerPlan.NextTime.Before(*nearTime){
		  	nearTime = &jobSchedulerPlan.NextTime
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

}
//初始化调度器
func InitScheduler(err error){
	G_scheduler = &Scheduler{
		JobEventChan:make(chan *common.JobEvent,1000),
		jobPlanTable:make(map[string]*common.JobSchedulerPlan),
	}
	go G_scheduler.SchedulerLoop()
	return
}