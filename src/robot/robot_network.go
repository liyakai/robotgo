package robot

import (
	//"errors"
	"io"
	"io/ioutil"
	"net"
	"robotgo/src/tools"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/xtaci/kcp-go"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type RNetwork struct {
	owner    *Robot
	id       uint64
	acc      string
	conn     net.Conn
	kcpconn  *kcp.KCP
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
	if network == "kcp" {
		kcpconn := kcp.NewKCP(0x01020304, func(buf []byte, size int) {
			glog.Info("kcp 封装后发送数据长度:", len(buf), "<--> size:", size)
			glog.Info("kcp 发送封装后的数据:", buf[0:32])
			gc.conn.Write(buf[0:size])
		}) // 			 gc.KcpDialWithOptions(address, 10, 3)
		// kcpconn.kcp.conv = 0x01020304
		gc.kcpconn = kcpconn

		network = "udp"
	}
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
	gc.kcpconn.ReleaseTX()
	glog.Infoln("[服务] 与网关服务器断开连接")
}

func (gc *RNetwork) KcpUpdate() {
	for i := 0; i < 10; i++ {
		gc.kcpconn.Update()
	}

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

func (gc *RNetwork) ReceiveMsgWithLen(len uint32) (data []byte) {
	read_time, _ := strconv.Atoi(tools.EnvGet("robot", "readtime"))
	gc.conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(read_time)))
	data, err := ioutil.ReadAll(io.LimitReader(gc.conn, int64(len)))
	if err != nil {
		glog.Errorln("gateClient 错误码", err.Error(), int64(len))
		//time.Sleep(time.Millisecond * 1000)
		return nil
	}
	return
}

func (gc *RNetwork) ReceiveKcpMsg() (data []byte) {
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
	glog.Info("成功读取消息 ReadAll", data[0:32])
	kcp_recv_len := gc.kcpconn.Input(data, true, true)
	glog.Info("成功读取消息 Input", data[0:32])
	gc.KcpUpdate()
	if kcp_recv_len > 0 {
		kcp_recv_len = gc.kcpconn.Recv(data)
		glog.Info("成功读取消息 Recv", data[0:32])
		gc.KcpUpdate()
		if kcp_recv_len > 0 {
			glog.Info("成功读取消息", data[0:32])
		}
	}
	gc.KcpUpdate()
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

func (gc *RNetwork) SendKcpMsg(data []byte) error {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("[异常] 报错 ", err, "\n", string(debug.Stack()))
		}
	}()
	if nil == gc.kcpconn {
		glog.Errorln("网络已经断开连接,无法发送")
		return errors.New("网络已经断开连接")
	}
	glog.Info("kcp 封装前发送数据长度:", len(data))
	gc.kcpconn.Send(data)
	gc.KcpUpdate()
	return nil
}
