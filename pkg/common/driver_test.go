package common

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

// Regression: concurrent gRPC calls crashed the node server with
// "fatal error: concurrent map writes" in the routine-logging interceptor
// (shared routineTimers map, unsynchronized). Run with -race.
func TestLogRoutineInterceptorConcurrent(t *testing.T) {
	ic := NewLogRoutineServerInterceptor(func(string) bool { return true })
	info := &grpc.UnaryServerInfo{FullMethod: "/csi.v1.Node/NodeUnpublishVolume"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil }

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := ic(context.Background(), nil, info, handler); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
