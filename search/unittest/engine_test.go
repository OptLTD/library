package unittest

import (
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/engine"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 创建测试用的 schema.Input
func createTestInput() *schema.Input {
	return &schema.Input{
		Model: &source.Model{
			UUKey:  "test_model",
			Source: "test_table",
			Search: "test_table",
		},
		Request: &request.Record{
			Model: "test_model",
			UUKey: "test_uukey",
		},
		Scope: map[string]any{},
		Groups: []source.Group{
			{
				UUKey: "basic",
				GType: consts.GTYPE_FLATTEN,
			},
		},
		Fields: []source.Field{
			{
				UUKey: "basic.name",
				Field: "name",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_STRINGS,
				Index: "name",
				Shown: true,
			},
			{
				UUKey: "basic.age",
				Field: "age",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_NUMERIC,
				Index: "age",
				Shown: true,
			},
			{
				UUKey: "basic.uukey",
				Field: "uukey",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_STRINGS,
				Index: "uukey",
				Shown: true,
			},
		},
		Unique: []string{"uukey"},
	}
}

// 创建测试用的 schema.Table
func createTestTable() *schema.Table {
	return &schema.Table{
		Model: &source.Model{
			UUKey:  "test_model",
			Source: "test_table",
			Search: "test_table",
		},
		Table: &source.Table{
			UUKey: "default",
		},
		Request: &request.Search{
			Model: "test_model",
			Page:  1,
			Size:  10,
		},
		Scope: map[string]any{},
		Groups: []source.Group{
			{
				UUKey: "basic",
				GType: consts.GTYPE_FLATTEN,
			},
		},
		Fields: []source.Field{
			{
				UUKey: "basic.name",
				Field: "name",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_STRINGS,
				Index: "name",
				Shown: true,
			},
			{
				UUKey: "basic.age",
				Field: "age",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_NUMERIC,
				Index: "age",
				Shown: true,
			},
		},
	}
}

func TestMemoryEngine_Using(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	assert.NotNil(t, memEngine)

	// 测试 Using 方法
	callable := &testCallable{}
	result := memEngine.Using(callable)
	assert.Equal(t, memEngine, result)
}

func TestMemoryEngine_First(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestInput()

	// 测试不存在的记录
	record := &respond.Record{
		UUKey: "not_exist",
		Model: "test_model",
		Request: map[string]any{
			"basic": map[string]any{
				"name": "Test",
			},
		},
		Default: map[string]any{
			"uukey": "not_exist",
		},
	}

	err := memEngine.First(skma, record)
	assert.NoError(t, err)
	assert.Nil(t, record.Storage)

	// 先存储一条记录
	record.UUKey = "test_001"
	record.Default["uukey"] = "test_001"
	record.Prepare = map[string]any{
		"basic": map[string]any{
			"name":  "Test User",
			"age":   25,
			"uukey": "test_001",
		},
	}
	err = memEngine.Store(skma, record)
	assert.NoError(t, err)

	// 再次查询
	record2 := &respond.Record{
		UUKey: "test_001",
		Model: "test_model",
		Request: map[string]any{
			"basic": map[string]any{
				"name": "Test",
			},
		},
		Default: map[string]any{
			"uukey": "test_001",
		},
	}
	err = memEngine.First(skma, record2)
	assert.NoError(t, err)
	assert.NotNil(t, record2.Storage)
}

func TestMemoryEngine_Store(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestInput()

	// 测试 INSERT
	record := &respond.Record{
		UUKey: "test_001",
		Model: "test_model",
		Request: map[string]any{
			"basic": map[string]any{
				"name": "Test User",
				"age":  25,
			},
		},
		Default: map[string]any{
			"uukey":      "test_001",
			"created_at": time.Now(),
		},
		Prepare: map[string]any{
			"basic": map[string]any{
				"name":  "Test User",
				"age":   25,
				"uukey": "test_001",
			},
		},
	}

	record.OpType = "INSERT"
	err := memEngine.Store(skma, record)
	assert.NoError(t, err)

	// 验证数据已存储
	record2 := &respond.Record{
		UUKey: "test_001",
		Model: "test_model",
		Default: map[string]any{
			"uukey": "test_001",
		},
	}
	err = memEngine.First(skma, record2)
	assert.NoError(t, err)
	assert.NotNil(t, record2.Storage)

	// 测试 UPDATE
	record.Request = map[string]any{
		"basic": map[string]any{
			"name": "Updated User",
			"age":  30,
		},
	}
	record.Prepare = map[string]any{
		"basic": map[string]any{
			"name":  "Updated User",
			"age":   30,
			"uukey": "test_001",
		},
	}
	record.OpType = "UPDATE"
	err = memEngine.Store(skma, record)
	assert.NoError(t, err)
}

func TestMemoryEngine_Select(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestInput()

	// 先存储几条记录
	records := []*respond.Record{
		{
			UUKey: "test_001",
			Model: "test_model",
			Default: map[string]any{
				"uukey": "test_001",
			},
			Prepare: map[string]any{
				"basic": map[string]any{
					"name":  "User 1",
					"age":   25,
					"uukey": "test_001",
				},
			},
		},
		{
			UUKey: "test_002",
			Model: "test_model",
			Default: map[string]any{
				"uukey": "test_002",
			},
			Prepare: map[string]any{
				"basic": map[string]any{
					"name":  "User 2",
					"age":   30,
					"uukey": "test_002",
				},
			},
		},
	}

	for _, rec := range records {
		rec.OpType = "INSERT"
		err := memEngine.Store(skma, rec)
		assert.NoError(t, err)
	}

	// 测试 Select
	selectRecords := []*respond.Record{
		{UUKey: "test_001", Model: "test_model"},
		{UUKey: "test_002", Model: "test_model"},
		{UUKey: "test_003", Model: "test_model"}, // 不存在的记录
	}

	err := memEngine.Select(skma, selectRecords)
	assert.NoError(t, err)
	assert.NotNil(t, selectRecords[0].Storage)
	assert.NotNil(t, selectRecords[1].Storage)
	assert.Nil(t, selectRecords[2].Storage)
}

func TestMemoryEngine_Upsert(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestInput()

	records := []*respond.Record{
		{
			UUKey: "test_001",
			Model: "test_model",
			Default: map[string]any{
				"uukey":      "test_001",
				"created_at": time.Now(),
			},
			Prepare: map[string]any{
				"basic": map[string]any{
					"name":  "User 1",
					"age":   25,
					"uukey": "test_001",
				},
			},
		},
		{
			UUKey: "test_002",
			Model: "test_model",
			Default: map[string]any{
				"uukey":      "test_002",
				"created_at": time.Now(),
			},
			Prepare: map[string]any{
				"basic": map[string]any{
					"name":  "User 2",
					"age":   30,
					"uukey": "test_002",
				},
			},
		},
	}

	err := memEngine.Upsert(skma, records)
	assert.NoError(t, err)

	// 验证数据已存储
	for _, rec := range records {
		record := &respond.Record{
			UUKey: rec.UUKey,
			Model: "test_model",
			Default: map[string]any{
				"uukey": rec.UUKey,
			},
		}
		err = memEngine.First(skma, record)
		assert.NoError(t, err)
		assert.NotNil(t, record.Storage)
	}
}

func TestMemoryEngine_Search(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestTable()

	// 先存储一些测试数据
	input := createTestInput()
	records := []*respond.Record{
		{
			UUKey: "test_001",
			Model: "test_model",
			Default: map[string]any{
				"uukey": "test_001",
			},
			Prepare: map[string]any{
				"name":  "Alice",
				"age":   25,
				"uukey": "test_001",
			},
		},
		{
			UUKey: "test_002",
			Model: "test_model",
			Default: map[string]any{
				"uukey": "test_002",
			},
			Prepare: map[string]any{
				"name":  "Bob",
				"age":   30,
				"uukey": "test_002",
			},
		},
		{
			UUKey: "test_003",
			Model: "test_model",
			Default: map[string]any{
				"uukey": "test_003",
			},
			Prepare: map[string]any{
				"name":  "Charlie",
				"age":   35,
				"uukey": "test_003",
			},
		},
	}

	for _, rec := range records {
		rec.OpType = "INSERT"
		err := memEngine.Store(input, rec)
		assert.NoError(t, err)
	}

	// 测试搜索所有记录
	result, err := memEngine.Search(skma)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(3), result.Count)
	assert.Equal(t, uint16(1), result.Page)
	assert.Equal(t, uint16(10), result.Size)
	assert.Len(t, result.Values, 3)

	// 测试带查询条件的搜索
	skma.Request.Query = map[string]any{
		"basic.name:EQ": "Alice",
	}
	result, err = memEngine.Search(skma)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.Count)
	assert.Len(t, result.Values, 1)

	// 测试分页
	skma.Request.Query = map[string]any{}
	skma.Request.Page = 1
	skma.Request.Size = 2
	result, err = memEngine.Search(skma)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), result.Count)
	assert.Len(t, result.Values, 2)
}

func TestMemoryEngine_SUMMARY(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestTable()

	// Pivot 目前简化实现为调用 Search
	result, err := memEngine.Search(skma)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMemoryEngine_QueryLogic(t *testing.T) {
	memEngine := engine.NewMemoryEngine()
	skma := createTestTable()
	input := createTestInput()

	// 存储测试数据
	records := []*respond.Record{
		{
			UUKey:   "test_001",
			Model:   "test_model",
			Default: map[string]any{"uukey": "test_001"},
			Prepare: map[string]any{
				"name": "Alice", "age": 25, "uukey": "test_001",
			},
		},
		{
			UUKey:   "test_002",
			Model:   "test_model",
			Default: map[string]any{"uukey": "test_002"},
			Prepare: map[string]any{
				"name": "Bob", "age": 30, "uukey": "test_002",
			},
		},
	}

	for _, rec := range records {
		rec.OpType = "INSERT"
		err := memEngine.Store(input, rec)
		assert.NoError(t, err)
	}

	// 测试 LIKE 查询
	skma.Request.Query = map[string]any{
		"basic.name:LIKE": "Ali",
	}
	result, err := memEngine.Search(skma)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Count, uint64(1))

	// 测试数值比较
	skma.Request.Query = map[string]any{
		"basic.age:GT": 25,
	}
	result, err = memEngine.Search(skma)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Count, uint64(1))

	// 测试 IN 查询
	skma.Request.Query = map[string]any{
		"basic.age:IN": []any{25, 30},
	}
	result, err = memEngine.Search(skma)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Count, uint64(2))
}

// testCallable 用于测试的回调实现
type testCallable struct{}

func (t *testCallable) BeforeUpsert(skma *schema.Input, record *respond.Record) error {
	return nil
}

func (t *testCallable) HandleUpsert(skma *schema.Input, record *respond.Record) error {
	return nil
}

func (t *testCallable) SearchResult(skma *schema.Table, values *respond.Result) error {
	return nil
}

func (t *testCallable) DigestResult(skma *schema.Digest, values *respond.Result) error {
	return nil
}
