#redigo-tree
---
Go Port of https://github.com/shimohq/ioredis-tree

## Install
---
```
go get -u github.com/kardianos/govendor
govendor sync -insecure -v
```

## Usage
---
```
import (
	tree "github.com/zhujiaqi/redigo-tree"
)

tree.TInsert("treename", "parent", "node", map[string]string{"index": "1000"})
```

## API & Examples

For complete API, check: https://github.com/shimohq/ioredis-tree

Please see redigotree\_test.go for examples. The tests are for reference only, they are not completed...for now.

To run the tests:    

```
go test
```
