package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	tojson "github.com/chentknba/pack/json"
	tolua  "github.com/chentknba/pack/lua"
	"github.com/tealeg/xlsx"
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

// cmd param
var do_cmd_dict   *bool
var do_cmd_errno  *bool
var do_cmd_action *bool

const (
	DICT_META_FILE = "dict_metas_auto.lua"
	ERRNO_FILE     = "errno_auto.lua"
	AUTO_FILE_DESC = "-- auto generated, modification is not permitted."
)

const help_msg = `
pack is a tool for execl -> json/lua. 

Usge:
	pack -option 

The options are:
	dict 		将execl配置生成dict_xxx.json, 同时生成dict_define.lua
	errno 		将错误码execl生成lua文件
	action 		将行为日志execl生成lua文件`

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
func genMeta(dict_name, execl_name string) (<-chan string, error) {
	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open %v err: %v\n", execl_name, err))
	}

	ch := make(chan string, 1)
	gen := tolua.NewMetagen(xlfile, dict_name)

	wg.Add(1)

	go func (){
		defer wg.Done()

		meta := gen.Done()

		ch <- meta
	}()

	return ch, nil
}

// genErrno 生成各错误码
func genErrno(execl_name string) {
	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		os.Exit(-1)
	}

	gen := tolua.NewErrnogen(xlfile)
	result := gen.Done()

	for file, content := range result {
		save(file, content)
	}
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
		ch, err := genMeta(dict_name, execl_name)
		if err != nil {
			fmt.Printf("unexpected genMeta fail, err: %v\n", err)
			continue
		}

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
	execl_name := "errno"

	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		return
	}

	var errno string
	errno += AUTO_FILE_DESC
	errno += "\r\n"
	errno += "\r\n"
	errno += "local errno = "
	errno += "\r\n"
	errno += "{"
	errno += "\r\n"

	gen := tolua.NewErrnogen(xlfile)
	result := gen.Done()

	for file, content := range result {
		e := errno

		e += content
		e += "\r\n"
		e += "}"
		e += "\r\n"
		e += "\r\n"
		e += "return errno"

		sf := svr_meta_path + file + ".lua"
		cf := cli_meta_path + file + ".lua"

		save(sf, e)
		save(cf, e)
	}
}

// cmd_genaction
func cmd_genaction() {
	execl_name := "action_log"
	file := execl_path + execl_name + ".xlsx"

	xlfile, err := xlsx.OpenFile(file)
	if err != nil {
		fmt.Printf("open %v err: %v\n", execl_name, err)
		return
	}

	gen := tolua.NewActiongen(xlfile)
	result := gen.Done()

	for file, content := range result {
		sf := svr_meta_path + file + ".lua"
		cf := cli_meta_path + file + ".lua"

		save(sf, content)
		save(cf, content)
	}
}

func init() {
	do_cmd_dict    = flag.Bool("dict", false, "generate execl 2o dict_json.")
	do_cmd_errno   = flag.Bool("errno", false, "generate errno.")
	do_cmd_action  = flag.Bool("action", false, "generate action log.")
}

func main() {
	if err := loadConf(); err != nil {
		fmt.Printf("load conf err: %v\n", err)
		return
	}

	flag.Parse()

	if !*do_cmd_dict && !*do_cmd_errno && !*do_cmd_action {
		fmt.Println(help_msg)
		return
	}

	if *do_cmd_dict {
		fmt.Println("start gen dict.")

		cmd_gendict()
		cmd_genmeta()
	}

	if *do_cmd_errno{
		fmt.Println("start gen errno.")
		cmd_generrno()
	}

	if *do_cmd_action {
		fmt.Println("start gen action log.")
		cmd_genaction()
	}

	fmt.Println("Done.")
}
