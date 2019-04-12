package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

/****************************************************************************************************
从开源接口showapi.com中获取成语数据 存进本地的json文本中
（输入一个关键字(keyword)和个数(rows) 查找该字的所有成语的信息）
****************************************************************************************************/
type Idiom struct {
	Title      string //成语名称
	Spell      string //拼写
	Content    string //解释
	Derivation string //典故
	Samples    string //例子
}

var (
	// Url           = `http://route.showapi.com/1196-1?showapi_appid=91194&showapi_sign=9f081d50fd6b40fba69b24c99cc5f7ac&keyword=肉&page=1&rows=20`
	Path          = `/usr/local/gopath/src/ChengYuApp/IdiomData/idiom.json`
	idiomsMap     map[string]Idiom
	stringBuilder bytes.Buffer //声明字符串缓冲
)

//从网上获取返回的json数据
func GetJson(url string) (jsonStr string) {
	// 获得网络数据
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http请求失败:", err)
	}
	//延时关闭IO资源
	defer resp.Body.Close()

	//resp.Body实现了Reader接口 对其进行数据读入
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取网络数据失败:", err)
	}

	// 将网络数据转化为string
	jsonStr = string(bytes)
	// fmt.Println(jsonStr)
	return jsonStr
}

//去除无用的信息 只保留成语名字
func IdiomName(jsonStr string) {
	tempMap := make(map[string]interface{})          //用来获得成语名字
	tempMap2 := make(map[string]interface{})         //用来获得成语信息
	var idiom = Idiom{}                              //结构体对象 用来放成语的信息
	idiomsMap = make(map[string]Idiom)               //成语名字与成语的所有信息
	err := json.Unmarshal([]byte(jsonStr), &tempMap) //将json解码到map类型
	if err != nil {
		fmt.Println(err)
	}
	dataSlice := tempMap["showapi_res_body"].(map[string]interface{})["data"].([]interface{}) //得到的是所有的成语名字的切片

	//读取原本文件里的json数据
	bytes, err := ioutil.ReadFile(Path)
	if err != nil {
		fmt.Println(err)
		return
	}
	// jsonSrt := string(bytes)
	// fmt.Println(jsonSrt)
	idiomsMapBase := make(map[string]Idiom)
	err = json.Unmarshal(bytes, &idiomsMapBase)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(idiomsMapBase)

	for _, v := range dataSlice {
		go func() {
			title := v.(map[string]interface{})["title"].(string)                                                                                      //此时得到的是成语名字
			idiomInfo := GetJson(`http://route.showapi.com/1196-2?showapi_appid=91194&showapi_sign=9f081d50fd6b40fba69b24c99cc5f7ac&keyword=` + title) //返回的是相应成语的信息
			err = json.Unmarshal([]byte(idiomInfo), &tempMap2)
			if err != nil {
				fmt.Println(err)
			}
			dataSlice2 := tempMap2["showapi_res_body"].(map[string]interface{})["data"].(map[string]interface{}) //成语的信息
			idiom.Title = dataSlice2["title"].(string)                                                           //将数据存进结构体
			idiom.Spell = dataSlice2["spell"].(string)
			idiom.Samples = dataSlice2["samples"].(string)
			idiom.Derivation = dataSlice2["derivation"].(string)
			idiom.Content = dataSlice2["content"].(string)
			idiomsMapBase[title] = idiom //成语名字对应的是所有的信息
		}()
		time.Sleep(1 * time.Second)
	}
	//运行完则将所有的成语信息传入idiomsMap里面了
	// fmt.Println(idiomsMapBase)
	WriteIdiomInfo(idiomsMapBase)
}

// 数据的本地持久化 将数据存进本地
func WriteIdiomInfo(idiomsMap map[string]Idiom) {
	dstFile, err := os.OpenFile(Path, os.O_CREATE|os.O_WRONLY, 0666) //新建打开路径下的文件,O_WRONLY 写 O_CREATE新建文件
	if err != nil {
		fmt.Println("打开文件失败：", err)
		return
	}
	defer dstFile.Close()

	encoder := json.NewEncoder(dstFile) //新建一个json的encoder
	err = encoder.Encode(idiomsMap)     //数据结构idiomsMap以json格式写入文件
	if err != nil {
		fmt.Println("下载json数据失败：", err)
		return
	} else {
		fmt.Println("下载完毕!")
	}

}

func main() {
	// 获得命令行参数 downIdiom.exe  -keyword 肉
	keywordInfo := [3]interface{}{"keyword", "未知关键字", "成语的关键字"}
	rowsInfo := [3]interface{}{"rows", "1", "成语的个数"}
	retValuesMap := GetCmdlineArgs(keywordInfo, rowsInfo)
	fmt.Println("下载数据中...")
	keyword := retValuesMap["keyword"].(string)
	rows := retValuesMap["rows"].(string)
	// fmt.Println(retValuesMap)
	baseUrl := "http://route.showapi.com/1196-1?showapi_appid=91194&showapi_sign=9f081d50fd6b40fba69b24c99cc5f7ac&keyword="
	stringBuilder.WriteString(baseUrl)
	stringBuilder.WriteString(keyword)
	stringBuilder.WriteString("&page=1&rows=")
	stringBuilder.WriteString(rows)
	// fmt.Println(stringBuilder.String())
	url := stringBuilder.String()
	jsonStr := GetJson(url) //返回json格式的数据
	IdiomName(jsonStr)      //存进数据

}
