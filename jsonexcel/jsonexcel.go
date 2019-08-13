package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/adamluo159/leaf/log"
)

var (
	Dir     string     = "cfg"
	JsonCfg *JsonExcel = &JsonExcel{
		json_interface_map: make(map[string]*jsonStruct),
	}
)

type JsonExcel struct {
	json_interface_map map[string]*jsonStruct //fname:jsonstruct
}
type jsonStruct struct {
	v        interface{}
	keyfield string
	vmap     reflect.Value
}

func (je *JsonExcel) Register(vmap interface{}) {
	vtype := reflect.TypeOf(vmap).Elem().Elem()
	if reflect.TypeOf(vmap).Kind() != reflect.Map {
		panic(fmt.Sprintf("vmap not type Map struct:%s", reflect.TypeOf(vtype).Elem().Name()))
	}
	ok := false
	fname := ""
	key_filed_name := ""
	mapvalue := reflect.New(vtype).Elem()
	mapvalue_type := mapvalue.Type()

	for i := 0; i < mapvalue.NumField(); i++ {
		fname, ok = mapvalue_type.Field(i).Tag.Lookup("file")
		if ok {
			key_filed_name = mapvalue_type.Field(i).Name
			break
		}
	}
	if key_filed_name == "" || fname == "" {
		panic(fmt.Sprintf("config struct:%s must set tag 'file':%s", vtype.Name(), fname))
	}
	je.json_interface_map[fname] = &jsonStruct{
		vmap:     reflect.ValueOf(vmap),
		keyfield: key_filed_name,
	}
}

func (je *JsonExcel) OnInit() {
	je.readJsonDir(Dir)
}
func (je *JsonExcel) Run(closeSig chan bool) {}
func (je *JsonExcel) OnDestroy()             {}
func (je *JsonExcel) Name() string {
	return "json excel"
}

func getfileName(path string) string {
	all_file_name := filepath.Base(path)
	return strings.Split(all_file_name, ".")[0]
}

func (je *JsonExcel) readJsonFile(filepath string, json_struct *jsonStruct) error {
	v_slice := make([]interface{}, 0)
	fbytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fbytes, &v_slice)
	if err != nil {
		return err
	}

	for _, v := range v_slice {
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		vtype := json_struct.vmap.Type().Elem().Elem()
		mapvalue := reflect.New(vtype)
		err = json.Unmarshal(bs, mapvalue.Interface())
		if err != nil {
			return err
		}
		kvalue := mapvalue.Elem().FieldByName(json_struct.keyfield)
		json_struct.vmap.SetMapIndex(kvalue, mapvalue)
	}

	return nil
}

func (je *JsonExcel) readJsonDir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fname := getfileName(path)
		json_struct, ok := je.json_interface_map[fname]

		if !ok {
			log.Release("not found json struct fname:%s", fname)
			return nil
		}

		err = je.readJsonFile(path, json_struct)
		if err != nil {
			panic(fmt.Sprintf("%v dir:%v fname:%s", err, dir, fname))
		}
		return nil
	})
}
