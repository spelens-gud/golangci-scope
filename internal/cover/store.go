package cover

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrServiceAlreadyRegistered = errors.New("service already registered")

type Store interface {
	// Add adds the given service to store
	Add(s ServiceUnderTest) error

	// Get returns the registered service information with the given service's name
	Get(name string) []string

	// Get returns all the registered service information as a map
	GetAll() map[string][]string

	// Init cleanup all the registered service information
	Init() error

	// Set stores the services information into internal state
	Set(services map[string][]string) error

	// Remove the service from the store by address
	Remove(addr string) error
}
type memoryStore struct {
	mu          sync.RWMutex
	servicesMap map[string][]string
}

// Add adds the given service to MemoryStore
func (l *memoryStore) Add(s ServiceUnderTest) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// load to memory
	if addrs, ok := l.servicesMap[s.Name]; ok {
		for _, addr := range addrs {
			if addr == s.Address {
				log.Printf("service registered already, name: %s, address: %s", s.Name, s.Address)
				return ErrServiceAlreadyRegistered
			}
		}
		addrs = append(addrs, s.Address)
		l.servicesMap[s.Name] = addrs
	} else {
		l.servicesMap[s.Name] = []string{s.Address}
	}

	return nil
}

// Get returns the registered service information with the given name
func (l *memoryStore) Get(name string) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.servicesMap[name]
}

// Get returns all the registered service information
func (l *memoryStore) GetAll() map[string][]string {
	res := make(map[string][]string)
	l.mu.RLock()
	defer l.mu.RUnlock()
	for k, v := range l.servicesMap {
		res[k] = append(make([]string, 0, len(v)), v...)
	}
	return res
}

// Init cleanup all the registered service information
// and the local persistent file
func (l *memoryStore) Init() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.servicesMap = make(map[string][]string, 0)
	return nil
}

func (l *memoryStore) Set(services map[string][]string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	newMap := make(map[string][]string)
	for k, v := range services {
		newMap[k] = append(make([]string, 0), v...)
	}
	l.servicesMap = newMap

	return nil
}

// Remove one service from the memory store
// if service is not fount, return "no service found" error
func (l *memoryStore) Remove(removeAddr string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	flag := false
	for name, addrs := range l.servicesMap {
		newAddrs := make([]string, 0)
		for _, addr := range addrs {
			if removeAddr != addr {
				newAddrs = append(newAddrs, addr)
			} else {
				flag = true
			}
		}
		// if no services left, remove by name
		if len(newAddrs) == 0 {
			delete(l.servicesMap, name)
		} else {
			l.servicesMap[name] = newAddrs
		}
	}

	if !flag {
		return fmt.Errorf("no service found: %s", removeAddr)
	}

	return nil
}
func NewMemoryStore() Store {
	return &memoryStore{
		servicesMap: make(map[string][]string, 0),
	}
}
