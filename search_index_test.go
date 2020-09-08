package searchindex

import (
	"fmt"
	"reflect"
	"testing"
)

type SymbolInfo struct {
	Symbol     string
	Exchange   string
	Instrument string
}

func sortFunc(i, j int, data interface{}) bool {
	if data.(SearchList)[i].Key == data.(SearchList)[j].Key {
		return data.(SearchList)[i].Data.(*SymbolInfo).Exchange < data.(SearchList)[j].Data.(*SymbolInfo).Exchange
	}
	return data.(SearchList)[i].Key < data.(SearchList)[j].Key
}

func TestSearchIndex(t *testing.T) {

	a_1 := SearchItem{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "1", Instrument: "A company"}}
	a_2 := SearchItem{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "2", Instrument: "A company"}}
	a_3 := SearchItem{Key: "A", Data: &SymbolInfo{Symbol: "A", Exchange: "3", Instrument: "A company"}}
	an := SearchItem{Key: "AN", Data: &SymbolInfo{Symbol: "AN", Exchange: "3", Instrument: "A company"}}
	ap := SearchItem{Key: "AP", Data: &SymbolInfo{Symbol: "AP", Exchange: "3", Instrument: "APPROVE company"}}
	ag := SearchItem{Key: "AG", Data: &SymbolInfo{Symbol: "AG", Exchange: "3", Instrument: "AG company"}}
	az := SearchItem{Key: "AZ", Data: &SymbolInfo{Symbol: "AZ", Exchange: "3"}}
	b1 := SearchItem{Key: "B", Data: &SymbolInfo{Symbol: "B", Exchange: "2", Instrument: "Company Betta"}}
	aa_1 := SearchItem{Key: "AA", Data: &SymbolInfo{Symbol: "AA", Instrument: "HA"}}
	a2 := SearchItem{Key: "A2", Data: &SymbolInfo{Symbol: "A2"}}

	var searchList SearchList
	searchList = append(searchList, &ag)
	searchList = append(searchList, &an)
	searchList = append(searchList, &ap)
	searchList = append(searchList, &az)
	searchList = append(searchList, &a_1)
	searchList = append(searchList, &b1)
	searchList = append(searchList, &a_3)
	searchList = append(searchList, &a_2)
	searchList = append(searchList, &aa_1)
	searchList = append(searchList, &a2)

	type TestData struct {
		Search   string
		Result   []SearchData
		Limit    int
		PageSize int
		Sort     func(i, j int, data interface{}) bool
	}

	data := []*TestData{
		{
			Search:   "AA",
			Result:   []SearchData{aa_1.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     sortFunc,
		},
		{
			Search:   "B",
			Result:   []SearchData{b1.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     nil,
		},
		{
			Search:   "A",
			Result:   []SearchData{a_1.Data, a_2.Data, a_3.Data, a2.Data, aa_1.Data, ag.Data, an.Data, ap.Data, az.Data},
			Limit:    100,
			PageSize: 10,
			Sort:     sortFunc,
		},
		{
			Search:   "A",
			Result:   []SearchData{a_1.Data, a_2.Data},
			Limit:    100,
			PageSize: 2,
			Sort:     sortFunc,
		},
		{
			Search:   "A",
			Result:   []SearchData{a_1.Data, a_2.Data, a_3.Data, a2.Data, aa_1.Data, ag.Data, an.Data, ap.Data},
			Limit:    100,
			PageSize: 8,
			Sort:     sortFunc,
		},
		{
			Search:   "B",
			Result:   []SearchData{b1.Data},
			Limit:    100,
			PageSize: 2,
			Sort:     sortFunc,
		},
	}

	for index, item := range data {
		searchIndex := NewSearchIndex(searchList, item.Limit, item.Sort, nil, true, nil)

		result := searchIndex.Search(SearchParams{Text: item.Search, OutputSize: item.PageSize, Matching: Beginning})

		if !reflect.DeepEqual(item.Result, result) {
			expected := ""
			for _, elem := range item.Result {
				expected += fmt.Sprintf("%v ", elem)
			}

			actual := ""
			for _, elem := range result {
				actual += fmt.Sprintf("%v ", elem)
			}
			t.Errorf("Test %d failed (TestSearchIndex).\nExpected: %v\nActual:   %v", index + 1, expected, actual)
		}
	}

}
