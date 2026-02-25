<script>
  const API_BASE = 'http://localhost:8080/api';
  let documentId = '';
  let loading = false;
  let results = null;
  let documents = [];
  let error = null;

  async function loadDocuments() {
    try {
      const res = await fetch(`${API_BASE}/documents`);
      const data = await res.json();
      documents = data.documents || [];
    } catch (e) {}
  }

  async function fetchResults() {
    loading = true;
    error = null;
    try {
      const res = await fetch(`${API_BASE}/documents/${documentId}/results`);
      results = await res.json();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  loadDocuments();
</script>

<h1>Review Extraction Results</h1>
{#if documents.length > 0}
  <div>
    <h2>Documents</h2>
    <table border="1" style="border-collapse:collapse;">
      <tr><th>ID</th><th>Filename</th><th>Status</th><th>Created</th><th>Action</th></tr>
      {#each documents as doc}
        <tr>
          <td style="padding:0.3rem">{doc.id}</td>
          <td style="padding:0.3rem">{doc.filename}</td>
          <td style="padding:0.3rem">{doc.status}</td>
          <td style="padding:0.3rem">{doc.created_at}</td>
          <td style="padding:0.3rem"><button on:click={() => { documentId = doc.id; fetchResults(); }}>View Results</button></td>
        </tr>
      {/each}
    </table>
  </div>
{/if}

<div style="margin-top:1rem">
  <label>Document ID: <input bind:value={documentId} placeholder="Enter document ID" /></label>
  <button on:click={fetchResults} disabled={loading || !documentId}>
    {loading ? 'Loading...' : 'Fetch Results'}
  </button>
</div>

{#if results}
  <h2>Results for {results.document_id}</h2>
  <p>Status: <strong>{results.status}</strong></p>
  {#if results.results && results.results.length > 0}
    <table border="1" style="border-collapse:collapse; margin-top:0.5rem">
      <tr><th>Field</th><th>Value</th><th>Confidence</th></tr>
      {#each results.results as r}
        <tr>
          <td style="padding:0.3rem">{r.name}</td>
          <td style="padding:0.3rem">{r.value}</td>
          <td style="padding:0.3rem">{(r.confidence * 100).toFixed(0)}%</td>
        </tr>
      {/each}
    </table>
  {:else}
    <p>No results yet. Extraction may still be in progress.</p>
  {/if}
{/if}

{#if error}
  <p style="color:red">Error: {error}</p>
{/if}
