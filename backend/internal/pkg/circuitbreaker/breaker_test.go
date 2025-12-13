package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

var errTest = errors.New("test error")

func TestNew(t *testing.T) {
	cb := New(5, time.Second)

	if cb == nil {
		t.Fatal("New() should not return nil")
	}

	if cb.threshold != 5 {
		t.Errorf("threshold = %d, want 5", cb.threshold)
	}

	if cb.timeout != time.Second {
		t.Errorf("timeout = %v, want %v", cb.timeout, time.Second)
	}

	if cb.state != StateClosed {
		t.Errorf("initial state = %v, want StateClosed", cb.state)
	}
}

func TestExecute_Success(t *testing.T) {
	cb := New(3, time.Second)

	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestExecute_SingleFailure(t *testing.T) {
	cb := New(3, time.Second)

	err := cb.Execute(func() error {
		return errTest
	})

	if err != errTest {
		t.Errorf("Execute() error = %v, want %v", err, errTest)
	}

	// 狀態應該仍然是 Closed
	if cb.state != StateClosed {
		t.Errorf("state = %v, want StateClosed after single failure", cb.state)
	}
}

func TestExecute_OpenAfterThreshold(t *testing.T) {
	cb := New(3, time.Second)

	// 連續失敗 3 次
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errTest
		})
	}

	// 狀態應該是 Open
	if cb.state != StateOpen {
		t.Errorf("state = %v, want StateOpen after threshold failures", cb.state)
	}

	// 下一次執行應該直接返回 ErrCircuitOpen
	err := cb.Execute(func() error {
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Execute() error = %v, want ErrCircuitOpen", err)
	}
}

func TestExecute_HalfOpenAfterTimeout(t *testing.T) {
	cb := New(3, 50*time.Millisecond)

	// 使電路斷開
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errTest
		})
	}

	// 等待超時
	time.Sleep(100 * time.Millisecond)

	// 下一次執行應該進入 HalfOpen 狀態
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// 成功後狀態應該是 Closed
	if cb.state != StateClosed {
		t.Errorf("state = %v, want StateClosed after successful execution in HalfOpen", cb.state)
	}
}

func TestExecute_ReopenAfterHalfOpenFailure(t *testing.T) {
	cb := New(3, 50*time.Millisecond)

	// 使電路斷開
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errTest
		})
	}

	// 等待超時
	time.Sleep(100 * time.Millisecond)

	// 在 HalfOpen 狀態下失敗
	_ = cb.Execute(func() error {
		return errTest
	})

	// 狀態應該回到 Open
	if cb.state != StateOpen {
		t.Errorf("state = %v, want StateOpen after HalfOpen failure", cb.state)
	}
}

func TestExecute_ResetAfterSuccess(t *testing.T) {
	cb := New(3, time.Second)

	// 失敗 2 次
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errTest
		})
	}

	// 成功 1 次
	_ = cb.Execute(func() error {
		return nil
	})

	// 失敗計數應該被重置
	if cb.failures != 0 {
		t.Errorf("failures = %d, want 0 after success", cb.failures)
	}
}

func TestExecute_Concurrent(t *testing.T) {
	cb := New(100, time.Second)
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	// 同時執行 1000 個請求
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := cb.Execute(func() error {
				if i%2 == 0 {
					return errTest
				}
				return nil
			})

			mu.Lock()
			if err == nil {
				successCount++
			} else if err == errTest {
				failCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// 確保沒有 panic 且計數合理
	total := successCount + failCount
	if total < 900 {
		t.Errorf("total executed = %d, expected at least 900", total)
	}
}

func TestState_Constants(t *testing.T) {
	if StateClosed != 0 {
		t.Errorf("StateClosed = %d, want 0", StateClosed)
	}

	if StateOpen != 1 {
		t.Errorf("StateOpen = %d, want 1", StateOpen)
	}

	if StateHalfOpen != 2 {
		t.Errorf("StateHalfOpen = %d, want 2", StateHalfOpen)
	}
}

func BenchmarkExecute_Success(b *testing.B) {
	cb := New(5, time.Second)

	for i := 0; i < b.N; i++ {
		_ = cb.Execute(func() error {
			return nil
		})
	}
}

func BenchmarkExecute_Concurrent(b *testing.B) {
	cb := New(1000, time.Second)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Execute(func() error {
				return nil
			})
		}
	})
}
