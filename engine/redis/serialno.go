package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/OptLTD/library/search/support"

	"github.com/alovn/go-bloomfilter"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/redis/go-redis/v9"
)

type SNOption struct {
	Kind  string
	Value string
}

type SNConfig struct {
	SnKind  string
	Options []string
}
type SerialNo struct {
	snkind string
	period string
	option []SNOption
	client *redis.Client
	filter bloomfilter.BloomFilter
}

const (
	GROWTH = "GROWTH"
	FILTER = "FILTER"
)

const (
	KIND_CONSTANT = "CONSTANT"
	KIND_DATETIME = "DATETIME"
	KIND_COUNTING = "COUNTING"
)

var ALL_KINDS = []string{
	KIND_CONSTANT,
	KIND_DATETIME,
	KIND_COUNTING,
}

func (self *SerialNo) Option(rules []string) []SNOption {
	result := []SNOption{}
	incExist := false
	for _, rule := range rules {
		if !strings.Contains(rule, ":") {
			continue
		}
		parts := strings.Split(rule, ":")
		kind := strings.ToUpper(parts[0])
		if !slice.Contain(ALL_KINDS, kind) {
			log.Println("SerialNo Kind Error:", kind)
			continue
		}
		if kind == KIND_COUNTING {
			incExist = true
		}
		option := SNOption{
			Kind: kind, Value: parts[1],
		}
		result = append(result, option)
	}
	if incExist == false {
		result = append(result, SNOption{
			Kind: KIND_COUNTING, Value: "5",
		})
	}
	return result
}
func NewSerialNo(client *redis.Client) *SerialNo {
	return &SerialNo{client: client}
}

func (self *SerialNo) Init(kind string, options []string) error {
	self.snkind = kind
	self.option = self.Option(options)
	self.period = time.Now().Format("200601")

	if self.client == nil {
		return errors.New("redis client not found")
	}

	self.filter = bloomfilter.NewRedisBloomFilter(
		self.client, self.key(FILTER), 100000,
	)
	return nil
}
func (self *SerialNo) Once() string {
	codes := self.Build(1)
	self.insert(codes[0])
	return codes[0]
}

func (self *SerialNo) Build(count int) []string {
	if self.client == nil || self.filter == nil {
		log.Println("serialno build skipped: serial is not initialized")
		return []string{}
	}
	if count <= 0 {
		return []string{}
	}
	code := self.parse(1)
	for i := 1; i <= 1000; i++ {
		if !self.exists(code) {
			break
		}
		self.insert(code)
		code = self.parse(1)
	}
	// 批量生成
	result := []string{}
	for i := 1; i <= count; i++ {
		code = self.parse(int64(i))
		if self.exists(code) {
			self.insert(code)
			count += 1
			continue
		}
		result = append(result, code)
	}
	return result
}

func (self *SerialNo) parse(base int64) string {
	growth, parts := int64(0), self.prefix()
	for _, option := range self.option {
		switch option.Kind {
		case "RANDOM":
			length, err := strconv.ParseInt(option.Value, 10, 64)
			length = support.If(length == 0 || err != nil, length, 5)
			parts = append(parts, random.RandNumeral(int(length)))
		case KIND_COUNTING:
			if growth, _ = strconv.ParseInt(option.Value, 10, 10); growth == 0 {
				growth = 5
			}
		}
	}

	if growth > 0 {
		// code := strings.Join(parts, "")
		// code += strings.Repeat("0", growth)
		// if !self.exists(code) {
		// 	self.setLast(1)
		// }
		noStr := fmt.Sprint(self.getLast() + base)
		noStr = strutil.PadStart(noStr, int(growth), "0")
		parts = append(parts, noStr)
	}
	return strings.Join(parts, "")
}

func (self *SerialNo) prefix() []string {
	parts := []string{}
	for _, option := range self.option {
		switch option.Kind {
		case KIND_CONSTANT:
			parts = append(parts, option.Value)
		case KIND_DATETIME:
			switch option.Value {
			case "ONLYDATE":
				parts = append(parts, time.Now().Format("20060102"))
			case "ONLYTIME":
				parts = append(parts, time.Now().Format("030405"))
			case "DATETIME":
				parts = append(parts, time.Now().Format("20060102030405"))
			}
		}
	}
	return parts
}

func (self *SerialNo) Insert(code string) bool {
	if self.client == nil || self.filter == nil {
		log.Println("serialno insert skipped: serial is not initialized")
		return false
	}
	if self.exists(code) {
		return true
	}
	return self.insert(code)
}

func (self *SerialNo) insert(code string) bool {
	err := self.filter.Put([]byte(code))
	if err != nil {
		log.Println("bloomfilter insert fail", err)
		return false
	}
	last := self.getLast()
	self.setLast(last + 1)
	return true
}

func (self *SerialNo) exists(code string) bool {
	exists, err := self.filter.MightContain([]byte(code))
	if exists && err == nil {
		return true
	}
	return false
}
func (self *SerialNo) getLast() int64 {
	ctx := context.Background()
	key := self.key(strings.Join(self.prefix(), ","))
	val, err := self.client.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		log.Println("key2 does not exist")
		return 0
	}

	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Println("key2 does not strconv")
		return 0
	}
	// log.Println("[GEN-SN] GET NUM => ", key, val)
	return num
}

func (self *SerialNo) setLast(last int64) {
	ctx := context.Background()
	key := self.key(strings.Join(self.prefix(), ","))
	err := self.client.Set(ctx, key, last, 0).Err()
	if err != nil {
		panic(err)
	}
	// log.Println("[GEN-SN] SET NUM => ", key, last)
}

func (self *SerialNo) key(kind string) string {
	parts := []string{
		"SN", self.snkind,
		self.period, kind,
	}
	return strings.Join(parts, ":")
}
