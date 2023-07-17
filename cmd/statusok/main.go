package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"golang.org/x/exp/slog"
	"net"
	"net/http"
	"os"
)

// DecodeBytes will decode hex string in src to bytes
func DecodeBytes(src []byte) ([]byte, error) {
	l := hex.DecodedLen(len(src))
	dst := make([]byte, l)
	n, err := hex.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	// go make sure n < len(dst)
	return dst[:n], nil
}

// 打桩使用，全部回复 200

func serveAny(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Inner-Code", "0")
	w.Header().Set("X-Executed-Duration", "0ms")
	w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Content-Length", "87211")
	// w.Header().Set("Connection", "keep-alive")

	_, _ = fmt.Fprintf(w, `
{"body":""}
`)
}

func listenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	// 不自定义 tcpNoDelay
	//if tcpLn, ok := ln.(*net.TCPListener); ok {
	//	ln = tcpNoDelayListener{
	//		TCPListener: tcpLn,
	//		NoDelay:     false,
	//	}
	//}
	serv := &http.Server{Addr: addr, Handler: handler}
	return serv.Serve(ln)
}

// 设置为 tcpNoDelay
type tcpNoDelayListener struct {
	*net.TCPListener
	NoDelay bool
}

func (ln tcpNoDelayListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	_ = tc.SetNoDelay(ln.NoDelay)
	return tc, nil
}

func main() {

	var addr string
	flag.StringVar(&addr, "addr", ":8100", "The addr of http file server")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	logger = logger.With("pid", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Serving HTTP", "addr", addr)

	h := http.NewServeMux()
	h.HandleFunc("/", serveAny)
	_ = cancel
	err := listenAndServe(ctx, addr, h)
	if err != nil {
		panic(err)
	}
}
