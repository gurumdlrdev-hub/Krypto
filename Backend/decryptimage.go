package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"net/http"
	"strconv"
)

func decryptImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("decrypt_image")
	if err != nil {
		http.Error(w, "Error getting the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ciphertext, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file: "+err.Error(), http.StatusInternalServerError)
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

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		http.Error(w, "Invalid encrypted file", http.StatusBadRequest)
		return
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		http.Error(w, "Decryption failed (Invalid Key or File corrupted): "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"decrypted_image.png\"")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(plaintext)))

	if _, err := w.Write(plaintext); err != nil {
		http.Error(w, "Unable to stream decrypted image: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
