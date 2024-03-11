package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

type Message struct {
	raw             []byte
	ID              uint16
	IsQuery         bool
	IsResponse      bool
	OpCode          string
	QuestionCount   uint16
	AnswerCount     uint16
	ServerCount     uint16
	AdditionalCount uint16
	Questions       []DnsQuestion
	Answers         []ResourceRecord
	NameServers     []ResourceRecord
	Extra           []ResourceRecord
}

type DnsQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type ResourceRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

func (r *ResourceRecord) Serialize() []byte {
	serialized := []byte{}
	serialized = append(serialized, serializeLabels(strings.Split(r.Name, "."))...)
	serialized = binary.BigEndian.AppendUint16(serialized, r.Type)
	serialized = binary.BigEndian.AppendUint16(serialized, r.Class)
	serialized = binary.BigEndian.AppendUint32(serialized, r.TTL)
	serialized = binary.BigEndian.AppendUint16(serialized, uint16(len(r.Data)))
	serialized = append(serialized, r.Data...)
	return serialized
}

func (d *Message) Deserialize(raw []byte) error {
	d.raw = bytes.TrimRight(raw, "\u0000")
	d.ID = binary.BigEndian.Uint16(raw[0:2])
	header := binary.BigEndian.Uint16(raw[2:4])

	d.IsQuery = header&0b1000000000000000 == 0b0000000000000000
	d.IsResponse = !d.IsQuery

	d.OpCode = "QUERY"
	if header&0b0000100000000000 == 0b0000100000000000 {
		d.OpCode = "IQUERY"
	}

	if header&0b0001000000000000 == 0b0001000000000000 {
		d.OpCode = "STATUS"
	}

	d.QuestionCount = binary.BigEndian.Uint16(raw[4:6])
	d.AnswerCount = binary.BigEndian.Uint16(raw[6:8])
	d.ServerCount = binary.BigEndian.Uint16(raw[8:10])
	d.AdditionalCount = binary.BigEndian.Uint16(raw[10:12])

	labelCursor := uint8(12)

	d.Questions = make([]DnsQuestion, d.QuestionCount)

	for i, q := range d.Questions {
		labels, read := deserializeLabels(raw[labelCursor:])
		labelCursor += read

		q.Name = strings.Join(labels, ".")
		q.Type = binary.BigEndian.Uint16(raw[labelCursor : labelCursor+2])
		labelCursor += 2
		q.Class = binary.BigEndian.Uint16(raw[labelCursor : labelCursor+2])
		labelCursor += 2

		// q was passed by value, so set this updated value to a specific index
		// in the slice
		d.Questions[i] = q
	}

	return nil
}

func (d *Message) Serialize() []byte {
	serialized := []byte{}

	serialized = binary.BigEndian.AppendUint16(serialized, d.ID)

	header := binary.BigEndian.Uint16(d.raw[2:4])
	// Set QR and AA fields to true
	header = header | 0b1000010010000000

	serialized = binary.BigEndian.AppendUint16(serialized, header)

	serialized = binary.BigEndian.AppendUint16(serialized, uint16(len(d.Questions)))
	serialized = binary.BigEndian.AppendUint16(serialized, uint16(len(d.Answers)))
	serialized = binary.BigEndian.AppendUint16(serialized, uint16(len(d.NameServers)))
	serialized = binary.BigEndian.AppendUint16(serialized, uint16(len(d.Extra)))

	for _, q := range d.Questions {
		serialized = append(serialized, serializeLabels(strings.Split(q.Name, "."))...)
		serialized = binary.BigEndian.AppendUint16(serialized, q.Type)
		serialized = binary.BigEndian.AppendUint16(serialized, q.Class)
	}

	for _, a := range d.Answers {
		serialized = append(serialized, a.Serialize()...)
	}

	return serialized
}

func (d *Message) QuestionType(q DnsQuestion) string {
	switch q.Type {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 3:
		return "MD"
	case 4:
		return "MF"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 12:
		return "PTR"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	}

	return ""
}

func (d *Message) AddAnswer(name string, answerType uint16, class uint16, data RData) {
	if d.Answers == nil {
		d.Answers = []ResourceRecord{}
	}

	s, err := data.Serialize()
	if err != nil {
		fmt.Printf("DnsMessage.AddAnswer: %s\n", err.Error())
		return
	}

	d.Answers = append(d.Answers, ResourceRecord{
		Name:  name,
		Type:  answerType,
		Class: class,
		TTL:   60,
		Data:  s,
	})
	d.AnswerCount += 1
}

// deserializeLabels reads a byte slice and parses the labels it contains.
// It returns the labels it found as a slice of strings, and returns the total
// number of bytes read while
func deserializeLabels(raw []byte) ([]string, uint8) {
	labels := []string{}
	cursor := uint8(0)

	for moreLabels := true; moreLabels; {
		if raw[cursor] == 0 {
			cursor += 1
			moreLabels = false
			continue
		}

		labelLength := raw[cursor]
		labelStart := cursor + 1
		labelEnd := labelStart + labelLength
		labels = append(labels, string(raw[labelStart:labelEnd]))
		cursor += labelLength + 1
	}

	labels = append(labels, "")

	return labels, cursor
}
