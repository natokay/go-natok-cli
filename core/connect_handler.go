package core

import (
	"github.com/kataras/golog"
	"net"
	"runtime/debug"
	"time"
)

// MsgHandler interface 消息处理接口
type MsgHandler interface {
	Success(*ConnectHandler)              //成功
	Error(*ConnectHandler)                //出错
	Encode(interface{}) []byte            //编码
	Decode([]byte) (interface{}, int)     //解码
	Receive(*ConnectHandler, interface{}) //接收
}

// Message struct 消息体对象
type Message struct {
	Type      byte
	SerialNum uint64
	Uri       string
	Data      []byte
}

// ConnectHandler struct 通道链接载体
type ConnectHandler struct {
	ReadTime    int64           //读取时间
	WriteTime   int64           //写入时间
	Active      bool            //是否活跃
	ReadBuf     []byte          //读取的内容
	Conn        net.Conn        //连接通道
	MsgHandler  MsgHandler      //消息句柄
	ConnHandler *ConnectHandler //连接句柄
}

// Write 消息写入
func (c *ConnectHandler) Write(msg interface{}) {
	if c.MsgHandler == nil {
		return
	}
	data := c.MsgHandler.Encode(msg)
	c.WriteTime = time.Now().Unix()
	c.Conn.Write(data)
}

// Listen 连接请求监听
func (c *ConnectHandler) Listen(conn net.Conn, msgHandler interface{}) {
	defer func() {
		if err := recover(); err != nil {
			c.MsgHandler.Error(c)
			golog.Warnf("Warn: %+v", err)
			debug.PrintStack()
		}
	}()

	if conn == nil {
		return
	}

	c.Conn = conn
	c.Active = true
	c.ReadTime = time.Now().Unix()
	c.MsgHandler = msgHandler.(MsgHandler)
	c.MsgHandler.Success(c)

	for {
		buf := make([]byte, 1024*64)
		if c.ReadBuf != nil && len(c.ReadBuf) > MaxPacketSize {
			golog.Error("Warn:  This conn is error ! Packet max than 4M !")
			c.Conn.Close()
		}

		n, err := c.Conn.Read(buf)
		if err != nil || n == 0 {
			golog.Errorf("Error:%v", err)
			//debug.PrintStack()
			c.Active = false
			c.MsgHandler.Error(c)
			break
		}

		c.ReadTime = time.Now().Unix()
		if c.ReadBuf == nil {
			c.ReadBuf = buf[0:n]
		} else {
			c.ReadBuf = append(c.ReadBuf, buf[0:n]...)
		}

		for {
			msg, n := c.MsgHandler.Decode(c.ReadBuf)
			if msg == nil {
				break
			}
			c.MsgHandler.Receive(c, msg)
			c.ReadBuf = c.ReadBuf[n:]
			if len(c.ReadBuf) == 0 {
				break
			}
		}

		if len(c.ReadBuf) > 0 {
			buf := make([]byte, len(c.ReadBuf))
			copy(buf, c.ReadBuf)
			c.ReadBuf = buf
		}
	}
}
