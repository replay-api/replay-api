#!/usr/bin/env python3
"""
PaddleOCR wrapper script for the leetgaming OCR pipeline.

Reads a PNG image from stdin or file path argument, runs PaddleOCR text
detection+recognition, and outputs structured JSON to stdout.

Usage:
    cat frame.png | python3 paddleocr_wrapper.py
    python3 paddleocr_wrapper.py --input frame.png
    python3 paddleocr_wrapper.py --input frame.png --preprocess

Output format:
    {"blocks": [{"text": "NAVI", "confidence": 0.97, "box": [[x1,y1],[x2,y2],[x3,y3],[x4,y4]]}, ...]}
"""

import sys
import json
import argparse
import os
import tempfile


def preprocess_image(img_path, output_path):
    """
    Preprocess image for better OCR on esports HUD text:
    1. Upscale small images (< 1080p) with Lanczos interpolation
    2. Increase contrast and sharpness
    3. Save as high-quality PNG
    """
    from PIL import Image, ImageEnhance

    img = Image.open(img_path)
    w, h = img.size

    # Only upscale if image is small (< 1080p height)
    if h < 1080:
        scale = 2
        img = img.resize((w * scale, h * scale), Image.LANCZOS)
    else:
        scale = 1

    # Enhance contrast (helps with colored text on colored backgrounds)
    enhancer = ImageEnhance.Contrast(img)
    img = enhancer.enhance(1.5)

    # Enhance sharpness (helps with anti-aliased game fonts)
    enhancer = ImageEnhance.Sharpness(img)
    img = enhancer.enhance(2.0)

    img.save(output_path, "PNG")
    return scale


def main():
    parser = argparse.ArgumentParser(description="PaddleOCR wrapper")
    parser.add_argument("--input", type=str, help="Input image file path (reads stdin if not provided)")
    parser.add_argument("--preprocess", action="store_true", help="Apply image preprocessing (upscale, contrast, sharpen)")
    args = parser.parse_args()

    # Suppress connectivity checks
    os.environ.setdefault("PADDLE_PDX_DISABLE_MODEL_SOURCE_CHECK", "True")

    cleanup_files = []

    if args.input:
        img_path = args.input
    else:
        # Read image from stdin and write to temp file
        image_data = sys.stdin.buffer.read()
        if not image_data:
            print(json.dumps({"blocks": []}))
            return

        tmp = tempfile.NamedTemporaryFile(suffix=".png", delete=False)
        tmp.write(image_data)
        tmp.close()
        img_path = tmp.name
        cleanup_files.append(img_path)

    try:
        # Apply preprocessing if requested
        if args.preprocess:
            preprocessed = tempfile.NamedTemporaryFile(suffix=".png", delete=False)
            preprocessed.close()
            cleanup_files.append(preprocessed.name)
            scale = preprocess_image(img_path, preprocessed.name)
            ocr_input = preprocessed.name
        else:
            scale = 1
            ocr_input = img_path

        from paddleocr import PaddleOCR

        ocr = PaddleOCR(lang="en")
        result = list(ocr.predict(ocr_input))

        blocks = []

        if result:
            r = result[0]
            for i, (text, confidence) in enumerate(zip(r["rec_texts"], r["rec_scores"])):
                if not text.strip():
                    continue

                poly = r["dt_polys"][i]
                box = poly.tolist() if hasattr(poly, "tolist") else poly
                # Ensure 4-point format and scale back to original coordinates
                if len(box) >= 4:
                    int_box = [[int(pt[0] / scale), int(pt[1] / scale)] for pt in box[:4]]
                else:
                    int_box = [[int(pt[0] / scale), int(pt[1] / scale)] for pt in box]

                blocks.append({
                    "text": text,
                    "confidence": float(confidence),
                    "box": int_box,
                })

        print(json.dumps({"blocks": blocks}))

    finally:
        for f in cleanup_files:
            if os.path.exists(f):
                os.unlink(f)


if __name__ == "__main__":
    main()
