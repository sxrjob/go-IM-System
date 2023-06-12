package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		for _, user := range server.OnlineMap {
			user.C <- msg
		}
	}
}
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//fmt.Println("conn sucess")
	user := NewUser(conn, server)

	//用户上线
	user.Online()

	//处理user输入的消息，进行广播
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.DeadLine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("read err:", err)
				return
			}
			msg := string(buf[:n-1])

			user.DoMsg(msg)
		}
	}()

	for {
		select {}
	}

}

func (server *Server) Start() {

	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("listen err:", err)
		return
	}

	//close
	defer listen.Close()

	go server.ListenMessage()
	//accept
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.accept err:", err)
			continue
		}
		go server.Handler(conn)
	}

}
