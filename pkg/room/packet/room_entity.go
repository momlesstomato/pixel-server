package packet

import (
	"fmt"
	"sort"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// UsersComposer sends entity list to client (s2c 3857).
type UsersComposer struct {
	// Entities stores all room entities to serialize.
	Entities []domain.RoomEntity
}

// PacketID returns the protocol packet identifier.
func (p UsersComposer) PacketID() uint16 { return UsersComposerID }

// Encode serializes room entity list.
func (p UsersComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Entities)))
	for _, e := range p.Entities {
		if err := encodeEntity(w, e); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeEntity writes one entity to the writer.
func encodeEntity(w *codec.Writer, e domain.RoomEntity) error {
	w.WriteInt32(int32(e.UserID))
	if err := w.WriteString(e.Username); err != nil {
		return err
	}
	if err := w.WriteString(e.Motto); err != nil {
		return err
	}
	if err := w.WriteString(e.Look); err != nil {
		return err
	}
	w.WriteInt32(int32(e.VirtualID))
	w.WriteInt32(int32(e.Position.X))
	w.WriteInt32(int32(e.Position.Y))
	if err := w.WriteString(fmt.Sprintf("%g", e.Position.Z)); err != nil {
		return err
	}
	w.WriteInt32(int32(e.BodyRotation))
	w.WriteInt32(1)
	if err := w.WriteString(e.Gender); err != nil {
		return err
	}
	w.WriteInt32(-1)
	w.WriteInt32(-1)
	if err := w.WriteString(""); err != nil {
		return err
	}
	if err := w.WriteString(""); err != nil {
		return err
	}
	w.WriteInt32(0)
	w.WriteBool(false)
	return nil
}

// UserUpdateComposer sends entity status updates (s2c 3559).
type UserUpdateComposer struct {
	// Entities stores entities with updated state.
	Entities []domain.RoomEntity
}

// PacketID returns the protocol packet identifier.
func (p UserUpdateComposer) PacketID() uint16 { return UserUpdateComposerID }

// Encode serializes entity status updates.
func (p UserUpdateComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Entities)))
	for _, e := range p.Entities {
		pos := e.Position
		if e.StepFrom != nil {
			pos = *e.StepFrom
		}
		w.WriteInt32(int32(e.VirtualID))
		w.WriteInt32(int32(pos.X))
		w.WriteInt32(int32(pos.Y))
		if err := w.WriteString(fmt.Sprintf("%g", pos.Z)); err != nil {
			return nil, err
		}
		w.WriteInt32(int32(e.HeadRotation))
		w.WriteInt32(int32(e.BodyRotation))
		status := encodeStatuses(e)
		if err := w.WriteString(status); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// encodeStatuses formats entity status map to protocol string.
func encodeStatuses(e domain.RoomEntity) string {
	keys := orderedStatusKeys(e.Statuses)
	result := "/"
	for _, k := range keys {
		v := e.Statuses[k]
		result += k + " " + v + "/"
	}
	return result
}

func orderedStatusKeys(statuses map[string]string) []string {
	keys := make([]string, 0, len(statuses))
	for key := range statuses {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := statusOrder(keys[i])
		right := statusOrder(keys[j])
		if left != right {
			return left < right
		}
		return keys[i] < keys[j]
	})
	return keys
}

func statusOrder(key string) int {
	switch key {
	case "flatctrl":
		return 10
	case "sign":
		return 20
	case "gst":
		return 30
	case "trd":
		return 40
	case "dance":
		return 50
	case "mv":
		return 90
	case "sit":
		return 100
	case "lay":
		return 110
	default:
		return 60
	}
}

// UserRemoveComposer notifies entity removal (s2c 3839).
type UserRemoveComposer struct {
	// VirtualID stores the removed entity virtual identifier.
	VirtualID int32
}

// PacketID returns the protocol packet identifier.
func (p UserRemoveComposer) PacketID() uint16 { return UserRemoveComposerID }

// Encode serializes the entity removal.
func (p UserRemoveComposer) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(fmt.Sprintf("%d", p.VirtualID)); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}


// DecodeMoveAvatar extracts walk destination from packet body.
func DecodeMoveAvatar(body []byte) []int {
	r := codec.NewReader(body)
	x, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	y, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	return []int{int(x), int(y)}
}
