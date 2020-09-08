# In-memory search index

## How to use

```go
import "github.com/twelvedata/searchindex"

// Values for indexation
searchList := SearchList{
    SearchItem{
        Key: "AAPL", 
        Data: &SymbolInfo{Symbol: "AAPL", Exchange: "NASDAQ", Instrument: "Apple Inc"},
    },
    SearchItem{
        Key: "AMZN", 
        Data: &SymbolInfo{Symbol: "AMZN", Exchange: "NASDAQ", Instrument: "Amazon.com Inc"},
    },
}

// Fill index
searchIndex := NewSearchIndex(searchList, 10, nil, nil, true, nil)

// Search
result := searchIndex.Search(SearchParams{
    Text: "aa", 
    OutputSize: 10, 
    Matching: searchindex.Beginning,
})
```

Run tests:

```bash
make test
```
