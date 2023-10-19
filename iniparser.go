/**
 *
 * @package       main
 * @author        YuanZhiGang <zackyuan@yeah.net>
 * @version       1.0.0
 * @copyright (c) 2013-2023, YuanZhiGang
 */

package main

import (
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type AppDbIni struct {
	Host    string `ini:"host"`
	Post    int    `ini:"port"`
	User    string `ini:"user"`
	Pass    string `ini:"pass"`
	Name    string `ini:"name"`
	Charset string `ini:"charset"`
}

type AppCacheIni struct {
	Host string `ini:"host"`
	Post int    `ini:"port"`
	Pass string `ini:"pass"`
}

type IniConfig struct {
	AppDbIni    `ini:"db"`
	AppCacheIni `ini:"cache"`
}

func loadIniConfig(filename string, ic interface{}) error {
	// 参数校验 ad必须是结构体指针类型
	t := reflect.TypeOf(ic)

	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		err := errors.New("ad param should be a struct pointer")
		return err
	}

	// 读取文件
	data_file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	// 将文件内容转换为字符串并输出
	//fmt.Println(string(data_file))
	// 将读取的文件内容进行分割为一行一行的
	slice_string := strings.Split(string(data_file), "\n")
	// 遍历每一行并输出
	var structName string
	for idx, line := range slice_string {

		//0.去除每行首尾的空格
		line := strings.TrimSpace(line)

		//1.跳过空行
		if line == "" {
			//fmt.Println("行号:", idx+1, "空行跳过")
			continue
		}

		//2.排除注释
		//strings.TrimSuffix() // 字符串是否以指点后缀结尾
		//strings.TrimPrefix() // 字符串是否以指点前缀开头
		//strings.HasPrefix() // HasPrefix测试字符串s是否以prefix开头。
		//strings.HasSuffix // HasSuffix测试字符串s是否以suffix结尾。
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			//fmt.Println("行号:", idx+1, "注释行跳过")
			continue
		}

		//3.筛选异常行，输出异常提示
		// 表明这个是节名的这一行,其他行是键值对内容
		if line[0] == '[' && line[len(line)-1] == ']' {
			section_content := strings.TrimSpace(line[1 : len(line)-1])

			//fmt.Println(idx+1, "section_content:", section_content)

			if section_content == "" {
				err := fmt.Errorf("行号:%d,节标签无内容", idx+1)
				return err
			}
			sectionName := section_content // 获取到了节名字符串

			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)

				//fmt.Println("Type:", field.Type)
				//fmt.Println("Name:", field.Name)
				//fmt.Println("Tag:", field.Tag)
				//fmt.Println("Index:", field.Index)
				//fmt.Println("PkgPath:", field.PkgPath)
				//fmt.Println("Anonymous:", field.Anonymous)
				//fmt.Println("Offset:", field.Offset)

				if field.Tag.Get("ini") == sectionName {
					structName = field.Name
					//fmt.Println(structName)
				}
			}

		} else { // 内容键值对形式的
			if strings.Contains(line, "[") || strings.Contains(line, "]") {
				err := fmt.Errorf("行号:%d,内容中包含'['或者']'字符", idx+1)
				return err
			}

			if !strings.Contains(line, "=") {
				err := fmt.Errorf("行号:%d,内容中不包含'='字符", idx+1)
				return err
			}

			if strings.Count(line, "=") != 1 {
				err := fmt.Errorf("行号:%d,内容中'='字符数量不为1", idx+1)
				return err
			}

			key_value := strings.Split(line, "=")
			content_key := strings.TrimSpace(key_value[0])
			content_value := strings.TrimSpace(key_value[1])

			//没有键或者没有值
			if content_key == "" || content_value == "" {
				err := fmt.Errorf("行号:%d,内容中没有键或者没有值", idx+1)
				return err
			}

			//fmt.Println(idx+1, "内容:key:", key_value[0], "value:", key_value[1])

			v := reflect.ValueOf(ic)
			// 通过名称获取嵌套结构的值信息
			valuestruct := v.Elem().FieldByName(structName)
			// 通过值信息获取嵌套结构的类型信息
			typestruct := valuestruct.Type()

			if valuestruct.Kind() != reflect.Struct {
				err := fmt.Errorf("type need is struct")
				return err
			}

			var struct_file_name string
			for i := 0; i < valuestruct.NumField(); i++ {
				st_field := typestruct.Field(i)

				// 找到了对应结构中对应的字段
				if st_field.Tag.Get("ini") == content_key {
					struct_file_name = st_field.Name
					break
				}
			}

			str_obj := valuestruct.FieldByName(struct_file_name)
			//fmt.Println(str_obj.Type(), str_obj.Kind(), struct_file_name)

			//给结构体赋值
			switch str_obj.Type().Kind() {
			case reflect.String:
				str_obj.SetString(content_value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				num, err := strconv.Atoi(content_value)
				if err != nil {
					err = fmt.Errorf("转换失败,err:%s", err)
					return err
				}

				str_obj.SetInt(int64(num))
			}

		}

	}

	return nil
}

func main() {
	// 1.自定义方式解析ini
	var ic IniConfig

	err := loadIniConfig("./etc/app.ini", &ic)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ic)

	// 使用ini解析包go-ini=> go get gopkg.in/ini.v1
	cfg, err := ini.Load("./etc/appmy.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	fmt.Println("App Mode:", cfg.Section("").Key("app_mode").String())
	fmt.Println("Data Path:", cfg.Section("paths").Key("data").String())

	fmt.Println(cfg.Section("db").Key("host").String())
	db_port, _ := cfg.Section("db").Key("port").Int()
	fmt.Println(db_port)

	fmt.Println(cfg.Section("cache").Key("host").String())
	fmt.Println(cfg.Section("cache").Key("port").String())

}
