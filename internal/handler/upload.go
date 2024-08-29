package handler

import (
	"github.com/meteormin/go-webserver/internal/templates"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var UploadDir = "uploads"

func Upload(baseDir string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		uploadHandler(writer, request, baseDir)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request, baseDir string) {
	if r.Method == http.MethodGet {
		t, err := template.ParseFS(templates.GetFS(), "upload.tmpl")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, map[string]interface{}{
			"UploadDir": UploadDir,
			"Host":      r.Host,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		// Handle file upload
		dir := r.FormValue("dir")
		if dir == "" {
			dir = "."
		}
		dir = filepath.Join(baseDir, UploadDir, dir)

		// Ensure directory exists
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a file to save the uploaded file
		filePath := filepath.Join(dir, header.Filename)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the uploaded file to the destination file
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/upload", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
