package parser

import (
	"strings"
)

// [NAME][*] => 从 source.Clicks 中收集所有以 [NAME] 为前缀的 key，作为展开结果
func ExpandClicks(keys []string, allKeys []string) []string {
	var expandedKeys []string
	for _, key := range keys {
		if !strings.HasSuffix(key, "[*]") {
			expandedKeys = append(expandedKeys, key)
			continue
		}

		var matched []string
		prefix := strings.TrimSuffix(key, "[*]")
		for _, mapKey := range allKeys {
			if strings.HasPrefix(mapKey, prefix) {
				matched = append(matched, mapKey)
			}
		}
		expandedKeys = append(expandedKeys, matched...)
	}

	return expandedKeys
}
