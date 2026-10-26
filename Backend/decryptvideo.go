package main

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func decryptBytes(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext is too short")
	}

	nonce := ciphertext[:nonceSize]
	actualData := ciphertext[nonceSize:]

	return aesgcm.Open(nil, nonce, actualData, nil)
}

func decryptVideoPipeline(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("decrypt_video")
	if err != nil {
		http.Error(w, "Error retrieving the ZIP file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempDir, err := os.MkdirTemp("", "viddecrypt-*")
	if err != nil {
		http.Error(w, "Could not create temporary folder.", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	zipData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read the ZIP file.", http.StatusInternalServerError)
		return
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), header.Size)
	if err != nil {
		http.Error(w, "Invalid ZIP file format", http.StatusBadRequest)
		return
	}

	framesDir := filepath.Join(tempDir, "frames")
	os.Mkdir(framesDir, 0755)
	audioPath := filepath.Join(tempDir, "restored_audio.wav")

	for _, zipFile := range zipReader.File {
		rc, err := zipFile.Open()
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(rc)
		rc.Close()

		if zipFile.Name == "audio.enc" {

			plainAudio, err := decryptBytes(data, audioSecretKey)
			if err != nil {
				http.Error(w, "Audio decryption failed: "+err.Error(), http.StatusBadRequest)
				return
			}
			os.WriteFile(audioPath, plainAudio, 0644)
		} else if strings.HasPrefix(zipFile.Name, "frames/") && strings.HasSuffix(zipFile.Name, ".enc") {

			plainFrame, err := decryptBytes(data, imageSecretKey)
			if err != nil {
				http.Error(w, "Image decryption failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			baseName := filepath.Base(zipFile.Name)
			pngName := strings.TrimSuffix(baseName, ".enc") + ".png"
			os.WriteFile(filepath.Join(framesDir, pngName), plainFrame, 0644)
		}
	}

	restoredVideoPath := filepath.Join(tempDir, "restored_output.mp4")
	frameInputPattern := filepath.Join(framesDir, "frame_%04d.png")

	cmd := exec.Command("ffmpeg",
		"-framerate", "30",
		"-i", frameInputPattern,
		"-i", audioPath,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-shortest",
		restoredVideoPath,
	)

	if err := cmd.Run(); err != nil {
		http.Error(w, "Unable to recreate the video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	videoBytes, err := os.ReadFile(restoredVideoPath)
	if err != nil {
		http.Error(w, "Unable to read the video file.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\"restored_video.mp4\"")
	w.Header().Set("Content-Type", "video/mp4")
	w.Write(videoBytes)
}
