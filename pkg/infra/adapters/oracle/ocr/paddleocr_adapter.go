package oracle_ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"os/exec"
	"time"

	oracle_out "github.com/replay-api/replay-api/pkg/domain/oracle/ports/out"
)

// PaddleOCRAdapter runs PaddleOCR as a subprocess to extract text from images
type PaddleOCRAdapter struct {
	pythonPath    string
	scriptPath    string
	ocrTimeout    time.Duration
}

// Compile-time interface check
var _ oracle_out.OCREnginePort = (*PaddleOCRAdapter)(nil)

// NewPaddleOCRAdapter creates a new PaddleOCR adapter
// scriptPath points to the Python script that wraps PaddleOCR
func NewPaddleOCRAdapter(pythonPath, scriptPath string, useGPU bool) *PaddleOCRAdapter {
	if pythonPath == "" {
		pythonPath = "python3"
	}

	return &PaddleOCRAdapter{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
		ocrTimeout: 60 * time.Second,
	}
}

// paddleOCRResult represents the JSON output from the PaddleOCR Python script
type paddleOCRResult struct {
	Blocks []paddleOCRBlock `json:"blocks"`
}

type paddleOCRBlock struct {
	Text       string     `json:"text"`
	Confidence float64    `json:"confidence"`
	Box        [4][2]int  `json:"box"` // 4 corner points [x,y]
}

// ExtractText runs PaddleOCR on the given image data and returns text blocks.
// If a region is specified, the image is cropped to that region first.
func (p *PaddleOCRAdapter) ExtractText(ctx context.Context, imageData []byte, region *oracle_out.Region) ([]oracle_out.TextBlock, error) {
	ctx, cancel := context.WithTimeout(ctx, p.ocrTimeout)
	defer cancel()

	// Crop image if region is specified
	processedData := imageData
	if region != nil {
		cropped, err := cropImage(imageData, region)
		if err != nil {
			slog.WarnContext(ctx, "image crop failed, using full image",
				slog.String("error", err.Error()),
			)
			// Fall back to full image
		} else {
			processedData = cropped
		}
	}

	// Write to temp file (PaddleOCR v3 works better with file paths)
	tmpFile, err := os.CreateTemp("", "ocr-frame-*.png")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(processedData); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	// Run PaddleOCR Python script with file path
	args := []string{p.scriptPath, "--input", tmpFile.Name()}
	// Enable preprocessing for cropped regions (small images benefit from upscale + enhance)
	if region != nil {
		args = append(args, "--preprocess")
	}

	cmd := exec.CommandContext(ctx, p.pythonPath, args...)
	cmd.Env = append(os.Environ(), "PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK=True")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("paddleocr failed: %w (stderr: %s)", err, stderr.String())
	}

	// Parse JSON output
	var result paddleOCRResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("paddleocr output parse failed: %w (raw: %s)", err, stdout.String())
	}

	// Convert to domain TextBlocks
	blocks := make([]oracle_out.TextBlock, 0, len(result.Blocks))
	for _, b := range result.Blocks {
		if b.Text == "" {
			continue
		}

		// Convert 4-point polygon to bounding box [x, y, width, height]
		minX, minY := b.Box[0][0], b.Box[0][1]
		maxX, maxY := b.Box[0][0], b.Box[0][1]
		for _, pt := range b.Box[1:] {
			if pt[0] < minX {
				minX = pt[0]
			}
			if pt[1] < minY {
				minY = pt[1]
			}
			if pt[0] > maxX {
				maxX = pt[0]
			}
			if pt[1] > maxY {
				maxY = pt[1]
			}
		}

		blocks = append(blocks, oracle_out.TextBlock{
			Text:       b.Text,
			Confidence: b.Confidence,
			BoundingBox: [4]int{minX, minY, maxX - minX, maxY - minY},
		})
	}

	slog.DebugContext(ctx, "OCR extraction complete",
		slog.Int("text_blocks", len(blocks)),
	)

	return blocks, nil
}

// cropImage crops a PNG image to the given region
func cropImage(data []byte, region *oracle_out.Region) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	rect := image.Rect(
		region.X, region.Y,
		region.X+region.Width, region.Y+region.Height,
	)

	// Type assert to SubImager (most image types support this)
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	si, ok := img.(subImager)
	if !ok {
		return nil, fmt.Errorf("image type %T does not support cropping", img)
	}

	cropped := si.SubImage(rect)

	var buf bytes.Buffer
	if err := png.Encode(&buf, cropped); err != nil {
		return nil, fmt.Errorf("encode cropped png: %w", err)
	}

	return buf.Bytes(), nil
}
