package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/OpenPrinting/goipp"
)

const (
	ippPort = 631
	rawPort = 9100

	printerName = "PiKul_VirtualPrinter"
	ippPath     = "/ipp/print"
	jobDir      = "jobs"
)

var (
	jobMutex sync.Mutex
	jobID    uint32
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Połączenie 9100:", conn.RemoteAddr())

	err := os.MkdirAll("jobs", 0755)
	if err != nil {
		fmt.Println("Błąd tworzenia katalogu jobs:", err)
		return
	}

	filename := fmt.Sprintf(
		"jobs/print_%s.bin",
		time.Now().Format("20060102_150405.000"),
	)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Błąd pliku:", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)

		if n > 0 {
			fmt.Printf("9100: odebrano %d bajtów\n", n)

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

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-uri-supported",
			goipp.TagURI,
			goipp.String(printerURI),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-name",
			goipp.TagName,
			goipp.String(printerName),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-info",
			goipp.TagText,
			goipp.String("PiKul Virtual Printer"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-location",
			goipp.TagText,
			goipp.String(""),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-make-and-model",
			goipp.TagText,
			goipp.String("PiKul Virtual Printer"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-state",
			goipp.TagEnum,
			goipp.Integer(3),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"printer-state-reasons",
			goipp.TagKeyword,
			goipp.String("none"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"document-format-default",
			goipp.TagMimeType,
			goipp.String("application/octet-stream"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"charset-configured",
			goipp.TagCharset,
			goipp.String("utf-8"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"charset-supported",
			goipp.TagCharset,
			goipp.String("utf-8"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"natural-language-configured",
			goipp.TagLanguage,
			goipp.String("en"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"generated-natural-language-supported",
			goipp.TagLanguage,
			goipp.String("en"),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"color-supported",
			goipp.TagBoolean,
			goipp.Boolean(false),
		),
	)

	resp.Printer.Add(
		goipp.MakeAttribute(
			"copies-supported",
			goipp.TagRange,
			goipp.Range{
				Lower: 1,
				Upper: 999,
			},
		),
	)

	writeIPPResponse(w, resp)
}

func handlePrintJob(
	w http.ResponseWriter,
	r *http.Request,
	req *goipp.Message,
) {
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

		http.Error(
			w,
			"Cannot create job directory",
			http.StatusInternalServerError,
		)

		return
	}

	filename := filepath.Join(
		jobDir,
		fmt.Sprintf(
			"job_%d.bin",
			id,
		),
	)

	file, err := os.Create(filename)

	if err != nil {
		fmt.Println("Błąd tworzenia pliku:", err)

		http.Error(
			w,
			"Cannot create job file",
			http.StatusInternalServerError,
		)

		return
	}

	defer file.Close()

	/*
		Decode() przeczytał komunikat IPP.

		Reszta r.Body to dokument.
	*/

	bytesWritten, err := io.Copy(
		file,
		r.Body,
	)

	if err != nil {
		fmt.Println("Błąd zapisu dokumentu:", err)

		http.Error(
			w,
			"Cannot save document",
			http.StatusInternalServerError,
		)

		return
	}

	fmt.Printf(
		"JOB %d: odebrano %d bajtów\n",
		id,
		bytesWritten,
	)

	fmt.Println("Plik:", filename)

	printerURI := getPrinterURI(r)

	jobURI := fmt.Sprintf(
		"%s/jobs/%d",
		printerURI,
		id,
	)

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-id",
			goipp.TagInteger,
			goipp.Integer(id),
		),
	)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-uri",
			goipp.TagURI,
			goipp.String(jobURI),
		),
	)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-name",
			goipp.TagName,
			goipp.String(
				fmt.Sprintf(
					"Job %d",
					id,
				),
			),
		),
	)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-state",
			goipp.TagEnum,
			goipp.Integer(9),
		),
	)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-state-reasons",
			goipp.TagKeyword,
			goipp.String("none"),
		),
	)

	writeIPPResponse(w, resp)
}

func handleValidateJob(
	w http.ResponseWriter,
	req *goipp.Message,
) {
	fmt.Println("IPP: Validate-Job")

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

	writeIPPResponse(w, resp)
}

func handleGetJobs(
	w http.ResponseWriter,
	req *goipp.Message,
) {
	fmt.Println("IPP: Get-Jobs")

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

	writeIPPResponse(w, resp)
}

func handleGetJobAttributes(
	w http.ResponseWriter,
	req *goipp.Message,
) {
	fmt.Println("IPP: Get-Job-Attributes")

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-state",
			goipp.TagEnum,
			goipp.Integer(9),
		),
	)

	resp.Job.Add(
		goipp.MakeAttribute(
			"job-state-reasons",
			goipp.TagKeyword,
			goipp.String("none"),
		),
	)

	writeIPPResponse(w, resp)
}

func handleCancelJob(
	w http.ResponseWriter,
	req *goipp.Message,
) {
	fmt.Println("IPP: Cancel-Job")

	resp := goipp.NewResponse(
		req.Version,
		goipp.StatusOk,
		req.RequestID,
	)

	addOperationAttributes(resp)

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

func writeIPPResponse(
	w http.ResponseWriter,
	resp *goipp.Message,
) {
	data, err := resp.EncodeBytes()

	if err != nil {
		fmt.Println(
			"Błąd kodowania odpowiedzi IPP:",
			err,
		)

		http.Error(
			w,
			"IPP encode error",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/ipp",
	)

	w.WriteHeader(
		http.StatusOK,
	)

	_, err = w.Write(data)

	if err != nil {
		fmt.Println(
			"Błąd wysyłania odpowiedzi IPP:",
			err,
		)

		return
	}

	fmt.Println(
		"IPP: wysłano odpowiedź:",
		len(data),
		"bajtów",
	)
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

	mux.HandleFunc(
		ippPath,
		handleIPP,
	)

	mux.HandleFunc(
		"/",
		handleIPP,
	)

	server := &http.Server{
		Addr: fmt.Sprintf(
			"0.0.0.0:%d",
			ippPort,
		),

		Handler: mux,

		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf(
		"Emulator IPP nasłuchuje na 0.0.0.0:%d...\n",
		ippPort,
	)

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println(
			"Nie można uruchomić 631:",
			err,
		)
	}
}