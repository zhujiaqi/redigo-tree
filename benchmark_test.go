package redigotree

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Suppress log output during benchmarks
	log.SetLevel(log.PanicLevel)
}

// Benchmark TInsert operations
func BenchmarkTInsert_Root(b *testing.B) {
	defer TDestroy("benchmark_insert", "root")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TInsert("benchmark_insert", "root", fmt.Sprintf("node_%d", i), nil)
	}
}

func BenchmarkTInsert_WithIndex(b *testing.B) {
	defer TDestroy("benchmark_insert_index", "root")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TInsert("benchmark_insert_index", "root", fmt.Sprintf("node_%d", i), map[string]string{"index": "0"})
	}
}

func BenchmarkTInsert_Nested(b *testing.B) {
	// Setup: create a 3-level tree
	TInsert("benchmark_nested", "root", "level1", nil)
	TInsert("benchmark_nested", "level1", "level2", nil)
	defer TDestroy("benchmark_nested", "root")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TInsert("benchmark_nested", "level2", fmt.Sprintf("leaf_%d", i), nil)
	}
}

// Benchmark TChildren operations
func BenchmarkTChildren_Shallow(b *testing.B) {
	// Setup: create a flat tree with 100 children
	TInsert("benchmark_children", "root", "root_node", nil)
	for i := 0; i < 100; i++ {
		TInsert("benchmark_children", "root_node", fmt.Sprintf("child_%d", i), nil)
	}
	defer TDestroy("benchmark_children", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TChildren("benchmark_children", "root_node", nil)
	}
}

func BenchmarkTChildren_Deep(b *testing.B) {
	// Setup: create a deep tree
	TInsert("benchmark_deep", "root", "root_node", nil)
	current := "root_node"
	for i := 0; i < 10; i++ {
		nodeName := fmt.Sprintf("level_%d", i)
		TInsert("benchmark_deep", current, nodeName, nil)
		current = nodeName
	}
	defer TDestroy("benchmark_deep", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TChildren("benchmark_deep", "level_5", nil)
	}
}

func BenchmarkTChildren_WithLevel(b *testing.B) {
	// Setup: create a multi-level tree
	TInsert("benchmark_level", "root", "root_node", nil)
	for i := 0; i < 50; i++ {
		TInsert("benchmark_level", "root_node", fmt.Sprintf("child_%d", i), nil)
		TInsert("benchmark_level", fmt.Sprintf("child_%d", i), fmt.Sprintf("grandchild_%d", i), nil)
	}
	defer TDestroy("benchmark_level", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TChildren("benchmark_level", "root_node", map[string]string{"level": "1"})
	}
}

// Benchmark TParents operations
func BenchmarkTParents_Shallow(b *testing.B) {
	// Setup: create a 2-level tree
	TInsert("benchmark_parents", "root", "root_node", nil)
	TInsert("benchmark_parents", "root_node", "leaf", nil)
	defer TDestroy("benchmark_parents", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TParents("benchmark_parents", "leaf")
	}
}

func BenchmarkTParents_Deep(b *testing.B) {
	// Setup: create a deep tree (10 levels)
	TInsert("benchmark_parents_deep", "root", "root_node", nil)
	current := "root_node"
	for i := 0; i < 10; i++ {
		nodeName := fmt.Sprintf("level_%d", i)
		TInsert("benchmark_parents_deep", current, nodeName, nil)
		current = nodeName
	}
	defer TDestroy("benchmark_parents_deep", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TParents("benchmark_parents_deep", "level_9")
	}
}

// Benchmark TPath operations
func BenchmarkTPath_Shallow(b *testing.B) {
	// Setup: create a flat tree
	TInsert("benchmark_path", "root", "root_node", nil)
	for i := 0; i < 20; i++ {
		TInsert("benchmark_path", "root_node", fmt.Sprintf("node_%d", i), nil)
	}
	defer TDestroy("benchmark_path", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TPath("benchmark_path", "root_node", "node_10")
	}
}

func BenchmarkTPath_Deep(b *testing.B) {
	// Setup: create a deep tree
	TInsert("benchmark_path_deep", "root", "root_node", nil)
	current := "root_node"
	for i := 0; i < 15; i++ {
		nodeName := fmt.Sprintf("level_%d", i)
		TInsert("benchmark_path_deep", current, nodeName, nil)
		current = nodeName
	}
	defer TDestroy("benchmark_path_deep", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TPath("benchmark_path_deep", "root_node", "level_14")
	}
}

// Benchmark TExists operations
func BenchmarkTExists_Exists(b *testing.B) {
	// Setup: create a tree
	TInsert("benchmark_exists", "root", "root_node", nil)
	TInsert("benchmark_exists", "root_node", "node", nil)
	defer TDestroy("benchmark_exists", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TExists("benchmark_exists", "node")
	}
}

func BenchmarkTExists_NotExists(b *testing.B) {
	// Setup: create a tree
	TInsert("benchmark_exists_not", "root", "root_node", nil)
	defer TDestroy("benchmark_exists_not", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TExists("benchmark_exists_not", "nonexistent")
	}
}

// Benchmark TRemove operations
func BenchmarkTRem_Single(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create a tree for each iteration
		TInsert("benchmark_rem", "root", "root_node", nil)
		for j := 0; j < 10; j++ {
			TInsert("benchmark_rem", "root_node", fmt.Sprintf("child_%d", j), nil)
		}
		b.StartTimer()

		TRem("benchmark_rem", "root_node", 1, "child_5")

		b.StopTimer()
		TDestroy("benchmark_rem", "root_node")
		b.StartTimer()
	}
}

// Benchmark TDestroy operations
func BenchmarkTDestroy_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create a small tree
		TInsert("benchmark_destroy_small", "root", "root_node", nil)
		for j := 0; j < 10; j++ {
			TInsert("benchmark_destroy_small", "root_node", fmt.Sprintf("child_%d", j), nil)
		}
		b.StartTimer()

		TDestroy("benchmark_destroy_small", "root_node")
	}
}

func BenchmarkTDestroy_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create a medium tree
		TInsert("benchmark_destroy_medium", "root", "root_node", nil)
		for j := 0; j < 50; j++ {
			TInsert("benchmark_destroy_medium", "root_node", fmt.Sprintf("child_%d", j), nil)
			for k := 0; k < 5; k++ {
				TInsert("benchmark_destroy_medium", fmt.Sprintf("child_%d", j), fmt.Sprintf("grandchild_%d_%d", j, k), nil)
			}
		}
		b.StartTimer()

		TDestroy("benchmark_destroy_medium", "root_node")
	}
}

func BenchmarkTDestroy_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create a large tree
		TInsert("benchmark_destroy_large", "root", "root_node", nil)
		for j := 0; j < 100; j++ {
			TInsert("benchmark_destroy_large", "root_node", fmt.Sprintf("child_%d", j), nil)
			for k := 0; k < 10; k++ {
				TInsert("benchmark_destroy_large", fmt.Sprintf("child_%d", j), fmt.Sprintf("grandchild_%d_%d", j, k), nil)
			}
		}
		b.StartTimer()

		TDestroy("benchmark_destroy_large", "root_node")
	}
}

// Benchmark TRename operations
func BenchmarkTRename(b *testing.B) {
	// Setup: create a tree
	TInsert("benchmark_rename", "root", "root_node", nil)
	TInsert("benchmark_rename", "root_node", "node", nil)
	defer TDestroy("benchmark_rename", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Rename back and forth
		if i%2 == 0 {
			TRename("benchmark_rename", "node", "node_new")
			b.StartTimer()
		} else {
			TRename("benchmark_rename", "node_new", "node")
			b.StartTimer()
		}
	}
}

// Benchmark TPrune operations
func BenchmarkTPrune(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create a tree
		TInsert("benchmark_prune", "root", "root_node", nil)
		TInsert("benchmark_prune", "root_node", "node", nil)
		TInsert("benchmark_prune", "node", "child", nil)
		b.StartTimer()

		TPrune("benchmark_prune", "node")

		b.StopTimer()
		TDestroy("benchmark_prune", "root_node")
		b.StartTimer()
	}
}

// Benchmark TMoveChildren operations
func BenchmarkTMoveChildren_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: create two trees
		TInsert("benchmark_move", "root1", "root1", nil)
		TInsert("benchmark_move", "root2", "root2", nil)
		for j := 0; j < 10; j++ {
			TInsert("benchmark_move", "root1", fmt.Sprintf("child_%d", j), nil)
		}
		b.StartTimer()

		TMoveChildren("benchmark_move", "root1", "root2", "APPEND")

		b.StopTimer()
		TDestroy("benchmark_move", "root1")
		TDestroy("benchmark_move", "root2")
		b.StartTimer()
	}
}

// Benchmark concurrent operations (simulated)
func BenchmarkTInsert_Concurrent(b *testing.B) {
	defer TDestroy("benchmark_concurrent", "root_node")

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			nodeName := fmt.Sprintf("node_%d_%d", i, b.N)
			TInsert("benchmark_concurrent", "root_node", nodeName, nil)
			i++
		}
	})
}

// Benchmark large tree queries
func BenchmarkTChildren_LargeTree(b *testing.B) {
	// Setup: create a large tree with 1000 nodes
	TInsert("benchmark_large", "root", "root_node", nil)
	for i := 0; i < 100; i++ {
		TInsert("benchmark_large", "root_node", fmt.Sprintf("branch_%d", i), nil)
		for j := 0; j < 10; j++ {
			TInsert("benchmark_large", fmt.Sprintf("branch_%d", i), fmt.Sprintf("leaf_%d_%d", i, j), nil)
		}
	}
	defer TDestroy("benchmark_large", "root_node")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TChildren("benchmark_large", "root_node", nil)
	}
}
