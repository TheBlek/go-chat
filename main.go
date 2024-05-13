package main

import (
    "net"
    "fmt"
    "encoding/binary"
    "strings"
    "bufio"
    "os"
    "strconv"
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
    addrs, err := net.InterfaceAddrs()
    
    fmt.Println("Select which ip address to use:")
    for index, addr := range addrs {
        switch a := addr.(type) {
        case *net.IPNet:
            fmt.Printf("[%v]: %v\n", index, a);
            break;
        }
    }
    reader := bufio.NewReader(os.Stdin)
    input, _ := reader.ReadString('\n')
    id, err := strconv.Atoi(input[:len(input)-1])
    if err != nil || id < 0 || id >= len(addrs) {
        fmt.Println("Invalid id");
        return;
    }
    addr := addrs[id].(*net.IPNet)

    ip := make(net.IP, len(addr.IP.To4()))
    binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(addr.IP.To4())|^binary.BigEndian.Uint32(net.IP(addr.Mask).To4()))
    broadcast := net.UDPAddr { IP: ip, Port: 8829, Zone: "" }

    conn, err := net.ListenPacket("udp4", ":8829")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

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
            if !from.(*net.UDPAddr).IP.Equal(addr.IP) {
                network <- Packet{ Message, from.String(), text }
            }
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
                    conn.WriteTo([]byte(prompt), &broadcast)
                }
            case packet := <- network:
                switch packet.Type {
                    case Message:
                        fmt.Printf("[%v]: %v\n", packet.Name, packet.Data)
                }
        }
    }
}
