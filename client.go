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
	flag       int // 当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}

	client.conn = conn
	return client
}

func (client *Client) DealRespinse() {
	io.Copy(os.Stdout, client.conn) // 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1、共聊模式")
	fmt.Println("2、私聊模式")
	fmt.Println("3、更新用户名")
	fmt.Println("-1、推出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围的数字<<<<")
		return false
	}
}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println("请输入聊天内容，exit退出...")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			_, err := client.conn.Write([]byte(chatMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) SelectUser() {
	sendMsg := "who"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>>>请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			chatMsg = "@" + remoteName + " " + chatMsg
			_, err := client.conn.Write([]byte(chatMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != -1 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP(默认为127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认为8888)")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}

	fmt.Println(">>>>> 链接服务器成功...")

	go client.DealRespinse()
	// 保持链接
	client.Run()
	fmt.Println(">>>>> 退出服务器完成...")
}
