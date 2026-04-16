package domain

import "encoding/json"

// InteractionData stores server-only item interaction metadata.
type InteractionData struct {
	// Teleporter stores optional teleporter pairing metadata.
	Teleporter *TeleporterData `json:"teleporter,omitempty"`
	// Dimmer stores optional room dimmer preset metadata.
	Dimmer *DimmerData `json:"dimmer,omitempty"`
	// Gift stores optional present reveal metadata.
	Gift *GiftData `json:"gift,omitempty"`
}

// TeleporterData stores teleporter partner metadata.
type TeleporterData struct {
	// RoomID stores the destination room identifier.
	RoomID int `json:"room_id"`
	// ItemID stores the destination item identifier.
	ItemID int `json:"item_id"`
}

// DimmerData stores room dimmer preset metadata.
type DimmerData struct {
	// Enabled reports whether the room dimmer is currently active.
	Enabled bool `json:"enabled"`
	// SelectedPresetID stores the current preset slot.
	SelectedPresetID int `json:"selected_preset_id"`
	// Presets stores all configured room dimmer presets.
	Presets []DimmerPresetData `json:"presets,omitempty"`
}

// DimmerPresetData stores one room dimmer preset definition.
type DimmerPresetData struct {
	// PresetID stores the stable preset slot identifier.
	PresetID int `json:"preset_id"`
	// Type stores the room dimmer effect identifier.
	Type int `json:"type"`
	// Color stores the preset RGB hex payload with leading hash.
	Color string `json:"color"`
	// Brightness stores the preset brightness value.
	Brightness int `json:"brightness"`
}

// GiftData stores present reveal metadata.
type GiftData struct {
	// DefinitionID stores the revealed internal furniture definition identifier.
	DefinitionID int `json:"definition_id"`
	// ProductCode stores the revealed client product code.
	ProductCode string `json:"product_code"`
	// PurchaserName stores the optional purchaser username.
	PurchaserName string `json:"purchaser_name,omitempty"`
	// PurchaserFigure stores the optional purchaser figure string.
	PurchaserFigure string `json:"purchaser_figure,omitempty"`
	// Message stores the optional gift note message.
	Message string `json:"message,omitempty"`
	// PetFigureString stores the optional revealed pet figure string.
	PetFigureString string `json:"pet_figure_string,omitempty"`
}

// ParseInteractionData decodes one server-only interaction payload.
func ParseInteractionData(raw string) (InteractionData, error) {
	if raw == "" {
		return InteractionData{}, nil
	}
	var data InteractionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return InteractionData{}, err
	}
	return data, nil
}

// Encode serializes one server-only interaction payload.
func (data InteractionData) Encode() (string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	if string(body) == "{}" {
		return "", nil
	}
	return string(body), nil
}
