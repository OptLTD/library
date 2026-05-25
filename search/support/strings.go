package support

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var Errorf = fmt.Errorf
var Sprint = fmt.Sprint
var Sprintf = fmt.Sprintf
var Printf = fmt.Printf
var Println = fmt.Println

var ToUpper = strings.ToUpper
var ToLower = strings.ToLower
var Replace = strings.Replace
var Split = strings.Split
var IsError = errors.Is

func ToJson(val any) string {
	json, err := json.Marshal(val)
	if err == nil {
		return string(json)
	}
	return ""
}
