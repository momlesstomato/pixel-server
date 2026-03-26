package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// IndexNode defines one catalog tree node for client serialization.
type IndexNode struct {
	// Visible stores whether the node appears in the client catalog.
	Visible bool
	// Icon stores the leaf icon sprite identifier.
	Icon int32
	// PageID stores the page identifier; -1 marks non-navigable nodes.
	PageID int32
	// PageName stores the page localization or link key.
	PageName string
	// Caption stores the page display caption.
	Caption string
	// OfferIDs stores item offer identifiers listed on this page.
	OfferIDs []int32
	// Children stores direct child nodes in display order.
	Children []IndexNode
}

// IndexPacket defines catalog.index (s2c 1032) payload.
type IndexPacket struct {
	// Root stores the root catalog tree node.
	Root IndexNode
	// NewItems stores whether new items have been added recently.
	NewItems bool
	// CatalogType stores the echoed mode string from the request.
	CatalogType string
}

// PacketID returns protocol packet identifier.
func (p IndexPacket) PacketID() uint16 { return IndexResponsePacketID }

// Encode serializes catalog index tree into packet body.
func (p IndexPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	if err := encodeIndexNode(writer, p.Root); err != nil {
		return nil, err
	}
	writer.WriteBool(p.NewItems)
	if err := writer.WriteString(p.CatalogType); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// encodeIndexNode writes one catalog tree node and its children.
func encodeIndexNode(w *codec.Writer, node IndexNode) error {
	w.WriteBool(node.Visible)
	w.WriteInt32(node.Icon)
	w.WriteInt32(node.PageID)
	if err := w.WriteString(node.PageName); err != nil {
		return err
	}
	if err := w.WriteString(node.Caption); err != nil {
		return err
	}
	w.WriteInt32(int32(len(node.OfferIDs)))
	for _, id := range node.OfferIDs {
		w.WriteInt32(id)
	}
	w.WriteInt32(int32(len(node.Children)))
	for _, child := range node.Children {
		if err := encodeIndexNode(w, child); err != nil {
			return err
		}
	}
	return nil
}
