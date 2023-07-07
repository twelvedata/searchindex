package searchindex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SymbolInfo struct {
	Symbol     string
	Exchange   string
	Instrument string
}

func sortFunc(i, j int, data SearchList[*SymbolInfo]) bool {
	if data[i].Key == data[j].Key {
		return data[i].Data.Exchange < data[j].Data.Exchange
	}
	return data[i].Key < data[j].Key
}

func TestSearchIndex(t *testing.T) {
	a11 := SearchItem[*SymbolInfo]{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "1", Instrument: "A company"}}
	a12 := SearchItem[*SymbolInfo]{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "2", Instrument: "A company"}}
	a13 := SearchItem[*SymbolInfo]{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "3", Instrument: "A company"}}
	an := SearchItem[*SymbolInfo]{Key: "AN", Data: &SymbolInfo{Symbol: "AN", Exchange: "3", Instrument: "A company"}}
	ap := SearchItem[*SymbolInfo]{Key: "AP", Data: &SymbolInfo{Symbol: "AP", Exchange: "3", Instrument: "APPROVE company"}}
	ag := SearchItem[*SymbolInfo]{Key: "AG", Data: &SymbolInfo{Symbol: "AG", Exchange: "3", Instrument: "AG company"}}
	az := SearchItem[*SymbolInfo]{Key: "AZ", Data: &SymbolInfo{Symbol: "AZ", Exchange: "3"}}
	b1 := SearchItem[*SymbolInfo]{Key: "B", Data: &SymbolInfo{Symbol: "B", Exchange: "2", Instrument: "Company Betta"}}
	aa11 := SearchItem[*SymbolInfo]{Key: "AA", Data: &SymbolInfo{Symbol: "AA", Instrument: "HA"}}
	a2 := SearchItem[*SymbolInfo]{Key: "A2", Data: &SymbolInfo{Symbol: "A2"}}

	var searchList SearchList[*SymbolInfo]
	searchList = append(searchList, &ag)
	searchList = append(searchList, &an)
	searchList = append(searchList, &ap)
	searchList = append(searchList, &az)
	searchList = append(searchList, &a11)
	searchList = append(searchList, &b1)
	searchList = append(searchList, &a13)
	searchList = append(searchList, &a12)
	searchList = append(searchList, &aa11)
	searchList = append(searchList, &a2)

	type TestData struct {
		Search   string
		Result   []*SymbolInfo
		Limit    int
		PageSize int
		Sort     func(i, j int, data SearchList[*SymbolInfo]) bool
	}

	data := []*TestData{
		{
			Search:   "AA",
			Result:   []*SymbolInfo{aa11.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     sortFunc,
		},
		{
			Search:   "B",
			Result:   []*SymbolInfo{b1.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     nil,
		},
		{
			Search:   "A",
			Result:   []*SymbolInfo{a11.Data, a12.Data, a13.Data, a2.Data, aa11.Data, ag.Data, an.Data, ap.Data, az.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     sortFunc,
		},
		{
			Search:   "A",
			Result:   []*SymbolInfo{a11.Data, a12.Data},
			Limit:    100,
			PageSize: 2,
			Sort:     sortFunc,
		},
		{
			Search:   "A",
			Result:   []*SymbolInfo{a11.Data, a12.Data, a13.Data, a2.Data, aa11.Data, ag.Data, an.Data, ap.Data},
			Limit:    100,
			PageSize: 8,
			Sort:     sortFunc,
		},
		{
			Search:   "B",
			Result:   []*SymbolInfo{b1.Data},
			Limit:    100,
			PageSize: 2,
			Sort:     sortFunc,
		},
	}

	for _, item := range data {
		searchIndex := NewSearchIndex(searchList, item.Limit, item.Sort, nil, true, nil)
		result := searchIndex.Search(SearchParams[*SymbolInfo]{Text: item.Search, OutputSize: item.PageSize, Matching: Beginning})

		assert.Equal(t, item.Result, result)
	}
}
