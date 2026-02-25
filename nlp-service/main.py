import base64
import io
import re
import logging
from typing import List

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

try:
    from pypdf import PdfReader
except ImportError:
    from PyPDF2 import PdfReader  # fallback

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="NLP Extraction Service")


class DataPoint(BaseModel):
    name: str
    description: str


class ExtractDataPointsRequest(BaseModel):
    document_content: str  # base64 encoded PDF
    data_points: List[DataPoint]


class ExtractionResult(BaseModel):
    name: str
    value: str
    confidence: float


class ExtractDataPointsResponse(BaseModel):
    results: List[ExtractionResult]


def extract_text_from_pdf(pdf_bytes: bytes) -> str:
    try:
        reader = PdfReader(io.BytesIO(pdf_bytes))
        text_parts = []
        for page in reader.pages:
            text = page.extract_text()
            if text:
                text_parts.append(text)
        return "\n".join(text_parts)
    except Exception as e:
        logger.error(f"PDF extraction error: {e}")
        return ""


def find_value_for_datapoint(text: str, name: str, description: str) -> tuple[str, float]:
    """
    Search for a value matching the data point using keyword proximity.
    Returns (value, confidence).
    """
    if not text:
        return "", 0.0

    # Normalize text
    sentences = re.split(r'[.\n]+', text)
    sentences = [s.strip() for s in sentences if s.strip()]

    # Build keywords from name and description
    keywords = set()
    for word in re.split(r'[\s_\-]+', name.lower()):
        if len(word) > 2:
            keywords.add(word)
    for word in re.split(r'[\s_\-]+', description.lower()):
        if len(word) > 3 and word not in {'the', 'and', 'for', 'this', 'that', 'with', 'from'}:
            keywords.add(word)

    best_sentence = ""
    best_score = 0.0

    for sentence in sentences:
        lower_sentence = sentence.lower()
        matched = sum(1 for kw in keywords if kw in lower_sentence)
        if matched == 0:
            continue
        score = matched / max(len(keywords), 1)

        # Boost score if sentence contains a number/date pattern (common for invoice fields)
        if re.search(r'\d', sentence):
            score = min(score + 0.1, 1.0)

        if score > best_score:
            best_score = score
            best_sentence = sentence

    if best_sentence:
        # Try to extract just the relevant part (after colon or label)
        colon_match = re.search(r'[:\-]\s*(.+)', best_sentence)
        if colon_match:
            value = colon_match.group(1).strip()
        else:
            # Trim to 200 chars
            value = best_sentence[:200].strip()

        # Cap confidence at 0.95
        confidence = min(best_score + 0.3, 0.95)
        return value, round(confidence, 2)

    return "", 0.0


@app.post("/extract-datapoints", response_model=ExtractDataPointsResponse)
async def extract_datapoints(request: ExtractDataPointsRequest):
    try:
        pdf_bytes = base64.b64decode(request.document_content)
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid base64 document content")

    text = extract_text_from_pdf(pdf_bytes)
    logger.info(f"Extracted {len(text)} characters from PDF")

    results = []
    for dp in request.data_points:
        value, confidence = find_value_for_datapoint(text, dp.name, dp.description)
        results.append(ExtractionResult(
            name=dp.name,
            value=value,
            confidence=confidence,
        ))
        logger.info(f"DataPoint '{dp.name}': value='{value[:50] if value else ''}', confidence={confidence}")

    return ExtractDataPointsResponse(results=results)


@app.get("/health")
async def health():
    return {"status": "ok"}
