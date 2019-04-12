/*
 * @Author: YeSiFan
 * @Date: 2019-04-09 22:19:35
 * @Last Modified by: YeSiFan
 * @Last Modified time: 2019-04-11 22:32:26
 */
/*
命令行可以输入昵称进行登录
all#消息                         向全体发送消息
昵称#消息                      私聊
exit                                退出
logs                                打印聊天记录
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

var (
	chanQuit         = make(chan bool, 0)
	downloadFileName string
	downloadFilePath string
)

func CHandleError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 处理发送数据
func HandleSend(conn net.Conn, name string) {
	// 发送昵称到服务端
	_, err := conn.Write([]byte(name))
	CHandleError(err)
	// 建立一个标准输入
	reader := bufio.NewReader(os.Stdin)
	for {
		// 读取每行数据
		lineBytes, _, err := reader.ReadLine()
		CHandleError(err)
		lineStr := string(lineBytes)
		// 上传文件
		// upload#文件名#要上传的文件的路径
		if strings.Index(lineStr, "upload") == 0 {
			strs := strings.Split(lineStr, "#")
			if len(strs) != 3 {
				fmt.Println("上传的格式有误")
				continue
			} else {
				fileName := strs[1] //读入文件名字
				filePath := strs[2] //拿到文件路径
				// 构造数据包
				dataPack := make([]byte, 0)
				//写入数据包头部（upoad#文件名）
				header := make([]byte, 100)
				copy(header, []byte("upload#"+fileName+"#"))
				dataPack = append(dataPack, header...)
				//写入数据包body(文件字节)
				bytes, err := ioutil.ReadFile(filePath)
				CHandleError(err)
				dataPack = append(dataPack, bytes...)
				// 写给服务端
				_, err = conn.Write(dataPack)
				CHandleError(err)
			}
		} else if strings.Index(lineStr, "download") == 0 {
			// 下载文件
			// download#要下载的文件名#要下载的文件的存放路径
			strs := strings.Split(lineStr, "#")
			if len(strs) != 3 {
				fmt.Println("下载的格式有误")
				continue
			} else {
				//读取文件名
				strs := strings.Split(lineStr, "#")
				downloadFileName = strs[1]
				// 读取要下载的地址
				downloadFilePath = strs[2]
				// 将数据传给服务端
				downloadInfo := []byte("download#" + downloadFileName + "#" + downloadFilePath)
				_, err = conn.Write(downloadInfo)
				// 客户端接收并存储

			}
		} else {
			// 如果输入exit则退出
			if lineStr == "exit" {
				_, err = conn.Write(lineBytes)
				CHandleError(err)
				os.Exit(0)
			} else {
				// 发送到服务端
				_, err = conn.Write(lineBytes)
				CHandleError(err)
			}
		}
	}
}

// 处理接收数据
func HandleReceive(conn net.Conn) {
	buffer := make([]byte, 1024) //缓冲区大小 读取数据
	for {
		n, err := conn.Read(buffer)
		if err != io.EOF {
			CHandleError(err)
		}
		//有东西时才显示
		if n > 0 {
			msg := string(buffer[:n])
			msgBytes := buffer[:n]
			fileNameLength := len(downloadFileName)
			filePathLength := len(downloadFilePath)
			filename := string(msgBytes[:fileNameLength])        //前50个字节为要下载的文件名
			pathname := string(msgBytes[50 : 50+filePathLength]) //50-100个字节为下载文件的保存地址的
			downloadBytes := msgBytes[100:]                      //后面的字节为下载文件
			fmt.Println(filename)
			fmt.Println(downloadFileName)
			fmt.Println(pathname)
			fmt.Println(downloadFilePath)
			if filename == downloadFileName {
				// 如果是服务端传递过来的下载的文件字节
				err := ioutil.WriteFile(downloadFilePath, downloadBytes, 0666)
				CHandleError(err)
				fmt.Println("下载成功！")
			} else {
				fmt.Println(msg)
			}

		}

	}

}

func main() {
	// 命令行参数携带昵称
	nameInfo := [3]interface{}{"name", "未命名", "昵称"}
	retValuesMap := GetCmdlineArgs(nameInfo)
	name := retValuesMap["name"].(string)

	// 拨号连接dial 获得connection
	conn, err := net.Dial("tcp", "127.0.0.1:8888")

	CHandleError(err)
	defer conn.Close()

	// 在一条独立的协程中输入 并发送消息
	go HandleSend(conn, name)
	// 在一条独立的协程中接收服务端消息
	go HandleReceive(conn)

	// 设堵塞 避免主协程结束导致子协程结束
	<-chanQuit
}
