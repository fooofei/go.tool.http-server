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
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"golang.org/x/exp/slog"
)

func handleUpload(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		fmt.Fprintf(w, "err = %v\n", err)
		return
	}

	var respContent = make(map[string]string)

	respContent["form"] = fmt.Sprintf("%v", r.PostForm)
	var file multipart.File
	var fileHeader *multipart.FileHeader
	if file, fileHeader, err = r.FormFile("file"); err != nil {
		fmt.Fprintf(w, "err = %v\n", err)
		return
	}
	respContent["filename"] = fileHeader.Filename
	var fileContent []byte
	if fileContent, err = io.ReadAll(file); err != nil {
		fmt.Fprintf(w, "err = %v\n", err)
		return
	}
	sumMD5 := md5.Sum(fileContent)
	sumSHA256 := sha256.Sum256(fileContent)

	respContent["filesize"] = fmt.Sprintf("%v", len(fileContent))
	respContent["md5"] = hex.EncodeToString(sumMD5[:])
	respContent["sha256"] = hex.EncodeToString(sumSHA256[:])

	// 这句话必须写在下面的前面，不然不生效
	w.Header().Set("Content-Type", "application/json")

	// w.cw.header  w.handlerHeader 是个坑
	json.NewEncoder(w).Encode(respContent)
}

func serve(ctx context.Context, addr string, handler http.Handler) error {
	var lc = net.ListenConfig{}
	var ln, err = lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	var serv = &http.Server{Addr: addr, Handler: handler}
	//baseDir := "" // "D:/0data/src/sim_http_server"
	//return serv.ServeTLS(ln, filepath.Join(baseDir, "certs/server.crt"),
	//	filepath.Join(baseDir, "certs/server.key")) // will build for tls
	var errCh = make(chan error, 2)
	go func() {
		var err2 = serv.Serve(ln)
		errCh <- err2
	}()
	select {
	case <-ctx.Done():
		var shutdownCtx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		serv.Shutdown(shutdownCtx)
		return nil
	case err = <-errCh:
		return err
	}
}

type countHandler struct {
	Count  atomic.Int64
	logger slog.Logger
}

func (h *countHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "<html>"+
		"<head>"+
		"<title>ZeroHTTPd: Unimplemented</title>"+
		"</head>"+
		"<body>"+
		"<h1>Bad Request (Unimplemented) %v </h1>"+
		"<p>Your client sent a request ZeroHTTPd did not understand and it is probably not your fault.</p>"+
		"</body>"+
		"</html>", h.Count)
	h.logger.Info("show count", "count", h.Count)
	h.Count.Add(1)
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8100", "The addr of http file server")
	flag.Parse()
	var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	logger = logger.With("pid", os.Getpid())
	var ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger.Info("Serving HTTP", "addr", addr)

	var h = http.FileServer(http.Dir("."))
	var m = http.NewServeMux()
	m.HandleFunc("/upload", handleUpload)
	m.Handle("/", h)
	var handlerWithLog = handlers.CombinedLoggingHandler(os.Stdout, m)

	if err := serve(ctx, addr, handlerWithLog); err != nil {
		panic(err)
	}
}
