package robot

import (
	"io/ioutil"
	"io"
	"net"
	"robotgo/src/tools"
	"runtime/debug"
	"strconv"
	"time"

	"errors"

	"github.com/golang/glog"
)

type RNetwork struct {
	owner    *Robot
	id       uint64
	acc      string
	conn     net.Conn
	gateaddr string
}

func NewNetWork() *RNetwork {
	gatecli := &RNetwork{}
	gatecli.gateaddr = tools.EnvGet("robot", "gateaddr")
	glog.Infoln("配置设置的网关服务器地址:", gatecli.gateaddr)
	return gatecli
}

func (gc *RNetwork) SetOwner(owner *Robot) {
	gc.owner = owner
}

func (gc *RNetwork) connect(address string) bool {
	network := tools.EnvGet("robot", "network")
	conn, err := net.Dial(network, address)
	if err != nil {
		glog.Errorln("连接网关失败.", address)
		return false
	}
	gc.conn = conn
	//gc.Start()
	glog.Info("连接成功 address:", address)
	return true
}

func (gc *RNetwork) OnClose() {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("[异常] 报错 ", err, "\n", string(debug.Stack()))
		}
	}()
	if nil == gc.conn {
		glog.Infoln("[服务] 已经与网关断开连接")
		return
	}
	gc.conn.Close()
	glog.Infoln("[服务] 与网关服务器断开连接")
}

func (gc *RNetwork) ParseMsg(data []byte, flag byte) bool {
	glog.Infoln("RNetwork ParseData:", data)

	return true
}

func (gc *RNetwork) ReceiveMsg() (data []byte) {
	//glog.Infoln("RNetwork 开始接收消息")
	read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
	gc.conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(read_time)))
	data, err := ioutil.ReadAll(gc.conn) //接收消息
	if err != nil {
		// glog.Errorln("gateClient 接收出错:", err.Error())
		//glog.Errorln("gateClient 接收出错:", data)
	}
	// glog.Infoln("gateClient 接收到的内容是")
	// glog.Info(data)
	return
}

func (gc *RNetwork) ReceiveMsgWithLen(len uint32) (data []byte)  {
	read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
	gc.conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(read_time)))
	data,err := ioutil.ReadAll(io.LimitReader(gc.conn, int64(len)))
	if err != nil {
		glog.Errorln("gateClient 错误码", err.Error(), int64(len))
		//time.Sleep(time.Millisecond * 1000)
		return nil
	}
	return
}

func (gc *RNetwork) SendMsg(data []byte) error {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("[异常] 报错 ", err, "\n", string(debug.Stack()))
		}
	}()
	if nil == gc.conn {
		glog.Errorln("网络已经断开连接,无法发送")
		return errors.New("网络已经断开连接")
	}
	_, err := gc.conn.Write(data)
	if err != nil {
		glog.Errorln("网络发送失败的原因为", err.Error())
		return err
	}
	return nil
}
