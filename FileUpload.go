package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// "time"
)

const maxUploadSize = 1024 * 1024 // 1 gb, easily modified if need be
const uploadPath = "./tmp"

func main() {
	http.HandleFunc("/upload", uploadFileHandler())

	// crutime := time.Now().Unix() // timeStamp, needs test 

	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files", fs))

	// Going to the index page will display the contents of the root directory
	http.Handle("/", http.FileServer(http.Dir(".")))


	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":8080", nil)) // not sure of the best place to keep files


}
/*
func postFile(filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWrite := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWrite.CreateFormFile("uploadFile" , filename)
	if err != nil{
		renderError(w, "ERROR_WRITING_TO_BUFFER", http.StatusBadRequest)
	}
	fileHandler, err := os.Open(filename)
	if err != nil{
		renderError(w, "ERROR_OPENING_FILE", http.StatusBadRequest)
	}
}
*/
func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_LARGE", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		/*fileType := r.PostFormValue("type")
		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return

		} */
		var Buf bytes.Buffer
		fileType := r.PostFormValue("type")
		file, header, err := r.FormFile("uploadFile")
		if err != nil{
			renderError(w, "ERROR_UPLOADING_FILE", http.StatusBadRequest)

		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		name := strings.Split(header.Filename, ".")
		fmt.Printf("File name %s\n", name[0])
		// Copy file data into buffer
		io.Copy(&Buf, file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)
		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}
		fileName := randToken(12)
		fileEndings, err := mime.ExtensionsByType(fileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}
