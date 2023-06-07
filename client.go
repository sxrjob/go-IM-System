package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client

}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器默认Ip地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器默认Port地址（8888）")
}

func (client *Client) menu() bool {
	var t int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&t)

	if t >= 0 && t <= 3 {
		client.flag = t
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
			break
		case 2:
			//私聊模式
			client.PrivateChat()
			break
		case 3:
			//更新用户名
			client.UpdateName()
			break
		}
	}
}

// 处理server回应的消息， 直接显示到标准输出即可
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上, 永久阻塞监听
	io.Copy(os.Stdout, client.conn)

}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>please input your new name: ")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.write.err:", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		sendMsg := chatMsg + "\n"
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.write.err:", err)
			break
		}
		chatMsg = ""
		//fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) getAllUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var chatMsg string
	var chatUser string
	// 输入聊天对象
	client.getAllUser()
	fmt.Println(">>>>请输入聊天对象[用户名]，exit退出:")
	fmt.Scanln(&chatUser)

	for chatUser != "exit" {
		//提示用户输入消息
		fmt.Println(">>>>请输入聊天内容，exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			sendMsg := "to|" + chatUser + "|" + chatMsg + "\n\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.write.err:", err)
				break
			}
			chatMsg = ""
			//fmt.Println(">>>>请输入聊天内容，exit退出.")
			fmt.Scanln(&chatMsg)
		}
		client.getAllUser()
		fmt.Println(">>>>请输入聊天对象[用户名]，exit退出:")
		fmt.Scanln(&chatUser)
	}
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println(">>>>>>>>>>>>>connect err....")
		return
	}

	//单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()

	fmt.Println(">>>>>>>>>>>>>connect complete....")

	//启动客户端业务
	client.Run()
}
