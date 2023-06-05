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
	Message   chan string
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		Message:   make(chan string),
		OnlineMap: make(map[string]*User),
	}
	return server
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// do handler
func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("连接建立成功")

	// 新建用户
	user := NewUser(conn, this)

	//用户上线
	user.Online()

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 用户下线
				user.DeadLine()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Error conn read: ", err)
				return
			}

			// 获取用户消息（去除’\n‘回车）
			msg := string(buf[:n-1])

			//将获取的消息广播
			this.BroadCast(user, msg)
		}
	}()
	//阻塞handler
	select {}
}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen is err:", err)
		return
	}

	//close listen socket
	defer listen.Close()

	//启动监听Message的goroutine
	go this.ListenMessage()

	//accept
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Listen accept err:", err)
			continue
		}
		go this.Handler(conn)

	}

}
