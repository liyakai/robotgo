package robot

import (
	"sync"
	"time"
)

type robotMgr struct {
	robots      map[string]*Robot
	mutex       sync.RWMutex
	maxRobotNum uint32
	freq        int64
	robotNum    uint32
	finish      bool
}

var robMgr *robotMgr

func RobotMgrGetMe() *robotMgr {
	if robMgr == nil {
		robMgr = &robotMgr{
			robots: make(map[string]*Robot),
		}
	}
	return robMgr
}

func (robm *robotMgr) Init(maxRobotNum uint32, freq int64) {
	robm.maxRobotNum = maxRobotNum
	robm.freq = freq
	go robm.timeAction()
}

func (robm *robotMgr) CreateRobot(index uint32) {
	robm.mutex.Lock()
	defer robm.mutex.Unlock()
	rob := newRobot(index)
	robm.robots[rob.uuid] = rob
}

func (robm *robotMgr) timeAction() {
	createTick := time.NewTicker(time.Millisecond * time.Duration(robm.freq))
	defer createTick.Stop()
	for {
		select {
		case <-createTick.C:
			if robm.robotNum < robm.maxRobotNum {
				robm.CreateRobot(robm.robotNum)
				robm.robotNum++
			}
		}
	}
}

func (robm *robotMgr) IsFinish() bool {
	return robm.finish
}
