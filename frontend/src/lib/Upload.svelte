<script>
  import { uploadDocument } from './api.js';

  export let onSuccess;

  let file = null;
  let dataPointsText = 'company name, invoice date, total amount, invoice number';
  let loading = false;
  let error = '';
  let success = '';

  function handleFileChange(event) {
    file = event.target.files[0] || null;
  }

  async function handleSubmit() {
    if (!file) {
      error = 'Please select a PDF file.';
      return;
    }

    const dataPoints = dataPointsText
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    if (dataPoints.length === 0) {
      error = 'Please enter at least one data point.';
      return;
    }

    loading = true;
    error = '';
    success = '';

    try {
      await uploadDocument(file, dataPoints);
      success = 'Document uploaded successfully!';
      setTimeout(() => onSuccess && onSuccess(), 1000);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="upload-container">
  <h2>Upload Document</h2>

  <div class="form-group">
    <label for="pdf-file">PDF File</label>
    <input
      id="pdf-file"
      type="file"
      accept=".pdf"
      on:change={handleFileChange}
      disabled={loading}
    />
  </div>

  <div class="form-group">
    <label for="data-points">Data Points to Extract <span class="hint">(comma-separated)</span></label>
    <textarea
      id="data-points"
      bind:value={dataPointsText}
      rows="3"
      placeholder="e.g. company name, invoice date, total amount"
      disabled={loading}
    ></textarea>
  </div>

  {#if error}
    <div class="alert alert-error">{error}</div>
  {/if}

  {#if success}
    <div class="alert alert-success">{success}</div>
  {/if}

  <button class="btn-primary" on:click={handleSubmit} disabled={loading}>
    {loading ? 'Uploading...' : 'Upload Document'}
  </button>
</div>

<style>
  .upload-container {
    max-width: 560px;
    margin: 0 auto;
    padding: 2rem;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  }

  h2 {
    margin: 0 0 1.5rem;
    color: #1e293b;
    font-size: 1.4rem;
  }

  .form-group {
    margin-bottom: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  label {
    font-weight: 600;
    color: #374151;
    font-size: 0.9rem;
  }

  .hint {
    font-weight: 400;
    color: #9ca3af;
    font-size: 0.8rem;
  }

  input[type="file"], textarea {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.95rem;
    font-family: inherit;
    transition: border-color 0.15s;
  }

  input[type="file"]:focus, textarea:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79,70,229,0.1);
  }

  textarea {
    resize: vertical;
  }

  .alert {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
  }

  .alert-error {
    background: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #f0fdf4;
    color: #16a34a;
    border: 1px solid #bbf7d0;
  }

  .btn-primary {
    background: #4f46e5;
    color: #fff;
    border: none;
    padding: 0.65rem 1.5rem;
    border-radius: 6px;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }

  .btn-primary:hover:not(:disabled) {
    background: #4338ca;
  }

  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
