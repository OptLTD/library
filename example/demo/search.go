package demo

// Search 串联 search 示例：Memory（core）+ MySQL driver（升级后 API）。
func Search() error {
	if err := SearchMemory(); err != nil {
		return err
	}
	return SearchMySQL()
}
