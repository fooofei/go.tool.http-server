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
	stdlog "log"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/fooofei/stdr"
	"github.com/go-logr/logr"
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
	//baseDir := "" // "D:/0data/src/sim_http_server"
	//return serv.ServeTLS(ln, filepath.Join(baseDir, "certs/server.crt"),
	//	filepath.Join(baseDir, "certs/server.key")) // will build for tls
	return serv.Serve(ln)
}

type countHandler struct {
	Count  int64
	logger logr.Logger
}

func (h *countHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprintf(w, "<html>"+
		"<head>"+
		"<title>ZeroHTTPd: Unimplemented</title>"+
		"</head>"+
		"<body>"+
		"<h1>Bad Request (Unimplemented) %v </h1>"+
		"<p>Your client sent a request ZeroHTTPd did not understand and it is probably not your fault.</p>"+
		"</body>"+
		"</html>", h.Count)
	h.logger.Info("show count", "count", h.Count)
	atomic.AddInt64(&h.Count, 1)
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8100", "The addr of http file server")
	flag.Parse()
	logger := stdr.New(stdlog.New(os.Stdout, "", stdlog.Lshortfile|stdlog.LstdFlags))
	logger = logger.WithValues("pid", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Serving HTTP", "addr", addr)

	h := http.FileServer(http.Dir("."))
	m := http.NewServeMux()
	m.HandleFunc("/upload", handleUpload)
	m.Handle("/", h)
	hLog := handlers.CombinedLoggingHandler(os.Stdout, m)

	_ = cancel
	err := listenAndServe(ctx, addr, hLog)
	if err != nil {
		panic(err)
	}
}
