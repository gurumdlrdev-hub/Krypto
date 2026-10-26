package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net/http"
	"strconv"
)

func encryptImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("encrypt_image")
	if err != nil {
		http.Error(w, "Error getting the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	plaintext, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read the file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := aes.NewCipher(imageSecretKey)
	if err != nil {
		http.Error(w, "Cipher creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		http.Error(w, "GCM creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		http.Error(w, "Nonce generation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	w.Header().Set("Content-Disposition", "attachment; filename=\"encrypted_image.enc\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(ciphertext)))

	if _, err := w.Write(ciphertext); err != nil {
		http.Error(w, "Unable to send encrypted file: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
