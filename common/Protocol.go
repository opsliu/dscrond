package common

import "encoding/json"

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