package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fileNames := []string{"test.m4a"}
	if len(os.Args) > 1 {
		fileNames = os.Args[1:]
	}

	fmt.Println(fileNames)

	for _, file := range fileNames {
		if !strings.HasSuffix(file, ".m4a") {
			file = file + ".m4a"
		}
		// Define the ffmpeg command and its arguments
		cmd := exec.Command("ffmpeg",
			"-f", "lavfi",
			"-i", "sine=frequency=1000:duration=1",
			"-c:a", "aac",
			"-b:a", "128k",
			"-metadata", "title=Test Title",
			"-metadata", "artist=Test Artist",
			"-metadata", "album=Test Album",
			"-metadata", "genre=Test Genre",
			file,
		)

		output, err := cmd.CombinedOutput()
		log.Printf("ffmpeg output:\n%s", string(output))
		if err != nil {
			log.Fatalf("ffmpeg command failed: %v", err)
		}

	}
}
