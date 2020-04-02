package store

// Store represents a read and write interface for a presisted key/value store.
type Store interface {
	// Open should enable writing and reading to the underlying store pointed to
	// by path.
	Open(path string) error

	// Close should close any underlying store and disable writing and reading to
	// it and release any allocated caches.
	Close()

	// Put should associate the given value with the given key and persist it to
	// the underlying store.
	Put(key string, value []byte) error

	// Get should return the value associated with the given key and nil if the
	// key isn't present in the underlying store.
	Get(key string) ([]byte, error)

	// Delete should remove any value associated with the given key.
	Delete(key string) error
}
