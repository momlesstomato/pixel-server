package main

import "gopkg.in/yaml.v3"

// Spec is the top-level structure of protocol.yaml.
type Spec struct {
	Packets struct {
		C2S []PacketDef `yaml:"c2s"`
		S2C []PacketDef `yaml:"s2c"`
	} `yaml:"packets"`
}

// PacketDef defines one packet entry from the protocol YAML.
type PacketDef struct {
	ID      uint16     `yaml:"id"`
	Name    string     `yaml:"name"`
	Realm   string     `yaml:"realm"`
	Summary string     `yaml:"summary"`
	Fields  []FieldDef `yaml:"fields"`
	Header  string     `yaml:"header"`
}

// FieldDef defines one packet field from the protocol YAML.
type FieldDef struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
}

// LoadSpec parses the YAML into a Spec.
func LoadSpec(data []byte) (*Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
