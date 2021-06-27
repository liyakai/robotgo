package tools

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var configData map[string]map[string]string

func EnvLoad(path string) bool {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = json.Unmarshal(file, &configData)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func EnvGet(table, key string) string {
	t, ok := configData[table]
	if !ok {
		return ""
	}
	val, ok := t[key]
	if !ok {
		return ""
	}
	return val
}

func EnvGlobal(key string) string {
	val, ok := configData["global"][key]
	if !ok {
		return ""
	}
	return val
}

func EnvSet(key1, key2, val string) {
	if kmap, ok := configData[key1]; ok {
		kmap[key2] = val
	}
}
