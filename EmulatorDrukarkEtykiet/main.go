package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/OpenPrinting/goipp"
	"github.com/evermile/go-zpl" //go get github.com/evermile/go-zpl
)

const (
	ippPort = 631
	rawPort = 9100

	printerName = "PiKul Virtual Printer"
	ippPath     = "/ipp/print"
	jobDir      = "jobs"
)

var (
	jobMutex sync.Mutex
	jobID    uint32
)

func handleConnection1(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Połączenie 9100:", conn.RemoteAddr())

	err := os.MkdirAll("jobs", 0755)
	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		return
	}

	filename := fmt.Sprintf("jobs/print_%s", time.Now().Format("20060102_150405.000"))

	file, err := os.Create(filename + ".bin")
	if err != nil {
		fmt.Println("Błąd pliku:", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 65536) // większy bufor tymczasowy

	for {
		n, err := conn.Read(buf)

		if n > 0 {
			fmt.Printf("9100: odebrano %d bajtów\n", n)
			// Sprawdza, czy to jest kod ZPL
			if len(buf) >= 2 && ((buf[0] == '^' && buf[1] == 'X') || buf[0] == '~') {
				fmt.Println("Wykryto kod ZPL, konwertowanie do PDF...")
				outputFile, err := createOutputWriter(filename + ".pdf")
				if err == nil {
					err := convertZPLToPDF(buf[:n], outputFile)
					if err != nil {
						fmt.Println("Błąd konwersji ZPL do PDF:", err)
					}
					if info, err := outputFile.Stat(); err == nil {
						fmt.Printf("Rozmiar pliku PDF: %d bajtów\n", info.Size())
					}
					outputFile.Close()
				}
			}

			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				fmt.Println("Błąd zapisu:", writeErr)
				return
			}
		}

		if err != nil {
			if err != io.EOF {
				fmt.Println("Błąd 9100:", err)
			}
			break
		}
	}

	fmt.Println("Rozłączono 9100:", conn.RemoteAddr())
}
func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Połączenie 9100:", conn.RemoteAddr())

	err := os.MkdirAll("jobs", 0755)
	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405.000")
	baseFilename := fmt.Sprintf("jobs/print_%s", timestamp)

	// Zapisujemy surowe dane do pliku binarnego
	binFile, err := os.Create(baseFilename + ".bin")
	if err != nil {
		fmt.Println("Błąd pliku:", err)
		return
	}
	defer binFile.Close()

	// Zbieramy wszystkie dane
	var allData bytes.Buffer
	buf := make([]byte, 65536)

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			allData.Write(buf[:n])
			binFile.Write(buf[:n])
			fmt.Printf("9100: odebrano %d bajtów (łącznie: %d)\n", n, allData.Len())
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("Błąd 9100:", err)
			}
			break
		}
	}

	data := allData.Bytes()
	fmt.Printf("Łącznie odebrano %d bajtów\n", len(data))

	if len(data) == 0 {
		fmt.Println("Brak danych")
		return
	}

	// Sprawdź czy to ZPL
	if len(data) >= 2 && ((data[0] == '^' && data[1] == 'X') || data[0] == '~') {
		fmt.Println("Wykryto kod ZPL, konwertowanie do PDF...")

		copies := splitZPLCopies(data)
		fmt.Printf("Znaleziono %d kopii ZPL\n", len(copies))

		if len(copies) == 0 {
			fmt.Println("Nie znaleziono kopii ZPL")
			return
		}

		// Grupuj po 50 etykiet na plik PDF
		groupSize := 50
		numGroups := (len(copies) + groupSize - 1) / groupSize // zaokrąglenie w górę
		fmt.Printf("Dzielę na %d grup po %d etykiet\n", numGroups, groupSize)

		for groupIdx := 0; groupIdx < numGroups; groupIdx++ {
			start := groupIdx * groupSize
			end := start + groupSize
			if end > len(copies) {
				end = len(copies)
			}

			group := copies[start:end]
			fmt.Printf("Grupa %d: etykiety %d-%d (%d sztuk)\n",
				groupIdx+1, start+1, end, len(group))

			// Połącz wszystkie etykiety z grupy w jeden ZPL
			var combinedZPL bytes.Buffer
			for _, zplData := range group {
				combinedZPL.Write(zplData)
			}

			// Zapisz jako pojedynczy PDF (z wieloma stronami)
			pdfFilename := fmt.Sprintf("%s_%d_%d.pdf", baseFilename, start+1, end)
			outputFile, err := os.Create(pdfFilename)
			if err != nil {
				fmt.Printf("Błąd tworzenia pliku: %v\n", err)
				continue
			}

			// Konwertuj połączony ZPL do PDF
			err = convertZPLToPDF(combinedZPL.Bytes(), outputFile)
			outputFile.Close()

			if err != nil {
				fmt.Printf("Błąd konwersji grupy %d: %v\n", groupIdx+1, err)
			} else {
				if info, err := os.Stat(pdfFilename); err == nil && info.Size() > 0 {
					fmt.Printf("✅ Zapisano: %s (%d bajtów, %d etykiet)\n",
						pdfFilename, info.Size(), len(group))
				} else {
					fmt.Printf("⚠️ Plik %s jest pusty!\n", pdfFilename)
				}
			}

			// Opóźnienie między grupami
			if groupIdx < numGroups-1 {
				time.Sleep(200 * time.Millisecond)
			}
		}

	} else {
		fmt.Printf("To nie jest ZPL - zapisano jako %s.bin\n", baseFilename)
		fmt.Printf("Pierwsze bajty: %x\n", data[:min(16, len(data))])
	}

	fmt.Println("Rozłączono 9100:", conn.RemoteAddr())
}

// Funkcja dzieląca strumień ZPL na pojedyncze kopie
func splitZPLCopies(data []byte) [][]byte {
	var copies [][]byte
	start := -1

	for i := 0; i < len(data)-3; i++ {
		// Szukaj początku: ^XA
		if data[i] == '^' && data[i+1] == 'X' && data[i+2] == 'A' {
			if start != -1 {
				// Szukaj końca poprzedniej: ^XZ
				for j := start; j < i-2; j++ {
					if data[j] == '^' && data[j+1] == 'X' && data[j+2] == 'Z' {
						copies = append(copies, data[start:j+3])
						break
					}
				}
			}
			start = i
		}
	}

	// Dodaj ostatnią kopię
	if start != -1 {
		for j := start; j < len(data)-2; j++ {
			if data[j] == '^' && data[j+1] == 'X' && data[j+2] == 'Z' {
				copies = append(copies, data[start:j+3])
				break
			}
		}
	}

	return copies
}

func getPrinterURI(r *http.Request) string {
	host := r.Host

	if host == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "localhost"
		}

		host = fmt.Sprintf(
			"%s:%d",
			hostname,
			ippPort,
		)
	}

	return "ipp://" + host + ippPath
}

func handleIPP(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("IPP:", r.Method, r.URL)
	fmt.Println("Klient:", r.RemoteAddr)
	fmt.Println("Host:", r.Host)
	fmt.Println("Content-Type:", r.Header.Get("Content-Type"))
	fmt.Println("Content-Length:", r.ContentLength)

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"IPP requires POST",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req goipp.Message

	err := req.Decode(r.Body)

	if err != nil {
		fmt.Println("Błąd dekodowania IPP:", err)

		/*
			Na tym etapie zwracamy zwykłe HTTP 400.
			Nie używamy żadnego nieznanego statusu goipp.
		*/

		http.Error(
			w,
			"Invalid IPP request",
			http.StatusBadRequest,
		)

		return
	}

	fmt.Println("IPP Version:", req.Version)
	fmt.Println("IPP Operation:", req.Code)
	fmt.Println("IPP Request ID:", req.RequestID)

	fmt.Println("----- IPP REQUEST -----")
	req.Print(os.Stdout, true)
	fmt.Println("-----------------------")

	switch goipp.Op(req.Code) {
	case goipp.OpGetPrinterAttributes:
		handleGetPrinterAttributes(w, r, &req)

	case goipp.OpPrintJob:
		//handleCancelJob(w, &req) //test
		handlePrintJob(w, r, &req)

	case goipp.OpValidateJob:
		handleValidateJob(w, &req)

	case goipp.OpGetJobs:
		handleGetJobs(w, &req)

	case goipp.OpGetJobAttributes:
		handleGetJobAttributes(w, &req)

	case goipp.OpCancelJob:
		handleCancelJob(w, &req)

	default:
		fmt.Println("Nieobsługiwana operacja IPP:", req.Code)

		/*
			Na razie odpowiadamy poprawnym komunikatem IPP
			o statusie OK, żeby nie uzależniać kodu od
			nazw statusów konkretnej wersji goipp.
		*/

		resp := goipp.NewResponse(
			req.Version,
			goipp.StatusOk,
			req.RequestID,
		)

		addOperationAttributes(resp)
		writeIPPResponse(w, resp)
	}

	fmt.Println("========================================")
}

func handleGetPrinterAttributes(
	w http.ResponseWriter,
	r *http.Request,
	req *goipp.Message,
) {
	fmt.Println("IPP: Get-Printer-Attributes")

	printerURI := getPrinterURI(r)

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

	// resp.Printer.Add(goipp.MakeAttribute(
	// 	"printer-state",
	// 	goipp.TagEnum,
	// 	goipp.Integer(3),
	// ))

	// resp.Printer.Add(goipp.MakeAttribute(
	// 	"printer-state-reasons",
	// 	goipp.TagKeyword,
	// 	goipp.String("none"),
	// ))
	// 	resp.Printer.Add(goipp.MakeAttribute(
	// 	"printer-is-accepting-jobs",
	// 	goipp.TagBoolean,
	// 	goipp.Boolean(true),
	// ))
	// resp.Printer.Add(goipp.MakeAttribute(
	// 	"queued-job-count",
	// 	goipp.TagInteger,
	// 	goipp.Integer(0),
	// ))

	// fmt.Println("----- IPP RESPONSE -----")
	// resp.Print(os.Stdout, true)
	// fmt.Println("-----------------------")
	// writeIPPResponse(w, resp)
	// return
	resp.Printer.Add(goipp.MakeAttribute(
		"document-format-default",
		goipp.TagMimeType,
		goipp.String("application/octet-stream"),
	))

	// resp.Printer.Add(goipp.MakeAttribute(
	// 	"printer-uri-supported",
	// 	goipp.TagURI,
	// 	goipp.String(printerURI),
	// ))
	resp.Printer.Add(goipp.MakeAttribute(
		"printer-uri-supported",
		goipp.TagURI,
		goipp.String("ipp://127.0.0.1:631/ipp/print"),
	))
	resp.Printer.Add(goipp.MakeAttribute(
		"printer-name",
		goipp.TagName,
		goipp.String(printerName),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-info",
		goipp.TagText,
		goipp.String("PiKul Virtual Printer"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-location",
		goipp.TagText,
		goipp.String(""),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-make-and-model",
		goipp.TagText,
		goipp.String("PiKul Virtual Printer"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-more-info",
		goipp.TagURI,
		goipp.String(printerURI),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-uuid",
		goipp.TagURI,
		goipp.String("urn:uuid:12345678-1234-1234-1234-123456789abc"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-state",
		goipp.TagEnum,
		goipp.Integer(3),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-state-reasons",
		goipp.TagKeyword,
		goipp.String("none"),
	))

	resp.Printer.Add(goipp.Attribute{
		Name: "ipp-versions-supported",
		Values: goipp.Values{
			{T: goipp.TagKeyword, V: goipp.String("1.1")},
			{T: goipp.TagKeyword, V: goipp.String("2.0")},
		},
	})

	resp.Printer.Add(goipp.Attribute{
		Name: "document-format-supported",
		Values: goipp.Values{
			{T: goipp.TagMimeType, V: goipp.String("application/octet-stream")},
			{T: goipp.TagMimeType, V: goipp.String("application/pdf")},
			{T: goipp.TagMimeType, V: goipp.String("text/plain")},
		},
	})

	resp.Printer.Add(goipp.MakeAttribute(
		"document-format-default",
		goipp.TagMimeType,
		goipp.String("text/plain"),
		//goipp.String("application/octet-stream"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"document-format-preferred",
		goipp.TagMimeType,
		goipp.String("application/octet-stream"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"charset-configured",
		goipp.TagCharset,
		goipp.String("utf-8"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"charset-supported",
		goipp.TagCharset,
		goipp.String("utf-8"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"natural-language-configured",
		goipp.TagLanguage,
		goipp.String("pl"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"generated-natural-language-supported",
		goipp.TagLanguage,
		goipp.String("pl"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"uri-authentication-supported",
		goipp.TagKeyword,
		goipp.String("none"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"uri-security-supported",
		goipp.TagKeyword,
		goipp.String("none"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"color-supported",
		goipp.TagBoolean,
		goipp.Boolean(false),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"copies-supported",
		goipp.TagRange,
		goipp.Range{
			Lower: 1,
			Upper: 999,
		},
	))

	resp.Printer.Add(goipp.Attribute{
		Name: "media-supported",
		Values: goipp.Values{
			{T: goipp.TagKeyword, V: goipp.String("iso_a4_210x297mm")},
			{T: goipp.TagKeyword, V: goipp.String("na_letter_8.5x11in")},
		},
	})

	resp.Printer.Add(goipp.Attribute{
		Name: "media-type-supported",
		Values: goipp.Values{
			{T: goipp.TagKeyword, V: goipp.String("stationery")},
		},
	})

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-dns-sd-name",
		goipp.TagName,
		goipp.String(printerName),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-firmware-name",
		goipp.TagName,
		goipp.String("PiKul Virtual Printer"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-firmware-string-version",
		goipp.TagText,
		goipp.String("1.0"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-device-id",
		goipp.TagText,
		goipp.String("MFG:PiKul;MDL:Virtual Printer;CMD:PDF,RAW;"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"printer-is-accepting-jobs",
		goipp.TagBoolean,
		goipp.Boolean(true),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"mopria-certified",
		goipp.TagKeyword,
		goipp.String("false"),
	))

	resp.Printer.Add(goipp.MakeAttribute(
		"mopria_certified",
		goipp.TagKeyword,
		goipp.String("false"),
	))

	// resp.Printer.Add(goipp.MakeAttribute(
	// 	"queued-job-count",
	// 	goipp.TagInteger,
	// 	goipp.Integer(0),
	// ))

	// // Deklaracja wsparcia dla rozdzielczości 300 DPI
	// resolution := goipp.MakeCollection(
	//     "x-resolution", goipp.TagInteger, goipp.Integer(300),
	//     "y-resolution", goipp.TagInteger, goipp.Integer(300),
	// )
	// printerAttrs.Add(goipp.MakeAttribute("printer-resolution-supported", goipp.TagCollection, resolution))
	// printerAttrs.Add(goipp.MakeAttribute("printer-resolution-default", goipp.TagCollection, resolution))

	fmt.Println("----- IPP RESPONSE -----")
	resp.Print(os.Stdout, true)
	fmt.Println("-----------------------")
	writeIPPResponse(w, resp)
}

// func createOutputWriter1(filename string) (io.Writer, error) {
// 	// Tworzy plik (lub nadpisuje jeśli istnieje)
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// UWAGA: plik MUSI być zamknięty po zakończeniu!
// 	return file, nil
// }

func createOutputWriter(filename string) (*os.File, error) {
	return os.Create(filename)
}

func mmToInches(mm int) int {
	inches := float64(mm) / 25.4 // 1 cal = 25.4 mm
	return int(inches + 0.5)     // zaokrąglenie do pełnego cala
}

func convertZPLToPDF(zplData []byte, outputFile io.Writer) error {
	width := mmToInches(70)  // 70mm w calach
	height := mmToInches(80) // 80mm w calach
	reader := bytes.NewReader(zplData)
	zpl.Convert(
		zpl.WithInput(reader),      // []byte z kodem ZPL
		zpl.WithOutput(outputFile), // io.Writer (np. plik)
		zpl.WithOutputFormat(zpl.PDF),
		zpl.WithWidth(width),   // Szerokość w calach
		zpl.WithHeight(height), // Wysokość w calach
		zpl.WithDensity(12),    // Gęstość (8 dpmm = 203 DPI) 12 - 300DPI
	)
	return nil
}

func handlePrintJob1(w http.ResponseWriter, r *http.Request, req *goipp.Message) {
	fmt.Println("IPP: Print-Job")

	jobMutex.Lock()
	jobID++
	id := jobID
	jobMutex.Unlock()

	err := os.MkdirAll(
		jobDir,
		0755,
	)

	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		http.Error(w, "Cannot create job directory", http.StatusInternalServerError)
		return
	}

	/* v.T (Type) values:
	   0x21 (33) → integer
	   0x22 (34) → boolean
	   0x23 (35) → enum
	   0x41 (65) → textWithoutLanguage
	   0x42 (66) → nameWithoutLanguage
	   0x44 (68) → keyword
	   0x45 (69) → uri
	*/

	jobName := fmt.Sprintf("job_%d", id)
	for _, attr := range req.Operation {
		if attr.Name == "job-name" && len(attr.Values) > 0 {
			v := attr.Values[0]
			jobName = fmt.Sprintf("%d-%v", v.T, v.V)
			break
		}
	}
	filename := filepath.Join(jobDir, jobName+".bin")

	// filename := filepath.Join(
	// 	jobDir,
	// 	fmt.Sprintf(
	// 		"job_%d.bin",
	// 		id,
	// 	),
	// )

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Błąd tworzenia pliku:", err)
		http.Error(w, "Cannot create job file", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	/*
		Decode() przeczytał komunikat IPP.
		Reszta r.Body to dokument.
	*/

	bytesWritten, err := io.Copy(file, r.Body)

	if err != nil {
		fmt.Println("Błąd zapisu dokumentu:", err)
		http.Error(w, "Cannot save document", http.StatusInternalServerError)
		return
	}

	fmt.Printf("JOB %d: odebrano %d bajtów\n", id, bytesWritten)
	fmt.Println("Plik:", filename)

	printerURI := getPrinterURI(r)
	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)

	addOperationAttributes(resp)

	jobURI := fmt.Sprintf("%s/jobs/%d", printerURI, id)
	resp.Job.Add(goipp.MakeAttribute("job-id", goipp.TagInteger, goipp.Integer(id)))
	resp.Job.Add(goipp.MakeAttribute("job-uri", goipp.TagURI, goipp.String(jobURI)))
	resp.Job.Add(goipp.MakeAttribute("job-name", goipp.TagName, goipp.String(fmt.Sprintf("Job %d", id))))
	resp.Job.Add(goipp.MakeAttribute("job-state", goipp.TagEnum, goipp.Integer(9)))
	resp.Job.Add(goipp.MakeAttribute("job-state-reasons", goipp.TagKeyword, goipp.String("none")))
	fmt.Println("----- IPP RESPONSE JOB-----")
	resp.Print(os.Stdout, true)
	fmt.Println("-----------------------")
	writeIPPResponse(w, resp)
}

func handlePrintJob2(w http.ResponseWriter, r *http.Request, req *goipp.Message) {
	fmt.Println("IPP: Print-Job")

	jobMutex.Lock()
	jobID++
	id := jobID
	jobMutex.Unlock()

	err := os.MkdirAll(
		jobDir,
		0755,
	)

	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		http.Error(w, "Cannot create job directory", http.StatusInternalServerError)
		return
	}

	/*
	   v.T (Type) values:
	   0x21 (33) → integer
	   0x22 (34) → boolean
	   0x23 (35) → enum
	   0x41 (65) → textWithoutLanguage
	   0x42 (66) → nameWithoutLanguage
	   0x44 (68) → keyword
	   0x45 (69) → uri
	*/

	jobName := fmt.Sprintf("job_%d", id)
	copies := 1

	// Przeszukaj OPERATION attributes
	for _, attr := range req.Operation {
		fmt.Printf("Operation Attr: %s, Type: 0x%X, Values: %v\n", attr.Name, attr.Values[0].T, attr.Values[0].V)
		if attr.Name == "job-name" && len(attr.Values) > 0 {
			v := attr.Values[0]
			jobName = fmt.Sprintf("%d-%v", v.T, v.V)
		}
	}

	// Przeszukaj JOB attributes (tam jest "copies")
	for _, attr := range req.Job {
		fmt.Printf("Job Attr: %s, Type: 0x%X, Values: %v\n", attr.Name, attr.Values[0].T, attr.Values[0].V)
		if attr.Name == "copies" && len(attr.Values) > 0 {
			v := attr.Values[0]
			if val, ok := v.V.(goipp.Integer); ok {
				copies = int(val)
				fmt.Printf("Ustawiono copies: %d\n", copies)
			}
		}
	}

	// Dla debug - wyświetl wszystkie atrybuty z Job
	fmt.Println("----- ATRYBUTY JOB -----")
	for _, attr := range req.Job {
		for _, v := range attr.Values {
			fmt.Printf("Job Attr: %s, Type: 0x%X, Value: %v\n", attr.Name, v.T, v.V)
		}
	}
	fmt.Println("------------------------")

	// Odczyt dokumentu
	documentData, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Błąd odczytu dokumentu:", err)
		http.Error(w, "Cannot read document", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Odebrano dokument: %d bajtów, kopie: %d\n", len(documentData), copies)

	// Zapisz wszystkie kopie
	for i := 1; i <= copies; i++ {
		var filename string
		if copies > 1 {
			filename = filepath.Join(jobDir, fmt.Sprintf("%s_copy_%d.bin", jobName, i))
		} else {
			filename = filepath.Join(jobDir, jobName+".bin")
		}

		err := os.WriteFile(filename, documentData, 0644)
		if err != nil {
			fmt.Printf("Błąd zapisu kopii %d: %v\n", i, err)
			http.Error(w, "Cannot save job file", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Zapisano kopię %d/%d: %s\n", i, copies, filename)
	}

	printerURI := getPrinterURI(r)
	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)

	addOperationAttributes(resp)

	jobURI := fmt.Sprintf("%s/jobs/%d", printerURI, id)
	resp.Job.Add(goipp.MakeAttribute("job-id", goipp.TagInteger, goipp.Integer(id)))
	resp.Job.Add(goipp.MakeAttribute("job-uri", goipp.TagURI, goipp.String(jobURI)))
	resp.Job.Add(goipp.MakeAttribute("job-name", goipp.TagName, goipp.String(fmt.Sprintf("Job %d", id))))
	resp.Job.Add(goipp.MakeAttribute("job-state", goipp.TagEnum, goipp.Integer(9)))
	resp.Job.Add(goipp.MakeAttribute("job-state-reasons", goipp.TagKeyword, goipp.String("none")))
	resp.Job.Add(goipp.MakeAttribute("copies", goipp.TagInteger, goipp.Integer(copies)))

	fmt.Println("----- IPP RESPONSE JOB-----")
	resp.Print(os.Stdout, true)
	fmt.Println("-----------------------")
	writeIPPResponse(w, resp)
}

func detectFileExtension(data []byte) string {
	if len(data) < 4 {
		return "bin"
	}

	// PDF
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "pdf"
	}

	// PNG
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// JPEG
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}

	// GIF
	if len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a") {
		return "gif"
	}

	// BMP
	if len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D {
		return "bmp"
	}

	// ZIP (Office, etc)
	if len(data) >= 4 && data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return "zip"
	}

	// PostScript
	if len(data) >= 4 && string(data[:4]) == "%!PS" {
		return "ps"
	}

	// PCL
	if len(data) >= 2 && data[0] == 0x1B && data[1] == 0x45 {
		return "pcl"
	}

	// Domyślnie
	return "bin"
}

func handlePrintJob(w http.ResponseWriter, r *http.Request, req *goipp.Message) {
	fmt.Println("IPP: Print-Job")

	jobMutex.Lock()
	jobID++
	id := jobID
	jobMutex.Unlock()

	err := os.MkdirAll(
		jobDir,
		0755,
	)

	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		http.Error(w, "Cannot create job directory", http.StatusInternalServerError)
		return
	}

	/*
	   v.T (Type) values:
	   0x21 (33) → integer
	   0x22 (34) → boolean
	   0x23 (35) → enum
	   0x41 (65) → textWithoutLanguage
	   0x42 (66) → nameWithoutLanguage
	   0x44 (68) → keyword
	   0x45 (69) → uri
	*/

	baseName := fmt.Sprintf("job_%d", id)
	copies := 1

	// ATRYBUTY OPERATION
	fmt.Println("----- ATRYBUTY OPERATION -----")
	for _, attr := range req.Operation {
		for _, v := range attr.Values {
			fmt.Printf("xAttrOperation: %s, Type: 0x%X, Value: %v\n", attr.Name, v.T, v.V)

			if attr.Name == "job-name" {
				jobName := fmt.Sprintf("%v", v.V)
				// Zamień wszystkie niedozwolone znaki na podkreślnik
				// Dozwolone: litery, cyfry, kropka, myślnik, podkreślnik
				reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
				baseName = reg.ReplaceAllString(jobName, "_")

				// Ogranicz do 255 znaków
				if len(baseName) > 255 {
					baseName = baseName[:255]
				}

				fmt.Printf("job-name (oczyszczony): %s\n", baseName)
			}
		}
	}
	fmt.Println("------------------------------")

	// ATRYBUTY JOB
	fmt.Println("----- ATRYBUTY JOB -----")
	for _, attr := range req.Job {
		for _, v := range attr.Values {
			fmt.Printf("xAttrJob: %s, Type: 0x%X, Value: %v\n", attr.Name, v.T, v.V)

			if attr.Name == "copies" {
				if val, ok := v.V.(goipp.Integer); ok {
					copies = int(val)
					fmt.Printf("Ustawiono copies: %d\n", copies)
				}
			}
		}
	}
	fmt.Println("------------------------------")

	// Odczyt dokumentu
	documentData, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Błąd odczytu dokumentu:", err)
		http.Error(w, "Cannot read document", http.StatusInternalServerError)
		return
	}

	// Wykryj rozszerzenie
	extension := detectFileExtension(documentData)
	fmt.Printf("Wykryto typ pliku: %s\n", extension)

	fmt.Printf("Odebrano dokument: %d bajtów, kopie: %d\n", len(documentData), copies)

	// Zapisz wszystkie kopie
	for i := 1; i <= copies; i++ {
		var filename string
		if copies > 1 {
			filename = filepath.Join(jobDir, fmt.Sprintf("%s_copy_%d.%s", baseName, i, extension))
		} else {
			filename = filepath.Join(jobDir, baseName+"."+extension)
		}

		err := os.WriteFile(filename, documentData, 0644)
		if err != nil {
			fmt.Printf("Błąd zapisu kopii %d: %v\n", i, err)
			http.Error(w, "Cannot save job file", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Zapisano kopię %d/%d: %s\n", i, copies, filename)
	}

	printerURI := getPrinterURI(r)
	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)

	addOperationAttributes(resp)

	jobURI := fmt.Sprintf("%s/jobs/%d", printerURI, id)
	resp.Job.Add(goipp.MakeAttribute("job-id", goipp.TagInteger, goipp.Integer(id)))
	resp.Job.Add(goipp.MakeAttribute("job-uri", goipp.TagURI, goipp.String(jobURI)))
	resp.Job.Add(goipp.MakeAttribute("job-name", goipp.TagName, goipp.String(fmt.Sprintf("Job %d", id))))
	resp.Job.Add(goipp.MakeAttribute("job-state", goipp.TagEnum, goipp.Integer(9)))
	resp.Job.Add(goipp.MakeAttribute("job-state-reasons", goipp.TagKeyword, goipp.String("none")))
	resp.Job.Add(goipp.MakeAttribute("copies", goipp.TagInteger, goipp.Integer(copies)))

	fmt.Println("----- IPP RESPONSE JOB-----")
	resp.Print(os.Stdout, true)
	fmt.Println("-----------------------")
	writeIPPResponse(w, resp)
}

func handleValidateJob(w http.ResponseWriter, req *goipp.Message) {
	fmt.Println("IPP: Validate-Job")

	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)
	addOperationAttributes(resp)
	writeIPPResponse(w, resp)
}

func handleGetJobs(w http.ResponseWriter, req *goipp.Message) {
	fmt.Println("IPP: Get-Jobs")

	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)
	addOperationAttributes(resp)
	writeIPPResponse(w, resp)
}

func handleGetJobAttributes(w http.ResponseWriter, req *goipp.Message) {
	fmt.Println("IPP: Get-Job-Attributes")

	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)

	addOperationAttributes(resp)

	resp.Job.Add(goipp.MakeAttribute("job-state", goipp.TagEnum, goipp.Integer(9)))
	resp.Job.Add(goipp.MakeAttribute("job-state-reasons", goipp.TagKeyword, goipp.String("none")))

	writeIPPResponse(w, resp)
}

// func handleCancelJob(
// 	w http.ResponseWriter,
// 	req *goipp.Message,
// ) {
// 	fmt.Println("IPP: Cancel-Job")

// 	resp := goipp.NewResponse(
// 		req.Version,
// 		goipp.StatusOk,
// 		req.RequestID,
// 	)

// 	addOperationAttributes(resp)

// 	writeIPPResponse(w, resp)
// }

func handleCancelJob(w http.ResponseWriter, req *goipp.Message) {
	fmt.Println("IPP: Cancel-Job")

	var jobUUID string

	for _, group := range req.Groups {
		for _, attr := range group.Attrs {
			if attr.Name == "job-uuid" && len(attr.Values) > 0 {
				jobUUID = fmt.Sprint(attr.Values[0].V)
			}
		}
	}

	fmt.Println("Cancel job UUID:", jobUUID)

	resp := goipp.NewResponse(req.Version, goipp.StatusOk, req.RequestID)

	addOperationAttributes(resp)

	resp.Groups.Add(goipp.Group{
		Tag: goipp.TagJobGroup,
		Attrs: goipp.Attributes{
			goipp.MakeAttribute(
				"job-uuid",
				goipp.TagURI,
				goipp.String(jobUUID),
			),
			goipp.MakeAttribute(
				"job-state",
				goipp.TagEnum,
				goipp.Integer(7),
			),
			goipp.MakeAttribute(
				"job-state-reasons",
				goipp.TagKeyword,
				goipp.String("none"),
			),
		},
	})

	writeIPPResponse(w, resp)
}
func addOperationAttributes(
	resp *goipp.Message,
) {
	resp.Operation.Add(
		goipp.MakeAttribute(
			"attributes-charset",
			goipp.TagCharset,
			goipp.String("utf-8"),
		),
	)

	resp.Operation.Add(
		goipp.MakeAttribute(
			"attributes-natural-language",
			goipp.TagLanguage,
			goipp.String("en"),
		),
	)
}

func writeIPPResponse(w http.ResponseWriter, resp *goipp.Message) {
	data, err := resp.EncodeBytes()

	if err != nil {
		fmt.Println("Błąd kodowania odpowiedzi IPP:", err)
		http.Error(w, "IPP encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/ipp")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(data)

	if err != nil {
		fmt.Println("Błąd wysyłania odpowiedzi IPP:", err)
		return
	}
	fmt.Println("IPP: wysłano odpowiedź:", len(data), "bajtów")
}

func main() {
	hostname, err := os.Hostname()

	if err != nil {
		fmt.Println(
			"Nie można pobrać hostname:",
			err,
		)

		return
	}

	fmt.Println("========================================")
	fmt.Println("PiKul Virtual Printer")
	fmt.Println("Hostname:", hostname)
	fmt.Println("Printer:", printerName)
	fmt.Println("========================================")

	/*
		9100 RAW
	*/

	listener, err := net.Listen(
		"tcp4",
		fmt.Sprintf(
			"0.0.0.0:%d",
			rawPort,
		),
	)

	if err != nil {
		fmt.Println(
			"Nie można uruchomić 9100:",
			err,
		)

		return
	}

	defer listener.Close()

	fmt.Printf(
		"Emulator RAW nasłuchuje na 0.0.0.0:%d...\n",
		rawPort,
	)

	go func() {
		for {
			conn, err := listener.Accept()

			if err != nil {
				fmt.Println(
					"Błąd 9100:",
					err,
				)

				continue
			}

			go handleConnection(conn)
		}
	}()

	/*
		631 IPP
	*/

	mux := http.NewServeMux()
	mux.HandleFunc(ippPath, handleIPP)
	mux.HandleFunc("/", handleIPP)

	server := &http.Server{
		Addr: fmt.Sprintf(
			"0.0.0.0:%d",
			ippPort,
		),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("Emulator IPP nasłuchuje na 0.0.0.0:%d...\n", ippPort)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Nie można uruchomić 631:", err)
	}
}
