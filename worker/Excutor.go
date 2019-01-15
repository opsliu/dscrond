package worker

import (
	"github.com/luckylgit/dscrond/common"
	"os/exec"
	"time"
	"math/rand"
)
//执行器
type Excutor struct {
}

var (
	G_excutor *Excutor
)

//执行任务
func (exc *Excutor) ExcuteJob(info *common.JobExcuteInfo) {
    go func() {
		var (
			cmd *exec.Cmd
			err error
			CmdOutput []byte
			result *common.JobExcuteResult
			jobLock *JobLock
		)
		result = &common.JobExcuteResult{
			ExcuteInfo:info,
			Output:make([]byte,0),
		}
		result.StartTime = time.Now()
		//初始化锁
		jobLock = G_jobMgr.CreateLock(info.Job.Name)

		time.Sleep(time.Duration(rand.Intn(1000))*time.Millisecond)
        if err = jobLock.TryLock();err != nil{
        	result.Err = err
        	result.EndTime = time.Now()
			result = common.BuildJobExecuteResult(info,CmdOutput,err)
		} else {
			result.StartTime = time.Now()
			//执行shell命令
			cmd = exec.CommandContext(info.CancelCtx,"C:\\cygwin64\\bin\\bash.exe","-c",info.Job.Command)
			CmdOutput,err = cmd.CombinedOutput()
			result = common.BuildJobExecuteResult(info,CmdOutput,err)
			result.EndTime = time.Now()
			G_scheduler.PushJobExcuteResult(result) //推送结果到chan
			//释放锁
			jobLock.UnLock()
		}
	}()
}


func InitExcutor()(err error) {
    G_excutor = &Excutor{}
	return
}