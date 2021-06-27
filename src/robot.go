package main

import (
	"fmt"
	"math/rand"
	"robotgo/src/tools"
	"strconv"
	"time"

	"github.com/golang/glog"
	b3 "github.com/liyakai/behavior3go"
	. "github.com/liyakai/behavior3go/config"
	. "github.com/liyakai/behavior3go/core"
	. "github.com/liyakai/behavior3go/loader"
)

type robotCmd struct {
	flag byte
	data []byte
}

type robot struct {
	id      uint64
	uuid    string
	pass    string
	zoneid  uint32
	url     string
	cookies string
	online  bool
	job     uint32
	gender  uint32
	name    string
}

func newRobot(index uint32) *robot {
	rbt := &robot{
		uuid: getRandStr(index, 0),
		pass: "123456",
	}
	rbt.online = true
	rbt.zoneid = 1
	rbt.job = 1
	rbt.gender = 1
	rbt.name = rbt.uuid //getRandStr(index,2)
	glog.Infoln("初始化机器人：", rbt.uuid)

	go rbt.RunAITree()
	return rbt
}

//

func getRandStr(index uint32, lenght int) string {
	data := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "D", "E", "F"} //  "B", "C" 可能触发屏蔽字 (博彩)
	retStr := tools.EnvGet("robot", "prefix_name")
	retStr += strconv.FormatUint(uint64(index), 10)
	for i := 0; i < lenght; i++ {
		retStr += data[rand.Intn(len(data))]
	}
	fmt.Println(retStr)
	//time.Sleep(time.Second*100)
	return retStr
}

func (rbt *robot) RunAITree() {
	projectConfig, ok := LoadRawProjectCfg("example.b3")
	if !ok {
		fmt.Println("LoadRawProjectCfg err")
		return
	}

	//自定义节点注册
	maps := b3.NewRegisterStructMaps()
	// maps.Register("Log", new(NodeLog))
	// maps.Register("Connected", new(Connected))
	// maps.Register("ConnectedRes", new(ConnectedRes))
	// maps.Register("ConnectServer", new(ConnectServer))
	// maps.Register("DisConnectServer", new(DisConnectServer))
	// maps.Register("SleepMS", new(SleepMS))
	// maps.Register("IsConnectGate", new(IsConnectGate))
	// maps.Register("RandomChooseOne", new(RandomChooseOne))
	// maps.Register("RandomExe", new(RandomExe))
	// maps.Register("SendBigByte", new(SendBigBytePb))
	// maps.Register("SendBigByteRes", new(SendBigBytePbRes))

	var firstTree *BehaviorTree
	//载入
	for _, v := range projectConfig.Data.Trees {
		tree := CreateBevTreeFromConfig(&v, maps)
		glog.Infoln(tree.GetTitile())
		if tree.GetTitile() == tools.EnvGet("robot", "tree") && firstTree == nil {
			firstTree = tree
			tree.Print()
		}
	}

	//输入板
	board := NewBlackboard()
	board.SetMem("robot", rbt)
	board.SetMem("is_connect_gate", false) // 设置为不在线
	big_byte_size, _ := strconv.Atoi(tools.EnvGet("robot", "big_byte_size"))
	board.SetMem("big_byte_size", int32(big_byte_size))
	board.SetMem("test_pkt_size", int32(big_byte_size))

	//循环每一帧
	for i := 0; i < 30000000; i++ {
		firstTree.Tick(i, board)
		time.Sleep(time.Millisecond * 100)
	}
}
