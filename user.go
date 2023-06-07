package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 新建用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线
func (this *User) Online() {
	// 用户上线，将用户加到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线信息
	this.server.BroadCast(this, "onLine")
}

// 用户下线
func (this *User) DeadLine() {
	// 用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户下线信息
	this.server.BroadCast(this, "deadLine")
}

// 用户发给自己的消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			sendMsg := "[" + user.Addr + "]" + user.Name + ": is onLine...\n"
			//this.C <- sendMsg
			this.SendMsg(sendMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|sxr
		newName := msg[7:]

		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("this name already exit! please change another one  \n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("you already change your name: " + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式: to|sxr|hello
		//1.获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("msg format is worry! please use \"to|sxr|hello\"format. \n")
			return
		}

		//2.根据用户名获取 User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg(" this userName is not exit!\n")
			return
		}

		//3.获取消息内容，通过user对象发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("msg is empty, please send again!\n")
			return
		}
		remoteUser.SendMsg(this.Name + ": " + content)

	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听管道中
func (this *User) ListenMessage() {
	for {
		mag := <-this.C
		this.conn.Write([]byte(mag + "\n"))
	}
}
