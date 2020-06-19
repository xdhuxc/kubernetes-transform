package util

import (
	"encoding/json"
	"strconv"

	"github.com/ghodss/yaml"
)

func Marshal(v interface{}) (string, string, error) {
	jsonDataInBytes, err := json.Marshal(v)
	if err != nil {
		return "", "", err
	}

	yamlDataInBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", "", err
	}

	return string(jsonDataInBytes), string(yamlDataInBytes), nil
}

// Convert2Strings 将 interface 类型的数组转换为字符串数组
func Convert2Strings(source []interface{}) []string {
	var target []string

	for _, item := range source {
		switch v := item.(type) {
		case string:
			target = append(target, v)
		case int:
			target = append(target, strconv.FormatInt(int64(v), 10))
		default:
			// maybe occur an error
			target = append(target, item.(string))
		}
	}

	return target
}

// Convert 函数将字符串数组转换为 interface 数组
func Convert2Interfaces(items []string) []interface{} {
	var targets []interface{}
	for _, item := range items {
		targets = append(targets, item)
	}

	return targets
}
