package main

import (
    "net"
    "fmt"
    "encoding/binary"
    "strings"
    "bufio"
    "os"
)

type PacketType int64

const (
    Message PacketType = 0
)

type Packet struct {
    Type PacketType
    Name string
    Data string
}

func (p *Packet) ToBytes() []byte {
    packet_text := fmt.Sprintf("%s:%s", p.Name, p.Data)
    packet := make([]byte, 8)
    binary.PutVarint(packet, int64(p.Type))
    return append(packet, []byte(packet_text)...)
}

func FromBytes(buffer []byte) Packet {
    kind, err := binary.Varint(buffer[:8])
    if err <= 0 {
        panic(err)
    }
    v := strings.SplitN(string(buffer[8:]), ":", 2) 
    return Packet { PacketType(kind), v[0], v[1] }
}

func main() {
    conn, err := net.ListenPacket("udp4", ":8829")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    fmt.Println(conn.LocalAddr())

    broadcast, err := net.ResolveUDPAddr("udp4", "192.168.10.255:8829")
    if err != nil {
        panic(err)
    }


    stdin := make(chan string)
    go func() {
        for {
            reader := bufio.NewReader(os.Stdin)
            text, _ := reader.ReadString('\n')
            stdin <- text[:len(text) - 1] // Remove \n
        }
    }()

    network := make(chan Packet)

    go func() {
        data := [2048]byte{0}
        for {
            size, from, err := conn.ReadFrom(data[:])
            if err != nil {
                return;
            }
            text := string(data[:size])
            network <- Packet{ Message, from.String(), text }
        }
    }()

    for {
        select {
            case prompt := <- stdin:
                if prompt[:3] == "/to" {
                    words := strings.SplitN(prompt, " ", 3)
                    if len(words) < 3 {
                        fmt.Println("Error: not enough words")
                        continue;
                    }
                    addr, err := net.ResolveUDPAddr("udp4", words[1] + ":8829")
                    if err != nil {
                        fmt.Printf("Error: failed to parse ip: %v", words[1])
                        continue;
                    }
                    conn.WriteTo([]byte(words[2]), addr)
                } else if prompt[:5] == "/quit" {
                    return;
                } else {
                    conn.WriteTo([]byte(prompt), broadcast)
                }
            case packet := <- network:
                switch packet.Type {
                    case Message:
                        fmt.Printf("[%v]: %v\n", packet.Name, packet.Data)
                }
        }
    }
}
