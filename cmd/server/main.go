package main

// 这里做成一份标准的如何建立 http server 的范例代码

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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

func listenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	serv := &http.Server{Addr: addr, Handler: handler}
	// serv.ServeTLS() will build for tls
	return serv.Serve(ln)
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8100", "The addr of http file server")
	flag.Parse()
	log.SetFlags(log.LstdFlags)
	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("Serving HTTP on %v", addr)

	h := http.FileServer(http.Dir("."))
	m := http.NewServeMux()
	m.HandleFunc("/upload", handleUpload)
	m.Handle("/", h)
	hLog := handlers.CombinedLoggingHandler(os.Stdout, m)

	_ = cancel
	log.Fatal(listenAndServe(ctx, addr, hLog))

}
