package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/OpenPrinting/goipp"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   IPP + SOCKET 9100")
	fmt.Println("========================================\n")

	// ========================================
	// 1. IPP PRZEZ HTTP (port 631)
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. IPP PRZEZ HTTP (CUPS)")
	fmt.Println("========================================\n")

	client := &http.Client{Timeout: 30 * time.Second}

	// Get-Printer-Attributes
	fmt.Println("--- Get-Printer-Attributes ---")

	req1 := goipp.NewRequest(0x0101, 0x000B, 1)
	req1.Operation.Add(goipp.MakeAttribute("attributes-charset", goipp.TagCharset, goipp.String("utf-8")))
	req1.Operation.Add(goipp.MakeAttribute("attributes-natural-language", goipp.TagLanguage, goipp.String("en")))
	req1.Operation.Add(goipp.MakeAttribute("printer-uri", goipp.TagURI, goipp.String("ipp://localhost:631/ipp/print")))
	req1.Operation.Add(goipp.MakeAttribute("requested-attributes", goipp.TagKeyword, goipp.String("all")))

	data1, _ := req1.EncodeBytes()
	fmt.Printf("Wysyłam: %d bajtów IPP\n", len(data1))

	resp1, err := client.Post("http://localhost:631/ipp/print", "application/ipp", bytes.NewReader(data1))
	if err != nil {
		fmt.Printf("Błąd: %v\n", err)
		return
	}
	defer resp1.Body.Close()

	body1, _ := io.ReadAll(resp1.Body)
	fmt.Printf("Otrzymano: %d bajtów\n", len(body1))

	var msg1 goipp.Message
	if err := msg1.Decode(bytes.NewReader(body1)); err != nil {
		fmt.Printf("Błąd dekodowania: %v\n", err)
	} else {
		fmt.Printf("Status IPP: 0x%04X ✅\n", msg1.Code)
	}

	// Print-Job przez HTTP
	fmt.Println("\n--- Print-Job przez HTTP ---")

	docContent := []byte(
		"====================================\n" +
			"TEST DOCUMENT - IPP HTTP\n" +
			"====================================\n" +
			"Data: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
			"Druk przez IPP/HTTP\n" +
			"====================================\n",
	)

	req2 := goipp.NewRequest(0x0101, 0x0002, 2)
	req2.Operation.Add(goipp.MakeAttribute("attributes-charset", goipp.TagCharset, goipp.String("utf-8")))
	req2.Operation.Add(goipp.MakeAttribute("attributes-natural-language", goipp.TagLanguage, goipp.String("en")))
	req2.Operation.Add(goipp.MakeAttribute("printer-uri", goipp.TagURI, goipp.String("ipp://localhost:631/ipp/print")))
	req2.Operation.Add(goipp.MakeAttribute("document-name", goipp.TagName, goipp.String("test_http.txt")))
	req2.Operation.Add(goipp.MakeAttribute("document-format", goipp.TagKeyword, goipp.String("text/plain")))
	req2.Operation.Add(goipp.MakeAttribute("job-name", goipp.TagName, goipp.String("Test HTTP")))

	reqData2, _ := req2.EncodeBytes()
	fullData2 := append(reqData2, docContent...)

	fmt.Printf("Wysyłam: %d bajtów (żądanie: %d + dokument: %d)\n",
		len(fullData2), len(reqData2), len(docContent))

	resp2, err := client.Post("http://localhost:631/ipp/print", "application/ipp", bytes.NewReader(fullData2))
	if err != nil {
		fmt.Printf("Błąd: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Otrzymano: %d bajtów\n", len(body2))

	var msg2 goipp.Message
	if err := msg2.Decode(bytes.NewReader(body2)); err != nil {
		fmt.Printf("Błąd dekodowania: %v\n", err)
	} else {
		fmt.Printf("Status IPP: 0x%04X\n", msg2.Code)
		if msg2.Code == 0x0000 {
			fmt.Println("✅ Drukowanie przez HTTP udane!")
		} else {
			fmt.Printf("❌ Błąd: 0x%04X\n", msg2.Code)
		}
	}

	// ========================================
	// 2. SOCKET NA PORCIE 9100 (RAW)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. SOCKET NA PORCIE 9100 (RAW)")
	fmt.Println("========================================\n")

	fmt.Println("--- Łączenie z portem 9100 ---")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:9100", 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Błąd połączenia na porcie 9100: %v\n", err)
		fmt.Println("   Sprawdź czy drukarka jest podłączona i nasłuchuje na porcie 9100")
		fmt.Println("   (To jest port dla protokołu RAW/JetDirect)")
	} else {
		defer conn.Close()
		fmt.Println("✅ Połączono z portem 9100\n")

		// Wysyłamy dane przez socket 9100
		rawData := []byte(
			"====================================\n" +
				"TEST - SOCKET 9100\n" +
				"====================================\n" +
				"Data: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
				"To jest test drukowania przez port 9100 (RAW socket)\n" +
				"====================================\n" +
				"\x0C", // Form Feed - koniec strony
		)

		fmt.Printf("Wysyłam przez socket 9100: %d bajtów\n", len(rawData))

		_, err = conn.Write(rawData)
		if err != nil {
			fmt.Printf("❌ Błąd wysyłania: %v\n", err)
		} else {
			fmt.Println("✅ Dane wysłane przez socket 9100")

			// Odczytujemy odpowiedź (jeśli jest)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			resp, err := io.ReadAll(conn)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					fmt.Println("   (Brak odpowiedzi - normalne dla RAW socket)")
				} else {
					fmt.Printf("   Błąd odczytu: %v\n", err)
				}
			} else if len(resp) > 0 {
				fmt.Printf("Odpowiedź: %d bajtów\n", len(resp))
				fmt.Printf("Hex: %x\n", resp[:min(64, len(resp))])
			}
		}
	}

	// ========================================
	// 3. SOCKET NA PORCIE 9100 - POSTSCRIPT
	// ========================================
	fmt.Println("\n--- Socket 9100 - PostScript ---")

	conn2, err := net.DialTimeout("tcp", "127.0.0.1:9100", 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Błąd: %v\n", err)
	} else {
		defer conn2.Close()
		fmt.Println("✅ Połączono")

		// Prosty PostScript
		psData := []byte(
			"%!PS-Adobe-3.0\n" +
				"%%Title: Test PS\n" +
				"%%Creator: Go\n" +
				"%%Pages: 1\n" +
				"/Times-Roman findfont 24 scalefont setfont\n" +
				"50 700 moveto\n" +
				"(Test PostScript przez socket 9100) show\n" +
				"50 650 moveto\n" +
				"(Data: " + time.Now().Format("2006-01-02 15:04:05") + ") show\n" +
				"showpage\n" +
				"%%EOF\n",
		)

		fmt.Printf("Wysyłam PostScript: %d bajtów\n", len(psData))

		_, err = conn2.Write(psData)
		if err != nil {
			fmt.Printf("❌ Błąd: %v\n", err)
		} else {
			fmt.Println("✅ PostScript wysłany przez socket 9100")
		}
	}

	// ========================================
	// 4. SOCKET NA PORCIE 9100 - PCL
	// ========================================
	fmt.Println("\n--- Socket 9100 - PCL ---")

	conn3, err := net.DialTimeout("tcp", "127.0.0.1:9100", 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Błąd: %v\n", err)
	} else {
		defer conn3.Close()
		fmt.Println("✅ Połączono")

		// Prosty PCL
		pclData := []byte(
			"\x1B%-12345X@PJL JOB\n" +
				"@PJL ENTER LANGUAGE=PCL\n" +
				"\x1BE\x1B&l0O\x1B&l1X\x1B&l1Z\x1B&l0U" +
				"\x1B(s0p12h0s0b4099T\x1B&a0R\x1B&a0C" +
				"Test PCL przez socket 9100\n" +
				"Data: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
				"\x1B%-12345X@PJL EOJ\n" +
				"\x1B%-12345X",
		)

		fmt.Printf("Wysyłam PCL: %d bajtów\n", len(pclData))

		_, err = conn3.Write(pclData)
		if err != nil {
			fmt.Printf("❌ Błąd: %v\n", err)
		} else {
			fmt.Println("✅ PCL wysłany przez socket 9100")
		}
	}

	// ========================================
	// PODSUMOWANIE
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("   PODSUMOWANIE")
	fmt.Println("========================================")
	fmt.Println("✅ IPP przez HTTP - komunikacja z CUPS")
	fmt.Println("✅ Socket 9100 - bezpośrednie drukowanie")
	fmt.Println("\nFormaty wysłane przez socket 9100:")
	fmt.Println("  - Tekst (RAW)")
	fmt.Println("  - PostScript")
	fmt.Println("  - PCL")
	fmt.Println("\nPort 9100 to standardowy port dla:")
	fmt.Println("  - JetDirect (HP)")
	fmt.Println("  - RAW printing")
	fmt.Println("  - Bezpośrednie drukowanie bez CUPS")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
