package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go [start seconds] [end seconds]")
		os.Exit(1)
	}

	start, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error parsing start seconds:", err)
		os.Exit(1)
	}

	end, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error parsing end seconds:", err)
		os.Exit(1)
	}

	timestamp := ""
	liveStreamFile := ""

	if len(os.Args) > 3 {
		// Use the provided livestream timestamp
		timestamp = os.Args[3]
		liveStreamFile = fmt.Sprintf("livestream-source/livestream_%s.ts", timestamp)
	} else {
		// Get the newest livestream file from the folder
		files, err := ioutil.ReadDir("livestream-source")
		if err != nil {
			fmt.Println("Error reading livestream files:", err)
			os.Exit(1)
		}

		// Sort files by modification time
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})

		// Get the newest livestream file
		if len(files) > 0 {
			newestFile := files[0]
			timestamp = parseTimestampFromFilename(newestFile.Name())
			liveStreamFile = filepath.Join("livestream-source", newestFile.Name())
		} else {
			fmt.Println("No livestream files found")
			os.Exit(1)
		}
	}

	output := fmt.Sprintf("clip_%s_%05d_%05d.mp4", timestamp, start, end)

	ok, err := ClipVideo(liveStreamFile, output, start, end)
	if err != nil {
		fmt.Println("Error clipping video:", err)
		os.Exit(1)
	}

	if !ok {
		fmt.Println("Failed to create clip")
		os.Exit(1)
	}

	fmt.Println("Clip created successfully:", output)
}

func ClipVideo(input, output string, start, end int) (bool, error) {
	cmd := exec.Command("ffmpeg", "-i", input, "-ss", strconv.Itoa(start), "-to", strconv.Itoa(end), "-c", "copy", "-bsf:a", "aac_adtstoasc", output)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("ffmpeg error: %w, output: %s", err, string(cmdOutput))
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		// The file was not created, so the clip was not successfully created
		return false, nil
	}

	return true, nil
}

func parseTimestampFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return ""
	}
	return strings.TrimSuffix(parts[2], ".ts")
}
