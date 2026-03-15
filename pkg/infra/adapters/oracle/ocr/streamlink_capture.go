package oracle_ocr

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
)

// StreamlinkCapture captures frames from live streams using streamlink + ffmpeg
type StreamlinkCapture struct {
	streamlinkPath string
	ffmpegPath     string
	quality        string
	captureTimeout time.Duration
}

var _ oracle_out.StreamCapturePort = (*StreamlinkCapture)(nil)

func NewStreamlinkCapture(quality string) *StreamlinkCapture {
	if quality == "" {
		quality = "720p,720p60,best"
	}

	streamlinkPath, _ := exec.LookPath("streamlink")
	if streamlinkPath == "" {
		streamlinkPath = "streamlink"
	}

	ffmpegPath, _ := exec.LookPath("ffmpeg")
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	return &StreamlinkCapture{
		streamlinkPath: streamlinkPath,
		ffmpegPath:     ffmpegPath,
		quality:        quality,
		captureTimeout: 30 * time.Second,
	}
}

func (s *StreamlinkCapture) CaptureFrame(ctx context.Context, streamURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.captureTimeout)
	defer cancel()

	streamlinkArgs := []string{
		streamURL,
		s.quality,
		"--stdout",
		"--hls-live-restart",
		"--stream-timeout", "15",
	}

	ffmpegArgs := []string{
		"-i", "pipe:0",
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-loglevel", "error",
		"pipe:1",
	}

	streamlinkCmd := exec.CommandContext(ctx, s.streamlinkPath, streamlinkArgs...)
	streamlinkOut, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("streamlink stdout pipe: %w", err)
	}

	var streamlinkStderr bytes.Buffer
	streamlinkCmd.Stderr = &streamlinkStderr

	ffmpegCmd := exec.CommandContext(ctx, s.ffmpegPath, ffmpegArgs...)
	ffmpegCmd.Stdin = streamlinkOut

	var ffmpegStdout bytes.Buffer
	var ffmpegStderr bytes.Buffer
	ffmpegCmd.Stdout = &ffmpegStdout
	ffmpegCmd.Stderr = &ffmpegStderr

	if err := streamlinkCmd.Start(); err != nil {
		return nil, fmt.Errorf("streamlink start failed: %w", err)
	}

	if err := ffmpegCmd.Start(); err != nil {
		_ = streamlinkCmd.Process.Kill()
		return nil, fmt.Errorf("ffmpeg start failed: %w", err)
	}

	ffmpegErr := ffmpegCmd.Wait()
	_ = streamlinkCmd.Process.Kill()
	_ = streamlinkCmd.Wait()

	if ffmpegErr != nil {
		slog.WarnContext(ctx, "ffmpeg failed",
			slog.String("stderr", ffmpegStderr.String()),
			slog.String("streamlink_stderr", streamlinkStderr.String()),
		)
		return nil, fmt.Errorf("ffmpeg failed: %w (stderr: %s)", ffmpegErr, ffmpegStderr.String())
	}

	frameData := ffmpegStdout.Bytes()
	if len(frameData) == 0 {
		return nil, fmt.Errorf("empty frame captured (streamlink stderr: %s)", streamlinkStderr.String())
	}

	slog.DebugContext(ctx, "frame captured",
		slog.String("stream_url", streamURL),
		slog.Int("frame_bytes", len(frameData)),
	)

	return frameData, nil
}

func (s *StreamlinkCapture) IsStreamLive(ctx context.Context, streamURL string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.streamlinkPath, "--json", streamURL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		output := stderr.String()
		if strings.Contains(output, "No playable streams") ||
			strings.Contains(output, "error: No streams found") ||
			strings.Contains(output, "Could not find any streams") {
			return false, nil
		}
		return false, fmt.Errorf("streamlink check failed: %w (stderr: %s)", err, output)
	}

	return stdout.Len() > 0, nil
}
