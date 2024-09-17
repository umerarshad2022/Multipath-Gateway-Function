//Main Host - Bidirectional Communication

package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

func main() {
    // Start the listener for incoming messages from the gateway
    go startListener("192.168.10.1:8080") // Host IP and port

    // Send messages to Node-B
    for {
        sendMessage("10.10.10.2:8080", "Hello from Host 1") // Destination IP and port
        time.Sleep(5 * time.Second)
    }
}

func startListener(address string) {
    listener, err := net.Listen("tcp", address)
    if err != nil {
        fmt.Println("Error starting TCP server:", err)
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Host is listening on", address)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err.Error() != "EOF" {
                fmt.Println("Error reading from connection:", err)
            }
            return
        }

        fmt.Println("Received message:", string(buffer[:n]))
    }
}

func sendMessage(address, message string) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        fmt.Println("Error connecting:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Println("Error sending message:", err)
        return
    }

    fmt.Println("Message sent:", message)
}
