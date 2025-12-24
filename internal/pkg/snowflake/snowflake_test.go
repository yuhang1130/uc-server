package snowflake

import (
	"sync"
	"testing"
)

// TestNewGenerator 测试创建生成器
func TestNewGenerator(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	if gen == nil {
		t.Fatal("Generator is nil")
	}
}

// TestGenerate 测试生成单个 ID
func TestGenerate(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.Generate()
	if id == 0 {
		t.Fatal("Generated ID is 0")
	}
	t.Logf("Generated ID: %d", id)
}

// TestGenerateUint 测试生成 uint64 ID
func TestGenerateUint(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.GenerateUint()
	if id == 0 {
		t.Fatal("Generated uint64 ID is 0")
	}
	t.Logf("Generated uint64 ID: %d", id)
}

// TestGenerateUnique 测试生成的 ID 唯一性
func TestGenerateUnique(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	ids := make(map[int64]bool)
	count := 10000

	for i := 0; i < count; i++ {
		id := gen.Generate()
		if ids[id] {
			t.Fatalf("Duplicate ID found: %d", id)
		}
		ids[id] = true
	}

	t.Logf("Generated %d unique IDs", count)
}

// TestGenerateConcurrent 测试并发生成 ID
func TestGenerateConcurrent(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := make(map[int64]bool)
	goroutines := 100
	idsPerGoroutine := 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := gen.Generate()
				mu.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID found in concurrent test: %d", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	expectedCount := goroutines * idsPerGoroutine
	if len(ids) != expectedCount {
		t.Fatalf("Expected %d unique IDs, got %d", expectedCount, len(ids))
	}
	t.Logf("Generated %d unique IDs concurrently", len(ids))
}

// TestGenerateBatch 测试批量生成 ID
func TestGenerateBatch(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	count := 100
	ids := gen.GenerateBatch(count)
	if len(ids) != count {
		t.Fatalf("Expected %d IDs, got %d", count, len(ids))
	}

	// 检查唯一性
	unique := make(map[int64]bool)
	for _, id := range ids {
		if unique[id] {
			t.Fatalf("Duplicate ID found in batch: %d", id)
		}
		unique[id] = true
	}
	t.Logf("Generated %d unique IDs in batch", count)
}

// TestParseID 测试解析 ID
func TestParseID(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.Generate()
	info := gen.ParseID(id)

	if info.ID != id {
		t.Fatalf("Parsed ID mismatch: expected %d, got %d", id, info.ID)
	}
	if info.Node != 1 {
		t.Fatalf("Parsed Node mismatch: expected 1, got %d", info.Node)
	}
	t.Logf("Parsed ID: %s", info.String())
}

// TestDefaultGenerator 测试默认生成器
func TestDefaultGenerator(t *testing.T) {
	err := InitDefault(1)
	if err != nil {
		t.Fatalf("Failed to init default generator: %v", err)
	}

	id := Generate()
	if id == 0 {
		t.Fatal("Generated ID is 0")
	}
	t.Logf("Generated ID from default generator: %d", id)
}

// TestGenerateString 测试生成字符串 ID
func TestGenerateString(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	idStr := gen.GenerateString()
	if idStr == "" {
		t.Fatal("Generated string ID is empty")
	}
	t.Logf("Generated string ID: %s", idStr)
}

// BenchmarkGenerate 性能测试
func BenchmarkGenerate(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Generate()
	}
}

// BenchmarkGenerateConcurrent 并发性能测试
func BenchmarkGenerateConcurrent(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = gen.Generate()
		}
	})
}
