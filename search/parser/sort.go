package parser

import (
	"sort"

	"github.com/OptLTD/library/search/source"
)

// sortFieldsByGroupThenSeq orders fields by group.seqno, then field.seqno, then uukey.
// Field seqno is typically per-group (each group starts at 1); sorting by field.seqno
// alone interleaves groups.
func sortFieldsByGroupThenSeq(fields []source.Field, groups map[string]source.Group) {
	if len(fields) <= 1 {
		return
	}
	gseq := map[string]uint16{}
	for key, g := range groups {
		id := g.UUKey
		if id == "" {
			id = key
		}
		gseq[id] = g.SeqNo
	}
	sort.SliceStable(fields, func(i, j int) bool {
		gi, gj := gseq[fields[i].Group], gseq[fields[j].Group]
		if gi != gj {
			return gi < gj
		}
		if fields[i].SeqNo != fields[j].SeqNo {
			return fields[i].SeqNo < fields[j].SeqNo
		}
		return fields[i].GetKey() < fields[j].GetKey()
	})
}

// SortFieldsByGroupThenSeq is the exported helper for service/setup layers.
func SortFieldsByGroupThenSeq(fields []source.Field, groups []source.Group) {
	if len(fields) <= 1 {
		return
	}
	m := make(map[string]source.Group, len(groups))
	for _, g := range groups {
		m[g.UUKey] = g
	}
	sortFieldsByGroupThenSeq(fields, m)
}
