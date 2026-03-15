package application

// perkMapping defines one client perk to permission mapping.
type perkMapping struct {
	// Code stores client perk code.
	Code string
	// Permission stores required permission string.
	Permission string
	// DeniedMessage stores denied reason.
	DeniedMessage string
}

var knownPerks = []perkMapping{
	{Code: "USE_GUIDE_TOOL", Permission: "perk.guide", DeniedMessage: "Requires guide role"},
	{Code: "GIVE_GUIDE_TOURS", Permission: "perk.guide.tours", DeniedMessage: "Requires guide role"},
	{Code: "JUDGE_CHAT_REVIEWS", Permission: "perk.chat_reviews", DeniedMessage: "Requires moderator"},
	{Code: "VOTE_IN_COMPETITIONS", Permission: "perk.competitions"},
	{Code: "CALL_ON_HELPERS", Permission: "perk.helpers"},
	{Code: "CITIZEN", Permission: "perk.citizen"},
	{Code: "TRADE", Permission: "perk.trade", DeniedMessage: "Requires Club membership"},
	{Code: "HEIGHTMAP_EDITOR_BETA", Permission: "perk.heightmap_editor"},
	{Code: "BUILDER_AT_WORK", Permission: "perk.builder"},
	{Code: "NAVIGATOR_ROOM_THUMBNAIL_CAMERA", Permission: "perk.room_thumbnail"},
	{Code: "CAMERA", Permission: "perk.camera", DeniedMessage: "Requires Club membership"},
	{Code: "MOUSE_ZOOM", Permission: "perk.mouse_zoom"},
	{Code: "NAVIGATOR_PHASE_TWO", Permission: "perk.navigator_v2"},
	{Code: "SAFE_CHAT", Permission: "perk.safe_chat"},
	{Code: "HABBO_CLUB_OFFER_BETA", Permission: "perk.club_offer"},
}
