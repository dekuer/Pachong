package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

/*******************************************************************************
从本地json文本中读取数据 然后根据命令行输入的参数(keyword)进行模糊查询或者精确查询
********************************************************************************/
var (
	Path = `/usr/local/gopath/src/ChengYuApp/IdiomData/idiom.json`
)

type Idiom struct {
	Title      string
	Spell      string
	Content    string
	Derivation string
	Samples    string
}

// 执行查询 传入查询模式和关键字
func SearchFor(model, keyword string) {
	var temp map[string]interface{}
	bytes, err := ioutil.ReadFile(Path)
	if err != nil {
		fmt.Println(err)
		return
	}
	// idiomJson := string(bytes)
	// fmt.Println(idiomJson)

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(temp)
	// fmt.Println(temp[keyword])

	idiom := Idiom{}

	if model == "ambiguous" {
		// 模糊查询
		fmt.Println()
		fmt.Println("模糊查询结果:")
		for title, idiomMap := range temp {
			if strings.Contains(title, keyword) {
				// 如果title中包含keyword 输出结果
				idiom.Title = idiomMap.(map[string]interface{})["Title"].(string)
				idiom.Spell = idiomMap.(map[string]interface{})["Spell"].(string)
				idiom.Content = idiomMap.(map[string]interface{})["Content"].(string)
				idiom.Derivation = idiomMap.(map[string]interface{})["Derivation"].(string)
				idiom.Samples = idiomMap.(map[string]interface{})["Samples"].(string)
				PrintIdiom(idiom)
				fmt.Println()
			}
		}
	} else if model == "accurate" {
		//精确查询
		fmt.Println()
		fmt.Println("精确查询结果:")
		for title, idiomMap := range temp {
			if keyword == title {
				// 如果关键字等于title的话
				idiom.Title = idiomMap.(map[string]interface{})["Title"].(string)
				idiom.Spell = idiomMap.(map[string]interface{})["Spell"].(string)
				idiom.Content = idiomMap.(map[string]interface{})["Content"].(string)
				idiom.Derivation = idiomMap.(map[string]interface{})["Derivation"].(string)
				idiom.Samples = idiomMap.(map[string]interface{})["Samples"].(string)
				PrintIdiom(idiom)
			}
		}
	} else {
		fmt.Println("非法输入")
		return
	}

}

//输出成语的信息
func PrintIdiom(idiom Idiom) {
	fmt.Println("成语名:", idiom.Title)
	fmt.Println("拼音:", idiom.Spell)
	fmt.Println("解释:", idiom.Content)
	fmt.Println("典故:", idiom.Derivation)
	fmt.Println("例句:", idiom.Samples)
}

func main() {
	modelInfo := [3]interface{}{"model", "未知命令", "ambiguous=模糊查询 accurate=精确查询"}
	keywordInfo := [3]interface{}{"keyword", "未知关键字", "查询的关键字"}
	retValueMap := GetCmdlineArgs(modelInfo, keywordInfo)
	model := retValueMap["model"].(string)
	keyword := retValueMap["keyword"].(string) //这个是关键字

	SearchFor(model, keyword) //打开本地文件 查找相应的成语信息
}
