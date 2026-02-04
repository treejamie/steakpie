package webhook

import (
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

	deliveryID := "test-delivery-123"
	repository := "test-repo"

	isNew, err := store.RecordEvent(deliveryID, repository)
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

	deliveryID := "test-delivery-456"
	repository := "test-repo"

	// First insert
	isNew, err := store.RecordEvent(deliveryID, repository)
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}
	if !isNew {
		t.Error("Expected first event to be new")
	}

	// Second insert (duplicate)
	isNew, err = store.RecordEvent(deliveryID, repository)
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
		repository string
	}{
		{"delivery-1", "repo-a"},
		{"delivery-2", "repo-a"},
		{"delivery-3", "repo-b"},
	}

	for _, e := range events {
		isNew, err := store.RecordEvent(e.deliveryID, e.repository)
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
	deliveryID := "concurrent-test"
	repository := "test-repo"

	done := make(chan bool)

	// Try to insert same event concurrently
	for i := 0; i < 5; i++ {
		go func() {
			_, err := store.RecordEvent(deliveryID, repository)
			if err != nil {
				t.Errorf("Concurrent insert failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should only have 1 event despite concurrent inserts
	count, err := store.Stats()
	if err != nil {
		t.Fatalf("Failed to query stats: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event after concurrent inserts, got %d", count)
	}
}
