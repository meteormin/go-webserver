package main

import (
	"github.com/meteormin/go-webserver/internal/handler"
	"log"
	"net/http"
)

func main() {
	// 기본 라우트: 파일 서빙
	http.HandleFunc("/wget", handler.Wget(handler.StaticDir))
	http.HandleFunc("/", handler.Static())

	// 서버 시작
	port := ":8080"
	log.Printf("Starting server on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
