package main

import (
	"fmt"
	"net"
	"os"

	"github.com/ktbartholomew/go-dns-server/v2/dns"
)

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

		fmt.Printf("%+v\n", m)
		conn.WriteToUDP(m.Serialize(), addr)
	}

}
