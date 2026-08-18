package concurrency

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// TestPipelineNaoVazaGoroutines garante que toda goroutine iniciada pelos
// helpers termina, inclusive sob cancelamento no meio do trabalho.
func TestPipelineNaoVazaGoroutines(t *testing.T) {
	antes := runtime.NumGoroutine()
	for range 10 {
		ctx, cancel := context.WithCancel(t.Context())
		go func() { // cancela no meio da execução para exercitar o caminho difícil
			time.Sleep(time.Millisecond)
			cancel()
		}()
		_, _ = pipelineDeQuadrados(ctx, make([]int, 10000), 4, 2)
		cancel()
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if runtime.NumGoroutine() <= antes+2 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("goroutines vazaram: antes=%d depois=%d", antes, runtime.NumGoroutine())
}

// TestConsumidorLentoNaoAcumula garante backpressure: com buffers pequenos e
// consumidor lento, o pipeline inteiro bloqueia dentro dos limites em vez de
// crescer goroutines ou memória (cenário "Database stage slows down").
func TestConsumidorLentoNaoAcumula(t *testing.T) {
	antes := runtime.NumGoroutine()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = pipelineDeQuadrados(t.Context(), make([]int, 5000), 8, 1)
	}()
	time.Sleep(20 * time.Millisecond) // deixa o pipeline encher os buffers
	depois := runtime.NumGoroutine()
	// 8 workers + source + redutor + harness: manter folga pequena e fixa.
	if depois-antes > 16 {
		t.Fatalf("goroutines cresceram além do orçamento: antes=%d depois=%d", antes, depois)
	}
	<-done
}
