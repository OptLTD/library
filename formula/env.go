package formula

// 环境类：业务 Combine 扁平键 → expr 可用的嵌套 map（如 demo.some）。

import (
	"reflect"
	"strings"
)

// FormulaEnv 将 Combine 转为 expr 可用的环境：无点的键原样放入（map 值浅拷贝一层，避免合并点号键时改动 Combine）；带点的键（如 demo.some）合并为嵌套 map，便于公式里写 demo.some。
func FormulaEnv(combine map[string]any) map[string]any {
	if combine == nil {
		return map[string]any{}
	}
	type dotEnt struct {
		parts []string
		val   any
	}
	var dots []dotEnt
	env := make(map[string]any, len(combine))
	for k, v := range combine {
		if strings.Contains(k, ".") {
			dots = append(dots, dotEnt{strings.Split(k, "."), v})
			continue
		}
		env[k] = envShallowCopyMapValue(v)
	}
	for _, d := range dots {
		envMergePath(env, d.parts, d.val)
	}
	return env
}

func envShallowCopyMapValue(v any) any {
	if m, ok := envAsStringMap(v); ok {
		cp := make(map[string]any, len(m))
		for k, val := range m {
			cp[k] = val
		}
		return cp
	}
	return v
}

// envAsStringMap 兼容 map[string]any、map[string]interface{} 及键为 string 的 reflect.Map（如 bson primitive.M）。
func envAsStringMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		out[iter.Key().String()] = iter.Value().Interface()
	}
	return out, true
}

func envMergePath(dest map[string]any, parts []string, value any) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		dest[parts[0]] = value
		return
	}
	key := parts[0]
	var sub map[string]any
	if cur, ok := dest[key]; ok {
		if m, ok := envAsStringMap(cur); ok {
			sub = make(map[string]any, len(m)+1)
			for k, val := range m {
				sub[k] = val
			}
		}
	}
	if sub == nil {
		sub = map[string]any{}
	}
	dest[key] = sub
	envMergePath(sub, parts[1:], value)
}
