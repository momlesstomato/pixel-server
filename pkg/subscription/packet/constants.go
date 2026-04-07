package packet

// GetSubscriptionPacketID defines packet identifier for user.get_subscription (c2s).
const GetSubscriptionPacketID uint16 = 3166

// SubscriptionResponsePacketID defines packet identifier for user.subscription (s2c).
const SubscriptionResponsePacketID uint16 = 954

// GetClubOffersPacketID defines packet identifier for catalog.get_club_offers.
const GetClubOffersPacketID uint16 = 3285

// GetProductOfferPacketID defines packet identifier for catalog.get_product_offer.
const GetProductOfferPacketID uint16 = 2594

// GetHCExtendOfferPacketID defines packet identifier for catalog.get_hc_extend_offer.
const GetHCExtendOfferPacketID uint16 = 2462

// GetClubGiftInfoPacketID defines packet identifier for catalog.get_club_gift_info.
const GetClubGiftInfoPacketID uint16 = 487

// GetKickbackInfoPacketID defines packet identifier for user.get_kickback_info.
const GetKickbackInfoPacketID uint16 = 869

// SelectClubGiftPacketID defines packet identifier for catalog.select_club_gift.
const SelectClubGiftPacketID uint16 = 2276

// GetDirectClubBuyAvailablePacketID defines packet identifier for catalog.get_direct_club_buy.
const GetDirectClubBuyAvailablePacketID uint16 = 801

// ClubOffersResponsePacketID defines packet identifier for catalog.club_offers (s2c).
const ClubOffersResponsePacketID uint16 = 2405

// HCExtendOfferResponsePacketID defines packet identifier for catalog.hc_extend_offer (s2c).
const HCExtendOfferResponsePacketID uint16 = 3964

// ClubGiftInfoResponsePacketID defines packet identifier for catalog.club_gift_info (s2c).
const ClubGiftInfoResponsePacketID uint16 = 619

// ClubGiftSelectedResponsePacketID defines packet identifier for catalog.club_gift_selected (s2c).
const ClubGiftSelectedResponsePacketID uint16 = 659

// KickbackInfoResponsePacketID defines packet identifier for user.kickback_info (s2c).
const KickbackInfoResponsePacketID uint16 = 3277

// DirectClubBuyAvailableResponsePacketID defines packet identifier for catalog.direct_sms_club_buy (s2c).
const DirectClubBuyAvailableResponsePacketID uint16 = 195
