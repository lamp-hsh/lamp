package comic

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"strings"
	"encoding/json"

	"example.com/m/auth"
)

type MoveFile struct {
	Name	string
}

func ComicResource(w http.ResponseWriter, r *http.Request) {
	if auth.VerifyPermission(r) {
		switch r.Method {
		case "GET":
			fileName := r.URL.Path
			if fileName == "" || fileName == "/" {
				http.Error(w, "File name not provided", http.StatusBadRequest)
				return
			}
			switch {
			case strings.Contains(fileName, ".webp"):
				ComicImage(w, r, fileName)
			default:
				http.Error(w, "File name not provided", http.StatusBadRequest)
			}
		case "PUT":
			storageToComicLog := StorageToComic(w, r)
			fmt.Println(storageToComicLog)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

func ComicImage(w http.ResponseWriter, r *http.Request, fileName string) {
	_, fileErr := os.Stat(fileName)
	if os.IsNotExist(fileErr) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, fileName)
}

func StorageToComic(w http.ResponseWriter, r *http.Request) string {
	fromPath := "/storage/"
	toPath := "/media/comic/"
	res := Response{1}
	var moveFile MoveFile
	json.NewDecoder(r.Body).Decode(&moveFile)

	_, fileErr := os.Stat(fromPath + moveFile.Name)
	if os.IsNotExist(fileErr) {
		json.NewEncoder(w).Encode(res)
		return "Target file not found"
	}

	srcFile, openErr := os.Open(fromPath + moveFile.Name)
    if openErr != nil {
        json.NewEncoder(w).Encode(res)
        return "Target file open failure"
    }
    defer srcFile.Close()

	dstFile, createErr := os.Create(toPath + moveFile.Name)
    if createErr != nil {
        json.NewEncoder(w).Encode(res)
        return "Destination file create failure"
    }
    defer dstFile.Close() 

	_, mvErr := io.Copy(dstFile, srcFile)
	if mvErr != nil {
		json.NewEncoder(w).Encode(res)
		return "Target file move failure"
	}

	rmErr := os.Remove(fromPath + moveFile.Name)
	if rmErr != nil {
		json.NewEncoder(w).Encode(res)
		return "Target file move failure"
	}

	res.Status = 0
	json.NewEncoder(w).Encode(res)
	return "Moved " + moveFile.Name + " into media/comic successfully."
}