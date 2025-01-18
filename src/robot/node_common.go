package robot

import (
	"encoding/binary"
	"math/rand"
	"robotgo/src/tools"
	"strconv"

	//"strconv"
	"time"

	"bytes"

	demo "robotgo/src/proto"

	"github.com/golang/glog"
	b3 "github.com/liyakai/behavior3go"
	. "github.com/liyakai/behavior3go/config"
	. "github.com/liyakai/behavior3go/core"
	"google.golang.org/protobuf/proto"
)

// 自定义action节点
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
	binary.Write(buffer, binary.LittleEndian, uint32(block_size))
	binary.Write(buffer, binary.LittleEndian, block)

	// glog.Infoln("\n\n发送数据块:", buffer.Bytes()[0:31])
	var sendErr error
	if network == "tcp" {
		sendErr = rbt.network.SendMsg(buffer.Bytes())
	} else if network == "udp" {
		sendErr = rbt.network.SendUdpMsg(buffer.Bytes())
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

	// readtime 时间后再Read
	read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
	if read_time > 0 {
		time.Sleep(time.Millisecond * time.Duration(read_time))
	}

	block_size := int(tick.Blackboard.GetMem("big_byte_size").(int32))
	network := tools.EnvGet("robot", "network")
	var rcv_len int
	var rcv_data []byte
	if network == "tcp" {
		rcv_data_len_data, rcv_data_len := rbt.network.ReceiveMsgWithLen(4)
		if 4 != rcv_data_len {
			glog.Infoln("接收数据块头部大小: ", rcv_data_len)
			return b3.FAILURE
		}
		//glog.Infoln("接收数据块头部大小: ", len(rcv_data_len))
		data_len := binary.LittleEndian.Uint32(rcv_data_len_data)
		rcv_data, rcv_len = rbt.network.ReceiveMsgWithLen((int)(data_len))
		//glog.Infoln("接收数据块body大小: ", rcv_len)
	} else if network == "udp" {
		var n int
		rcv_data, n = rbt.network.ReceiveMsgFromUdp()
		rcv_len = int(n) - 4
	} else if network == "kcp" {
		rcv_data, rcv_len = rbt.network.ReceiveKcpMsg()
		rcv_len = rcv_len - 4
	}
	// glog.Infoln("接收数据块大小", rcv_len)
	// glog.Infoln("接收数据块内容:", rcv_data[0:32])
	if rcv_len != block_size {
		glog.Infoln(rbt.name, "接收数据块大小 :", rcv_len, " block_size:", block_size)
		glog.Infoln(rbt.name, "接收数据块大小", len(rcv_data))
		if rcv_len >= 32 {
			glog.Infoln(rbt.name, "接收数据块内容:", rcv_data[0:32])
		}
	}

	// glog.Info("收到数据:", rcv_data[0:31])
	return b3.SUCCESS
}

// 发送 SendProtocolBytes
type SendProtocolBytesReq struct {
	Action
}

func (this *SendProtocolBytesReq) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *SendProtocolBytesReq) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	network := tools.EnvGet("robot", "network")
	// block_size := tick.Blackboard.GetMem("big_byte_size").(int32)
	// block := make([]byte, block_size)
	// for i := int32(0); i < block_size; i++ {
	// 	block[i] = byte(i)
	// }
	req := &demo.GetUserRequest{
		UserId: 955,
	}
	req_size := proto.Size(req)

	buffer := new(bytes.Buffer)
	// binary.Write(buffer, binary.LittleEndian, uint32(block_size)+4)
	binary.Write(buffer, binary.LittleEndian, uint32(req_size)+20)
	// 写入消息头
	binary.Write(buffer, binary.LittleEndian, uint8(0xde))      // magic
	binary.Write(buffer, binary.LittleEndian, uint8(0x01))      // version
	binary.Write(buffer, binary.LittleEndian, uint8(0x01))      // serialize_type
	binary.Write(buffer, binary.LittleEndian, uint8(0x01))      // msg_type
	binary.Write(buffer, binary.LittleEndian, uint32(9527))        // seq_num
	binary.Write(buffer, binary.LittleEndian, uint32(3417382616))        // func_id
	binary.Write(buffer, binary.LittleEndian, uint32(req_size)) // length
	binary.Write(buffer, binary.LittleEndian, uint32(0))        // attach_length
	// binary.Write(buffer, binary.LittleEndian, block)
	// 写入请求消息

	reqData, err := proto.Marshal(req)
	if err != nil {
		glog.Errorln("序列化请求消息失败:", err)
		return b3.FAILURE
	}
	binary.Write(buffer, binary.LittleEndian, reqData)

	// glog.Infoln("\n\n发送数据块:", buffer.Bytes()[0:31])
	var sendErr error
	if network == "tcp" {
		sendErr = rbt.network.SendMsg(buffer.Bytes())
	} else if network == "udp" {
		sendErr = rbt.network.SendUdpMsg(buffer.Bytes())
	} else if network == "kcp" {
		sendErr = rbt.network.SendKcpMsg(buffer.Bytes())
	}

	if sendErr != nil {
		return b3.FAILURE
	}
	// glog.Infoln("robot:",rbt.name," SendProtocolBytesReq, network:", network)
	return b3.SUCCESS
}

// 接收 SendProtocolBytes 返回的消息
type SendProtocolBytesRet struct {
	Action
}

func (this *SendProtocolBytesRet) Initialize(setting *BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *SendProtocolBytesRet) OnTick(tick *Tick) b3.Status {
	rbt := tick.Blackboard.GetMem("robot").(*Robot)
	// readtime 时间后再Read
	read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
	if read_time > 0 {
		time.Sleep(time.Millisecond * time.Duration(read_time))
	}

	// block_size := int(tick.Blackboard.GetMem("big_byte_size").(int32))
	network := tools.EnvGet("robot", "network")
	var rcv_len int
	var rcv_data []byte
	if network == "tcp" {
		rcv_data_len_data, rcv_data_len := rbt.network.ReceiveMsgWithLen(4)
		if 4 != rcv_data_len {
			glog.Infoln("robot:", rbt.name ," 接收数据块头部大小: ", rcv_data_len)
			return b3.FAILURE
		}
		data_len := binary.LittleEndian.Uint32(rcv_data_len_data)
		// glog.Info("robot:", rbt.name ," 接收数据块大小: ", data_len, " 头部数据内容 %x:",rcv_data_len_data)
		rcv_data, rcv_len = rbt.network.ReceiveMsgWithLen((int)(data_len))
		//glog.Infoln("接收数据块body大小: ", rcv_len)
	} else if network == "udp" {
		var n int
		rcv_data, n = rbt.network.ReceiveMsgFromUdp()
		rcv_len = int(n) - 4
	} else if network == "kcp" {
		rcv_data, rcv_len = rbt.network.ReceiveKcpMsg()
		rcv_len = rcv_len - 4
	}
	// glog.Infoln("接收数据块大小", rcv_len)
	// glog.Infof("接收数据块内容: %x", rcv_data)

	// 解析消息头
	if len(rcv_data) < 20 {
		glog.Errorln("消息头长度不足20字节")
		return b3.FAILURE
	}

	header := struct {
		Magic         uint8
		Version       uint8  
		SerializeType uint8
		ErrCode       uint8
		MsgType       uint8
		SeqNum        uint32
		Length        uint32
		AttachLength  uint32
	}{}

	header.Magic = rcv_data[0]
	header.Version = rcv_data[1]
	header.SerializeType = rcv_data[2] 
	header.ErrCode = rcv_data[3]
	header.MsgType = rcv_data[4]
	header.SeqNum = binary.LittleEndian.Uint32(rcv_data[8:12])
	header.Length = binary.LittleEndian.Uint32(rcv_data[12:16])
	header.AttachLength = binary.LittleEndian.Uint32(rcv_data[16:20])

	// glog.Infof("消息头: Magic=%d Version=%d SerializeType=%d ErrCode=%d MsgType=%d SeqNum=%d Length=%d AttachLength=%d",
	// 	header.Magic, header.Version, header.SerializeType, header.ErrCode, header.MsgType, header.SeqNum, header.Length, header.AttachLength)
	// 跳过消息头20字节
	resp := &demo.GetUserResponse{}
	err := proto.Unmarshal(rcv_data[20:20+header.Length], resp)
	if err != nil {
		glog.Errorln("反序列化响应消息失败:", err)
		return b3.FAILURE
	}
	// 获取附件数据
	if header.AttachLength > 0 {
		// attachData := rcv_data[20+header.Length : 20+header.Length+header.AttachLength]
		// glog.Infof("robot:%s, 附件数据: %s", rbt.name, attachData)
	}

	// glog.Info("robot:", rbt.name ," 收到结果:", resp)
	return b3.SUCCESS
}
