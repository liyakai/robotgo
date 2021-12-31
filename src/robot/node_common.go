package robot

import (
	"encoding/binary"
	"math/rand"
	"robotgo/src/tools"
	"strconv"

	//"strconv"
	"time"

	"bytes"

	"github.com/golang/glog"
	b3 "github.com/liyakai/behavior3go"
	. "github.com/liyakai/behavior3go/config"
	. "github.com/liyakai/behavior3go/core"
)

//自定义action节点
type NodeLog struct {
	Action
	info string
}

func (this *NodeLog) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
	this.info = setting.GetPropertyAsString("info")
}

func (this *NodeLog) OnTick(tick *Tick) b3.Status {
	glog.Infoln("nodelog:", tick.GetLastSubTree(), this.info)
	return b3.SUCCESS
}

// sleep ms
type SleepMS struct {
	Action
	sleeptop  int
	sleepbase int
}

func (this *SleepMS) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
	this.sleeptop = setting.GetPropertyAsInt("sleeptop")
	this.sleepbase = setting.GetPropertyAsInt("sleepbase")
	//this.sleepms = str_ms[3:len(str_ms)-3]	// 前后减去三个字节是为了去掉引号

}

func (this *SleepMS) OnTick(tick *Tick) b3.Status {
	//glog.Infoln("执行 Sleep 节点")
	intsleepms := rand.Intn(this.sleeptop-this.sleepbase) + this.sleepbase
	//glog.Infoln("休眠:", tick.GetLastSubTree(), strconv.Itoa(intsleepms), "ms")
	time.Sleep(time.Millisecond * time.Duration(intsleepms))
	return b3.SUCCESS
}

// 连接服务器
type ConnectServer struct {
	Action
	gateaddr string
}

func (this *ConnectServer) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gateaddr = setting.GetPropertyAsString("gateaddr")
}

func (this *ConnectServer) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	var gateaddr string
	if len(tools.EnvGet("robot", "gateaddr")) > 0 {
		gateaddr = rbt.network.gateaddr
		glog.Infoln("配置的网关地址不为空,优先使用配置的网关地址:", gateaddr)
	} else {
		gateaddr = this.gateaddr
		glog.Infoln("配置的网关地址为空,使用行为树节点的网关地址:", gateaddr)
	}
	if !rbt.network.connect(gateaddr) {
		glog.Infoln("[连接网关] 连接网关服失败", rbt.uuid)
		return b3.FAILURE
	}
	tick.Blackboard.SetMem("is_online", true) // 设置为不在线
	glog.Infoln("设置在线状态为:", tick.Blackboard.GetMem("is_online").(bool))
	return b3.SUCCESS
}

// 关闭服务器连接
type DisConnectServer struct {
	Action
}

func (this *DisConnectServer) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *DisConnectServer) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	rbt.network.OnClose()
	tick.Blackboard.SetMem("is_online", false)       // 设置为不在线
	tick.Blackboard.SetMem("is_connect_gate", false) // 设置为断开连接
	glog.Infoln("设置TCP连接状态为:", tick.Blackboard.GetMem("is_online").(bool))
	return b3.SUCCESS
}

// 自定义条件节点 判断是否有连接
type IsConnectGate struct {
	Decorator
}

func (this *IsConnectGate) Initialize(setting *BTNodeCfg) {
	this.Decorator.Initialize(setting)
}

func (this *IsConnectGate) OnTick(tick *Tick) b3.Status {
	is_connect_gate := tick.Blackboard.GetMem("is_connect_gate").(bool)
	if is_connect_gate == false {
		if this.GetChild() == nil {
			return b3.FAILURE
		}
		var status = this.GetChild().Execute(tick)
		if status == b3.SUCCESS {
			tick.Blackboard.SetMem("is_connect_gate", true) // 设置为在线
			glog.Infoln("设置TCP连接状态为:", tick.Blackboard.GetMem("is_online").(bool))
		}
		return status
	}
	return b3.SUCCESS
}

// 自定义 随机选择一个子节点执行
type RandomChooseOne struct {
	Composite
}

func (this *RandomChooseOne) Initialize(setting *BTNodeCfg) {
	this.Composite.Initialize(setting)
}

func (this *RandomChooseOne) OnTick(tick *Tick) b3.Status {
	total_count := this.GetChildCount()
	exe_child := rand.Intn(total_count)
	var status = this.GetChild(exe_child).Execute(tick)
	return status
}

// 自定义 根据配置概率决定一个子节点是否执行
type RandomExe struct {
	Composite
	probability int
	denominator int
}

func (this *RandomExe) Initialize(setting *BTNodeCfg) {
	this.Composite.Initialize(setting)
	this.probability = setting.GetPropertyAsInt("probability")
	this.denominator = setting.GetPropertyAsInt("denominator")
	if this.denominator <= 0 || this.probability < 0 {
		this.denominator = 10000
		this.probability = 0
	}
}

func (this *RandomExe) OnTick(tick *Tick) b3.Status {
	exe_child := rand.Intn(this.denominator)
	if exe_child < this.probability {
		for i := 0; i < this.GetChildCount(); i++ {
			var status = this.GetChild(i).Execute(tick)
			if status != b3.SUCCESS {
				return status
			}
		}
	}
	return b3.SUCCESS
}

// 发送 SendBigByte
type SendBigByte struct {
	Action
}

func (this *SendBigByte) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *SendBigByte) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	network := tools.EnvGet("robot", "network")
	block_size := tick.Blackboard.GetMem("big_byte_size").(int32)
	block := make([]byte, block_size)
	for i := int32(0); i < block_size; i++ {
		block[i] = byte(i)
	}
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, uint32(block_size)+4)
	binary.Write(buffer, binary.LittleEndian, block)

	// glog.Infoln("\n\n发送数据块:", buffer.Bytes()[0:31])
	var sendErr error
	if network == "tcp" || network == "udp" {
		sendErr = rbt.network.SendMsg(buffer.Bytes())
	} else if network == "kcp" {
		sendErr = rbt.network.SendKcpMsg(buffer.Bytes())
	}

	if sendErr != nil {
		return b3.FAILURE
	}
	return b3.SUCCESS
}

// 接收 SendBigByte 返回的消息
type SendBigByteRes struct {
	Action
}

func (this *SendBigByteRes) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *SendBigByteRes) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	block_size := tick.Blackboard.GetMem("big_byte_size").(int32)
	network := tools.EnvGet("robot", "network")
	var rcv_len int32
	var rcv_data []byte
	if network == "tcp" {
		rcv_data_len := rbt.network.ReceiveMsgWithLen(4)
		if 4 != len(rcv_data_len) {
			glog.Infoln("接收数据块头部大小: ", len(rcv_data_len))
			return b3.FAILURE
		}
		//glog.Infoln("接收数据块头部大小: ", len(rcv_data_len))
		data_len := binary.LittleEndian.Uint32(rcv_data_len) - 4
		rcv_data = rbt.network.ReceiveMsgWithLen(data_len)
		rcv_len = int32(len(rcv_data))
		read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
		// 模拟read超时时间,跟udp,kcp对齐,
		time.Sleep(time.Millisecond * time.Duration(read_time))
		//glog.Infoln("接收数据块body大小: ", rcv_len)
	} else if network == "udp" {
		rcv_data = rbt.network.ReceiveMsg()
		rcv_len = int32(len(rcv_data)) - 4
	} else if network == "kcp" {
		rcv_data, rcv_len = rbt.network.ReceiveKcpMsg()
		rcv_len = rcv_len - 4
	}
	// glog.Infoln("接收数据块大小", rcv_len)
	// glog.Infoln("接收数据块内容:", rcv_data[0:32])
	if rcv_len != block_size {
		glog.Infoln("接收数据块大小", len(rcv_data))
		if rcv_len >= 32 {
			glog.Infoln("接收数据块内容:", rcv_data[0:32])
		}
	}

	// glog.Info("收到数据:", rcv_data[0:31])
	return b3.SUCCESS
}
