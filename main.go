package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/handlers"
    "os"
)

func main() {
    var port string
    flag.StringVar(&port, "port", "8100", "The port of http file server")
    flag.Parse()
    log.SetFlags(log.LstdFlags)

    addr := fmt.Sprintf("0.0.0.0:%v", port)

    log.Printf("Serving HTTP on %v", addr)
    h := http.FileServer(http.Dir("."))
    hlog := handlers.LoggingHandler(os.Stdout, h)
    log.Fatal(http.ListenAndServe(addr, hlog))
}