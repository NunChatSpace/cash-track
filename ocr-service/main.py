#!/usr/bin/env python3
"""EasyOCR HTTP service for Thai bank slip text extraction."""

import io
import logging
from typing import Optional

import easyocr
from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="OCR Service", version="1.0.0")

# Initialize EasyOCR reader (loads model on startup)
reader: Optional[easyocr.Reader] = None


@app.on_event("startup")
async def startup_event():
    global reader
    logger.info("Loading EasyOCR models for Thai and English...")
    reader = easyocr.Reader(["th", "en"], gpu=False)
    logger.info("EasyOCR models loaded successfully")


@app.get("/health")
async def health_check():
    return {"status": "healthy", "models_loaded": reader is not None}


@app.post("/ocr")
async def extract_text(file: UploadFile = File(...)):
    """Extract text from uploaded image using EasyOCR."""
    if reader is None:
        raise HTTPException(status_code=503, detail="OCR models not loaded yet")

    if not file.content_type or not file.content_type.startswith("image/"):
        raise HTTPException(status_code=400, detail="File must be an image")

    try:
        contents = await file.read()
        image = Image.open(io.BytesIO(contents))

        # Convert to RGB if necessary
        if image.mode != "RGB":
            image = image.convert("RGB")

        # Run OCR
        logger.info(f"Processing image: {file.filename}")
        results = reader.readtext(image)

        # Combine all detected text
        text_lines = [result[1] for result in results]
        combined_text = "\n".join(text_lines)

        logger.info(f"Extracted {len(text_lines)} text blocks")

        return JSONResponse(
            content={
                "text": combined_text,
                "blocks": [
                    {"text": result[1], "confidence": float(result[2])}
                    for result in results
                ],
            }
        )

    except Exception as e:
        logger.error(f"OCR error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8001)
