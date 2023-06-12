package main

import (
	"net"
	"strings"
)

type User struct {
	Addr   string
	Name   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	remoteAddr := conn.RemoteAddr().String()
	user := &User{
		Addr:   remoteAddr,
		Name:   remoteAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "is online...")

}

func (user *User) DeadLine() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "is deadline...")

}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}

func (user *User) SendOwnMsg(msg string) {
	user.conn.Write([]byte(msg))
}

func (this *User) DoMsg(msg string) {
	// 查询所有在线用户
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			sendMsg := "[" + user.Addr + "]" + user.Name + " is online...\n"
			this.SendOwnMsg(sendMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//用户重命名----消息格式：rename|sxr
		newName := msg[7:]

		_, ok := this.server.OnlineMap[newName]

		if ok {
			this.SendOwnMsg("this name already exit! please change another one \n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.Name = newName
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.SendOwnMsg("you already change your name: " + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//私聊---消息格式: to|sxr|hi

		remoteName := strings.Split(msg, "|")[1]
		content := strings.Split(msg, "|")[2]

		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendOwnMsg("this user is not exit! \n")
			return
		}

		if content != "" {
			remoteUser.SendOwnMsg(this.Name + ": " + content)
		} else {
			this.SendOwnMsg("msg is empty, please send again \n")
			return
		}

	} else {
		this.server.BroadCast(this, msg)
	}
}
