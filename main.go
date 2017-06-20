package main

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"sync"

	tojson "github.com/chentknba/pack/json"
	tolua  "github.com/chentknba/pack/lua"
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

const (
	DICT_META_FILE = "dict_metas_auto.lua"
	ERRNO_FILE     = "errno_auto.lua"
	AUTO_FILE_DESC = "-- auto generated, modification is not permitted."
)

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

func save(file, content string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
	}()

	f.WriteString(content)

	return err
}

// genDict 生成一般配置
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

	scaner := tojson.NewScaner(xlfile, dict_name)

	svrdict, clidict := scaner.GetOutput()

	if err := save(sout, svrdict); err != nil {
		fmt.Printf("write svr dict fail, err: %v\n", err)
	}

	if err := save(cout, clidict); err != nil {
		fmt.Printf("write cli dict fail, err: %v\n", err)
	}
}

// genMeta 生成配置的dict_define
func genMeta(dict_name, execl_name string) <-chan string {
	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		os.Exit(-1)
	}

	ch := make(chan string)
	gen := tolua.NewMetagen(xlfile, dict_name)

	wg.Add(1)

	go func (){
		defer wg.Done()

		meta := gen.Done()

		ch <- meta
	}()

	return ch
}

// genErrno 生成各错误码
func genErrno(lua_name, execl_name string) {
	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		os.Exit(-1)
	}

	gen := tolua.NewErrnogen(xlfile, lua_name)

	str := gen.Done()

	var errno string

	errno += AUTO_FILE_DESC
	errno += "\r\n"
	errno += "\r\n"
	errno += "local errno = "
	errno += "\r\n"
	errno += "{"
	errno += "\r\n"

	errno += str
	errno += "\r\n"

	errno += "}"
	errno += "\r\n"
	errno += "\r\n"
	errno += "return errno"

	// sout := svr_meta_path + ERRNO_FILE
	// cout := cli_meta_path + ERRNO_FILE

	// if err := save(sout, errno); err != nil {
	// 	fmt.Printf("write svr errno fail, err: %v\n", err)
	// }

	// if err := save(cout, errno); err != nil {
	// 	fmt.Printf("write cli errno fail, err: %v\n", err)
	// }
}

// cmd_gendict 生成各配置文件对应的json
func cmd_gendict() {
	for dict_name, execl_name := range dict_cfg {
		wg.Add(1)

		go genDict(dict_name, execl_name)
	}

	wg.Wait()
}

// cmd_genmeta 生成 dict_define.lua
func cmd_genmeta() {
	var meta string
	meta += AUTO_FILE_DESC
	meta += "\r\n"
	meta += "\r\n"
	meta += "local DICT_DEFINE = "
	meta += "\r\n"
	meta += "{"

	for dict_name, execl_name := range dict_cfg {
		ch := genMeta(dict_name, execl_name)

		meta += "\r\n"
		meta += <-ch
		meta += ","
		meta += "\r\n"
	}

	wg.Wait()

	meta += "\r\n"
	meta += "}"
	meta += "\r\n"
	meta += "\r\n"
	meta += "return DICT_DEFINE"

	sout := svr_meta_path + DICT_META_FILE
	cout := cli_meta_path + DICT_META_FILE

	if err := save(sout, meta); err != nil {
		fmt.Printf("write svr meta fail, err: %v\n", err)
	}

	if err := save(cout, meta); err != nil {
		fmt.Printf("write cli meta fail, err: %v\n", err)
	}
}

// cmd_generrno 生成 errno.lua
func cmd_generrno() {
}

func main() {
	if err := loadConf(); err != nil {
		fmt.Printf("load conf err: %v\n", err)
		return
	}

	cmd_gendict()

	cmd_genmeta()

	cmd_generrno()

	fmt.Println("Done.")
}
