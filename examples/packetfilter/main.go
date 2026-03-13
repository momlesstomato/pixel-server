package main

import sdk "github.com/momlesstomato/pixel-sdk"

// blockedPacketID defines the packet id that this filter blocks.
const blockedPacketID uint16 = 9999

// packetFilter cancels specific inbound packets.
type packetFilter struct {
	server sdk.Server
}

// Manifest returns plugin identity metadata.
func (p *packetFilter) Manifest() sdk.Manifest {
	return sdk.Manifest{Name: "packet-filter", Author: "pixelsv", Version: "1.0.0"}
}

// Enable subscribes to inbound packet events and cancels blocked packets.
func (p *packetFilter) Enable(server sdk.Server) error {
	p.server = server
	server.Events().Subscribe(func(e *sdk.PacketReceived) {
		if e.PacketID == blockedPacketID {
			server.Logger().Printf("blocked packet %d from connection %s", e.PacketID, e.ConnID)
			e.Cancel()
		}
	})
	server.Logger().Printf("packet-filter plugin enabled (blocking packet %d)", blockedPacketID)
	return nil
}

// Disable cleans up plugin resources.
func (p *packetFilter) Disable() error {
	p.server.Logger().Printf("packet-filter plugin disabled")
	return nil
}

// NewPlugin is the exported symbol used by the .so loader.
func NewPlugin() sdk.Plugin {
	return &packetFilter{}
}

func main() {}
