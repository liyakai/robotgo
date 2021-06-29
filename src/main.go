package main

import (
	"flag"
	"fmt"
	"math/rand"
	. "robotgo/src/robot"
	"robotgo/src/tools"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/golang/glog"
)

var (
	robotNum  = flag.Int("num", 1, "机器人数量")
	loginFreq = flag.Int("freq", 10, "机器人登录频率")
	logdir    = flag.String("logdir", "", "Log file dir")
	config    = flag.String("config", "robot.json", "config path")
	seednum   = flag.Int("seednum", 1, "随机种子")
)

func main() {
	flag.Parse()
	defer func() {
		if err := recover(); err != nil {
			glog.Error("[异常] 报错 ", err, "\n", string(debug.Stack()))
		}
	}()
	robot_num, login_req := ConfigParam()
	glog.Info("[启动] 开始机器人测试")

	RobotMgrGetMe().Init(uint32(robot_num), int64(login_req))
	glog.Info("[启动] 完成 ")
	for {
		if RobotMgrGetMe().IsFinish() {
			return
		}

		time.Sleep(time.Second * 5)
	}
}

// 配置参数
func ConfigParam() (int, int) {
	if !tools.EnvLoad(*config) {
		glog.Error("加载配置文件失败")
		return 0, 0
	}

	loglevel := tools.EnvGlobal("loglevel")
	fmt.Println("loglevel:" + loglevel)
	if loglevel != "" {
		flag.Lookup("stderrthreshold").Value.Set(loglevel)
	}
	// 是否只输出到标准输出中
	logtostderr := tools.EnvGlobal("logtostderr")
	fmt.Println("logtostderr:" + logtostderr)
	if logtostderr != "" {
		flag.Lookup("logtostderr").Value.Set(logtostderr)
	}

	// 日志输出到日志中时,是否也输出到标准错误输出
	logtostderr = tools.EnvGlobal("alsologtostderr")
	glog.Info("alsologtostderr:" + logtostderr)
	if logtostderr != "" {
		flag.Lookup("alsologtostderr").Value.Set(logtostderr)
	}

	// 日志目录
	if *logdir != "" {
		flag.Lookup("log_dir").Value.Set(*logdir)
	} else {
		flag.Lookup("log_dir").Value.Set(tools.EnvGlobal("logdir"))
	}
	glog.Info("log_dir:" + flag.Lookup("log_dir").Value.String())
	// 调整并发
	runtime.GOMAXPROCS(runtime.NumCPU())

	rand.Seed(int64(*seednum))
	// 机器人数量
	robot_num, err := strconv.Atoi(tools.EnvGet("robot", "robot_num"))
	if err != nil {
		glog.Errorln("[启动] 请检查配置config.json. 是否有robot robot_num 项")
		return 0, 0
	}
	if *robotNum != 1 {
		robot_num = *robotNum
	}
	glog.Info("机器人数量:" + strconv.Itoa(robot_num))

	// 机器人登录频率
	login_freq, err := strconv.Atoi(tools.EnvGet("robot", "login_freq"))
	if err != nil {
		glog.Errorln("[启动] 请检查配置config.json. 是否有robot login_freq 项")
		return 0, 0
	}
	if *loginFreq != 10 {
		login_freq = *loginFreq
	}
	glog.Info("机器人登录频率:" + strconv.Itoa(login_freq))
	glog.Flush()
	return robot_num, login_freq
}
