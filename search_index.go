package searchindex

import (
	"github.com/iancoleman/orderedmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"reflect"
	"regexp"
	s "sort"
	"strings"
	"unicode"
)

type SearchIndexInterface interface {
	AppendData(data SearchList)
	Search(params SearchParams) []SearchData
}

type SearchIndex struct {
	SearchIndexInterface
	index          Index
	limit          int
	preprocessFunc func(key string, stopWords map[string]bool) []string
	sortFunc       func(i, j int, data interface{}) bool
	indexParts     bool
	stopWords      map[string]bool
}

type Index struct {
	children *orderedmap.OrderedMap
	key      string
	data     SearchList
}

const (
	Strict    = iota
	Beginning = iota
)

type SearchParams struct {
	Text        string
	OutputSize  int
	Matching    int
	StartValues []SearchData
}

type SearchData interface{}

type SearchItem struct {
	Key  string
	Data SearchData
}
type SearchList []*SearchItem

func defaultSortFunc(i, j int, data interface{}) bool {
	return data.(SearchList)[i].Key < data.(SearchList)[j].Key
}

func defaultPreprocessFunc(key string, stopWords map[string]bool) []string {
	// Replace punctuation to spaces
	rePunctuation := regexp.MustCompile("[`'\".,:;\\?!+\\-–*=<>_~@#№$%^&()|/\\\\]")
	// By default we remove special symbols, because we need searches BTCUSD and BTC-USD get BTC/USD key as result
	processed := rePunctuation.ReplaceAllString(key, "")

	// Replace double spaces to single space
	reSpaces := regexp.MustCompile("\\s+")
	processed = reSpaces.ReplaceAllString(processed, " ")

	processed = strings.Trim(processed, " ")
	processed = strings.ToLower(processed)

	// Replace "São, Österreich" to "Sao, Osterreich"
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	processed, _, _ = transform.String(t, processed)

	parts := strings.Split(processed, " ")

	// Exclude stop words
	var result []string
	for _, part := range parts {
		if _, ok := stopWords[part]; !ok {
			result = append(result, part)
		}
	}

	return result
}

func NewSearchIndex(
	data SearchList,
	limit int,
	sort func(i, j int, data interface{}) bool,
	preprocess func(key string, stopWords map[string]bool) []string,
	indexParts bool,
	stopWords []string,
) SearchIndexInterface {
	preprocessFunc := preprocess
	if preprocessFunc == nil {
		preprocessFunc = defaultPreprocessFunc
	}

	sortFunc := defaultSortFunc
	if sort != nil {
		sortFunc = sort
	}

	// Prepare stop words
	sw := make(map[string]bool)
	for _, word := range stopWords {
		parts := preprocessFunc(word, make(map[string]bool))
		for _, part := range parts {
			sw[part] = true
		}
	}

	// Create and fill index with initial data
	searchIndex := &SearchIndex{
		index: Index{
			children: orderedmap.New(),
		},
		limit:          limit,
		preprocessFunc: preprocessFunc,
		sortFunc:       sortFunc,
		indexParts:     indexParts,
		stopWords:      sw,
	}
	searchIndex.AppendData(data)

	return searchIndex
}

func (c SearchIndex) AppendData(data SearchList) {
	// Copy original data
	copied := copyOriginalData(data)

	// Preprocess keys
	var preprocessed SearchList
	for _, item := range copied {
		sortedParts := c.preprocessFunc(item.Key, c.stopWords)
		for j, _ := range sortedParts {
			d := *item
			copiedItem := &d
			copiedItem.Key = strings.Join(sortedParts[j:], " ")
			preprocessed = append(preprocessed, copiedItem)
			if !c.indexParts {
				break
			}
		}
	}

	// Sort
	s.SliceStable(preprocessed, func(i, j int) bool {
		return c.sortFunc(i, j, preprocessed)
	})

	// Group by key
	itemsByKey := orderedmap.New()
	for _, item := range preprocessed {
		current, ok := itemsByKey.Get(item.Key)
		if !ok {
			itemsByKey.Set(item.Key, SearchList{item})
		} else {
			current = append(current.(SearchList), item)
			itemsByKey.Set(item.Key, current)
		}
	}

	for _, key := range itemsByKey.Keys() {
		item, _ := itemsByKey.Get(key)
		addToIndex(&c.index, key, key, item.(SearchList))
	}
}

func copyOriginalData(data SearchList) SearchList {
	copied := make(SearchList, len(data))
	for i, _ := range data {
		d := *data[i]
		copied[i] = &d
	}
	return copied
}

func addToIndex(index *Index, keyTail string, key string, data SearchList) {
	if len(keyTail) == 0 {
		index.key = key
		index.data = data
		return
	}
	first := keyTail[:1]
	tail := keyTail[1:]
	idx, ok := index.children.Get(first)
	if !ok {
		idx = &Index{
			children: orderedmap.New(),
		}
		index.children.Set(first, idx)
	}
	addToIndex(idx.(*Index), tail, key, data)
}

func (c SearchIndex) Search(params SearchParams) []SearchData {
	outputSize := params.OutputSize
	if outputSize == 0 || outputSize > c.limit || outputSize <= 0 {
		outputSize = c.limit
	}

	start := make(map[uintptr]bool)
	for _, item := range params.StartValues {
		ptr := reflect.ValueOf(item).Pointer()
		start[ptr] = true
	}

	// Start search
	data := c.searchInIndex(
		&c.index,
		strings.Join(c.preprocessFunc(params.Text, c.stopWords), " "),
		params.Matching,
		outputSize-len(params.StartValues),
		start,
	)

	// And append result after start
	result := make([]SearchData, len(params.StartValues))
	copy(result, params.StartValues)
	result = append(result, data...)

	return result
}

func (c SearchIndex) searchInIndex(index *Index, key string, matching int, outputSize int, start map[uintptr]bool) []SearchData {
	if key == "" {
		found := make(map[uintptr]bool)
		searched := c.searchList(index, make(SearchList, 0), matching, outputSize, found, start)
		return c.getData(searched)
	}
	idx, ok := index.children.Get(key[:1])
	if !ok {
		return make([]SearchData, 0)
	}
	return c.searchInIndex(idx.(*Index), key[1:], matching, outputSize, start)
}

func (c SearchIndex) searchList(index *Index, items SearchList, matching int, outputSize int, found map[uintptr]bool, start map[uintptr]bool) SearchList {
	if (outputSize > 0 && len(items) >= outputSize) || outputSize == 0 {
		return items
	}
	if index.data != nil {
		for _, item := range index.data {
			// Check data in found, because we do not need to add duplicates in result
			ptr := reflect.ValueOf(item.Data).Pointer()
			if _, exists := found[ptr]; !exists {
				if _, exists := start[ptr]; !exists {
					items = append(items, item)
					found[ptr] = true
					if outputSize > 0 && len(items) >= outputSize {
						return items
					}
				}
			}
		}
	}
	if len(index.children.Keys()) == 0 {
		return items
	}
	if matching == Beginning {
		for _, key := range index.children.Keys() {
			idx, _ := index.children.Get(key)
			items = c.searchList(idx.(*Index), items, matching, outputSize, found, start)
		}
	}
	return items
}

func (c SearchIndex) getData(data SearchList) []SearchData {
	result := make([]SearchData, len(data))
	for i, item := range data {
		result[i] = item.Data
	}
	return result
}
