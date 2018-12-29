package common

type Job struct {
	Name string `json:"name"` //定时任务的名字
	Command string `json:"command"` //定时任务的命令
	CronExpr string `json:"cronExpr"` //定时任务的crontab表达式
}