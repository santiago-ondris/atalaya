package poller

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type sourceStub struct {
	batch  domain.EventBatch
	cursor domain.Cursor
}

func (source *sourceStub) FetchEvents(_ context.Context, cursor domain.Cursor) (domain.EventBatch, error) {
	source.cursor = cursor
	return source.batch, nil
}

type storeStub struct {
	checkpoint domain.Cursor
	imported   domain.EventBatch
	attempts   int
}

func (store *storeStub) Checkpoint(context.Context, uuid.UUID) (domain.Cursor, error) {
	return store.checkpoint, nil
}
func (store *storeStub) ImportBatch(_ context.Context, _ uuid.UUID, batch domain.EventBatch) (int, error) {
	store.imported = batch
	return len(batch.Events), nil
}
func (store *storeStub) RecordAttempt(context.Context, uuid.UUID, error) error {
	store.attempts++
	return nil
}

func TestPollUsesCheckpointAndImportsBatch(t *testing.T) {
	source := &sourceStub{batch: domain.EventBatch{Events: []domain.Event{{SourceEventID: "event-1"}}, NextCursor: domain.Cursor{Value: "next"}}}
	storage := &storeStub{checkpoint: domain.Cursor{Value: "current"}}
	p := New(source, storage, uuid.New(), "prensap", "backend", time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))
	p.poll(context.Background())
	if source.cursor.Value != "current" {
		t.Fatalf("expected stored cursor, got %q", source.cursor.Value)
	}
	if len(storage.imported.Events) != 1 || storage.imported.NextCursor.Value != "next" {
		t.Fatalf("unexpected imported batch: %#v", storage.imported)
	}
	if storage.attempts != 0 {
		t.Fatalf("unexpected failed attempts: %d", storage.attempts)
	}
}
