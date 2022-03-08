package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	MapLock   sync.RWMutex

	// 消息广播channel
	Message chan string
}

// NewServer 创建server实例
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// 监听message 广播消息channel 的 goroutine, 一旦有消息就广播给所有在线user

func (serv *Server) ListenMessage() {
	for {
		msg := <-serv.Message
		serv.MapLock.Lock()

		// 消息发送给全部在线user
		for _, cli := range serv.OnlineMap {
			cli.C <- msg
		}
		serv.MapLock.Unlock()
	}
}

// BroadCast 广播消息
func (serv *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	serv.Message <- sendMsg
}

func (serv *Server) Handle(conn net.Conn) {
	user := NewUser(conn, serv)
	// 用户上线后，将用户加入onlineMap中

	user.Online()
	// 接收客户端发送的消息

	isActive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read error", err)
				return
			}
			// 提取用户消息 去除(\n)
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			isActive <- true
		}
	}()

	for {
		select {
		case <-isActive:
		// 当前用户是活跃的，重置定时器
		// 不做任何事情，为了激活select 更新下面的定时器
		case <-time.After(10 * time.Hour):
			user.SendMsg("Time out")

			close(user.C)

			err := conn.Close()
			if err != nil {
				return
			}

			return
		}

	}

}

func (serv *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serv.Ip, serv.Port))
	if err != nil {
		fmt.Println("net.Listen:", err)
		return
	}
	// close listen
	defer listener.Close()
	go serv.ListenMessage()
	// accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accept err:", err)
			continue
		}
		go serv.Handle(conn)
	}

}
