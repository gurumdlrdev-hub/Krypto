package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net/http"
)

func encryptAudio(w http.ResponseWriter, r *http.Request) {

	file, _, err := r.FormFile("encrypt_audio")
	if err != nil {
		http.Error(w, "Error retrieving the audio file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read the audio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := aes.NewCipher(audioSecretKey)
	if err != nil {
		http.Error(w, "Cipher creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		http.Error(w, "GCM creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		http.Error(w, "Nonce creation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ciphertext := aesgcm.Seal(nonce, nonce, audioData, nil)

	w.Header().Set("Content-Disposition", "attachment; filename=\"encrypted_audio.bin\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	w.Write(ciphertext)
}
