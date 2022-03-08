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
	mode       int
}

func NewClient(ip string, port int) *Client {

	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		mode:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))

	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}

	client.conn = conn
	return client
}
func (c *Client) SelectOnlineUser() {
	_, err := c.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("conn.Write error", err)
		return
	}
}

func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	fmt.Println(">>>>>>>>Please input a username, exit to logout")
	c.SelectOnlineUser()
	_, err := fmt.Scanln(&remoteName)
	if err != nil {
		return
	}

	for remoteName != "exit" {
		fmt.Println(">>>>>>>>Please input chat message, exit to logout")
		_, err := fmt.Scanln(chatMsg)
		if err != nil {
			return
		}

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write error", err)
				}
			}
			// 继续发送第二条，重置消息
			chatMsg = ""
			fmt.Println(">>>>>>>>Please input chat message, exit to logout")
			_, err = fmt.Scanln(chatMsg)
			if err != nil {
				return
			}

		}
		// 继续和其他用户私聊
		fmt.Println(">>>>>>>>Please input a username, exit to logout")
		c.SelectOnlineUser()
		_, err = fmt.Scanln(&remoteName)
		if err != nil {
			return
		}

	}
}
func (c *Client) Run() {
	for c.mode != 0 {
		for c.menu() != true {

		}
		switch c.mode {
		case 1:
			c.PublicChat()
		case 2:
			c.PrivateChat()
		case 3:
			c.UpdateName()
		}
	}
}

func (c *Client) DealResponse() {
	// 一旦c.conn 有数据，就打印到标准输出，且永久阻塞监听
	_, err := io.Copy(os.Stdout, c.conn)
	if err != nil {
		return
	}

	// 等价于
	for {
		buf := make([]byte, 4096)
		read, err := c.conn.Read(buf)
		if err != nil {
			return
		}
		fmt.Println(read)
	}
}

func (c *Client) menu() bool {

	var mode int

	fmt.Println("1.Public mode")
	fmt.Println("2.Private mode")
	fmt.Println("3.Change username")
	fmt.Println("0.Exit")

	_, err := fmt.Scanln(&mode)
	if err != nil {
		return false
	}

	if mode >= 0 && mode < 4 {
		c.mode = mode
		return true
	} else {
		fmt.Println("Please input correct number")
		return false
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println("Please input username")
	_, err := fmt.Scanln(&c.Name)
	if err != nil {
		return false
	}
	sendMsg := "rename|" + c.Name + "\n"

	_, err = c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error", err)
		return false
	}

	return true

}

// PublicChat 公聊模式
func (c *Client) PublicChat() {
	var chatMsg string
	fmt.Println("Please input chat message, input exit to logout")

	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error", err)
				break
			}
		}
		chatMsg = ""
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}
	}

}

var ServerIp string
var ServerPort int

func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "Set server ip address")
	flag.IntVar(&ServerPort, "port", 8888, "Set server port")
}

func main() {
	flag.Parse()

	client := NewClient(ServerIp, ServerPort)

	if client == nil {
		fmt.Println(">>>>>>> Connection server failed....")
		return
	}

	// 单独开启一个goroutine, 处理server回执消息

	go client.DealResponse()
	fmt.Println(">>>>>>> Connection server successful....")

	client.Run()

}
