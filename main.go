package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	ffmpegPath = "ffmpeg" // Укажите путь к ffmpeg
)

func main() {
	http.HandleFunc("/stream", streamHandler)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	// Добавляем заголовки CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Обрабатываем предварительный запрос
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	mkvURL := r.URL.Query().Get("url")
	if mkvURL == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	// Создаем временный файл для хранения MP4
	tempFile := "output.mp4"

	cmd := exec.Command(ffmpegPath, "-i", mkvURL, "-codec:v", "libx264", "-codec:a", "aac", "-movflags", "frag_keyframe+empty_moov", "-f", "mp4", tempFile)
	var stderrOutput strings.Builder
	cmd.Stderr = &stderrOutput

	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to run ffmpeg: %v\n%s", err, stderrOutput.String()), http.StatusInternalServerError)
		return
	}

	// Открываем временный файл
	file, err := os.Open(tempFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open temp file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Получаем размер файла и модифицированное время
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file info: %v", err), http.StatusInternalServerError)
		return
	}
	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime()

	// Обработка заголовка Range
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, tempFile, modTime, io.NewSectionReader(file, 0, fileSize))
	} else {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeContent(w, r, tempFile, modTime, file)
	}
}
