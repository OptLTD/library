package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/engine"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"
	"strings"
	"sync/atomic"
	"time"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/shopspring/decimal"
)

func NewEngine(client *elasticsearch.Client) engine.IEngine {
	return &Engine{client: client}
}

type Engine struct {
	handles []engine.ICallable
	client  *elasticsearch.Client
}

func (self *Engine) Using(handle engine.ICallable) engine.IEngine {
	if self.handles == nil {
		self.handles = []engine.ICallable{}
	}

	self.handles = append(self.handles, handle)
	return self
}

func (self *Engine) First(skma *schema.Input, record *respond.Record) error {
	return nil
}

func (self *Engine) Store(skma *schema.Input, record *respond.Record) error {
	// Build the request body.
	data, err := json.Marshal(record.Storage)
	if err != nil {
		log.Printf("Error marshaling document: %s", err)
		return err
	}

	// 处理回掉
	for _, handle := range self.handles {
		err := handle.BeforeUpsert(skma, record)
		if err != nil {
			return err
		}
	}

	// Set up the request object.
	req := esapi.IndexRequest{
		DocumentID: record.UUKey,

		Index: skma.Model.Search,
		Body:  bytes.NewReader(data),
		// Refresh: "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), self.client)
	if err == nil && res != nil {
		defer res.Body.Close()
	}
	defer res.Body.Close()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
		return err
	} else {
		log.Println("Success index document: ", res)
	}
	// 处理回掉
	for _, handle := range self.handles {
		err := handle.HandleUpsert(skma, record)
		if err != nil {
			log.Println("handle upsert error:", err)
		}
	}
	return nil
}

func (self *Engine) Upsert(skma *schema.Input, records []*respond.Record) error {
	size := len(records)
	if size == 0 {
		return support.UpsertEmptyRecord
	}

	var success uint64
	start := time.Now().UTC()
	config := esutil.BulkIndexerConfig{
		Index:         skma.Model.Search, // The default index name
		Client:        self.client,       // The Elasticsearch client
		NumWorkers:    runtime.NumCPU(),  // The number of worker goroutines
		FlushBytes:    int(5e+6),         // The flush threshold in bytes
		FlushInterval: 1 * time.Second,   // The periodic flush interval
	}

	// 处理回掉
	for _, h := range self.handles {
		for i := 0; i < size; i++ {
			err := h.BeforeUpsert(skma, records[i])
			if err != nil {
				return err
			}
		}
	}

	bulk, _ := esutil.NewBulkIndexer(config)
	for i := 0; i < size; i++ {
		record := records[i]
		// Build the request body.
		data, err := json.Marshal(record.Storage)
		if err != nil {
			log.Fatalf("Error format document: %s", err)
			continue
		}
		// log.Println("upsert record info:", string(data))
		err = bulk.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				DocumentID: record.GetUUKey(skma),

				Action: "index",
				Body:   bytes.NewReader(data),
				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&success, 1)
				},
				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Fatalf("Bulk Add Unexpected error: %s", err)
		}
	}
	if err := bulk.Close(context.Background()); err != nil {
		log.Fatalf("Bulk Close Unexpected error: %s", err)
		return err
	} else {
		stats := bulk.Stats()
		dur := time.Since(start)
		log.Printf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(stats.NumFlushed)), dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(stats.NumFlushed))),
		)
	}

	// 处理回掉
	for _, h := range self.handles {
		for i := 0; i < size; i++ {
			err := h.HandleUpsert(skma, records[i])
			if err != nil {
				log.Println("handle upsert err:", err)
			}
		}
	}

	return nil
}

func (self *Engine) Select(schema *schema.Input, record []*respond.Record) error {
	return nil
}

func (self *Engine) Update(skma *schema.Input, data map[string]any) error {
	// 检查 Scope 是否为空
	if len(skma.Scope) == 0 {
		return fmt.Errorf("update scope cannot be empty")
	}

	// 检查是否包含有效的 corp_id
	id, has := skma.Scope[consts.FIELD_CORP_ID]
	if !has || support.Bool(id) == false {
		return fmt.Errorf("update scope must contain corp_id")
	}

	mergedQuery, _ := schema.BuildQuery(skma.Scope)
	query := self.buildQuery(consts.LOGIC_SUBAND, &mergedQuery)

	// 更新前先 count，超过 1000 条则阻止
	var countBody bytes.Buffer
	countRequestBody := map[string]any{
		"query": query,
	}
	if err := json.NewEncoder(&countBody).Encode(countRequestBody); err != nil {
		return fmt.Errorf("error encoding count query: %s", err)
	}

	countRes, err := self.client.Count(
		self.client.Count.WithContext(context.Background()),
		self.client.Count.WithIndex(skma.Model.Search),
		self.client.Count.WithBody(&countBody),
	)
	if err != nil {
		return fmt.Errorf("error counting documents: %s", err)
	}
	defer countRes.Body.Close()

	if countRes.IsError() {
		var e map[string]any
		if err := json.NewDecoder(countRes.Body).Decode(&e); err != nil {
			return fmt.Errorf("error parsing count response: %s", err)
		}
		return fmt.Errorf("count failed: %s", e["error"])
	}

	var countResp map[string]any
	if err := json.NewDecoder(countRes.Body).Decode(&countResp); err != nil {
		return fmt.Errorf("error decoding count response: %s", err)
	}

	count := int64(countResp["count"].(float64))
	if count > 1000 {
		return fmt.Errorf("update would affect %d records, exceeds limit of 1000", count)
	}

	// 构建更新脚本，使用 params 传递值
	scriptParams := map[string]any{}
	scriptParts := []string{}
	for key, value := range data {
		paramKey := fmt.Sprintf("param_%s", key)
		scriptParams[paramKey] = value
		scriptParts = append(scriptParts, fmt.Sprintf("ctx._source.%s = params.%s", key, paramKey))
	}
	script := strings.Join(scriptParts, "; ")

	var body bytes.Buffer
	requestBody := map[string]any{
		"query": query,
		"script": map[string]any{
			"source": script,
			"lang":   "painless",
			"params": scriptParams,
		},
	}
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		log.Printf("Error encoding update query: %s", err)
		return err
	}

	req := esapi.UpdateByQueryRequest{
		Index: []string{skma.Model.Search},
		Body:  &body,
	}

	res, err := req.Do(context.Background(), self.client)
	if err != nil {
		log.Printf("Error updating documents: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("Error parsing response: %s", err)
		} else {
			log.Printf("[%s] %s: %s", res.Status(),
				e["error"].(map[string]any)["type"],
				e["error"].(map[string]any)["reason"],
			)
		}
		return fmt.Errorf("update failed with status: %s", res.Status())
	}
	return nil
}

func (self *Engine) Search(skma *schema.Table) (*respond.Result, error) {
	aggrs := self.buildBasic(skma, "")
	// 分桶聚合
	pivot := skma.BuildDigest()
	groups := pivot.GroupBy
	counts := pivot.CountFn
	if len(groups)+len(counts) == 0 {
		return nil, nil
	}
	for _, group := range groups {
		option, method := map[string]any{}, consts.VALUE_TERMS
		field, ok := slice.Find(skma.Fields, func(idx int, item source.Field) bool {
			return item.UUKey == group.Index
		})
		if !ok || field.UUKey == "" {
			continue
		}
		switch field.FType {
		case consts.FTYPE_DATETIME:
			method = consts.VALUE_HIST2
			// group.Format：day=按天，month=按月（空或未识别时按天，与前端默认一致）
			switch group.Format {
			case "month":
				option["format"] = "yyyy-MM"
				option["interval"] = "month"
			default:
				option["format"] = "yyyy-MM-dd"
				option["interval"] = "day"
			}
			option["min_doc_count"] = 1
		}
		aggrs = []source.CountFn{{
			Label: group.Index, Index: group.Index,
			Func: method, Option: option, Items: aggrs,
		}}
	}

	// skma.Totals = totals
	request := skma.Request
	// Build the request body.
	var body bytes.Buffer
	merge := skma.BuildQuery()
	query := map[string]any{
		"query": self.buildQuery(consts.LOGIC_SUBAND, &merge),
		"aggs":  self.buildAggrs(aggrs),
	}
	if order := request.Order; order != nil {
		query["sort"] = []map[string]any{{
			order.Field: map[string]any{
				"order": order.Order,
			},
		}}
	}
	// query = map[string]any{} // debug
	if err := json.NewEncoder(&body).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
		return nil, err
	}
	log.Printf("ElasticSearch Query: %s", body.String())

	// Perform the search request.
	size := int(request.Size)
	from := int(request.Page-1) * size
	res, err := self.client.Search(
		self.client.Search.WithTrackTotalHits(true),
		self.client.Search.WithContext(context.Background()),
		self.client.Search.WithIndex(skma.Model.Search),
		self.client.Search.WithTimeout(time.Second*10),
		self.client.Search.WithBody(&body),
		self.client.Search.WithFrom(from),
		self.client.Search.WithSize(size),
		self.client.Search.WithPretty(),
	)
	if err != nil {
		log.Printf("ElasticSearch Respond Error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Printf("ElasticSearch Parsing Error: %s", err)
		} else {
			// Print the response status and error information.
			log.Printf("[%s] %s: %s", res.Status(),
				e["error"].(map[string]any)["type"],
				e["error"].(map[string]any)["reason"],
			)
		}
		return nil, err
	}

	var resp map[string]any
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		log.Printf("ElasticSearch Parsing Error: %s", err)
		return nil, err
	}

	// Print the response status, number of results, and request duration.
	hits1 := resp["hits"].(map[string]any)
	total := hits1["total"].(map[string]any)
	value := total["value"].(float64)
	tooks := resp["took"].(float64)
	hits2 := hits1["hits"].([]any)
	log.Printf("ElasticSearch Result: [%s] %d hits; took: %dms", res.Status(), int(value), int(tooks))

	// result
	values := []map[string]any{}
	for _, hit := range hits2 {
		source := hit.(map[string]any)
		values = append(values, source["_source"].(map[string]any))
	}

	// 统计值
	totalValues := map[string]any{}
	aggr, ok := resp["aggregations"].(map[string]any)
	if ok {
		totalValues = self.decodeAggrs(aggr, aggrs)
	}
	result := &respond.Result{
		Page: request.Page,
		Size: request.Size,

		Totals: totalValues,
		Count:  uint64(value),
		Values: values,
	}

	// 处理回掉
	for _, handle := range self.handles {
		handle.SearchResult(skma, result)
	}
	return result, nil
}

func (self *Engine) Digest(skma *schema.Digest) (*respond.Result, error) {
	meta := skma.Table.BuildDigest()
	pivot := slice.Map(meta.GroupBy, func(_ int, g source.GroupBy) string {
		return g.Index
	})

	result, err := self.Search(skma.Table)
	if err != nil {
		return result, err
	}

	values := self.expandTotals(pivot, result.Totals)
	totals := slice.Reduce(values, func(idx int, carry map[string]any, curr map[string]any) map[string]any {
		for k, val := range curr {
			if k == "_total_" {
				continue
			}
			v := float64(0)
			if take, ok := carry[k]; ok {
				v = take.(float64)
			}
			switch val := val.(type) {
			case string:
				v += 1
			case float64:
				v += val
			case int64:
				v += float64(val)
			default:
				log.Println("Total Reduce Error:", k, val)
			}
			carry[k], _ = decimal.NewFromFloat(v).Round(2).Float64()
		}

		_total_, ok := carry["_total_"].(map[string]any)
		if !ok {
			carry["_total_"] = curr["_total_"]
			return carry
		}
		if total, ok := curr["_total_"].(map[string]any); ok {
			for k, val := range total {
				v := float64(0)
				if take, ok := _total_[k]; ok {
					v = take.(float64)
				}
				switch val := val.(type) {
				case string:
					v += 1
				case float64:
					v += val
				case int64:
					v += float64(val)
				default:
					log.Println("Total Reduce Error:", k, val)
				}
				_total_[k], _ = decimal.NewFromFloat(v).Round(2).Float64()
			}
			carry["_total_"] = _total_
		}
		return carry
	}, map[string]any{})

	result.Values, result.Totals = values, totals
	result.Count = uint64(len(values))

	// 处理回掉
	for _, handle := range self.handles {
		handle.DigestResult(skma, result)
	}
	return result, nil
}

func (self *Engine) expandTotals(tokens []string, totals map[string]any) []map[string]any {
	if len(tokens) == 0 || len(totals) == 0 {
		return []map[string]any{}
	}
	first, tokens := tokens[0], append([]string{}, tokens[1:]...)
	other := maputil.OmitByKeys(totals, []string{first})
	total, ok := other["_total_"].(map[string]any)
	if !ok {
		total = map[string]any{}
	}

	values := []map[string]any{}
	if item, ok := totals[first]; ok {
		items := item.([]any)
		count := len(items)
		for i := 0; i < count; i++ {
			value := items[i].(map[string]any)
			value = maputil.Merge(value, other)
			inner := value["_total_"].(map[string]any)
			value["_total_"] = maputil.Merge(inner, total)
			if len(tokens) == 0 {
				values = append(values, value)
				continue
			}
			childs := self.expandTotals(tokens, value)
			values = append(values, childs...)
		}
	}
	return values
}

func (self *Engine) buildField(fields *[]source.Field) []string {
	if len(*fields) == 0 {
		return []string{"*"}
	}
	slices := []string{}
	for _, field := range *fields {
		slices = append(slices, field.Field)
	}
	return slices
}

func (self *Engine) buildQuery(logic string, queries *[]schema.Query) map[string]any {
	result := []map[string]any{}
	for _, query := range *queries {
		var where map[string]any
		switch strings.ToUpper(query.Logic) {
		case consts.LOGIC_EQUALSTO:
			where = map[string]any{"term": map[string]any{query.Field: query.Value}}
		case consts.LOGIC_STR_LIKE:
			where = map[string]any{"match": map[string]any{query.Field: query.Value}}
		case consts.LOGIC_INCLUDES:
			where = map[string]any{"terms": map[string]any{query.Field: query.Value}}
		case consts.LOGIC_CONTAINS:
			where = map[string]any{"terms": map[string]any{query.Field: query.Value}}
		case consts.LOGIC_LESTHAN:
			where = map[string]any{
				"range": map[string]any{
					query.Field: map[string]any{
						"lt": query.Value,
					},
				},
			}
		case consts.LOGIC_GREATER:
			where = map[string]any{
				"range": map[string]any{
					query.Field: map[string]any{
						"gt": query.Value,
					},
				},
			}
		case consts.LOGIC_LESS_EQ:
			where = map[string]any{
				"range": map[string]any{
					query.Field: map[string]any{
						"lte": query.Value,
					},
				},
			}
		case consts.LOGIC_GRAT_EQ:
			where = map[string]any{
				"range": map[string]any{
					query.Field: map[string]any{
						"gte": query.Value,
					},
				},
			}
		case consts.LOGIC_BETWEEN:
			values := query.Value.([]any)
			where = map[string]any{"range": map[string]any{
				query.Field: map[string]any{
					"gte": values[0],
					"lte": values[1],
				},
			}}
		case consts.LOGIC_EXISTS:
			where = map[string]any{"exists": map[string]any{"field": query.Field}}
		case consts.LOGIC_VAL_NULL:
			query := schema.Query{Logic: consts.LOGIC_EXISTS, Field: query.Field}
			where = self.buildQuery(consts.LOGIC_SUBNOT, &[]schema.Query{query})
		case consts.LOGIC_NOT_NULL:
			query := schema.Query{Logic: consts.LOGIC_EXISTS, Field: query.Field}
			where = self.buildQuery(consts.LOGIC_SUBAND, &[]schema.Query{query})
		case consts.LOGIC_SUBRAW:
			continue
		case consts.LOGIC_SUBOR:
			where = self.buildQuery(consts.LOGIC_SUBOR, query.Items)
		case consts.LOGIC_SUBAND:
			where = self.buildQuery(consts.LOGIC_SUBAND, query.Items)
		case consts.LOGIC_SUBNOT:
			where = self.buildQuery(consts.LOGIC_SUBNOT, query.Items)
		default:
			// 异常逻辑
		}
		result = append(result, where)
	}
	if len(result) == 0 {
		return map[string]any{}
	}
	switch logic {
	case consts.LOGIC_SUBOR:
		return map[string]any{"bool": map[string]any{"should": result}}
	case consts.LOGIC_SUBAND:
		return map[string]any{"bool": map[string]any{"must": result}}
	case consts.LOGIC_SUBNOT:
		return map[string]any{"bool": map[string]any{"must_not": result}}
	}
	return map[string]any{}
}

func (self *Engine) buildBasic(skma *schema.Table, nested string) []source.CountFn {
	totals := []source.CountFn{}
	if nested == "" {
		for _, group := range skma.Groups {
			if group.Extra.Nested == false {
				continue
			}
			child := self.buildBasic(skma, group.UUKey)
			if len(child) > 0 {
				totals = append(totals, source.CountFn{
					Label: group.UUKey, Index: group.UUKey,
					Func: consts.VALUE_NESTED, Items: child,
					Option: map[string]any{"path": group.UUKey},
				})
			}
		}
	}
	for i := 0; i < len(skma.Fields); i++ {
		field := skma.Fields[i]
		if field.Shown == false || field.Index == "" {
			continue
		}
		if nested != "" && field.Group != nested {
			continue
		}
		// todo fixed nested
		switch field.GetDataType() {
		case consts.DTYPE_INTEGER, consts.DTYPE_TINYINT,
			consts.DTYPE_LONGINT, consts.DTYPE_SCALED,
			consts.DTYPE_DECIMAL, consts.DTYPE_EXPENSE:
			totals = append(totals, source.CountFn{
				Label: field.UUKey, Index: field.Index,
				Func: consts.VALUE_SUM, Option: nil,
			})
		case consts.DTYPE_SERIALNO, consts.DTYPE_RELATION,
			consts.DTYPE_OPTIONAL, consts.DTYPE_WORKFLOW:
			totals = append(totals, source.CountFn{
				Label: field.UUKey, Index: field.Index,
				Func: consts.VALUE_CNT, Option: nil,
			})
		case consts.DTYPE_KEYWORDS, consts.DTYPE_SUBJECT,
			consts.DTYPE_X_EMAIL, consts.DTYPE_X_PHONE:
			totals = append(totals, source.CountFn{
				Label: field.UUKey, Index: field.Index,
				Func: consts.VALUE_CNT, Option: nil,
			})
		case consts.DTYPE_DATETIME, consts.DTYPE_ONLYDATE:
			continue
		case consts.DTYPE_DOC_FILE, consts.DTYPE_IMG_FILE,
			consts.DTYPE_LONGTEXT, consts.DTYPE_RICHTEXT:
			continue
		case consts.DTYPE_LOCATION:
			log.Println(
				"cannot aggregate so skip location field",
				field.UUKey, "of model", skma.Model.UUKey,
			)
			continue
		default:
			log.Println("unrecongize:", field.UUKey, "of", skma.Model.UUKey)
		}
	}
	return totals
}
func (self *Engine) buildAggrs(totals []source.CountFn) map[string]any {
	result := map[string]any{}
	if len(totals) == 0 {
		return result
	}
	for _, item := range totals {
		child := map[string]any{}
		if len(item.Items) > 0 {
			child = self.buildAggrs(item.Items)
		}
		aggr := map[string]any{"field": item.Index}
		aggr = maputil.Merge(item.Option, aggr)
		switch strings.ToUpper(item.Func) {
		case consts.VALUE_AVG:
			result[item.Label] = map[string]any{
				"avg": map[string]any{"field": item.Index},
			}
		case consts.VALUE_SUM:
			result[item.Label] = map[string]any{
				"sum": map[string]any{"field": item.Index},
			}
		case consts.VALUE_MAX:
			result[item.Label] = map[string]any{
				"max": map[string]any{"field": item.Index},
			}
		case consts.VALUE_MIN:
			result[item.Label] = map[string]any{
				"min": map[string]any{"field": item.Index},
			}
		case consts.VALUE_CNT:
			result[item.Label] = map[string]any{
				"value_count": map[string]any{"field": item.Index},
			}
		case consts.VALUE_UNQ:
			result[item.Label] = map[string]any{
				"cardinality": map[string]any{"field": item.Index},
			}

		case consts.VALUE_NESTED:
			result[item.Label] = map[string]any{
				// 子聚合
				"aggs": child,
				// nested聚合
				"nested": map[string]any{
					"path": item.Index,
				},
			}
		case consts.VALUE_TERMS:
			result[item.Label] = map[string]any{
				// 子聚合
				"aggs": child,
				// 桶聚合
				"terms": aggr,
			}
		case consts.VALUE_RANGE:
			result[item.Label] = map[string]any{
				// 子聚合
				"aggs": child,
				// 桶聚合
				"range": aggr,
			}
		case consts.VALUE_HIST1:
			result[item.Label] = map[string]any{
				// 子聚合
				"aggs": child,
				// 桶聚合
				"histogram": aggr,
			}
		case consts.VALUE_HIST2:
			result[item.Label] = map[string]any{
				// 子聚合
				"aggs": child,
				// 桶聚合
				"date_histogram": aggr,
			}
		default:
			// 异常逻辑
		}
	}

	return result
}

func (self *Engine) decodeAggrs(values map[string]any, totals []source.CountFn) map[string]any {
	result := map[string]any{}
	if len(totals) == 0 {
		return result
	}
	for _, item := range totals {
		switch strings.ToUpper(item.Func) {
		case consts.VALUE_AVG, consts.VALUE_SUM, consts.VALUE_MAX, consts.VALUE_MIN, consts.VALUE_CNT, consts.VALUE_UNQ:
			if val, ok := values[item.Label]; ok {
				result[item.Label], _ = val.(map[string]any)["value"]
				switch result[item.Label].(type) {
				case float64:
					v := result[item.Label].(float64)
					result[item.Label], _ = decimal.NewFromFloat(v).Round(2).Float64()
				}
			}
		case consts.VALUE_NESTED:
			if val, ok := values[item.Label].(map[string]any); ok {
				value := self.decodeAggrs(val, item.Items)
				result = maputil.Merge(result, value)
			}
		case consts.VALUE_TERMS, consts.VALUE_RANGE, consts.VALUE_HIST1, consts.VALUE_HIST2:
			if val, ok := values[item.Label]; ok {
				buckets, _ := val.(map[string]any)["buckets"]
				result1 := []any{}
				for _, bucket := range buckets.([]any) {
					decode := bucket.(map[string]any)
					value := self.decodeAggrs(decode, item.Items)
					if key, ok := decode["key"]; ok {
						value[item.Index] = key
					}
					if key, ok := decode["key_as_string"]; ok {
						value[item.Index] = key
					}
					value["_total_"] = map[string]any{
						item.Index: decode["doc_count"],
					}
					result1 = append(result1, value)
				}
				result[item.Label] = result1
			}
		default:
			// 异常逻辑
		}
	}

	return result
}
