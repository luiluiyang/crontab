package master

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/chris/crontab/common"
)

// APIServer struct
type APIServer struct {
	httpServer *http.Server
}

var (
	// 单例对象
	G_apiServer *APIServer
)

//保存任务接口
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)
	// 1. 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 2.取表单中的job字段
	postJob = r.PostForm.Get("job")
	// 3.反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	// 4.保存到etcd
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}
	// 5.返回正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
	return
}

// 删除任务接口
// POST /job/delete name=job1
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		name   string
		oldJob *common.Job
		bytes  []byte
	)

	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 删除的任务名
	name = r.PostForm.Get("name")

	// 去删除任务
	if oldJob, err = G_jobMgr.DeleteJob(name); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
	return
}

// 列举所有crontab任务
func handleJobList(w http.ResponseWriter, r *http.Request) {
	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)

	if jobList, err = G_jobMgr.ListJobs(); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		w.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
	return
}

// 强制杀死某个路由
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		name  string
		bytes []byte
	)

	// 解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	// 要杀死的任务名
	name = r.PostForm.Get("name")

	// 杀死任务
	if err = G_jobMgr.KillJob(name); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		w.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		w.Write(bytes)
	}
}

//初始化服务
func InitAPIServer() (err error) {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		staticDir     http.Dir
		staticHandler http.Handler
	)

	// 配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)

	// 静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler)) //   ./webroot/index.html

	// 启动TCP监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.APIPort)); err != nil {
		return
	}

	// 创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.APIReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.APIWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	//赋值单例
	G_apiServer = &APIServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)

	return
}
