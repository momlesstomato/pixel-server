package ws

import (
	"github.com/gofiber/contrib/websocket"
	"pixelsv/pkg/codec"
)

// writeDisconnectReason writes one disconnect.reason frame to one websocket connection.
func writeDisconnectReason(conn *websocket.Conn, reason int32) error {
	return conn.WriteMessage(websocket.BinaryMessage, disconnectReasonFrame(reason))
}

// disconnectReasonFrame encodes one disconnect.reason frame payload.
func disconnectReasonFrame(reason int32) []byte {
	writer := codec.NewWriter(8)
	writer.WriteInt32(reason)
	return codec.EncodeFrame(4000, writer.Bytes())
}
