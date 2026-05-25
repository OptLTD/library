package support

import (
	"fmt"
)

var EntryNotExsit = fmt.Errorf("Load Error: %s", "EntryNotExsit")
var ConfigNotExsit = fmt.Errorf("Load Error: %s", "ConfigNotExsit")
var ErrorConfigClient = fmt.Errorf("Load Error: %s", "ErrorConfigClient")
var ErrorConfigFields = fmt.Errorf("Load Error: %s", "ErrorConfigFields")
var ErrorConfigTables = fmt.Errorf("Load Error: %s", "ErrorConfigTables")
var ErrorModelDriver = fmt.Errorf("Load Error: %s", "ErrorModelDriver")
var ErrorModelSource = fmt.Errorf("Load Error: %s", "ErrorModelSource")
var ErrorInputField = fmt.Errorf("Load Error: %s", "ErrorInputField")
var ErrorTableField = fmt.Errorf("Load Error: %s", "ErrorTableField")

var UnexpectedFormat = fmt.Errorf("Data Error: %s", "UnexpectFormat")
var UpsertEmptyRecord = fmt.Errorf("Upsert Error: %s", "UpsertEmptyRecord")
var AggrByEmptyParams = fmt.Errorf("Upsert Error: %s", "AggrByEmptyParams")

var RecordsFoundWhenFirst = fmt.Errorf("Upsert Error: %s", "RecordsFoundWhenFirst")
