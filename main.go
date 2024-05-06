package main

import (
    "net"
    "fmt"
)

func main() {
    conn, err := net.ListenPacket("udp4", ":8829")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    broadcast_addr, err := net.ResolveUDPAddr("udp4", "192.168.10.255:8829")
    if err != nil {
        panic(err)
    }
    conn.WriteTo([]byte("Hellope"), broadcast_addr)

    data := [2048]byte{0}
    size, from, err := conn.ReadFrom(data[:])
    if err != nil {
        panic(err)
    }
    fmt.Printf("Received %v from %v\n", string(data[:size]), from)
    size, from, err = conn.ReadFrom(data[:])
    if err != nil {
        panic(err)
    }
    fmt.Printf("Received %v from %v\n", string(data[:size]), from)
}
