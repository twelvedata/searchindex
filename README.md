# In-memory search index

![](illustration.png)

## Features

- Indexation using simple tree
- Search by beginning of string
- Exact search
- Suitable for non-long strings

More about in [article on Medium](https://medium.com/twelve-data/in-memory-text-search-index-for-quotes-on-go-5243adc62c26)

## How to use

```go
package main

import (
    "fmt"
    "github.com/twelvedata/searchindex"
)

type SymbolInfo struct {
    Symbol     string
    Exchange   string
    Instrument string
}

func main() {
    // Values for indexation
    searchList := searchindex.SearchList[*SymbolInfo]{
        &searchindex.SearchItem[*SymbolInfo]{
            Key: "AAPL",
            Data: &SymbolInfo{Symbol: "AAPL", Exchange: "NASDAQ", Instrument: "Apple Inc"},
        },
        &searchindex.SearchItem[*SymbolInfo]{
            Key: "AMZN",
            Data: &SymbolInfo{Symbol: "AMZN", Exchange: "NASDAQ", Instrument: "Amazon.com Inc"},
        },
    }

    // Fill index
    searchIndex := searchindex.NewSearchIndex(searchList, 10, nil, nil, true, nil)

    // Search
    result := searchIndex.Search(searchindex.SearchParams[*SymbolInfo]{
        Text: "aa",
        OutputSize: 10,
        Matching: searchindex.Beginning,
    })

    fmt.Println(result[0])
}
```

Run tests:

```bash
make test
```
