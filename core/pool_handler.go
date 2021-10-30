package core

import (
	"crypto/tls"
	"github.com/kataras/golog"
	"net"
	"sync"
)

// NatokPool interface NATOK连接池处理接口
type NatokPool interface {
	Add(*PoolHandler) (*ConnectHandler, error) //添加进连接池
	Remove(*ConnectHandler)                    //移除出连接池
	IsActive(*ConnectHandler) bool             //连接是否活跃
}

type ConnPooler struct {
	Addr string
	Conf *tls.Config
}

// PoolHandler struct
type PoolHandler struct {
	Mu    sync.Mutex        //锁
	Pool  NatokPool         //池
	Conns []*ConnectHandler //连接
}

// GetConn 获取连接
func (p *PoolHandler) GetConn() (*ConnectHandler, error) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if len(p.Conns) == 0 {
		return nil, nil
	}
	conn := p.Conns[len(p.Conns)-1]
	p.Conns = p.Conns[:len(p.Conns)-1]
	if p.Pool.IsActive(conn) {
		return conn, nil
	} else {
		return nil, nil
	}
}

// Push 放入连接池
func (p *PoolHandler) Push(conn *ConnectHandler) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.Conns = append(p.Conns, conn)
}

// Pull 从连接池取出
func (p *PoolHandler) Pull() (*ConnectHandler, error) {
	for {
		if len(p.Conns) == 0 {
			conn, err := p.Pool.Add(p)
			if err != nil {
				return nil, err
			}
			return conn, nil
		} else {
			conn, err := p.GetConn()
			if conn != nil {
				return conn, err
			}
		}
	}
}

// Add 添加进连接池
func (p *ConnPooler) Add(pool *PoolHandler) (*ConnectHandler, error) {
	var conn net.Conn
	var err error

	if p.Conf != nil {
		conn, err = tls.Dial("tcp", p.Addr, p.Conf)
	} else {
		conn, err = net.Dial("tcp", p.Addr)
	}

	if err != nil {
		golog.Errorf("Error:+v%", err)
		return nil, err
	}

	natokServerHandler := &NatokServerHandler{PoolHandler: pool}
	connHandler := &ConnectHandler{
		Active:     true,
		Conn:       conn,
		MsgHandler: interface{}(natokServerHandler).(MsgHandler),
	}
	natokServerHandler.ConnHandler = connHandler
	natokServerHandler.HeartBeat()
	go func() {
		connHandler.Listen(conn, natokServerHandler)
	}()
	return connHandler, nil
}

// Remove 移除出连接池
func (p *ConnPooler) Remove(connHandler *ConnectHandler) {
	connHandler.Conn.Close()
}

// IsActive 连接是否活跃
func (p *ConnPooler) IsActive(connHandler *ConnectHandler) bool {
	return connHandler.Active
}
