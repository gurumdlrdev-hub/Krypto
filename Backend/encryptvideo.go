package main

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func encryptBytes(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aesgcm.Seal(nonce, nonce, data, nil), nil
}

func encryptVideoPipeline(w http.ResponseWriter, r *http.Request) {

	file, _, err := r.FormFile("encrypt_video")
	if err != nil {
		http.Error(w, "Error retrieving video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempDir, err := os.MkdirTemp("", "vidcrypto-*")
	if err != nil {
		http.Error(w, "Could not create temporary folder.", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	videoPath := filepath.Join(tempDir, "input.mp4")
	tempVideoFile, _ := os.Create(videoPath)
	io.Copy(tempVideoFile, file)
	tempVideoFile.Close()

	durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	out, err := durationCmd.Output()
	if err != nil {
		http.Error(w, "Unable to calculate video duration.", http.StatusBadRequest)
		return
	}
	duration, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if duration > 20.0 {
		http.Error(w, "The video length must be under 20 seconds!", http.StatusBadRequest)
		return
	}

	audioPath := filepath.Join(tempDir, "extracted_audio.wav")
	exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "pcm_s16le", audioPath).Run()

	framesDir := filepath.Join(tempDir, "frames")
	os.Mkdir(framesDir, 0755)
	framePattern := filepath.Join(framesDir, "frame_%04d.png")
	exec.Command("ffmpeg", "-i", videoPath, "-an", "-vf", "fps=30", framePattern).Run()

	w.Header().Set("Content-Disposition", "attachment; filename=\"encrypted_vault.zip\"")
	w.Header().Set("Content-Type", "application/zip")
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	if audioBytes, err := os.ReadFile(audioPath); err == nil {
		encAudio, _ := encryptBytes(audioBytes, audioSecretKey)
		writer, _ := zipWriter.Create("audio.enc")
		writer.Write(encAudio)
	}

	frameFiles, _ := os.ReadDir(framesDir)
	for i, frameEntry := range frameFiles {
		frameBytes, _ := os.ReadFile(filepath.Join(framesDir, frameEntry.Name()))
		encFrame, _ := encryptBytes(frameBytes, imageSecretKey)

		writer, _ := zipWriter.Create(fmt.Sprintf("frames/frame_%04d.enc", i+1))
		writer.Write(encFrame)
	}
}
