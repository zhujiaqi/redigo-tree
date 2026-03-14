# redigo-tree

**TL;DR**: A Go-based library for Redis tree data structures, leveraging Redigo for high-concurrency backend systems and distributed hierarchical data management. Implements atomic tree operations via Lua scripting with messagepack serialization, enabling O(1) sibling insertions, O(log n) path traversal, and efficient subtree manipulation in a Redis-backed environment.

---

Go Port of https://github.com/shimohq/ioredis-tree

## Technique & Data Model

### Core Architecture

**redigo-tree** implements a Redis-backed tree data structure using Lua scripting for atomic operations. The library leverages Redis's single-threaded execution model to ensure consistency across tree manipulations without requiring distributed locks.

### Data Model

#### Key Structure
```
{tree_name}::{node_id}        → Node data (messagepack encoded)
{tree_name}::{node_id}::P     → Parent references (Redis Set)
```

#### Node Storage Format
Each node is stored as a messagepack-encoded array containing:
- `node_id`: Unique identifier for the node
- `hasChild`: Boolean flag indicating whether the node has children
- `children`: Array of child node metadata (node_id, hasChild, nested children)

#### Parent Index
A Redis Set storing all parent nodes for a given node, enabling efficient upward traversal (TParents operation).

### Lua Scripting Strategy

All tree operations are implemented as Lua scripts that execute atomically on the Redis server:

1. **tinsert**: Insert a node as a child of a parent (with optional index/position)
2. **tchildren**: Retrieve children of a node (with optional depth limit)
3. **tparents**: Traverse upward to get all ancestor nodes
4. **tpath**: Get the complete path from one node to another
5. **trem**: Remove a specific child from a parent
6. **tmrem**: Remove a node from all parents (multi-parent support)
7. **tdestroy**: Recursively delete a node and all descendants
8. **texists**: Check if a node exists
9. **trename**: Rename a node
10. **tprune**: Remove all children from a node
11. **tmovechildren**: Move all children from one node to another

### Atomicity Guarantees

By executing tree operations as Lua scripts, redigo-tree ensures:
- **Atomicity**: Each operation completes without interruption
- **Consistency**: Tree structure remains valid after each operation
- **Isolation**: Concurrent operations do not interfere with each other
- **No Distributed Locks**: Redis single-threaded execution provides natural locking

### Performance Characteristics

| Operation    | Time Complexity | Notes                              |
|--------------|-----------------|------------------------------------|
| TInsert      | O(1)*           | *Amortized, depends on sibling count |
| TChildren    | O(n)            | n = number of children             |
| TParents     | O(log n)        | n = tree depth                     |
| TPath        | O(log n)        | n = tree depth                     |
| TExists      | O(1)            | Direct key lookup                  |
| TRem         | O(n)            | n = sibling count                  |
| TDestroy     | O(n)            | n = subtree size                   |
| TRename      | O(1)            | Direct key rename                  |
| TPrune       | O(n)            | n = children count                 |
| TMoveChildren| O(n)            | n = children being moved           |

---

## Install

```bash
go get -u github.com/kardianos/govendor
govendor sync -insecure -v
```

Or with Go modules:
```bash
go mod init your-module
go get github.com/zhujiaqi/redigo-tree
```

---

## Usage

```go
import (
	tree "github.com/zhujiaqi/redigo-tree"
)

// Insert a node
tree.TInsert("treename", "parent", "node", map[string]string{"index": "1000"})

// Get children
children := tree.TChildren("treename", "parent", nil)

// Get parent hierarchy
parents := tree.TParents("treename", "node")

// Get path between nodes
path := tree.TPath("treename", "root", "target")

// Check existence
exists := tree.TExists("treename", "node")

// Remove a node
tree.TRem("treename", "parent", 1, "node")

// Destroy entire subtree
tree.TDestroy("treename", "root")
```

---

## Performance Benchmarks

**Environment**: Apple M4, Darwin ARM64, Redis local instance
**Benchmark Duration**: 50ms per test

### Insert Operations

| Benchmark              | Operations | Time per Op | Memory | Allocations |
|------------------------|-----------:|------------:|-------:|------------:|
| TInsert_Root           | 218        | 241,726 ns  | 5.47 KB| 14          |
| TInsert_WithIndex      | 609        | 286,335 ns  | 5.48 KB| 15          |
| TInsert_Nested         | 616        | 291,544 ns  | 5.47 KB| 15          |
| TInsert_Concurrent     | 1154       | 267,011 ns  | 6.22 KB| 17          |

### Query Operations

| Benchmark              | Operations | Time per Op | Memory | Allocations |
|------------------------|-----------:|------------:|-------:|------------:|
| TChildren_Shallow      | 362        | 153,877 ns  | 19.3 KB| 614         |
| TParents_Shallow       | 1431       | 40,186 ns   | 1.30 KB| 17          |
| TParents_Deep          | 1426       | 40,195 ns   | 1.28 KB| 17          |
| TPath_Shallow          | 1470       | 41,930 ns   | 3.16 KB| 13          |
| TPath_Deep             | 963        | 59,852 ns   | 4.17 KB| 57          |
| TExists_Exists         | 1468       | 38,658 ns   | 1.31 KB| 11          |
| TExists_NotExists      | 1486       | 39,255 ns   | 1.31 KB| 11          |

### Key Insights

1. **TInsert Performance**: Concurrent insertions show better throughput due to parallel execution, with only ~10% overhead compared to sequential inserts.

2. **TChildren Memory**: Higher memory allocation (614 allocs/op) is expected as it constructs TreeNode structures for all children. Consider using the `level` option to limit depth.

3. **TParents Efficiency**: Consistent ~40μs performance regardless of tree depth, making it suitable for deep hierarchy traversal.

4. **TExists Optimization**: Fastest query operation at ~39μs, ideal for existence checks before expensive operations.

5. **TPath Scaling**: Deep path traversal (15 levels) shows ~40% increase in latency compared to shallow trees, but remains efficient.

---

## API & Examples

For complete API reference, check: https://github.com/shimohq/ioredis-tree

Please see `redigotree_test.go` for working examples of all operations.

To run the tests:

```bash
go test
```

To run benchmarks:

```bash
go test -bench=. -benchmem
```

---

## License

See LICENSE file for details.
