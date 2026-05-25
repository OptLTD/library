package request

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Account struct {
	Lang string `json:"lang"`
	Name string `json:"name"`
	// 任意类型
	UUID any `json:"uuid"`
	Team any `json:"team"`
	Corp any `json:"corp"`
}

func NewAccount(user map[string]any) *Account {
	acc := &Account{}
	for key, val := range user {
		switch key {
		case "name":
			acc.Name = fmt.Sprint(val)
		case "lang":
			acc.Lang = fmt.Sprint(val)
		case "id":
			acc.UUID = val
		case "team_id":
			acc.Team = val
		case "corp_id":
			acc.Corp = val
		}
	}
	return acc
}

func ParseLogin(login string) *Account {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil
	}
	if strings.HasPrefix(login, "{") {
		var acc Account
		if err := json.Unmarshal([]byte(login), &acc); err != nil {
			return nil
		}
		if acc.Name != "" || acc.Corp != nil {
			return &acc
		}
		return nil
	}
	parts := strings.SplitN(login, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}
	name := strings.TrimSpace(parts[0])
	corpStr := strings.TrimSpace(parts[1])
	corp, err := strconv.ParseUint(corpStr, 10, 64)
	if err != nil {
		return nil
	}
	return &Account{
		Name: name,
		UUID: name,
		Corp: corp,
	}
}
