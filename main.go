// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"cloud.google.com/go/translate"
	"golang.org/x/text/language"

	vision "cloud.google.com/go/vision/apiv1"
)

// Response definiation for response API
type Response struct {
	ORIGINAL string `json:"original_message"`
	ENGLISH string `json:"english_message"`
}

const maxMemory = 2 * 1024 * 1024 // 2 megabytes.

// Function to translater 
func translateText(targetLanguage, text string) (string, error) {
	// text := "The Go Gopher is cute"
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
			return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
			return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
			return "", fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
			return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
}

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

			message := "No text found."
			message_translate := "Untranslated text"

			if len(annotations) != 0 {
				message = annotations[0].Description
			}

			message_translate, _ = translateText("en", message)

			res := Response{}
			res.ORIGINAL = message
			res.ENGLISH = message_translate

			resJSON, err := json.Marshal(res)
			if err != nil {
				log.Fatalf("Failed to parse json: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resJSON)
		}
	}
}

func main() {
	http.HandleFunc("/upload", getText)
	http.ListenAndServe(":8080", nil)
}
