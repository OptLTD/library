package search

// Run 串联 search 示例：Memory（engine/memory）+ MySQL（engine/mysql）。
func Run() error {
	if err := SearchMemory(); err != nil {
		return err
	}
	return SearchMySQL()
}
