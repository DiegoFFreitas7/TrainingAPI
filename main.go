// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	vision "cloud.google.com/go/vision/apiv1"
)

const maxMemory = 2 * 1024 * 1024 // 2 megabytes.

func getText(w http.ResponseWriter, r *http.Request) {

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

			ctx := context.Background()
			client, err := vision.NewImageAnnotatorClient(ctx)

			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			file, _ := h.Open()

			image, err := vision.NewImageFromReader(file)
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
	http.HandleFunc("/upload", getText)
	http.ListenAndServe(":8080", nil)
}
