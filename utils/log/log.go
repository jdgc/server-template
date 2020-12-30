package log

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetPrefix("[LOG] ")
	log.SetOutput(os.Stdout)
}

func Print(s string) {
	log.Println(s)
}

func PrintRequest(r *http.Request) {
	output := fmt.Sprintf("[%s: %s] %s", r.Method, r.URL.String(), r.RemoteAddr)
	log.Println(output)
}
