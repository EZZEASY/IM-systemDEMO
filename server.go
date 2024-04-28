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
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
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

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		// 将msg发送给所有在线的用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + "：" + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("链接建立成功")
	user := NewUser(conn, this)
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端佛送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n])
			user.DoMessage(msg)
			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户活跃，应该重置定时器
			// 不执行任何事情，为了激活select，使下方的case条件被执行
		case <-time.After(time.Second * 300): //执行时时间自动重置
			//已经超时，强制下线
			user.SendMsg("超时下线")
			close(user.C)
			conn.Close()
			return // 或  runtime.Goexit()
		}
	}
}

func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("ner.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 监听msg广播消息的逻辑
	go this.ListenMessage()

	for {
		// accept msg
		conn, err := listener.Accept()
		// do handler
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go this.Handler(conn)
	}
}
