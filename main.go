package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func handleUpload(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		_, _ = fmt.Fprintf(w, "err = %v\n", err)
		return
	}

	respContent := make(map[string]string)

	respContent["form"] = fmt.Sprintf("%v", r.PostForm)
	f, fh, err := r.FormFile("file")
	if err != nil {
		_, _ = fmt.Fprintf(w, "err = %v\n", err)
		return
	}
	respContent["filename"] = fh.Filename
	fall, err := ioutil.ReadAll(f)
	if err != nil {
		_, _ = fmt.Fprintf(w, "err = %v\n", err)
		return
	}
	sumMD5 := md5.Sum(fall)
	sumSHA256 := sha256.Sum256(fall)

	respContent["filesize"] = fmt.Sprintf("%v", len(fall))
	respContent["md5"] = hex.EncodeToString(sumMD5[:])
	respContent["sha256"] = hex.EncodeToString(sumSHA256[:])

	// 这句话必须写在下面的前面，不然不生效
	w.Header().Set("Content-Type", "application/json")

	// w.cw.header  w.handlerHeader 是个坑
	_ = json.NewEncoder(w).Encode(respContent)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8100", "The port of http file server")
	flag.Parse()
	log.SetFlags(log.LstdFlags)

	addr := fmt.Sprintf("0.0.0.0:%v", port)

	log.Printf("Serving HTTP on %v", addr)

	h := http.FileServer(http.Dir("."))
	m := http.NewServeMux()
	m.HandleFunc("/upload", handleUpload)
	m.Handle("/", h)
	hLog := handlers.CombinedLoggingHandler(os.Stdout, m)

	log.Fatal(http.ListenAndServe(addr, hLog))
}
