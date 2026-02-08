package webhook

import (
	"fmt"
	"testing"
)

func TestNewEventStore(t *testing.T) {
	// Test in-memory database creation
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	// Verify table exists by querying
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 events, got %d", count)
	}
}

func TestRecordEvent_NewEvent(t *testing.T) {
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	isNew, err := store.RecordEvent("test-delivery-123", "latest", 675688875, "sha256:abc123", "test-repo")
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	if !isNew {
		t.Error("Expected event to be new")
	}

	// Verify it was recorded
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event, got %d", count)
	}
}

func TestRecordEvent_DuplicateEvent(t *testing.T) {
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	// First insert
	isNew, err := store.RecordEvent("test-delivery-456", "latest", 675688875, "sha256:abc123", "test-repo")
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}
	if !isNew {
		t.Error("Expected first event to be new")
	}

	// Second insert (same delivery ID)
	isNew, err = store.RecordEvent("test-delivery-456", "latest", 675688875, "sha256:abc123", "test-repo")
	if err != nil {
		t.Fatalf("Failed to record duplicate event: %v", err)
	}
	if isNew {
		t.Error("Expected event to be duplicate")
	}

	// Verify only one was recorded
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event, got %d", count)
	}
}

func TestRecordEvent_DifferentDeliveryIDs(t *testing.T) {
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	events := []struct {
		deliveryID string
		tag        string
		versionID  int64
		sha        string
		repository string
	}{
		{"delivery-1", "latest", 100, "sha256:aaa", "repo-a"},
		{"delivery-2", "latest", 200, "sha256:bbb", "repo-a"},
		{"delivery-3", "latest", 300, "sha256:ccc", "repo-b"},
	}

	for _, e := range events {
		isNew, err := store.RecordEvent(e.deliveryID, e.tag, e.versionID, e.sha, e.repository)
		if err != nil {
			t.Fatalf("Failed to record event %s: %v", e.deliveryID, err)
		}
		if !isNew {
			t.Errorf("Expected event %s to be new", e.deliveryID)
		}
	}

	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 events, got %d", count)
	}
}

func TestRecordEvent_ContentDedup(t *testing.T) {
	store, err := NewEventStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	// First insert
	isNew, err := store.RecordEvent("delivery-aaa", "latest", 675688875, "sha256:abc123", "test-repo")
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}
	if !isNew {
		t.Error("Expected first event to be new")
	}

	// Second insert — different delivery ID, same content
	isNew, err = store.RecordEvent("delivery-bbb", "latest", 675688875, "sha256:abc123", "test-repo")
	if err != nil {
		t.Fatalf("Failed to record content-duplicate event: %v", err)
	}
	if isNew {
		t.Error("Expected second event to be detected as content duplicate")
	}

	// Verify only one was recorded
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event, got %d", count)
	}
}

func TestRecordEvent_ConcurrentInserts(t *testing.T) {
	// Use file-based database for concurrent testing
	// :memory: doesn't work well with concurrent access
	tmpDB := t.TempDir() + "/test_db.sqlite"
	store, err := NewEventStore(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer store.Close()

	// Test concurrent inserts don't cause issues
	done := make(chan bool)
	successCount := make(chan int, 5)

	// Try to insert same content concurrently with different delivery IDs
	for i := 0; i < 5; i++ {
		go func(idx int) {
			isNew, err := store.RecordEvent(
				fmt.Sprintf("concurrent-delivery-%d", idx),
				"latest", 675688875, "sha256:abc123", "test-repo",
			)
			// SQLITE_BUSY is acceptable in concurrent scenarios
			if err == nil && isNew {
				successCount <- 1
			} else {
				successCount <- 0
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	successful := 0
	for i := 0; i < 5; i++ {
		<-done
		successful += <-successCount
	}

	// At least one insert should have succeeded
	if successful < 1 {
		t.Errorf("Expected at least 1 successful insert, got %d", successful)
	}

	// Should only have 1 event despite concurrent inserts (content dedup)
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event after concurrent inserts, got %d", count)
	}
}
