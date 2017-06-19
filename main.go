package main

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"sync"

	tojson "github.com/chentknba/pack/json"
)

var wg sync.WaitGroup

var cfg = map[string]string{}
var dict_cfg = map[string]string{}

var execl_path string
var desc_json string
var svr_dict_path string
var cli_dict_path string
var svr_meta_path string
var cli_meta_path string


func loadConf() error {
	bytes, err := ioutil.ReadFile("conf.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	execl_path = cfg["execl_path"]
	desc_json = execl_path + "desc.json"

	bytes, err = ioutil.ReadFile(desc_json)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &dict_cfg); err != nil {
		return nil
	}

	svr_dict_path = cfg["svr_dict_path"]
	cli_dict_path = cfg["cli_dict_path"]

    svr_meta_path = cfg["svr_meta_path"]
    cli_meta_path = cfg["cli_meta_path"]

	return nil
}

// 生成一般配置
func genDict(dict_name, execl_name string) {
	defer wg.Done()

	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		return
	}

	sout := svr_dict_path + dict_name + ".json"
	cout := cli_dict_path + dict_name + ".json"

	sf, err := os.OpenFile(sout, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open sout file fail, err: %v\n", err)
		return
	}

	defer sf.Close()

	cf, err := os.OpenFile(cout, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("open cout file fail, err: %v\n", err)
		return
	}

	defer cf.Close()

	scaner := tojson.NewScaner(xlfile, dict_name)

	svrdict, clidict, sm, cm := scaner.GetOutput()

	sf.WriteString(svrdict)
	cf.WriteString(clidict)

    fmt.Println(sm)
    fmt.Println(cm)
}

// 生成配置dict_define
func genMetaLua() {
    defer wg.Done()
}

// 生成错误码lua
func genErrnoLua() {
}

// 生成log事件定义lua
func genEventLua() {
}

func main() {
	if err := loadConf(); err != nil {
		fmt.Printf("load conf err: %v\n", err)
		return
	}


	for dict_name, execl_name := range dict_cfg {
		wg.Add(1)
		go genDict(dict_name, execl_name)
	}

    go func() {
        wg.Add(1)
        genMetaLua()
    }()

	wg.Wait()

	fmt.Println("Done.")
}
