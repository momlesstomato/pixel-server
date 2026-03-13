package realtime

import (
	"context"
	"time"

	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	packetprofile "github.com/momlesstomato/pixel-server/pkg/user/packet/profile"
	"go.uber.org/zap"
)

// settingsPatch defines one staged settings patch payload.
type settingsPatch struct {
	volumeSystem *int
	volumeFurni  *int
	volumeTrax   *int
	oldChat      *bool
	roomInvites  *bool
}

// handleSettingsVolume processes one user.settings_volume packet.
func (runtime *Runtime) handleSettingsVolume(connID string, userID int, body []byte) error {
	packet := packetprofile.UserSettingsVolumePacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packet.PacketID(), err)
	}
	system, furni, trax := int(packet.VolumeSystem), int(packet.VolumeFurni), int(packet.VolumeTrax)
	runtime.queuePending(connID, userID, settingsPatch{volumeSystem: &system, volumeFurni: &furni, volumeTrax: &trax})
	return nil
}

// handleSettingsRoomInvites processes one user.settings_room_invites packet.
func (runtime *Runtime) handleSettingsRoomInvites(connID string, userID int, body []byte) error {
	packet := packetprofile.UserSettingsRoomInvitesPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, runtime.packetIDs.SettingsRoomInvites, err)
	}
	value := packet.Enabled
	runtime.queuePending(connID, userID, settingsPatch{roomInvites: &value})
	return nil
}

// handleSettingsOldChat processes one user.settings_old_chat packet.
func (runtime *Runtime) handleSettingsOldChat(connID string, userID int, body []byte) error {
	packet := packetprofile.UserSettingsOldChatPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, runtime.packetIDs.SettingsOldChat, err)
	}
	value := packet.Enabled
	runtime.queuePending(connID, userID, settingsPatch{oldChat: &value})
	return nil
}

// queuePending merges one settings patch and schedules a debounced flush.
func (runtime *Runtime) queuePending(connID string, userID int, patch settingsPatch) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	entry, found := runtime.pending[connID]
	if !found {
		entry = &pendingSettings{userID: userID}
		runtime.pending[connID] = entry
	}
	if patch.volumeSystem != nil {
		entry.patch.volumeSystem = patch.volumeSystem
	}
	if patch.volumeFurni != nil {
		entry.patch.volumeFurni = patch.volumeFurni
	}
	if patch.volumeTrax != nil {
		entry.patch.volumeTrax = patch.volumeTrax
	}
	if patch.oldChat != nil {
		entry.patch.oldChat = patch.oldChat
	}
	if patch.roomInvites != nil {
		entry.patch.roomInvites = patch.roomInvites
	}
	if entry.timer != nil {
		entry.timer.Stop()
	}
	entry.timer = time.AfterFunc(runtime.debounce, func() { runtime.flushPending(context.Background(), connID) })
}

// flushPending flushes staged settings writes for one connection.
func (runtime *Runtime) flushPending(ctx context.Context, connID string) {
	runtime.mutex.Lock()
	entry, found := runtime.pending[connID]
	if found {
		delete(runtime.pending, connID)
	}
	runtime.mutex.Unlock()
	if !found {
		return
	}
	if entry.timer != nil {
		entry.timer.Stop()
	}
	patch := userdomain.SettingsPatch{
		VolumeSystem: entry.patch.volumeSystem, VolumeFurni: entry.patch.volumeFurni,
		VolumeTrax: entry.patch.volumeTrax, OldChat: entry.patch.oldChat, RoomInvites: entry.patch.roomInvites,
	}
	if _, err := runtime.service.SaveSettings(ctx, entry.userID, patch); err != nil {
		runtime.logger.Warn("settings debounce flush failed", zap.String("conn_id", connID), zap.Error(err))
	}
}
