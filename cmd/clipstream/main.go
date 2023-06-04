package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	urlPtr := flag.String("u", "", "URL")
	flag.Parse()

	if *urlPtr == "" {
		fmt.Println("You must provide a URL (e.g. -u https://www.youtube.com/watch?v=12345)")
		os.Exit(1)
	}

	// if it is a youtube URL, get the stream URL, otherwise use yt-dlp -g {url} to get the stream URL

	url := *urlPtr

	if strings.HasPrefix(url, "https://www.youtube.com") {
		// Get the stream URL
		streamURL, err := GetStreamURL(url)
		if err != nil {
			fmt.Println("Error getting stream URL:", err)
			os.Exit(1)
		}

		url = streamURL
	}

	// Unix timestamp for this run
	timestamp := time.Now().Unix()

	err := os.MkdirAll("livestream-source", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	err = os.MkdirAll("clips", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	liveStreamFile := fmt.Sprintf("livestream-source/livestream_%d.ts", timestamp)

	go DownloadStream(url, liveStreamFile)

	clipLength := 10

	i := 0
	for {
		start := i * clipLength
		end := start + clipLength

		duration, err := GetVideoDuration(liveStreamFile)
		if err != nil {
			if !strings.Contains(err.Error(), "No such file or directory") {
				fmt.Println("Error getting video duration:", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		if duration < end {
			time.Sleep(1 * time.Second)
			continue
		}

		output := fmt.Sprintf("clips/clip_%d_%05d_%05d.mp4", timestamp, start, end)

		ok, err := ClipVideo(liveStreamFile, output, start, end)
		if err != nil {
			fmt.Println("Error clipping video:", err)
		}

		if ok {
			i++
			fmt.Printf("clipped %s\n", output)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func DownloadStream(url, output string) {
	cmd := exec.Command("ffmpeg", "-i", url, "-c", "copy", output)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error downloading stream:", err)
	}
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

func GetVideoDuration(file string) (int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", file)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %w, output: %s", err, string(cmdOutput))
	}

	// Parse the duration from the output
	durationStr := strings.TrimSpace(string(cmdOutput))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing duration: %w", err)
	}

	return int(duration), nil
}

func GetStreamURL(ytUrl string) (string, error) {
	cmd := exec.Command("yt-dlp", "-g", ytUrl)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp error: %w, output: %s", err, string(cmdOutput))
	}

	// Parse the stream URL from the output
	streamURL := strings.TrimSpace(string(cmdOutput))

	return streamURL, nil
}
