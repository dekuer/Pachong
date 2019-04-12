package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	rePhone = `1[358]\d{9}`
	// reEmail = `[1-9]\d{4,}@qq.com`
	reEmail = `\w+@\w+\.[a-z]{1,3}`
	reLink  = `https?://www\.\w+\.[a-z]{2,5}`
	// 350204 1988 04 10 2014
	reID = `([1-6][1-7]\d{4})((19[1-9][1-9])|(201[0-9]))((0[1-9])|(1[0-2]))((0[1-9])|(1[0-9])|(2[0-9])|(3[0-1]))(\d{4})`
	// <img src="http://img.dwstatic.com/www/1808/397914339429/1553823856007.jpg"  alt="一位怕狗的外卖小哥的求救。。">
	// reImage = `<img.+?src="(http.+?)"`
	reImage = `<img.+?src="(http.+?\.((jpg)|(png)|(jpeg)|(gif)))(.*)?"(.+)?alt="(.*)?"` //根据地址查找图片
	// img标签中的alt属性
	reAlt = `alt="(.+?)"`
	// reImageName = `http://(.+)?\.((jpg)|(png)|(jpeg)|(gif))` //根据原本的文件名查找图片
	reImageName = `\w+?\.((jpg)|(png)|(jpeg)|(gif))`

	chSem      = make(chan int, 10)
	chSemAsync = make(chan string, 10)
	downloadWG sync.WaitGroup //等待组
	randomMT   sync.Mutex     //互斥锁
)

func HandleErr(err error, when string) {
	if err != nil {
		fmt.Println(when, err)
		os.Exit(1)
	}
}

// 去重
func RemoveRepet(arr [][]string) [][]string {
	newarr := make([][]string, 0) //新建一个二维切片

	for i := range arr {
		flag := true
		for j := range newarr {
			if newarr[j][0] == arr[i][0] {
				flag = false
				break
			}
		}
		if flag {
			newarr = append(newarr, arr[i])
		}
	}
	// fmt.Println(newarr)
	return newarr
}

//爬取手机号
func Phone() {
	resp, err := http.Get("http://www.tiantianxieye.com")
	HandleErr(err, `http://www.tiantianxieye.com`)
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// fmt.Println(html)
	re := regexp.MustCompile(rePhone)               //得到正则表达式的对象
	allString := re.FindAllStringSubmatch(html, -1) //查找结果 -1匹配全部
	for _, v := range allString {
		fmt.Println(v[0])
	}
}

// 爬取邮箱
func Email() {
	resp, err := http.Get("http://tieba.baidu.com/p/1634694280")
	HandleErr(err, `http://tieba.baidu.com/p/1634694280`)
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// fmt.Println(html)
	re := regexp.MustCompile(reEmail)
	allString := re.FindAllStringSubmatch(html, -1)
	for _, v := range allString {
		fmt.Println(v[0])
	}
}

// 爬取链接
func Link() {
	resp, err := http.Get("https://www.hao123.com/")
	HandleErr(err, `https://www.hao123.com/`)
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// println(html)
	re := regexp.MustCompile(reLink)
	allString := re.FindAllStringSubmatch(html, -1)
	removeString := RemoveRepet(allString)
	// fmt.Println(removeString)
	for _, v := range removeString {
		fmt.Println(v[0])
	}
}

func ID() {
	resp, err := http.Get("http://baijiahao.baidu.com/s?id=1601965978904022957&wfr=spider&for=pc")
	HandleErr(err, `http://baijiahao.baidu.com/s?id=1601965978904022957&wfr=spider&for=pc`)
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// fmt.Print(html)
	re := regexp.MustCompile(reID)
	allString := re.FindAllStringSubmatch(html, -1)
	for _, v := range allString {
		fmt.Println(v[0])
	}
}

func Image() []string {
	resp, _ := http.Get("https://www.mzitu.com/page/2/")
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// fmt.Println(html)
	imgUrl := make([]string, 0)
	re := regexp.MustCompile(reImage)
	allString := re.FindAllStringSubmatch(html, -1)
	for _, v := range allString {
		imgUrl = append(imgUrl, v[1]) //url地址
		// fmt.Println(v[0], "\n", GetImgNameFromTag(v[0])) //v[0]为整串的img标签 v[1]为http地址 v[3]为alt
	}
	// fmt.Println(imgTag)
	return imgUrl
}

// 同步下载图片 文件名字用随机数
func DownLoadImage(imageUrl []string) {
	for _, v := range imageUrl {
		resp, err := http.Get(v)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		Imgbytes, _ := ioutil.ReadAll(resp.Body)
		filename := `/usr/local/gopath/src/Pachong/image/` + GetRandomName() + `.jpg`
		err = ioutil.WriteFile(filename, Imgbytes, 777) //写入文件
		if err != nil {
			fmt.Println(filename, "下载失败")
		} else {
			fmt.Println(filename, "下载成功")
		}
	}
}

// 异步下载图片 用随机数
func DownLoadImageAsync(imageUrl []string) {
	downloadWG.Add(1) //设置计数器为1
	// fmt.Println("loading…")
	go func() {
		chSem <- 123
		DownLoadImage(imageUrl)
		<-chSem
		downloadWG.Done() //每次-1
	}()
	downloadWG.Wait() //在计数器为0前一直等待
}

//生成[start,end)的 随机数
func GetRandomInt(start, end int) int {
	randomMT.Lock()                   //加锁
	defer randomMT.Unlock()           //解锁
	<-time.After(1 * time.Nanosecond) //阻塞1纳秒
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := start + r.Intn(end-start)
	return ret
}

// 生成时间戳+随机数的文件名
func GetRandomName() string {
	timestamp := strconv.Itoa(int(time.Now().UnixNano())) //转换为字符串
	randomNum := strconv.Itoa(GetRandomInt(100, 1000))    //100-999之间的随机数
	return timestamp + randomNum
}

// 传进去img标签里面提取出alt        有alt使用alt作为文件名 没有的话使用时间戳+随机数作为文件名
func GetImgNameFromTag(imgTag string) string {
	re := regexp.MustCompile(reAlt)
	rets := re.FindAllStringSubmatch(imgTag, 1)
	if len(rets) > 0 {
		return rets[0][1]
	} else {
		return GetRandomName()
	}

}

// image文件名用alt
func ImageForAlt() []string {
	resp, _ := http.Get("https://www.7160.com/meinvmingxing/list_5_1.html")
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	// fmt.Println(html)
	imgInfo := make([]string, 0)
	re := regexp.MustCompile(reImage)
	allString := re.FindAllStringSubmatch(html, -1)
	for _, v := range allString {
		imgInfo = append(imgInfo, v[0]) //url地址
		// fmt.Println(v[0], "\n", GetImgNameFromTag(v[0])) //v[0]为整串的img标签 v[1]为http地址 v[3]为alt
	}
	// fmt.Println(imgInfo)
	return imgInfo
}

//同步下载 文件名字用alt
func DownLoadImageForAlt(imageInfo []string) {
	for _, v := range imageInfo {

		re := regexp.MustCompile(reImage)
		allString := re.FindAllStringSubmatch(v, -1)

		resp, err := http.Get(allString[0][1])
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		Imgbytes, _ := ioutil.ReadAll(resp.Body)
		filename := `/usr/local/gopath/src/Pachong/image/` + GetImgNameFromTag(allString[0][0]) + allString[0][4]
		err = ioutil.WriteFile(filename, Imgbytes, 777) //写入文件
		if err != nil {
			fmt.Println(filename, "下载失败")
		} else {
			fmt.Println(filename, "下载成功")
		}
	}
}

//同步下载 文件名字用链接中的文件名
func DownLoadImageForLink(imageInfo []string) {
	for _, v := range imageInfo {
		re := regexp.MustCompile(reImage)
		allString := re.FindAllStringSubmatch(v, -1)

		reTwo := regexp.MustCompile(reImageName)
		linkFileName := reTwo.FindAllStringSubmatch(allString[0][1], -1)

		resp, err := http.Get(allString[0][1])
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		Imgbytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(allString)
		// fmt.Println(linkFileName[0][1])
		// filename := `/usr/local/gopath/src/Pachong/image/` + GetImgNameFromTag(allString[0][0]) + `.jpg`
		filename := `/usr/local/gopath/src/Pachong/image/` + linkFileName[0][0]
		err = ioutil.WriteFile(filename, Imgbytes, 777) //写入文件
		if err != nil {
			fmt.Println(filename, "下载失败")
		} else {
			fmt.Println(filename, "下载成功")
		}
	}
}

//抓取有分页的图片
func PageUrl() {
	var baseUrl string = `http://www.52doutuwang.com/page/`
	var Url string
	for i := 1; i < 20; i++ {
		Url = baseUrl + strconv.Itoa(i) + "?owbijs=oqtev1"
		ImagePaging(Url)
	}
}

func ImagePaging(url string) {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	imgUrl := make([]string, 0)
	re := regexp.MustCompile(reImage)
	allString := re.FindAllStringSubmatch(html, -1)
	for _, v := range allString {
		imgUrl = append(imgUrl, v[1]) //url地址
	}
	fmt.Println("download…")
	// DownLoadImage(imgUrl)
	DownLoadImageAsync(imgUrl)
}

//异步抓取有分页的图片
func PageUrlAsync() {
	var baseUrl string = `http://www.52doutuwang.com/page/`
	var Url string
	for i := 1; i < 11; i++ {
		Url = baseUrl + strconv.Itoa(i) + "?owbijs=oqtev1"
		ImagePagingAsync(Url)
	}
}

func ImagePagingAsync(url string) {
	chSemAsync <- url
	go func() {
		urlout := <-chSemAsync
		resp, _ := http.Get(urlout)
		defer resp.Body.Close()
		bytes, _ := ioutil.ReadAll(resp.Body)
		html := string(bytes)
		imgUrl := make([]string, 0)
		re := regexp.MustCompile(reImage)
		allString := re.FindAllStringSubmatch(html, -1)
		for _, v := range allString {
			imgUrl = append(imgUrl, v[1]) //url地址
		}
		fmt.Println("download…")
		DownLoadImage(imgUrl)
	}()
	// DownLoadImageAsync(imgUrl)

}

func main() {
	// Phone()
	// Email()
	// Link()
	// ID()f

	/*单纯的爬虫*/
	// imageUrl := Image()
	// DownLoadImage(imageUrl)
	// DownLoadImageAsync(imageUrl)

	/*用alt作为名字的*/
	// imageInfo := ImageForAlt()//alt作为
	// DownLoadImageForAlt(imageInfo)
	// DownLoadImageForLink(imageInfo)

	/*有分页的图片*/
	// PageUrl()

	/*开启异步的*/
	PageUrlAsync()
	time.Sleep(time.Second * 10)
}
