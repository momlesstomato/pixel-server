package interceptor

import (
	"sync"

	"pixelsv/pkg/plugin"
)

// registration tracks one packet hook registration.
type registration struct {
	// interceptor is the owning interceptor instance.
	interceptor *Interceptor
	// headerID is the target packet header.
	headerID uint16
	// id is the registration identifier.
	id uint64
	// pre reports pre-handler vs post-handler registration.
	pre bool
	// all reports all-header vs one-header registration.
	all bool
	// once guarantees one-time unregistration.
	once sync.Once
}

// Unsubscribe removes the hook registration.
func (r *registration) Unsubscribe() {
	if r == nil {
		return
	}
	r.once.Do(func() {
		if r.interceptor != nil {
			r.interceptor.remove(r.headerID, hookEntry{id: r.id, pre: r.pre, all: r.all})
		}
	})
}

// noopRegistration is returned for ignored hook registrations.
type noopRegistration struct{}

// Unsubscribe is a no-op for ignored registrations.
func (n noopRegistration) Unsubscribe() {
	_ = n
}

// add inserts one hook registration.
func (i *Interceptor) add(headerID uint16, hook plugin.PacketHook, pre bool, all bool) plugin.Registration {
	if i == nil || hook == nil {
		return noopRegistration{}
	}
	id := i.nextID.Add(1)
	i.mu.Lock()
	if all {
		if pre {
			i.beforeAll[id] = hook
		} else {
			i.afterAll[id] = hook
		}
	} else if pre {
		if _, ok := i.before[headerID]; !ok {
			i.before[headerID] = make(map[uint64]plugin.PacketHook)
		}
		i.before[headerID][id] = hook
	} else {
		if _, ok := i.after[headerID]; !ok {
			i.after[headerID] = make(map[uint64]plugin.PacketHook)
		}
		i.after[headerID][id] = hook
	}
	i.mu.Unlock()
	return &registration{interceptor: i, headerID: headerID, id: id, pre: pre, all: all}
}

// remove deletes one hook registration.
func (i *Interceptor) remove(headerID uint16, entry hookEntry) {
	if i == nil || entry.id == 0 {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if entry.all {
		if entry.pre {
			delete(i.beforeAll, entry.id)
		} else {
			delete(i.afterAll, entry.id)
		}
		return
	}
	if entry.pre {
		set := i.before[headerID]
		delete(set, entry.id)
		if len(set) == 0 {
			delete(i.before, headerID)
		}
		return
	}
	set := i.after[headerID]
	delete(set, entry.id)
	if len(set) == 0 {
		delete(i.after, headerID)
	}
}

var _ plugin.Registration = (*registration)(nil)
var _ plugin.Registration = noopRegistration{}
