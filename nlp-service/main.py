import base64

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from extractor import extract_data_points

app = FastAPI(title="NLP PDF Extractor", version="1.0.0")


class ExtractRequest(BaseModel):
    pdf_base64: str
    data_points: list[str]


class ExtractResponse(BaseModel):
    results: dict[str, str]


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/extract", response_model=ExtractResponse)
def extract(request: ExtractRequest) -> ExtractResponse:
    try:
        pdf_bytes = base64.b64decode(request.pdf_base64)
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid base64-encoded PDF.")

    if not pdf_bytes:
        raise HTTPException(status_code=400, detail="PDF content is empty.")

    try:
        results = extract_data_points(pdf_bytes, request.data_points)
    except Exception as exc:
        raise HTTPException(status_code=422, detail=f"Extraction failed: {exc}") from exc

    return ExtractResponse(results=results)
