import re
import fitz  # PyMuPDF
import spacy

# Load spaCy model once at import time; fall back gracefully if unavailable.
try:
    _nlp = spacy.load("en_core_web_sm")
except OSError:
    _nlp = None

# ---------------------------------------------------------------------------
# Regex patterns
# ---------------------------------------------------------------------------

_DATE_PATTERNS = [
    r"\b\d{4}-\d{2}-\d{2}\b",                                    # YYYY-MM-DD
    r"\b\d{1,2}/\d{1,2}/\d{2,4}\b",                              # MM/DD/YYYY or M/D/YY
    r"\b(?:January|February|March|April|May|June|July|August|"
    r"September|October|November|December)\s+\d{1,2},?\s+\d{4}\b",  # Month DD, YYYY
    r"\b\d{1,2}\s+(?:January|February|March|April|May|June|July|"
    r"August|September|October|November|December)\s+\d{4}\b",    # DD Month YYYY
    r"\b\d{1,2}-(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-\d{2,4}\b",
]

_AMOUNT_PATTERNS = [
    r"[$€£¥]\s*\d{1,3}(?:,\d{3})*(?:\.\d{2})?",                 # $1,234.56
    r"\b\d{1,3}(?:,\d{3})*(?:\.\d{2})?\s*(?:USD|EUR|GBP|CAD|AUD)\b",  # 1,234.56 USD
]

# Labels that typically precede a company / organisation name in a document.
_COMPANY_LABEL_RE = re.compile(
    r"(?:company|client|vendor|from|to|bill\s+to|sold\s+to|supplier|"
    r"employer|organization|organisation)\s*[:\-]\s*(.+)",
    re.IGNORECASE,
)


def _extract_text(pdf_bytes: bytes) -> str:
    """Return concatenated text from every page of the PDF."""
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    pages = [page.get_text() for page in doc]
    doc.close()
    return "\n".join(pages)


def _find_dates(text: str) -> list[str]:
    matches: list[str] = []
    for pattern in _DATE_PATTERNS:
        matches.extend(re.findall(pattern, text, re.IGNORECASE))
    return matches


def _find_amounts(text: str) -> list[str]:
    matches: list[str] = []
    for pattern in _AMOUNT_PATTERNS:
        matches.extend(re.findall(pattern, text))
    return matches


def _find_company_names(text: str) -> list[str]:
    # 1. Try structured label patterns first.
    label_matches = _COMPANY_LABEL_RE.findall(text)
    results = [m.strip().splitlines()[0].strip() for m in label_matches if m.strip()]

    # 2. Fall back to spaCy NER ORG entities.
    if not results and _nlp is not None:
        doc = _nlp(text[:100_000])  # cap to avoid memory issues
        results = [ent.text.strip() for ent in doc.ents if ent.label_ == "ORG"]

    return results


def _find_by_label(text: str, label: str) -> str:
    """Generic extraction: find *label* in text and return the value that follows it."""
    pattern = re.compile(
        r"(?i)" + re.escape(label) + r"\s*[:\-]?\s*(.+)",
    )
    match = pattern.search(text)
    if match:
        # Return only the first line of whatever follows.
        return match.group(1).strip().splitlines()[0].strip()
    return ""


# ---------------------------------------------------------------------------
# Public entry point
# ---------------------------------------------------------------------------

def extract_data_points(pdf_bytes: bytes, data_points: list[str]) -> dict[str, str]:
    """Extract requested *data_points* from *pdf_bytes* and return a mapping."""
    text = _extract_text(pdf_bytes)
    results: dict[str, str] = {}

    for dp in data_points:
        key = dp.lower().strip()
        value = ""

        if re.search(r"\bdate\b", key):
            found = _find_dates(text)
            value = found[0] if found else ""

        elif re.search(r"\b(amount|total|price|cost|sum|balance|due)\b", key):
            found = _find_amounts(text)
            value = found[0] if found else ""

        elif re.search(r"\b(company|client|vendor|supplier|organization|organisation)\b", key):
            found = _find_company_names(text)
            value = found[0] if found else ""

        else:
            # Generic: look for the label verbatim in the document.
            value = _find_by_label(text, dp)

        results[dp] = value

    return results
