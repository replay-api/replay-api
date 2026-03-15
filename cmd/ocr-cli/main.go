package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
	oracle_services "github.com/replay-api/replay-api/pkg/domain/oracle/services"
	oracle_ocr "github.com/replay-api/replay-api/pkg/infra/adapters/oracle/ocr"
	replay_common "github.com/replay-api/replay-common/pkg/replay"
)

type ScoredFrame struct {
	Timestamp string                       `json:"timestamp"`
	RawBlocks []oracle_out.TextBlock       `json:"raw_blocks"`
	Score     *oracle_services.ParsedScore `json:"score,omitempty"`
	Error     string                       `json:"error,omitempty"`
	FrameFile string                       `json:"frame_file,omitempty"`
}

func main() {
	videoURL := flag.String("url", "", "YouTube video URL (required)")
	timestamp := flag.String("ts", "", "Timestamp to capture (HH:MM:SS or MM:SS)")
	scan := flag.Bool("scan", false, "Scan the video at intervals to find scoreboards")
	scanInterval := flag.Int("interval", 30, "Seconds between captures when scanning")
	scanStart := flag.String("scan-start", "00:05:00", "Start timestamp for scanning")
	scanEnd := flag.String("scan-end", "", "End timestamp for scanning (default: video duration)")
	game := flag.String("game", "cs2", "Game ID (cs2, csgo, vlrnt)")
	cropRegion := flag.String("crop", "", "Crop region as x,y,width,height (e.g. 100,50,800,200)")
	saveFrames := flag.String("save-frames", "", "Directory to save captured frames as PNG files")
	pythonPath := flag.String("python", "python3", "Path to python3 binary")
	ocrScript := flag.String("ocr-script", "", "Path to paddleocr_wrapper.py (auto-detected if empty)")
	useGPU := flag.Bool("gpu", false, "Use GPU for PaddleOCR inference")
	verbose := flag.Bool("v", false, "Verbose logging")
	jsonOut := flag.Bool("json", false, "Output results as JSON")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ocr-cli: Extract CS2 scores from YouTube VODs using OCR\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  ocr-cli -url <youtube-url> -ts HH:MM:SS          # Single frame\n")
		fmt.Fprintf(os.Stderr, "  ocr-cli -url <youtube-url> -scan                  # Scan video\n")
		fmt.Fprintf(os.Stderr, "  ocr-cli -url <youtube-url> -scan -interval 60     # Custom interval\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *videoURL == "" {
		flag.Usage()
		os.Exit(1)
	}

	logLevel := slog.LevelWarn
	if *verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	if *ocrScript == "" {
		*ocrScript = findOCRScript()
	}

	vodCapture := oracle_ocr.NewVodCapture()
	ocrEngine := oracle_ocr.NewPaddleOCRAdapter(*pythonPath, *ocrScript, *useGPU)
	parser := oracle_services.NewOCRScoreParser()

	gameID := replay_common.GameIDKey(*game)
	ctx := context.Background()

	var region *oracle_out.Region
	if *cropRegion != "" {
		r, err := parseCropRegion(*cropRegion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid crop region %q: %v\n", *cropRegion, err)
			os.Exit(1)
		}
		region = r
	} else {
		// Auto-set CS2 scoreboard HUD region (top-center where round score is displayed)
		// For 720p (1280x720), the scoreboard is at approximately (400,0)-(880,80)
		switch gameID {
		case replay_common.CS2_GAME_ID, replay_common.CSGO_GAME_ID:
			region = &oracle_out.Region{X: 350, Y: 0, Width: 530, Height: 80}
			fmt.Fprintf(os.Stderr, "Auto-cropping to CS2 scoreboard HUD region (%d,%d %dx%d)\n",
				region.X, region.Y, region.Width, region.Height)
		}
	}

	if *scan {
		results := runScan(ctx, vodCapture, ocrEngine, parser, *videoURL, gameID, region,
			*scanStart, *scanEnd, *scanInterval, *saveFrames)
		outputResults(results, *jsonOut)
	} else if *timestamp != "" {
		result := processSingleFrame(ctx, vodCapture, ocrEngine, parser, *videoURL, *timestamp, gameID, region, *saveFrames)
		outputResults([]ScoredFrame{result}, *jsonOut)
	} else {
		fmt.Fprintln(os.Stderr, "No timestamp specified, capturing at 00:01:00...")
		result := processSingleFrame(ctx, vodCapture, ocrEngine, parser, *videoURL, "00:01:00", gameID, region, *saveFrames)
		outputResults([]ScoredFrame{result}, *jsonOut)
	}
}

func processSingleFrame(
	ctx context.Context,
	capture *oracle_ocr.VodCapture,
	ocrEngine *oracle_ocr.PaddleOCRAdapter,
	parser *oracle_services.OCRScoreParser,
	videoURL, timestamp string,
	gameID replay_common.GameIDKey,
	region *oracle_out.Region,
	saveDir string,
) ScoredFrame {
	result := ScoredFrame{Timestamp: timestamp}

	fmt.Fprintf(os.Stderr, "Capturing frame at %s...\n", timestamp)
	frameData, err := capture.CaptureFrameAt(ctx, videoURL, timestamp)
	if err != nil {
		result.Error = fmt.Sprintf("capture failed: %v", err)
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", result.Error)
		return result
	}
	fmt.Fprintf(os.Stderr, "Frame captured (%d bytes)\n", len(frameData))

	if saveDir != "" {
		result.FrameFile = saveFrame(frameData, saveDir, timestamp)
	}

	fmt.Fprintln(os.Stderr, "Running OCR...")
	blocks, err := ocrEngine.ExtractText(ctx, frameData, region)
	if err != nil {
		result.Error = fmt.Sprintf("OCR failed: %v", err)
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", result.Error)
		return result
	}

	result.RawBlocks = blocks
	fmt.Fprintf(os.Stderr, "OCR found %d text blocks:\n", len(blocks))
	for _, b := range blocks {
		fmt.Fprintf(os.Stderr, "   [%.0f%%] %q\n", b.Confidence*100, b.Text)
	}

	score, err := parser.ParseScoreboard(blocks, gameID)
	if err != nil {
		result.Error = fmt.Sprintf("score parse failed: %v", err)
		fmt.Fprintf(os.Stderr, "WARN: %s\n", result.Error)
		return result
	}

	result.Score = score
	printScore(timestamp, score)
	return result
}

func runScan(
	ctx context.Context,
	capture *oracle_ocr.VodCapture,
	ocrEngine *oracle_ocr.PaddleOCRAdapter,
	parser *oracle_services.OCRScoreParser,
	videoURL string,
	gameID replay_common.GameIDKey,
	region *oracle_out.Region,
	scanStart, scanEnd string,
	intervalSec int,
	saveDir string,
) []ScoredFrame {
	if scanEnd == "" {
		fmt.Fprintln(os.Stderr, "Getting video duration...")
		dur, err := capture.GetVideoDuration(ctx, videoURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: Could not get duration: %v, defaulting to 2 hours\n", err)
			dur = 7200
		}
		scanEnd = formatTimestamp(int(dur))
		fmt.Fprintf(os.Stderr, "Video duration: %s\n", scanEnd)
	}

	startSec := timestampToSeconds(scanStart)
	endSec := timestampToSeconds(scanEnd)

	var timestamps []string
	for s := startSec; s <= endSec; s += intervalSec {
		timestamps = append(timestamps, formatTimestamp(s))
	}

	fmt.Fprintf(os.Stderr, "Scanning %d positions from %s to %s (every %ds)...\n",
		len(timestamps), scanStart, scanEnd, intervalSec)

	var results []ScoredFrame
	scoresFound := 0

	for i, ts := range timestamps {
		fmt.Fprintf(os.Stderr, "\n--- [%d/%d] %s ---\n", i+1, len(timestamps), ts)
		result := processSingleFrame(ctx, capture, ocrEngine, parser, videoURL, ts, gameID, region, saveDir)
		results = append(results, result)
		if result.Score != nil {
			scoresFound++
		}
	}

	fmt.Fprintf(os.Stderr, "\n========================================\n")
	fmt.Fprintf(os.Stderr, "Scan complete: %d/%d frames had parseable scores\n", scoresFound, len(timestamps))

	if scoresFound > 0 {
		fmt.Fprintf(os.Stderr, "\nScores found:\n")
		for _, r := range results {
			if r.Score != nil {
				printScore(r.Timestamp, r.Score)
			}
		}
	}

	return results
}

func printScore(timestamp string, score *oracle_services.ParsedScore) {
	if score.TeamAName != "" && score.TeamBName != "" {
		fmt.Fprintf(os.Stderr, "SCORE [%s] %s %d - %d %s", timestamp, score.TeamAName, score.TeamAScore, score.TeamBScore, score.TeamBName)
	} else {
		fmt.Fprintf(os.Stderr, "SCORE [%s] %d - %d", timestamp, score.TeamAScore, score.TeamBScore)
	}
	if score.MapName != "" {
		fmt.Fprintf(os.Stderr, " (Map: %s)", score.MapName)
	}
	fmt.Fprintln(os.Stderr)
}

func outputResults(results []ScoredFrame, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(results)
		return
	}
	for _, r := range results {
		if r.Score != nil {
			if r.Score.TeamAName != "" {
				fmt.Printf("%s\t%s %d - %d %s", r.Timestamp, r.Score.TeamAName, r.Score.TeamAScore, r.Score.TeamBScore, r.Score.TeamBName)
			} else {
				fmt.Printf("%s\t%d - %d", r.Timestamp, r.Score.TeamAScore, r.Score.TeamBScore)
			}
			if r.Score.MapName != "" {
				fmt.Printf("\t%s", r.Score.MapName)
			}
			fmt.Println()
		}
	}
}

func saveFrame(data []byte, dir, timestamp string) string {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Warn("failed to create frame directory", slog.String("error", err.Error()))
		return ""
	}
	safeName := strings.ReplaceAll(timestamp, ":", "-")
	path := filepath.Join(dir, fmt.Sprintf("frame_%s.png", safeName))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.Warn("failed to save frame", slog.String("error", err.Error()))
		return ""
	}
	fmt.Fprintf(os.Stderr, "Frame saved: %s\n", path)
	return path
}

func findOCRScript() string {
	candidates := []string{
		"scripts/paddleocr_wrapper.py",
		"replay-api/scripts/paddleocr_wrapper.py",
		"../scripts/paddleocr_wrapper.py",
		"../../scripts/paddleocr_wrapper.py",
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "scripts", "paddleocr_wrapper.py"),
			filepath.Join(exeDir, "..", "scripts", "paddleocr_wrapper.py"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "scripts/paddleocr_wrapper.py"
}

func parseCropRegion(s string) (*oracle_out.Region, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("expected 4 values (x,y,width,height), got %d", len(parts))
	}
	vals := make([]int, 4)
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid value %q at position %d: %w", p, i, err)
		}
		vals[i] = v
	}
	return &oracle_out.Region{X: vals[0], Y: vals[1], Width: vals[2], Height: vals[3]}, nil
}

func timestampToSeconds(ts string) int {
	parts := strings.Split(ts, ":")
	switch len(parts) {
	case 3:
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		s, _ := strconv.Atoi(parts[2])
		return h*3600 + m*60 + s
	case 2:
		m, _ := strconv.Atoi(parts[0])
		s, _ := strconv.Atoi(parts[1])
		return m*60 + s
	default:
		s, _ := strconv.Atoi(ts)
		return s
	}
}

func formatTimestamp(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
