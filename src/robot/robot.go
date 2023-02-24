package robot

import (
	"math/rand"
	"robotgo/src/tools"
	"strconv"
	"time"

	"github.com/golang/glog"
	. "github.com/liyakai/behavior3go/config"
	. "github.com/liyakai/behavior3go/core"
	. "github.com/liyakai/behavior3go/loader"
)

type robotCmd struct {
	flag byte
	data []byte
}

type Robot struct {
	network *RNetwork
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

func newRobot(index uint32) *Robot {
	rbt := &Robot{
		uuid:    getRandStr(index, 0),
		pass:    "123456",
		network: NewNetWork(),
	}
	rbt.network.SetOwner(rbt)
	rbt.online = true
	rbt.zoneid = 1
	rbt.job = 1
	rbt.gender = 1
	rbt.name = rbt.uuid
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
	return retStr
}

func (rbt *Robot) RunAITree() {
	projectConfig, ok := LoadRawProjectCfg(tools.EnvGet("robot", "tree_file"))
	if !ok {
		glog.Infoln("LoadRawProjectCfg err")
		return
	}

	//自定义节点注册
	maps := RegisterNodes()

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
	board := RegisterBlackBoard()
	board.SetMem("robot", rbt)
	//循环每一帧
	cycle_num, err := strconv.Atoi(tools.EnvGet("robot", "cycle_num"))
	if nil != err {
		glog.Infoln("解析配置参数 cycle_num 失败.")
	}
	for i := 0; i < cycle_num; i++ {
		firstTree.Tick(i, board)
		time.Sleep(time.Millisecond * 100)
	}
}
