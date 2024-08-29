package handler

import (
	"github.com/meteormin/go-webserver/internal/templates"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var StaticDir = "web"

func Static() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		path := r.URL.Path
		if path != "/" && strings.HasSuffix(path, "/") {
			log.Printf("Check has index.html in %s", filepath.Join(path, "index.html"))
			if _, err := os.Stat(filepath.Join(StaticDir, path, "index.html")); err == nil {
				log.Printf("has index.html in %s", path)
				path = filepath.Join(path, "index.html")
			}
		}
		fullPath := filepath.Join(StaticDir, path)
		if isDir(fullPath) {
			log.Printf("%s is dir", fullPath)
			serveDir(w, r, fullPath, r.URL.Path)
		} else {
			log.Printf("%s is file", fullPath)
			serveFile(w, r, fullPath)
		}
	}
}

// 파일이 디렉토리인지 확인
func isDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// 디렉토리 내용을 HTML로 제공
func serveDir(w http.ResponseWriter, r *http.Request, path string, urlPath string) {
	files, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, "Unable to read directory", http.StatusInternalServerError)
		return
	}

	var items []map[string]interface{}
	for _, file := range files {
		name := file.Name()
		url := filepath.Join(urlPath, name)
		item := map[string]interface{}{
			"Class": ifElse(file.IsDir(), "directory", "file"),
			"URL":   url,
			"Name":  name,
		}
		items = append(items, item)
		log.Printf("item: {class: %s, url: %s, name: %s}\n", item["Class"], item["URL"], item["Name"])
	}

	// 현재 경로와 호스트 처리
	currentPath := urlPath
	if currentPath == "" {
		currentPath = "~"
	} else if strings.HasSuffix(currentPath, "/") {
		currentPath = currentPath[:len(currentPath)-1]
	}

	host := r.Host
	tmpl, err := template.ParseFS(templates.GetFS(), "list.tmpl")
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Path":  currentPath,
		"Items": items,
		"Host":  host,
	}

	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render directory listing: "+err.Error(), http.StatusInternalServerError)
	}
}

// 파일 내용을 raw string으로 제공
func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(content)
	if err != nil {
		return
	}
}

// ifElse is a helper function for templates
func ifElse(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}
