// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
)

const maxMemory = 2 * 1024 * 1024 // 2 megabytes.

// ReadDoc is an HTTP Cloud Function with a request parameter.
func uploadImage(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		log.Printf("Error parsing form: %v", err)
		return
	}

	defer func() {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			http.Error(w, "Error cleaning up form files", http.StatusInternalServerError)
			log.Printf("Error cleaning up form files: %v", err)
		}
	}()

	for _, headers := range r.MultipartForm.File {
		for _, h := range headers {
			file, _ := h.Open()
			ctx := context.Background()
			scClient, err := storage.NewClient(ctx)
			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}

			t := time.Now()

			bucket := "alvardevlp07.appspot.com"
			fileName := t.Format("20060102150405") + "-" + h.Filename

			wc := scClient.Bucket(bucket).Object(fileName).NewWriter(ctx)
			if _, err = io.Copy(wc, file); err != nil {
				fmt.Println("Error copying")
			}
			if err := wc.Close(); err != nil {
				fmt.Println("Conexion closed")
				fmt.Println(err)
			}

			client, err := vision.NewImageAnnotatorClient(ctx)

			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			image := vision.NewImageFromURI("gs://alvardevlp07.appspot.com/" + fileName)
			annotations, err := client.DetectTexts(ctx, image, nil, 10)

			if err != nil {
				log.Fatalf("Failed to detect text: %v", err)
			}

			if len(annotations) == 0 {
				fmt.Fprintf(w, "No text found.")
			} else {
				fmt.Fprintf(w, annotations[0].Description)
			}

		}
	}
}

func main() {
	http.HandleFunc("/upload", uploadImage)
	http.ListenAndServe(":8080", nil)
}
