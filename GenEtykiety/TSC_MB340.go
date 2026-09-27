package main

import (
	"fmt"
	"net"
	"time"
)

func socketSendReceive(address string, data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// wyślij
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	// timeout na odpowiedź
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// odbierz
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	fmt.Println(fmt.Sprintf("Odpowiedź: bajtów(%d), response: %s", n, string(buf[:n])))
	return buf[:n], nil
}

// func main() {
// 	response, err := socketSendReceive(
// 		"192.168.20.158:9100",
// 		[]byte{0x1B, 0x21, 0x53, 0x0D, 0x0A},
// 	)
// 	if err != nil {
// 		fmt.Println("Błąd:", err)
// 		return
// 	}

// 	fmt.Printf("Odpowiedź: % X\n", response)
// }
