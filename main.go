package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ktbartholomew/go-dns-server/v2/dns"
)

func xxd(b []byte) {
	const bytesPerLine = 16
	for i, byteVal := range b {
		// Print offset
		if i%bytesPerLine == 0 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%07x: ", i)
		}

		// Print byte in hex
		fmt.Printf("%02x ", byteVal)

		// Print ASCII representation
		if i%bytesPerLine == bytesPerLine-1 {
			fmt.Print(" ")
			for j := i - (bytesPerLine - 1); j <= i; j++ {
				if b[j] >= 32 && b[j] <= 126 {
					fmt.Printf("%c", b[j])
				} else {
					fmt.Print(".")
				}
			}
		}
	}

	// Print remaining ASCII characters for the last line if necessary
	remainingBytes := len(b) % bytesPerLine
	if remainingBytes != 0 {
		spaces := (bytesPerLine - remainingBytes) * 3
		fmt.Print("   ")
		fmt.Print(strings.Repeat(" ", spaces))
		for i := len(b) - remainingBytes; i < len(b); i++ {
			if b[i] >= 32 && b[i] <= 126 {
				fmt.Printf("%c", b[i])
			} else {
				fmt.Print(".")
			}
		}
	}

	fmt.Println()
}

func trimZeroBytes(b []byte) []byte {
	lastIndex := len(b) - 1
	for lastIndex >= 0 && b[lastIndex] == 0 {
		lastIndex--
	}
	return b[:lastIndex+1]
}

func main() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 5553})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("listening for DNS queries on :5553")

	for {
		msg := make([]byte, 512)

		_, addr, err := conn.ReadFromUDP(msg[0:])
		if err != nil {
			fmt.Println(err.Error())
		}

		msg = trimZeroBytes(msg)

		xxd(msg)
		m := &dns.Message{}
		m.Deserialize(msg)

		if m.IsQuery && len(m.Questions) > 0 {
			fmt.Printf("%s %d %s\n", m.Questions[0].Name, m.Questions[0].Class, m.QuestionType(m.Questions[0]))
		}

		q := m.Questions[0]

		switch q.Type {
		case 1:
			m.AddAnswer(q.Name, q.Type, q.Class, dns.AData{IPAddr: "192.168.7.93"})
		case 5:
			m.AddAnswer(q.Name, q.Type, q.Class, dns.CNAMEData{Name: "canonical.example.com."})
		}

		xxd(m.Serialize())

		conn.WriteToUDP(m.Serialize(), addr)
	}

}
