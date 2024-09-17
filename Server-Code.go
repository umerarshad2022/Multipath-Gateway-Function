// server
package main

import (
    "fmt"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "github.com/google/gopacket/pcap"
    "net"
    "os"
    "sync"
    "syscall"
)

const (
    IPPROTO_TCP   = 6
    TCP_MULTIPATH = 42 
)

type NATEntry struct {
    InternalIP   string
    InternalPort layers.TCPPort
    ExternalIP   string
    ExternalPort layers.TCPPort
}

var natTable = make(map[string]NATEntry) // NAT table to track internal to external mappings
var externalIP = "100.87.50.2"           // External IP address of the gateway

func main() {
    var wg sync.WaitGroup

    // Capture traffic on ens5
    wg.Add(1)
    go func() {
        defer wg.Done()
        captureTraffic("ens5", handleCapturedPacket)
    }()

    // Listen for incoming connections from the host
    wg.Add(1)
    go func() {
        defer wg.Done()
        startListener("10.10.10.1:8080", handleHostConnection)
    }()

    // Listen for incoming connections from the server on interface 1
    wg.Add(1)
    go func() {
        defer wg.Done()
        startListener("100.87.50.2:8080", handleServerConnection)
    }()

    wg.Wait()
}

// Capturing traffic from a specific interface and handling it
func captureTraffic(interfaceName string, packetHandler func(gopacket.Packet)) {
    handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
    if err != nil {
        fmt.Println("Error opening device", interfaceName, ":", err)
        os.Exit(1)
    }
    defer handle.Close()

    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    for packet := range packetSource.Packets() {
        packetHandler(packet)
    }
}

// Process the captured packet and forward it to the server
func handleCapturedPacket(packet gopacket.Packet) {
    ipLayer := packet.Layer(layers.LayerTypeIPv4)
    tcpLayer := packet.Layer(layers.LayerTypeTCP)
    if ipLayer == nil || tcpLayer == nil {
     //   fmt.Println("Error: IP or TCP layer not found in captured packet")
        return
    }

    ipPacket, _ := ipLayer.(*layers.IPv4)
    tcpPacket, _ := tcpLayer.(*layers.TCP)

    // Perform NAT (modify only the source IP and forward the packet)
    translatedPacket := performNAT(ipPacket, tcpPacket, "192.168.10.2", externalIP)
    forwardToServer(translatedPacket)
}

// Start a listener for establishing incoming connections
func startListener(address string, handler func(net.Conn)) {
    listener, err := net.Listen("tcp", address)
    if err != nil {
        fmt.Println("Error starting TCP server on", address, ":", err)
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Gateway is listening on", address)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection on", address, ":", err)
            continue
        }

        go handler(conn)
    }
}

// Handle incoming connections from the host (internal network)
func handleHostConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err.Error() != "EOF" {
                fmt.Println("Error reading from host connection:", err)
            }
            return
        }

        packet := gopacket.NewPacket(buffer[:n], layers.LayerTypeIPv4, gopacket.Default)
        ipLayer := packet.Layer(layers.LayerTypeIPv4)
        tcpLayer := packet.Layer(layers.LayerTypeTCP)
        if ipLayer == nil || tcpLayer == nil {
          //  fmt.Println("Error: IP or TCP layer not found")
            return
        }
        ipPacket := ipLayer.(*layers.IPv4)
        tcpPacket := tcpLayer.(*layers.TCP)

        // Perform NAT and forward the message to the server using MPTCP
        translatedData := performNAT(ipPacket, tcpPacket, "10.10.10.1", externalIP)
        forwardToServer(translatedData)
    }
}

// Handle incoming connections from the server (external network)
func handleServerConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err.Error() != "EOF" {
                fmt.Println("Error reading from server connection:", err)
            }
            return
        }

        packet := gopacket.NewPacket(buffer[:n], layers.LayerTypeIPv4, gopacket.Default)
        ipLayer := packet.Layer(layers.LayerTypeIPv4)
        tcpLayer := packet.Layer(layers.LayerTypeTCP)
        if ipLayer == nil || tcpLayer == nil {
          //  fmt.Println("Error: IP or TCP layer not found")
            return
        }
        ipPacket := ipLayer.(*layers.IPv4)
        tcpPacket := tcpLayer.(*layers.TCP)

        // Perform NAT (reverse) and forward the message to the host
        translatedData := performNAT(ipPacket, tcpPacket, externalIP, "10.10.10.1")
        forwardToHost(translatedData)
    }
}

// Forward traffic to the server with MPTCP
func forwardToServer(packet []byte) {
    conn, err := dialMPTCP("100.87.50.1:8080", "ens8") // Primary interface for MPTCP
    if err != nil {
        fmt.Println("Error connecting to server via MPTCP:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write(packet)
    if err != nil {
        fmt.Println("Error forwarding packet via MPTCP:", err)
        return
    }

    fmt.Println("Packet forwarded via MPTCP")
}

// Forward traffic back to the host
func forwardToHost(packet []byte) {
    conn, err := net.Dial("tcp", "10.10.10.2:8080")
    if err != nil {
        fmt.Println("Error connecting to host:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write(packet)
    if err != nil {
        fmt.Println("Error forwarding to host:", err)
        return
    }

    fmt.Println("Packet forwarded to host")
}

// Dial using MPTCP function
func dialMPTCP(address, iface string) (net.Conn, error) {
    raddr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        return nil, err
    }

    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, IPPROTO_TCP)
    if err != nil {
        return nil, err
    }

    ifaceByteArray := []byte(iface)
    if err = syscall.SetsockoptString(fd, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, string(ifaceByteArray)); err != nil {
        syscall.Close(fd)
        return nil, err
    }

    if err = syscall.SetsockoptInt(fd, IPPROTO_TCP, TCP_MULTIPATH, 1); err != nil {
        syscall.Close(fd)
        return nil, err
    }

    sockaddr := &syscall.SockaddrInet4{Port: raddr.Port}
    copy(sockaddr.Addr[:], raddr.IP.To4())

    if err = syscall.Connect(fd, sockaddr); err != nil {
        syscall.Close(fd)
        return nil, err
    }

    file := os.NewFile(uintptr(fd), "")
    conn, err := net.FileConn(file)
    if err != nil {
        file.Close()
        return nil, err
    }

    return conn, nil
}

// Perform basic NAT by modifying source and destination IP addresses
func performNAT(ipPacket *layers.IPv4, tcpPacket *layers.TCP, fromIP, toIP string) []byte {
    // Change the source or destination IP based on whether the packet is incoming or outgoing
    if ipPacket.SrcIP.String() == fromIP {
        ipPacket.SrcIP = net.ParseIP(toIP)
    } else if ipPacket.DstIP.String() == fromIP {
        ipPacket.DstIP = net.ParseIP(toIP)
    }

    // Serialize packet back to bytes without checksums
    buffer := gopacket.NewSerializeBuffer()
    options := gopacket.SerializeOptions{} // No checksum recalculations
    err := gopacket.SerializeLayers(buffer, options, ipPacket, tcpPacket)
    if err != nil {
        fmt.Println("Error serializing packet:", err)
        return nil
    }
    return buffer.Bytes()
}
