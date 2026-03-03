package memory

import (
	"context"
	"sync"

	"pixel-server/pkg/user"
)

// WardrobeRepo is an in-memory fake that implements user.WardrobeRepository.
type WardrobeRepo struct {
	mu    sync.RWMutex
	slots map[int32][]*user.WardrobeOutfit
}

// NewWardrobeRepo returns a ready-to-use in-memory WardrobeRepo.
func NewWardrobeRepo() *WardrobeRepo {
	return &WardrobeRepo{slots: make(map[int32][]*user.WardrobeOutfit)}
}

func (r *WardrobeRepo) GetByUser(_ context.Context, userID int32) ([]*user.WardrobeOutfit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	outfits := r.slots[userID]
	out := make([]*user.WardrobeOutfit, len(outfits))
	for i, o := range outfits {
		cp := *o
		out[i] = &cp
	}
	return out, nil
}

func (r *WardrobeRepo) Save(_ context.Context, outfit *user.WardrobeOutfit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing := r.slots[outfit.UserID]
	for i, o := range existing {
		if o.SlotID == outfit.SlotID {
			cp := *outfit
			existing[i] = &cp
			return nil
		}
	}
	cp := *outfit
	r.slots[outfit.UserID] = append(existing, &cp)
	return nil
}

// compile-time interface check.
var _ user.WardrobeRepository = (*WardrobeRepo)(nil)
