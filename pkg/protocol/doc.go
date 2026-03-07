// Package protocol contains generated packet contracts from the pixel-protocol spec.
//
//go:generate go run ../../tools/protogen -spec ../../vendor/pixel-protocol/spec/protocol.yaml -out . -realm handshake-security -direction c2s
//go:generate go run ../../tools/protogen -spec ../../vendor/pixel-protocol/spec/protocol.yaml -out . -realm session-connection -direction c2s
package protocol
