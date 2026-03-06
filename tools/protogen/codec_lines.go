package main

import (
	"fmt"
	"strings"
)

// goTypeFor maps one protocol field into the generated Go field type.
func goTypeFor(field fieldSpec) (string, error) {
	base, ok := map[string]string{
		"boolean": "bool",
		"int32":   "int32",
		"uint16":  "uint16",
		"uint32":  "uint32",
		"string":  "string",
		"bytes":   "[]byte",
	}[field.Type]
	if !ok {
		return "", fmt.Errorf("unsupported field type: %s", field.Type)
	}
	if field.Required {
		return base, nil
	}
	return "*" + base, nil
}

// encodeLine emits one encode method statement for one field.
func encodeLine(field fieldSpec) (string, error) {
	fieldID := fieldName(field.Name)
	value := "p." + fieldID
	if field.Required {
		return encodeWriteLine(field.Type, value)
	}
	writeLine, err := encodeWriteLine(field.Type, "*"+value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("if p.%s != nil { %s }", fieldID, writeLine), nil
}

// encodeWriteLine emits one primitive write call for a value expression.
func encodeWriteLine(kind string, value string) (string, error) {
	switch kind {
	case "boolean":
		return "writer.WriteBool(" + value + ")", nil
	case "int32":
		return "writer.WriteInt32(" + value + ")", nil
	case "uint16":
		return "writer.WriteUint16(" + value + ")", nil
	case "uint32":
		return "writer.WriteUint32(" + value + ")", nil
	case "string":
		return "writer.WriteString(" + value + ")", nil
	case "bytes":
		return "writer.WriteBytes(" + value + ")", nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", kind)
	}
}

// decodeLine emits one decode method statement for one field.
func decodeLine(field fieldSpec) (string, error) {
	fieldID := fieldName(field.Name)
	target := "packet." + fieldID
	if field.Required {
		return decodeReadLine(field.Type, target)
	}
	valueID := fieldID + "Value"
	readLine, err := decodeReadLine(field.Type, valueID)
	if err != nil {
		return "", err
	}
	valueType, err := goTypeFor(field)
	if err != nil {
		return "", err
	}
	valueType = strings.TrimPrefix(valueType, "*")
	return fmt.Sprintf("if reader.Remaining() > 0 { var %s %s; %s; %s = &%s }", valueID, valueType, readLine, target, valueID), nil
}

// decodeReadLine emits one primitive read expression with error check.
func decodeReadLine(kind string, target string) (string, error) {
	switch kind {
	case "boolean":
		return target + ", err = reader.ReadBool(); if err != nil { return nil, err }", nil
	case "int32":
		return target + ", err = reader.ReadInt32(); if err != nil { return nil, err }", nil
	case "uint16":
		return target + ", err = reader.ReadUint16(); if err != nil { return nil, err }", nil
	case "uint32":
		return target + ", err = reader.ReadUint32(); if err != nil { return nil, err }", nil
	case "string":
		return target + ", err = reader.ReadString(); if err != nil { return nil, err }", nil
	case "bytes":
		return target + ", err = reader.ReadBytes(reader.Remaining()); if err != nil { return nil, err }", nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", kind)
	}
}
