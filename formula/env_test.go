package formula

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormulaEnv_dotKeysBecomeNested(t *testing.T) {
	env := FormulaEnv(map[string]any{
		"demo.some": "123",
		"other":     1,
	})
	demo, ok := env["demo"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "123", demo["some"])
	assert.Equal(t, 1, env["other"])
}

func TestFormulaEnv_nestedAndDotMerge(t *testing.T) {
	env := FormulaEnv(map[string]any{
		"demo": map[string]any{"a": 1},
		"demo.b": 2,
	})
	demo := env["demo"].(map[string]any)
	assert.Equal(t, 1, demo["a"])
	assert.Equal(t, 2, demo["b"])
}

func TestFormulaEnv_BuildUsesDemoDotSome(t *testing.T) {
	out, err := Build(`demo.some + 1`, FormulaEnv(map[string]any{
		"demo.some": 10,
	}))
	assert.NoError(t, err)
	assert.InDelta(t, 11.0, out.(float64), 1e-6)
}
