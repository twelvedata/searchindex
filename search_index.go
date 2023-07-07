package searchindex

import (
	"reflect"
	"regexp"
	s "sort"
	"strings"
	"unicode"

	"github.com/iancoleman/orderedmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type SearchIndexInterface[T any] interface {
	AppendData(data SearchList[T])
	Search(params SearchParams[T]) []T
}

type SearchIndex[T any] struct {
	SearchIndexInterface[T]
	index          Index[T]
	limit          int
	preprocessFunc func(key string, stopWords map[string]bool) []string
	sortFunc       func(i, j int, data SearchList[T]) bool
	indexParts     bool
	stopWords      map[string]bool
}

type Index[T any] struct {
	children *orderedmap.OrderedMap
	key      string
	data     SearchList[T]
}

const (
	Strict    = iota
	Beginning = iota
)

type SearchParams[T any] struct {
	Text        string
	OutputSize  int
	Matching    int
	StartValues []T
}

type SearchItem[T any] struct {
	Key  string
	Data T
}
type SearchList[T any] []*SearchItem[T]

func defaultSortFunc[T any](i, j int, data SearchList[T]) bool {
	return data[i].Key < data[j].Key
}

func defaultPreprocessFunc(key string, stopWords map[string]bool) []string {
	// Replace punctuation to spaces
	rePunctuation := regexp.MustCompile("[`'\".,:;\\?!+\\-–*=<>_~@#№$%^&()|/\\\\]")
	processed := rePunctuation.ReplaceAllString(key, " ")

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

func NewSearchIndex[T any](
	data SearchList[T],
	limit int,
	sort func(i, j int, data SearchList[T]) bool,
	preprocess func(key string, stopWords map[string]bool) []string,
	indexParts bool,
	stopWords []string,
) SearchIndexInterface[T] {
	preprocessFunc := preprocess
	if preprocessFunc == nil {
		preprocessFunc = defaultPreprocessFunc
	}

	sortFunc := defaultSortFunc[T]
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
	searchIndex := &SearchIndex[T]{
		index: Index[T]{
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

func (c SearchIndex[T]) AppendData(data SearchList[T]) {
	// Copy original data
	copied := copyOriginalData(data)

	// Preprocess keys
	var preprocessed SearchList[T]
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
			itemsByKey.Set(item.Key, SearchList[T]{item})
		} else {
			current = append(current.(SearchList[T]), item)
			itemsByKey.Set(item.Key, current)
		}
	}

	for _, key := range itemsByKey.Keys() {
		item, _ := itemsByKey.Get(key)
		addToIndex(&c.index, key, key, item.(SearchList[T]))
	}
}

func copyOriginalData[T any](data SearchList[T]) SearchList[T] {
	copied := make(SearchList[T], len(data))
	for i, _ := range data {
		d := *data[i]
		copied[i] = &d
	}
	return copied
}

func addToIndex[T any](index *Index[T], keyTail string, key string, data SearchList[T]) {
	if len(keyTail) == 0 {
		index.key = key
		index.data = data
		return
	}
	first := keyTail[:1]
	tail := keyTail[1:]
	idx, ok := index.children.Get(first)
	if !ok {
		idx = &Index[T]{
			children: orderedmap.New(),
		}
		index.children.Set(first, idx)
	}
	addToIndex(idx.(*Index[T]), tail, key, data)
}

func (c SearchIndex[T]) Search(params SearchParams[T]) []T {
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
	result := make([]T, len(params.StartValues))
	copy(result, params.StartValues)
	result = append(result, data...)

	return result
}

func (c SearchIndex[T]) searchInIndex(index *Index[T], key string, matching int, outputSize int, start map[uintptr]bool) []T {
	if key == "" {
		found := make(map[uintptr]bool)
		searched := c.searchList(index, make(SearchList[T], 0), matching, outputSize, found, start)
		return c.getData(searched)
	}
	idx, ok := index.children.Get(key[:1])
	if !ok {
		return make([]T, 0)
	}
	return c.searchInIndex(idx.(*Index[T]), key[1:], matching, outputSize, start)
}

func (c SearchIndex[T]) searchList(index *Index[T], items SearchList[T], matching int, outputSize int, found map[uintptr]bool, start map[uintptr]bool) SearchList[T] {
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
			items = c.searchList(idx.(*Index[T]), items, matching, outputSize, found, start)
		}
	}
	return items
}

func (c SearchIndex[T]) getData(data SearchList[T]) []T {
	result := make([]T, len(data))
	for i, item := range data {
		result[i] = item.Data
	}
	return result
}
