package main

import (
	"context"
	//	"database/sql"
	//	"encoding/json"
	"flag"
	//	"fmt"
	//	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	//	u "github.com/pprisn/test_masops/server/utils"
	//	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "templates/index.html")
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.RemoteAddr, "\t", r.Method, "\t", r.URL)
		h.ServeHTTP(w, r)
	})
}

func Report1(w http.ResponseWriter, r *http.Request) {
	url := "http://upload.wikimedia.org/wikipedia/en/b/bc/Wiki.png"
	timeout := time.Duration(5) * time.Second
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	//copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
	w.Header().Set("Content-Disposition", "attachment; filename=Wiki.png")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))
	//stream the body to the client without fully loading it into memory
	io.Copy(w, resp.Body)
}

func main() {
	var dir string

	flag.StringVar(&dir, "dir", "./static/", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	router.Use(Middleware)

	srv := &http.Server{
		Handler: router,
		Addr:    "localhost:3000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//log.Fatal(srv.ListenAndServe())

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
