package packet

// GetIndexPacketID defines packet identifier for catalog.get_index (c2s).
const GetIndexPacketID uint16 = 1195

// IndexResponsePacketID defines packet identifier for catalog.index (s2c).
const IndexResponsePacketID uint16 = 1032

// GetPagePacketID defines packet identifier for catalog.get_page (c2s).
const GetPagePacketID uint16 = 412

// PageResponsePacketID defines packet identifier for catalog.page (s2c).
const PageResponsePacketID uint16 = 804

// PurchasePacketID defines packet identifier for catalog.purchase.
const PurchasePacketID uint16 = 3492

// PurchaseGiftPacketID defines packet identifier for catalog.purchase_gift.
const PurchaseGiftPacketID uint16 = 1411

// RedeemVoucherPacketID defines packet identifier for catalog.redeem_voucher.
const RedeemVoucherPacketID uint16 = 339

// CheckGiftablePacketID defines packet identifier for catalog.check_giftable.
const CheckGiftablePacketID uint16 = 1347

// GetGiftWrappingConfigPacketID defines packet identifier for catalog.get_gift_wrapping_config.
const GetGiftWrappingConfigPacketID uint16 = 418

// PurchaseOKPacketID defines packet identifier for catalog.purchase_ok (s2c).
const PurchaseOKPacketID uint16 = 869

// PurchaseErrorPacketID defines packet identifier for catalog.purchase_error (s2c).
const PurchaseErrorPacketID uint16 = 1404

// PurchaseNotAllowedPacketID defines packet identifier for catalog.purchase_not_allowed (s2c).
const PurchaseNotAllowedPacketID uint16 = 3770

// GiftWrappingConfigResponsePacketID defines packet identifier for catalog.gift_wrapping_config (s2c).
const GiftWrappingConfigResponsePacketID uint16 = 2234
