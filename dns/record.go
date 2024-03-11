package dns

import (
	"strconv"
	"strings"
)

type RData interface {
	Serialize() ([]byte, error)
}

type AData struct {
	IPAddr string
}

func (d AData) Serialize() ([]byte, error) {
	octets := strings.Split(d.IPAddr, ".")
	out := []byte{}

	for _, o := range octets {
		b, err := strconv.Atoi(o)
		if err != nil {
			return nil, err
		}

		out = append(out, uint8(b))
	}

	return out, nil
}

type CNAMEData struct {
	Name string
}

func (d CNAMEData) Serialize() ([]byte, error) {
	return serializeLabels(strings.Split(d.Name, ".")), nil
}

func serializeLabels(labels []string) []byte {
	raw := []byte{}

	for _, l := range labels {
		raw = append(
			raw, append([]byte{uint8(len(l))}, []byte(l)...)...,
		)
	}

	return raw
}
