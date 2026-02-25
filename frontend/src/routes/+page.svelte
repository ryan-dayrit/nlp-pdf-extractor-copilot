<script>
  const API_BASE = 'http://localhost:8080/api';
  let file = null;
  let uploading = false;
  let result = null;
  let error = null;

  async function upload() {
    if (!file) return;
    uploading = true;
    error = null;
    const form = new FormData();
    form.append('file', file);
    try {
      const res = await fetch(`${API_BASE}/documents`, { method: 'POST', body: form });
      result = await res.json();
    } catch (e) {
      error = e.message;
    } finally {
      uploading = false;
    }
  }
</script>

<h1>Upload PDF Document</h1>
<input type="file" accept=".pdf" on:change={e => file = e.target.files[0]} />
<button on:click={upload} disabled={uploading || !file}>
  {uploading ? 'Uploading...' : 'Upload'}
</button>
{#if result}
  <p>✅ Uploaded! Document ID: <strong>{result.document_id}</strong></p>
  <p>Go to <a href="/datapoints">Data Points</a> to configure extraction.</p>
{/if}
{#if error}
  <p style="color:red">Error: {error}</p>
{/if}
