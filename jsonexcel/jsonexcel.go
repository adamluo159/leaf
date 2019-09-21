package jsonexcel

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
	Dir              string                 = "cfg"
	jsonInterfaceMap map[string]*jsonStruct = make(map[string]*jsonStruct)
)

type jsonStruct struct {
	v        interface{}
	keyfield string
	vmap     reflect.Value
}

func getfileName(path string) string {
	allFileName := filepath.Base(path)
	return strings.Split(allFileName, ".")[0]
}

func Init() {
	filepath.Walk(Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fname := getfileName(path)
		jsonStructgo, ok := jsonInterfaceMap[fname]

		if !ok {
			log.Release("not found json struct fname:%s", fname)
			return nil
		}

		err = readJSONFile(path, jsonStructgo)
		if err != nil {
			panic(fmt.Sprintf("%v dir:%v fname:%s", err, Dir, fname))
		}
		return nil
	})
}

func Register(vmap interface{}) {
	vtype := reflect.TypeOf(vmap).Elem().Elem()
	if reflect.TypeOf(vmap).Kind() != reflect.Map {
		panic(fmt.Sprintf("vmap not type Map struct:%s", reflect.TypeOf(vtype).Elem().Name()))
	}
	ok := false
	fname := ""
	keyFiledName := ""
	mapvalue := reflect.New(vtype).Elem()
	mapvalueTypego := mapvalue.Type()

	for i := 0; i < mapvalue.NumField(); i++ {
		fname, ok = mapvalueTypego.Field(i).Tag.Lookup("file")
		if ok {
			keyFiledName = mapvalueTypego.Field(i).Name
			break
		}
	}
	if keyFiledName == "" || fname == "" {
		panic(fmt.Sprintf("config struct:%s must set tag 'file':%s", vtype.Name(), fname))
	}
	jsonInterfaceMap[fname] = &jsonStruct{
		vmap:     reflect.ValueOf(vmap),
		keyfield: keyFiledName,
	}
}

func readJSONFile(filepath string, jsonStruct *jsonStruct) error {
	vSlice := make([]interface{}, 0)
	fbytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fbytes, &vSlice)
	if err != nil {
		return err
	}

	for _, v := range vSlice {
		bs, err := json.Marshal(v)
		if err != nil {
			return err
		}
		vtype := jsonStruct.vmap.Type().Elem().Elem()
		mapvalue := reflect.New(vtype)
		err = json.Unmarshal(bs, mapvalue.Interface())
		if err != nil {
			return err
		}
		kvalue := mapvalue.Elem().FieldByName(jsonStruct.keyfield)
		if kvalue.Kind() == reflect.Int && kvalue.Int() == 0 {
			panic(fmt.Sprintf("表格:%s 主键:%s 值不能为0 ", filepath, jsonStruct.keyfield))
		}
		jsonStruct.vmap.SetMapIndex(kvalue, mapvalue)
	}

	return nil
}
