module example

go 1.25.0

require (
	github.com/OptLTD/library/engine/memory v0.0.0
	github.com/OptLTD/library/engine/mysql v0.0.0
	github.com/OptLTD/library/formula v0.0.0
	github.com/OptLTD/library/jsmodule v1.2.2
	github.com/OptLTD/library/jsrunner v1.2.2
	github.com/OptLTD/library/search v1.2.2
)

require (
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/duke-git/lancet/v2 v2.3.7 // indirect
	github.com/expr-lang/expr v1.17.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/google/pprof v0.0.0-20250302191652-9094ed2288e7 // indirect
	github.com/grafana/sobek v0.0.0-20260309140132-c198b3f43d96 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/text v0.34.0 // indirect
	gorm.io/gorm v1.25.11 // indirect
)

replace (
	github.com/OptLTD/library/engine/memory => ../engine/memory
	github.com/OptLTD/library/engine/mysql => ../engine/mysql
	github.com/OptLTD/library/formula => ../formula
	github.com/OptLTD/library/jsmodule => ../jsmodule
	github.com/OptLTD/library/jsrunner => ../jsrunner
	github.com/OptLTD/library/search => ../search
)
