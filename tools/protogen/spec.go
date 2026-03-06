package main

import "go.yaml.in/yaml/v3"

// specFile is the root structure of the protocol specification.
type specFile struct {
	// Packets groups packet definitions by direction.
	Packets packetGroup `yaml:"packets"`
}

// packetGroup stores c2s and s2c packet lists.
type packetGroup struct {
	// C2S defines client-to-server packets.
	C2S []packetSpec `yaml:"c2s"`
	// S2C defines server-to-client packets.
	S2C []packetSpec `yaml:"s2c"`
}

// packetSpec is one packet declaration from the YAML spec.
type packetSpec struct {
	// ID is the wire header identifier.
	ID uint16 `yaml:"id"`
	// Name is the canonical packet name.
	Name string `yaml:"name"`
	// Realm is the bounded-context packet realm.
	Realm string `yaml:"realm"`
	// Summary is the short packet description.
	Summary string `yaml:"summary"`
	// Fields are packet payload fields in wire order.
	Fields []fieldSpec `yaml:"fields"`
}

// fieldSpec is one packet field declaration.
type fieldSpec struct {
	// Name is the payload field name.
	Name string `yaml:"name"`
	// Type is the protocol primitive or composite type.
	Type string `yaml:"type"`
	// Required indicates whether the field is mandatory.
	Required bool `yaml:"required"`
}

// decodeSpec unmarshals YAML input bytes into a specFile.
func decodeSpec(raw []byte) (specFile, error) {
	spec := specFile{}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return spec, err
	}
	return spec, nil
}

// selectPackets filters packets by realm and direction.
func selectPackets(spec specFile, realm string, direction string) []packetSpec {
	source := spec.Packets.C2S
	if direction == "s2c" {
		source = spec.Packets.S2C
	}
	filtered := make([]packetSpec, 0, len(source))
	for _, packet := range source {
		if packet.Realm == realm {
			filtered = append(filtered, packet)
		}
	}
	return filtered
}
