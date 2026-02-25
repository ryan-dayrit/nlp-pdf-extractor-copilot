# nlp-pdf-extractor-copilot

An automated, event-driven pipeline for extracting structured data points from PDF documents using NLP — built with GitHub Copilot using Claude Sonnet 4.6 .

Upload a PDF and a list of fields you want extracted (e.g. `"invoice_total"`, `"vendor_name"`). The system processes the document asynchronously via Kafka and returns the extracted key-value pairs through a REST/gRPC API.

# dev time 
16m 3s 

---

## Architecture

```
                          ┌─────────────┐
                          │   Frontend  │  :80  (Svelte + Nginx)
                          └──────┬──────┘
                                 │  HTTP (REST)
                          ┌──────▼──────┐
                          │ grpc-service│  :50051 (gRPC)
                          │             │  :8080  (HTTP/JSON REST)
                          └──────┬──────┘
                    publish │         ▲ update results
                          ┌──────▼──────┐
                          │    Kafka    │  :9092 / :29092
                          │ (document-  │
                          │  uploads)   │
                          └──────┬──────┘
                        consume │
                          ┌──────▼──────┐
                          │  consumer   │  (Go Kafka consumer)
                          └──────┬──────┘
                                 │  POST /extract
                          ┌──────▼──────┐
                          │ nlp-service │  :8000  (FastAPI + PyMuPDF)
                          └─────────────┘
```

### Services

| Service | Language | Image / Source | Port(s) |
|---|---|---|---|
| `zookeeper` | — | `confluentinc/cp-zookeeper:7.6.0` | 2181 |
| `kafka` | — | `confluentinc/cp-kafka:7.6.0` | 9092 (internal), 29092 (host) |
| `kafka-init` | — | `confluentinc/cp-kafka:7.6.0` | — |
| `nlp-service` | Python (FastAPI) | `./nlp-service` | 8000 |
| `grpc-service` | Go | `./grpc-service` | 50051, 8080 |
| `consumer` | Go | `./consumer` | — |
| `frontend` | Svelte | `./frontend` | 80 |

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24
- [Docker Compose](https://docs.docker.com/compose/install/) ≥ 2 (bundled with Docker Desktop)

---

## Quick Start

```bash
git clone https://github.com/your-org/nlp-pdf-extractor-copilot.git
cd nlp-pdf-extractor-copilot
docker-compose up --build
```

Once all containers are healthy, open **http://localhost** in your browser.

---

## API Documentation — gRPC Service

The gRPC service exposes both a **gRPC** interface (port `50051`) and a plain **HTTP/JSON REST** interface (port `8080`). The frontend and consumer communicate over HTTP.

### `POST /documents`

Upload a PDF for processing.

**Content-Type:** `multipart/form-data`

| Field | Type | Required | Description |
|---|---|---|---|
| `file` | file | ✅ | The PDF file to upload |
| `data_points` | string (JSON array) | ✅ | Fields to extract, e.g. `["invoice_total","vendor_name"]` |

**Response `201 Created`:**
```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending"
}
```

---

### `GET /documents`

List all uploaded documents.

**Response `200 OK`:**
```json
{
  "documents": [
    {
      "document_id": "550e8400-...",
      "filename": "invoice.pdf",
      "status": "completed"
    }
  ]
}
```

---

### `GET /documents/{id}/datapoints`

Retrieve extraction results for a specific document.

**Response `200 OK`:**
```json
{
  "document_id": "550e8400-...",
  "status": "completed",
  "results": {
    "invoice_total": "€1,250.00",
    "vendor_name": "Acme Corp"
  }
}
```

**Response `404 Not Found`** if the document ID does not exist.

---

### `POST /documents/{id}/datapoints`

Update extraction results for a document (called internally by the consumer).

**Request body:**
```json
{
  "results": {
    "invoice_total": "€1,250.00",
    "vendor_name": "Acme Corp"
  }
}
```

**Response `200 OK`:**
```json
{ "status": "updated" }
```

---

## API Documentation — NLP Service

### `GET /health`

Health check.

**Response `200 OK`:** `{"status": "ok"}`

### `POST /extract`

Extract data points from a PDF.

**Request body:**
```json
{
  "pdf_base64": "<base64-encoded PDF bytes>",
  "data_points": ["invoice_total", "vendor_name"]
}
```

**Response `200 OK`:**
```json
{
  "results": {
    "invoice_total": "€1,250.00",
    "vendor_name": "Acme Corp"
  }
}
```

---

## Development Setup

### nlp-service (Python / FastAPI)

```bash
cd nlp-service
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

### grpc-service (Go)

```bash
cd grpc-service
go mod download
KAFKA_BROKERS=localhost:29092 go run .
# gRPC on :50051, HTTP on :8080
```

### consumer (Go)

```bash
cd consumer
go mod download
KAFKA_BROKERS=localhost:29092 \
NLP_SERVICE_URL=http://localhost:8000 \
GRPC_SERVICE_URL=http://localhost:8080 \
go run .
```

### frontend (Svelte)

```bash
cd frontend
npm install
npm run dev
# dev server on http://localhost:5173
```

> **Tip:** Start Kafka locally (or via `docker-compose up zookeeper kafka`) before running the Go services individually.

---

## Environment Variables

| Variable | Service | Default (Docker) | Description |
|---|---|---|---|
| `KAFKA_BROKERS` | grpc-service, consumer | `kafka:9092` | Comma-separated Kafka broker addresses |
| `NLP_SERVICE_URL` | consumer | `http://nlp-service:8000` | Base URL of the NLP extraction service |
| `GRPC_SERVICE_URL` | consumer | `http://grpc-service:8080` | Base URL of the gRPC HTTP gateway |

---
