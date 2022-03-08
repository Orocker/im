package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
	// 当前用户连接的server句柄
	server *Server
}

// NewUser 创建用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMsg()
	return user
}

// ListenMsg 监听当前User channel，一旦有消息就直接发送给对端客户端
func (user *User) ListenMsg() {
	for {
		msg := <-user.C
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
}

// Online 用户上线业务
func (user *User) Online() {
	user.server.MapLock.Lock()
	// 用户上线后，将用户加入onlineMap中
	user.server.OnlineMap[user.Name] = user
	user.server.MapLock.Unlock()
	// 广播用户上线消息
	user.server.BroadCast(user, "is online")

}

// Offline 用户下线业务
func (user *User) Offline() {
	user.server.MapLock.Lock()
	// 用户上线后，将用户移除onlineMap
	delete(user.server.OnlineMap, user.Name)
	user.server.MapLock.Unlock()
	// 广播用户上线消息
	user.server.BroadCast(user, "is offline")
}

// DoMessage 处理用户发送的消息
func (user *User) DoMessage(msg string) {
	//输入指定指令查询当前在线的用户

	if msg == "who" {
		user.server.MapLock.Lock()
		for _, onlineUser := range user.server.OnlineMap {
			onlineMsg := "[" + onlineUser.Addr + "]" + onlineUser.Name + ": is online\n"
			user.SendMsg(onlineMsg)
		}
		user.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// rename|newName 指令修改当前用户名
		newName := strings.Split(msg, "|")[1]

		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("This name is existed ")
		} else {
			user.server.MapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.MapLock.Unlock()
			user.Name = newName

			user.SendMsg("Your new user name is :" + newName + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 私聊,消息格式 to|userName

		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("The message format is incorrect")
			return
		}
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("The username does not exist")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("The message can not be empty")
			return
		}

		remoteUser.SendMsg(user.Name + ":" + content + "\n")

	} else {
		user.server.BroadCast(user, msg)
	}
}

// SendMsg 给当前user对应的客户端发送消息
func (user *User) SendMsg(msg string) {
	_, err := user.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}
