package retention

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type stubStore struct {
	purged int64
	err    error
}

func (s *stubStore) PurgeOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	return s.purged, s.err
}

func TestRetentionWorkerPurgesEvents(t *testing.T) {
	store := &stubStore{purged: 42}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	worker := NewWorker(store, 90, 10*time.Millisecond, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	worker.Start(ctx)

	if store.purged != 42 {
		t.Fatalf("expected 42 purged events, got %d", store.purged)
	}
}

func TestRetentionWorkerDisabled(t *testing.T) {
	store := &stubStore{purged: 0}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	worker := NewWorker(store, 0, 10*time.Millisecond, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	worker.Start(ctx)
}
