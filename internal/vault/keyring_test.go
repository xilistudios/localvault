package vault

import (
	"fmt"
	"sort"
	"testing"
)

// MockKeyring is an in-memory SecretStore for testing
type MockKeyring struct {
	store map[string][]byte
}

func NewMockKeyring() *MockKeyring {
	return &MockKeyring{store: make(map[string][]byte)}
}

func (m *MockKeyring) Get(key string) ([]byte, error) {
	data, ok := m.store[key]
	if !ok {
		return nil, fmt.Errorf("secret %q not found", key)
	}
	return data, nil
}

func (m *MockKeyring) Set(key string, data []byte) error {
	m.store[key] = data
	return nil
}

func (m *MockKeyring) Remove(key string) error {
	delete(m.store, key)
	return nil
}

func (m *MockKeyring) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.store))
	for k := range m.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// Verify MockKeyring satisfies SecretStore interface
var _ SecretStore = (*MockKeyring)(nil)

// --- Tests ---

func TestSecretKey(t *testing.T) {
	got := SecretKey("myapp", "dev", "DATABASE_URL")
	want := "localvault.myapp.dev.DATABASE_URL"
	if got != want {
		t.Errorf("SecretKey() = %q, want %q", got, want)
	}
}

func TestSecretKeyPrefix(t *testing.T) {
	got := SecretKeyPrefix("myapp", "dev")
	want := "localvault.myapp.dev."
	if got != want {
		t.Errorf("SecretKeyPrefix() = %q, want %q", got, want)
	}
}

func TestMockKeyringSetGetRoundTrip(t *testing.T) {
	mk := NewMockKeyring()
	key := "localvault.myapp.dev.DATABASE_URL"
	value := []byte("postgres://localhost:5432/mydb")

	if err := mk.Set(key, value); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := mk.Get(key)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Get() = %q, want %q", got, value)
	}
}

func TestMockKeyringGetMissingKey(t *testing.T) {
	mk := NewMockKeyring()
	_, err := mk.Get("nonexistent")
	if err == nil {
		t.Fatal("Get() on missing key should return error")
	}
}

func TestMockKeyringRemoveExisting(t *testing.T) {
	mk := NewMockKeyring()
	key := "localvault.myapp.dev.API_KEY"
	value := []byte("secret123")

	if err := mk.Set(key, value); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := mk.Remove(key); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	_, err := mk.Get(key)
	if err == nil {
		t.Fatal("Get() after Remove() should return error")
	}
}

func TestMockKeyringRemoveMissingKeyIsNoop(t *testing.T) {
	mk := NewMockKeyring()
	// Removing a key that doesn't exist should not error
	if err := mk.Remove("nonexistent"); err != nil {
		t.Fatalf("Remove() on missing key should be no-op, got error: %v", err)
	}
}

func TestMockKeyringKeysSorted(t *testing.T) {
	mk := NewMockKeyring()

	mk.Set("localvault.myapp.dev.Z_LAST", []byte("1"))
	mk.Set("localvault.myapp.dev.A_FIRST", []byte("2"))
	mk.Set("localvault.myapp.dev.M_MIDDLE", []byte("3"))

	keys, err := mk.Keys()
	if err != nil {
		t.Fatalf("Keys() error: %v", err)
	}

	want := []string{
		"localvault.myapp.dev.A_FIRST",
		"localvault.myapp.dev.M_MIDDLE",
		"localvault.myapp.dev.Z_LAST",
	}
	if len(keys) != len(want) {
		t.Fatalf("Keys() returned %d keys, want %d", len(keys), len(want))
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, k, want[i])
		}
	}
}
