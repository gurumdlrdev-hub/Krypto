package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"net/http"
)

func decryptAudio(w http.ResponseWriter, r *http.Request) {

	file, _, err := r.FormFile("decrypt_audio")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ciphertext, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Could not read the file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := aes.NewCipher(audioSecretKey)
	if err != nil {
		http.Error(w, "Cipher creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		http.Error(w, "GCM creation error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		http.Error(w, "Invalid file format (file is too small)", http.StatusBadRequest)
		return
	}

	nonce := ciphertext[:nonceSize]
	actualCiphertext := ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		http.Error(w, "Unable to decrypt (incorrect key or data corrupted): "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"restored_audio.mp3\"")
	w.Header().Set("Content-Type", "audio/mpeg")

	w.Write(plaintext)
}
