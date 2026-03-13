package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRaces(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		races, err := LoadRaces("/nonexistent/path.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(races) != 0 {
			t.Errorf("expected empty slice, got %d races", len(races))
		}
	})

	t.Run("file is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "races.json")
		os.WriteFile(path, []byte(""), 0644)

		races, err := LoadRaces(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(races) != 0 {
			t.Errorf("expected empty slice, got %d races", len(races))
		}
	})

	t.Run("valid races file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "races.json")
		content := `[
			{"id": 1, "name": "Race 1", "start_date": "2024-01-01T10:00:00Z"},
			{"id": 2, "name": "Race 2", "start_date": "2024-06-01T10:00:00Z"}
		]`
		os.WriteFile(path, []byte(content), 0644)

		races, err := LoadRaces(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(races) != 2 {
			t.Errorf("expected 2 races, got %d", len(races))
		}
		if races[0].Name != "Race 1" {
			t.Errorf("expected first race name to be Race 1, got %s", races[0].Name)
		}
	})
}

func TestSaveRaces(t *testing.T) {
	t.Run("saves and sorts by start_date descending", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "races.json")

		races := []Race{
			{ID: 1, Name: "Early Race", StartDate: "2024-01-01T10:00:00Z"},
			{ID: 2, Name: "Late Race", StartDate: "2024-06-01T10:00:00Z"},
			{ID: 3, Name: "Mid Race", StartDate: "2024-03-01T10:00:00Z"},
		}

		err := SaveRaces(path, races)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		loaded, err := LoadRaces(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(loaded) != 3 {
			t.Errorf("expected 3 races, got %d", len(loaded))
		}
		if loaded[0].Name != "Late Race" {
			t.Errorf("expected first race to be Late Race, got %s", loaded[0].Name)
		}
		if loaded[1].Name != "Mid Race" {
			t.Errorf("expected second race to be Mid Race, got %s", loaded[1].Name)
		}
		if loaded[2].Name != "Early Race" {
			t.Errorf("expected third race to be Early Race, got %s", loaded[2].Name)
		}
	})

	t.Run("atomic write", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "races.json")

		races := []Race{{ID: 1, Name: "Test Race", StartDate: "2024-01-01T10:00:00Z"}}

		err := SaveRaces(path, races)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Size() == 0 {
			t.Error("expected non-empty file")
		}
	})
}

func TestLoadSync(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		state, err := LoadSync("/nonexistent/path.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.LastSyncedAt != 0 || state.LastActivityEpoch != 0 {
			t.Errorf("expected zero state, got %+v", state)
		}
	})

	t.Run("valid sync file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "sync.json")
		content := `{"last_synced_at": 1234567890, "last_activity_epoch": 1234567890}`
		os.WriteFile(path, []byte(content), 0644)

		state, err := LoadSync(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.LastSyncedAt != 1234567890 {
			t.Errorf("expected LastSyncedAt 1234567890, got %d", state.LastSyncedAt)
		}
		if state.LastActivityEpoch != 1234567890 {
			t.Errorf("expected LastActivityEpoch 1234567890, got %d", state.LastActivityEpoch)
		}
	})
}

func TestSaveSync(t *testing.T) {
	t.Run("saves sync state", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "sync.json")

		state := SyncState{
			LastSyncedAt:      1234567890,
			LastActivityEpoch: 1234567890,
		}

		err := SaveSync(path, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		loaded, err := LoadSync(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loaded.LastSyncedAt != 1234567890 {
			t.Errorf("expected LastSyncedAt 1234567890, got %d", loaded.LastSyncedAt)
		}
	})

	t.Run("atomic write", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "sync.json")

		state := SyncState{LastSyncedAt: 123, LastActivityEpoch: 456}

		err := SaveSync(path, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Size() == 0 {
			t.Error("expected non-empty file")
		}
	})
}
