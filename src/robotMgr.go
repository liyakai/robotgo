package main

import (
	"sync"
	"time"
)

type robotMgr struct {
	robots      map[string]*robot
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
			robots: make(map[string]*robot),
		}
	}
	return robMgr
}

func (robm *robotMgr) init(maxRobotNum uint32, freq int64) {
	robm.maxRobotNum = maxRobotNum
	robm.freq = freq
	go robm.timeAction()
}

func (robm *robotMgr) createRobot(index uint32) {
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
				robm.createRobot(robm.robotNum)
				robm.robotNum++
			}
		}
	}
}

func (robm *robotMgr) isFinish() bool {
	return robm.finish
}
