package conn

import "github.com/zalando/go-keyring"

const serviceName = "db-client"

// Keyring abstracts credential storage.
type Keyring interface {
	Set(service, name, password string) error
	Get(service, name string) (string, error)
	Delete(service, name string) error
}

// OSKeyring uses the system keyring via go-keyring.
type OSKeyring struct{}

// Set stores a credential.
func (OSKeyring) Set(service, name, password string) error {
	return keyring.Set(service, name, password)
}

// Get retrieves a credential.
func (OSKeyring) Get(service, name string) (string, error) {
	return keyring.Get(service, name)
}

// Delete removes a credential.
func (OSKeyring) Delete(service, name string) error {
	return keyring.Delete(service, name)
}

// MemoryKeyring stores credentials in memory (for testing).
type MemoryKeyring struct {
	store map[string]string
}

// NewMemoryKeyring creates a MemoryKeyring.
func NewMemoryKeyring() *MemoryKeyring {
	return &MemoryKeyring{store: make(map[string]string)}
}

// Set stores a credential in memory.
func (m *MemoryKeyring) Set(service, name, password string) error {
	m.store[service+"/"+name] = password
	return nil
}

// Get retrieves a credential from memory.
func (m *MemoryKeyring) Get(service, name string) (string, error) {
	p, ok := m.store[service+"/"+name]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return p, nil
}

// Delete removes a credential from memory.
func (m *MemoryKeyring) Delete(service, name string) error {
	key := service + "/" + name
	if _, ok := m.store[key]; !ok {
		return keyring.ErrNotFound
	}
	delete(m.store, key)
	return nil
}

// CredentialStore manages passwords via a Keyring backend.
type CredentialStore struct {
	keyring Keyring
}

// NewCredentialStore creates a CredentialStore with the given Keyring.
func NewCredentialStore(kr Keyring) *CredentialStore {
	return &CredentialStore{keyring: kr}
}

// SetPassword stores a password for a connection.
func (c *CredentialStore) SetPassword(connName, password string) error {
	return c.keyring.Set(serviceName, connName, password)
}

// GetPassword retrieves a password for a connection.
func (c *CredentialStore) GetPassword(connName string) (string, error) {
	return c.keyring.Get(serviceName, connName)
}

// DeletePassword removes a password for a connection.
func (c *CredentialStore) DeletePassword(connName string) error {
	return c.keyring.Delete(serviceName, connName)
}
