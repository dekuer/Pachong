/*
 * @Author: YeSiFan
 * @Date: 2019-04-07 10:42:52
 * @Last Modified by: YeSiFan
 * @Last Modified time: 2019-04-11 22:31:41
 */
/*单聊 多聊 上线通知 昵称*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

var (
	// 可以用地址或者昵称为键
	// allClientsMap = make(map[string]net.Conn)
	selfName      string
	allClientsMap = make(map[string]*Client)
	//所有群
	allGroupsMap map[string]*Group
	basePath     = `/usr/local/gopath/src/ChatRoom3.0/uploads/`
)

func init() {
	allGroupsMap = make(map[string]*Group)
	allGroupsMap["实例群"] = NewGroup("实例群", &Client{name: "管理员"})
}

type Client struct {
	conn net.Conn
	name string
	addr string
}

// 处理错误
func SHandleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ioWithConn(client *Client, selfName string) {
	buffer := make([]byte, 1024) //缓存区
	// clientAddr := client.conn.RemoteAddr().String() //获取客户端地址

	// 不断接收客户端消息
	for {
		n, err := client.conn.Read(buffer) //读取缓冲区数据
		if err != io.EOF {
			SHandleError(err)
		}

		// 如果发送的消息不为空
		if n > 0 {
			msgBytes := buffer[:n]

			if bytes.Index(msgBytes, []byte("upload")) == 0 {
				/*处理文件上传*/
				// 拿到数据包(文件名)
				// upload#filename#
				msgStr := string(msgBytes[:100])
				fileName := strings.Split(msgStr, "#")[1]

				// 拿到数据包身体(文件字节)
				fileBytes := msgBytes[100:]

				// 将文件字节写入指定位置
				err := ioutil.WriteFile(basePath+fileName, fileBytes, 0666)
				SHandleError(err)
				SendMsg2Client("文件上传成功", client)
			} else if bytes.Index(msgBytes, []byte("download")) == 0 {
				/*处理文件下载*/
				// 拿到数据
				// download#要下载的文件名#文件要下载存放的地址
				// download#madefake.txt#/usr/local/gopath/src/ChatRoom3.0/downloads/mabi.txt
				msgStr := string(msgBytes)
				fileName := strings.Split(msgStr, "#")[1]
				filePath := strings.Split(msgStr, "#")[2]
				//服务端向客户端写出对应文件的字节，客户端接收并存储
				bytes, err := ioutil.ReadFile(basePath + fileName)
				SHandleError(err)
				downloadBytes := make([]byte, 0)
				header := make([]byte, 50)
				body := make([]byte, 50)
				copy(header, []byte(fileName))
				copy(body, []byte(filePath))
				downloadBytes = append(downloadBytes, header...)
				downloadBytes = append(downloadBytes, body...)
				downloadBytes = append(downloadBytes, bytes...)
				client.conn.Write(downloadBytes) //最终发送的是 要下载的文件名+文件保存地址+文件字节流
			} else {
				/*处理字符消息*/
				msg := string(buffer[:n])
				// privatechat私聊则不显示
				privatechat := "#"

				// 如果不是私聊则显示客户端姓名和发送过来的内容
				if strings.Contains(msg, privatechat) == false && msg != "exit" && msg != "logs" {
					fmt.Println(client.name, "：", msg)
					// 将客户端说的每一句话都记录在以他的名字命名的文件里面
					WritMsgToFile(msg, client)
					SendInfo(msg, selfName, client)
				} else if strings.Contains(msg, privatechat) == true && msg != "exit" && msg != "logs" {
					// 如果是私聊的话
					WritMsgToFile(msg, client)
					SendInfo(msg, selfName, client)
				} else if msg == "exit" {
					// 如果客户端退出 则从所有的客户端map里面删除
					fmt.Println(client.name + "离开聊天室")
					delete(allClientsMap, client.name)
				} else if msg == "logs" {
					logsInfoByte, err := ioutil.ReadFile("/usr/local/gopath/src/ChatRoom2.0/logs/" + client.name + ".log")
					SHandleError(err)

					SendLogsInfo(client, logsInfoByte)
					// fmt.Println(logsInfo)
				}
			}

		}
	}

}

// 发送聊天历史
func SendLogsInfo(client *Client, logsInfo []byte) {
	_, err := client.conn.Write(logsInfo)
	SHandleError(err)
}

//// 向客户端发送消息
func SendInfo(msg, selfName string, client *Client) {

	strs := strings.Split(msg, "#") //通过#分隔
	if len(strs) > 1 {
		header := strs[0]
		body := strs[1]
		// 使用昵称定位客户端connection

		switch header {
		case "all":
			// 群发
			for nickname, client := range allClientsMap {
				// fmt.Println(nickname, selfName)
				if selfName != nickname {

					client.conn.Write([]byte(selfName + "发来消息：" + body))
				}
			}
		case "group_setup":
			//建群#群昵称
			if _, ok := allGroupsMap[body]; !ok {
				// 要创建的群不存在就建群
				newGroup := NewGroup(body, client)
				//将新群添加到所有群集合
				allGroupsMap[body] = newGroup
				// 通知群主建群成功
				SendMsg2Client("建群成功", client)
			} else {
				// 要创建的群存在
				SendMsg2Client("要创建的群已经存在", client)
			}
		case "group_info":
			if body == "all" {
				// 查看所有群信息
				info := ""
				for _, group := range allGroupsMap {
					info += group.String() + "\n"
				}
				SendMsg2Client(info, client)
			} else {
				// 查看单个群消息
				if group, ok := allGroupsMap[body]; ok {
					SendMsg2Client(group.String(), client)
				} else {
					SendMsg2Client("查无此群", client)
				}
			}
		case "group_join":
			// 申请加群
			// 如果群不存在
			group, ok := allGroupsMap[body]
			if !ok {
				SendMsg2Client("查无此群", client)
			} else {
				// 向群主申请加群
				SendMsg2Client(client.name+"申请加入群"+body+"是否同意?", group.Owner)
				SendMsg2Client("申请已发送,请等待群主审核", client)
			}
		case "group_joinreply":
			// 加群申请的回复
			// group_joinreply#no@zhangsan@东方艺术殿堂
			strs := strings.Split(body, "@")
			answer := strs[0]        //yes or no
			applicantName := strs[1] //申请人
			groupName := strs[2]     //申请的群昵称

			group, ok1 := allGroupsMap[groupName]       //判断是否合法的群昵称
			toWhom, ok2 := allClientsMap[applicantName] //判断是否合法的申请人
			//自动执行加群申请
			if ok1 && ok2 {
				NewGroupJoinReply(client, toWhom, group, answer).AutoRun()
			}

		default:
			//单独发
			for nickname, client := range allClientsMap {
				// 如果目标地址和客户端地址一样的话就发送
				if nickname == header {
					client.conn.Write([]byte(selfName + "发来消息：" + body))
					break
				}
			}
		}

	}
}

// 发送消息给客户端
func SendMsg2Client(msg string, client *Client) {
	client.conn.Write([]byte(msg))
}

//保存聊天记录 文件以用户名命名
func WritMsgToFile(msg string, client *Client) {
	// 打开文件
	// fmt.Println(client.name)
	file, err := os.OpenFile("/usr/local/gopath/src/ChatRoom2.0/logs/"+client.name+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	SHandleError(err)
	defer file.Close()
	// 写入消息
	logMsg := fmt.Sprintln(time.Now().Format("2006-01-02 15:04:05"), msg)
	// fmt.Println(logmsg)
	file.Write([]byte(logMsg))
}

func main() {

	// 建立服务端监听
	listener, err := net.Listen("tcp", "127.0.0.1:8888")
	SHandleError(err)
	// 最后依次关闭所有连接
	defer func() {
		// clientsMap的类型是make(map[string]net.Conn)
		for _, client := range allClientsMap {
			client.conn.Write([]byte("服务器进行维护!")) //告诉每个客户端的信息
		}
		listener.Close() //关闭连接
	}()
	// 循环接入所有客户端
	for {
		conn, err := listener.Accept()
		SHandleError(err)
		clientAddr := conn.RemoteAddr()

		// 给聊天室发送上线

		// 接收保存昵称
		var clientName string
		buffer := make([]byte, 1024)
		for {
			// 读取客户端昵称
			n, err := conn.Read(buffer)
			SHandleError(err)
			if n > 0 {
				clientName = string(buffer[:n])
				fmt.Println(clientName + "进入聊天室")
				selfName = clientName
				break
			}
		}

		//将每一个客户端地址和conn都丢入map
		// allClientsMap[clientAddr.String()] = conn

		client := Client{conn, clientName, clientAddr.String()}
		allClientsMap[clientName] = &client

		// 在单独的协程中与每个女朋友聊天
		go ioWithConn(&client, selfName)
	}
	//优雅退出
}
