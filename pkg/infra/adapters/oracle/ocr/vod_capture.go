package oracle_ocr

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
)

// VodCapture captures frames from YouTube VODs (past recordings) using yt-dlp + ffmpeg.
// Unlike StreamlinkCapture which targets live streams, this adapter handles video-on-demand
// content by downloading and seeking to specific timestamps.
type VodCapture struct {
	ytdlpPath      string
	ffmpegPath     string
	captureTimeout time.Duration
}

// Compile-time interface check
var _ oracle_out.StreamCapturePort = (*VodCapture)(nil)

// NewVodCapture creates a new VOD capture adapter
func NewVodCapture() *VodCapture {
	ytdlpPath, _ := exec.LookPath("yt-dlp")
	if ytdlpPath == "" {
		ytdlpPath = "yt-dlp"
	}

	ffmpegPath, _ := exec.LookPath("ffmpeg")
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &VodCapture{
		ytdlpPath:      ytdlpPath,
		ffmpegPath:     ffmpegPath,
		captureTimeout: 60 * time.Second,
	}
}

// CaptureFrame extracts a single frame from a VOD at the beginning of the video.
// For timestamp-specific capture, use CaptureFrameAt.
func (v *VodCapture) CaptureFrame(ctx context.Context, videoURL string) ([]byte, error) {
	return v.CaptureFrameAt(ctx, videoURL, "")
}

// CaptureFrameAt extracts a single frame from a YouTube VOD at the given timestamp.
// timestamp format: "HH:MM:SS" or "MM:SS" or seconds (e.g. "01:23:45", "83:45", "5025")
// If timestamp is empty, captures a frame from 10 seconds in (to skip intros).
func (v *VodCapture) CaptureFrameAt(ctx context.Context, videoURL string, timestamp string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, v.captureTimeout)
	defer cancel()

	if timestamp == "" {
		timestamp = "00:00:10"
	}

	// Step 1: Get the direct stream URL via yt-dlp (720p preferred, good balance of quality vs processing speed)
	ytdlpArgs := []string{
		"-f", "bestvideo[height<=720]+bestaudio/best[height<=720]/best",
		"--get-url",
		"--no-warnings",
		"--no-playlist",
		videoURL,
	}

	ytdlpCmd := exec.CommandContext(ctx, v.ytdlpPath, ytdlpArgs...)
	var ytdlpOut, ytdlpErr bytes.Buffer
	ytdlpCmd.Stdout = &ytdlpOut
	ytdlpCmd.Stderr = &ytdlpErr

	if err := ytdlpCmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp get URL failed: %w (stderr: %s)", err, ytdlpErr.String())
	}

	// yt-dlp outputs one URL per line; first line is video stream
	streamURL := bytes.TrimSpace(bytes.SplitN(ytdlpOut.Bytes(), []byte("\n"), 2)[0])
	if len(streamURL) == 0 {
		return nil, fmt.Errorf("yt-dlp returned empty stream URL")
	}

	slog.DebugContext(ctx, "got direct stream URL",
		slog.String("video_url", videoURL),
		slog.Int("url_length", len(streamURL)),
	)

	// Step 2: Use ffmpeg to seek to timestamp and extract one frame as PNG
	ffmpegArgs := []string{
		"-ss", timestamp,
		"-i", string(streamURL),
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-loglevel", "error",
		"pipe:1",
	}

	ffmpegCmd := exec.CommandContext(ctx, v.ffmpegPath, ffmpegArgs...)
	var ffmpegStdout, ffmpegStderr bytes.Buffer
	ffmpegCmd.Stdout = &ffmpegStdout
	ffmpegCmd.Stderr = &ffmpegStderr

	if err := ffmpegCmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg frame extract failed: %w (stderr: %s)", err, ffmpegStderr.String())
	}

	frameData := ffmpegStdout.Bytes()
	if len(frameData) == 0 {
		return nil, fmt.Errorf("empty frame captured at timestamp %s", timestamp)
	}

	slog.DebugContext(ctx, "VOD frame captured",
		slog.String("video_url", videoURL),
		slog.String("timestamp", timestamp),
		slog.Int("frame_bytes", len(frameData)),
	)

	return frameData, nil
}

// CaptureFramesMulti extracts frames at multiple timestamps, returning them in order.
// Useful for scanning a VOD for scoreboards at regular intervals.
func (v *VodCapture) CaptureFramesMulti(ctx context.Context, videoURL string, timestamps []string) ([][]byte, error) {
	frames := make([][]byte, 0, len(timestamps))

	for _, ts := range timestamps {
		frame, err := v.CaptureFrameAt(ctx, videoURL, ts)
		if err != nil {
			slog.WarnContext(ctx, "frame capture failed at timestamp, skipping",
				slog.String("timestamp", ts),
				slog.String("error", err.Error()),
			)
			continue
		}
		frames = append(frames, frame)
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("no frames captured from %d timestamps", len(timestamps))
	}

	return frames, nil
}

// IsStreamLive always returns false for VODs (they are not live).
func (v *VodCapture) IsStreamLive(ctx context.Context, streamURL string) (bool, error) {
	return false, nil
}

// GetVideoDuration returns the duration of a YouTube video in seconds using yt-dlp.
func (v *VodCapture) GetVideoDuration(ctx context.Context, videoURL string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	args := []string{
		"--get-duration",
		"--no-warnings",
		"--no-playlist",
		videoURL,
	}

	cmd := exec.CommandContext(ctx, v.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("yt-dlp get duration failed: %w", err)
	}

	durationStr := bytes.TrimSpace(stdout.Bytes())
	return parseDuration(string(durationStr))
}

// parseDuration parses yt-dlp duration format "HH:MM:SS" or "MM:SS" to seconds
func parseDuration(s string) (float64, error) {
	var d time.Duration
	// yt-dlp outputs "H:MM:SS" or "MM:SS"
	parts := bytes.Split([]byte(s), []byte(":"))

	switch len(parts) {
	case 3:
		var h, m, sec int
		_, err := fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
		if err != nil {
			return 0, fmt.Errorf("parse duration %q: %w", s, err)
		}
		d = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	case 2:
		var m, sec int
		_, err := fmt.Sscanf(s, "%d:%d", &m, &sec)
		if err != nil {
			return 0, fmt.Errorf("parse duration %q: %w", s, err)
		}
		d = time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	default:
		var sec int
		_, err := fmt.Sscanf(s, "%d", &sec)
		if err != nil {
			return 0, fmt.Errorf("parse duration %q: %w", s, err)
		}
		d = time.Duration(sec) * time.Second
	}

	return d.Seconds(), nil
}
