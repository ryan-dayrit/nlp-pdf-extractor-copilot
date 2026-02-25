const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export async function uploadDocument(file, dataPoints) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('data_points', JSON.stringify(dataPoints));

  const response = await fetch(`${API_BASE}/documents`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Upload failed: ${response.status} ${text}`);
  }

  return response.json();
}

export async function listDocuments() {
  const response = await fetch(`${API_BASE}/documents`);

  if (!response.ok) {
    throw new Error(`Failed to list documents: ${response.status}`);
  }

  return response.json();
}

export async function getDataPoints(documentId) {
  const response = await fetch(`${API_BASE}/documents/${documentId}/datapoints`);

  if (!response.ok) {
    throw new Error(`Failed to get data points: ${response.status}`);
  }

  return response.json();
}
