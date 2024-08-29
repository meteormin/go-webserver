package handler

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var WgetDir = "downloads"

func Wget(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
			post(w, r, basePath)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

func post(w http.ResponseWriter, r *http.Request, basePath string) {
	var jsonMap map[string]interface{}
	jsonData := r.Body
	err := json.NewDecoder(jsonData).Decode(&jsonMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	queryUrl := jsonMap["url"].(string)
	parsedUrl, err := url.Parse(queryUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	downloadPath := filepath.Join(basePath, WgetDir, parsedUrl.Host, parsedUrl.Path)
	err = downloadPage(queryUrl, downloadPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, strings.Replace(downloadPath, "web", "", 1), http.StatusFound)
}

// downloadFile downloads a file from the specified URL and saves it to the given filepath.
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// downloadPage downloads the HTML page and its resources, saving them in the specified directory.
func downloadPage(pageUrl string, baseDir string) error {
	resp, err := http.Get(pageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return saveContentBasedOnMimeType(pageUrl, baseDir)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	// Parse the URL to handle relative paths correctly
	parsedUrl, err := url.Parse(pageUrl)
	if err != nil {
		return err
	}

	var downloadResource func(*html.Node)
	downloadResource = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var resourceUrl string
			var resourceAttr string

			// Identify the resource type and attribute to download
			if n.Data == "img" {
				for i := range n.Attr {
					if n.Attr[i].Key == "src" {
						resourceAttr = "src"
						resourceUrl = n.Attr[i].Val
						break
					}
				}
			} else if n.Data == "link" {
				for i := range n.Attr {
					if n.Attr[i].Key == "rel" && n.Attr[i].Val == "stylesheet" {
						resourceAttr = "href"
						resourceUrl = n.Attr[i].Val
						break
					}
				}
			} else if n.Data == "script" {
				for i := range n.Attr {
					if n.Attr[i].Key == "src" {
						resourceAttr = "src"
						resourceUrl = n.Attr[i].Val
						break
					}
				}
			}

			// If a resource was found, download it
			if resourceUrl != "" {
				// Resolve the relative URL against the base URL
				resourceUrlParsed, err := url.Parse(resourceUrl)
				if err != nil {
					fmt.Println("Skipping resource:", resourceUrl, "Error:", err)
					return
				}

				// Create the full URL for the resource
				fullUrl := parsedUrl.ResolveReference(resourceUrlParsed)
				relPath := filepath.Join(baseDir, resourceUrlParsed.Path)

				// Ensure the directory exists
				err = os.MkdirAll(filepath.Dir(relPath), os.ModePerm)
				if err != nil {
					return
				}

				fmt.Println("Downloading resource:", fullUrl.String(), "to", relPath)
				if err := downloadFile(relPath, fullUrl.String()); err != nil {
					fmt.Println("Error downloading resource:", fullUrl.String(), "Error:", err)
					return
				}

				// Update the HTML node to point to the local resource path
				for i := range n.Attr {
					if n.Attr[i].Key == resourceAttr {
						n.Attr[i].Val = resourceUrlParsed.Path
					}
				}
			}
		}

		// Recursively process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			downloadResource(c)
		}
	}

	// Create the base directory
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return err
	}

	// Download resources and update HTML
	downloadResource(doc)

	// Save the updated HTML to a file
	htmlFile := filepath.Join(baseDir, "index.html")
	out, err := os.Create(htmlFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := html.Render(out, doc); err != nil {
		return err
	}

	return nil
}

// saveContentBasedOnMimeType downloads content and saves it to a file based on its MIME type.
func saveContentBasedOnMimeType(contentUrl string, baseDir string) error {
	resp, err := http.Get(contentUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the URL
	parsedUrl, err := url.Parse(contentUrl)
	if err != nil {
		return err
	}

	// Determine the MIME type and file extension
	contentType := resp.Header.Get("Content-Type")
	var fileExtension string
	switch contentType {
	case "image/jpeg":
		fileExtension = ".jpg"
	case "image/png":
		fileExtension = ".png"
	case "image/gif":
		fileExtension = ".gif"
	case "text/html":
		fileExtension = ".html"
	default:
		// Extract extension from URL if content type is not explicitly handled
		fileExtension = filepath.Ext(parsedUrl.Path)
		if fileExtension == "" {
			fileExtension = ".bin" // Default to .bin for unknown types
		}
	}

	// Create the file path
	filePath := filepath.Join(baseDir, parsedUrl.Host, parsedUrl.Path)

	// Ensure the directory exists
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	// Save the file
	fmt.Println("Saving file:", filePath)
	if err := downloadFile(filePath, contentUrl); err != nil {
		return err
	}

	return nil
}
