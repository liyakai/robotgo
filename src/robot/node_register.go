package robot

import (
	"robotgo/src/tools"
	"strconv"

	b3 "github.com/liyakai/behavior3go"
	. "github.com/liyakai/behavior3go/core"
)

func RegisterBlackBoard() *Blackboard {
	//输入板
	board := NewBlackboard()
	board.SetMem("is_connect_gate", false) // 设置为不在线
	big_byte_size, _ := strconv.Atoi(tools.EnvGet("robot", "big_byte_size"))
	board.SetMem("big_byte_size", int32(big_byte_size))
	board.SetMem("test_pkt_size", int32(big_byte_size))
	return board
}

func RegisterNodes() *b3.RegisterStructMaps {
	//自定义节点注册
	maps := b3.NewRegisterStructMaps()
	maps.Register("Log", new(NodeLog))
	maps.Register("ConnectServer", new(ConnectServer))
	maps.Register("DisConnectServer", new(DisConnectServer))
	maps.Register("SleepMS", new(SleepMS))
	maps.Register("IsConnectGate", new(IsConnectGate))
	maps.Register("RandomChooseOne", new(RandomChooseOne))
	maps.Register("RandomExe", new(RandomExe))
	maps.Register("SendBigByte", new(SendBigByte))
	maps.Register("SendBigByteRes", new(SendBigByteRes))
	maps.Register("SendProtocolBytesReq", new(SendProtocolBytesReq))
	maps.Register("SendProtocolBytesRet", new(SendProtocolBytesRet))
	return maps
}
