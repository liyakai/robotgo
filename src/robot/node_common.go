package robot

import (
	"math/rand"
	"robotgo/src/tools"
	"strconv"
	"time"

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
	intsleepms := rand.Intn(this.sleeptop) + this.sleepbase
	glog.Infoln("休眠:", tick.GetLastSubTree(), strconv.Itoa(intsleepms*100), "ms")
	time.Sleep(100 * time.Millisecond * time.Duration(intsleepms))
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
	return b3.FAILURE
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
	// rbt := tick.Blackboard.GetMem("robot").(*Robot)
	// var bblock tools.PBigByteBlock
	// bblock.Size = tick.Blackboard.GetMem("big_byte_size").(int32)
	// bblock.ByteBlock = make([]byte, bblock.Size)
	// packetBin, err := bblock.GetPacketBin()
	// if err != nil {
	// 	glog.Infoln("生成 SendBigByte 包失败", err.Error())
	// 	return b3.FAILURE
	// }
	// glog.Infoln(packetBin)
	// sendErr := rbt.network.SendMsg(packetBin)
	// if sendErr != nil {
	// 	return b3.FAILURE
	// }
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
	// rbt := tick.Blackboard.GetMem("robot").(*Robot)
	// rcv_data := rbt.network.ReceiveMsg()
	// packets, err := tools.FindPacket("BigByteBlock", rcv_data)
	// if err != nil {
	// 	glog.Errorln("BigByteBlock 解包失败")
	// 	return b3.FAILURE
	// }
	// for i := 0; i < len(packets); i++ {
	// 	// 类型断言
	// 	pkt, ok := packets[i].(*tools.PBigByteBlock)
	// 	if !ok {
	// 		glog.Errorln("PBigByteBlock 类型断言失败", reflect.TypeOf(packets[0]))
	// 		return b3.FAILURE
	// 	}
	// 	glog.Infoln("接收数据块大小", pkt.Size)
	// 	return b3.SUCCESS
	// }
	return b3.FAILURE
}
