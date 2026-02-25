<script>
  import { onMount, onDestroy } from 'svelte';
  import { getDataPoints } from './api.js';

  export let document;
  export let onBack;

  let result = null;
  let error = '';
  let intervalId;

  async function load() {
    try {
      result = await getDataPoints(document.document_id);
      error = '';
      if (result.status === 'completed' || result.status === 'failed') {
        clearInterval(intervalId);
      }
    } catch (e) {
      error = e.message;
    }
  }

  onMount(() => {
    load();
    intervalId = setInterval(load, 3000);
  });

  onDestroy(() => clearInterval(intervalId));

  $: resultEntries = result && result.results ? Object.entries(result.results) : [];

  const STATUS_STYLES = {
    pending:    'background:#fefce8;color:#a16207;border:1px solid #fde68a;',
    processing: 'background:#eff6ff;color:#1d4ed8;border:1px solid #bfdbfe;',
    completed:  'background:#f0fdf4;color:#15803d;border:1px solid #bbf7d0;',
    failed:     'background:#fef2f2;color:#dc2626;border:1px solid #fecaca;',
  };
</script>

<div class="results-container">
  <button class="btn-back" on:click={onBack}>← Back to Documents</button>

  <h2>Extraction Results</h2>

  <div class="meta">
    <div class="meta-row">
      <span class="meta-label">Document ID</span>
      <span class="meta-value mono">{document.document_id}</span>
    </div>
    <div class="meta-row">
      <span class="meta-label">Filename</span>
      <span class="meta-value">{document.filename || '—'}</span>
    </div>
    {#if result}
      <div class="meta-row">
        <span class="meta-label">Status</span>
        <span class="badge" style={STATUS_STYLES[result.status] || 'background:#f3f4f6;color:#374151;border:1px solid #d1d5db;'}>
          {result.status}
        </span>
      </div>
    {/if}
  </div>

  {#if error}
    <div class="alert alert-error">{error}</div>
  {/if}

  {#if !result && !error}
    <div class="loading">Loading results…</div>
  {:else if result && result.status !== 'completed'}
    <div class="info">
      {#if result.status === 'failed'}
        Extraction failed for this document.
      {:else}
        Extraction in progress. This page will refresh automatically.
      {/if}
    </div>
  {:else if result && resultEntries.length > 0}
    <table>
      <thead>
        <tr>
          <th>Data Point</th>
          <th>Extracted Value</th>
        </tr>
      </thead>
      <tbody>
        {#each resultEntries as [key, value]}
          <tr>
            <td class="key-cell">{key}</td>
            <td>{value !== null && value !== undefined ? value : '—'}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if result && result.status === 'completed'}
    <div class="info">No results available.</div>
  {/if}
</div>

<style>
  .results-container {
    max-width: 720px;
    margin: 0 auto;
    padding: 2rem;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  }

  .btn-back {
    background: none;
    border: none;
    color: #4f46e5;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    padding: 0;
    margin-bottom: 1.25rem;
    display: inline-block;
  }

  .btn-back:hover {
    color: #4338ca;
    text-decoration: underline;
  }

  h2 {
    margin: 0 0 1.25rem;
    color: #1e293b;
    font-size: 1.4rem;
  }

  .meta {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    padding: 1rem 1.25rem;
    margin-bottom: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .meta-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    min-width: 100px;
  }

  .meta-value {
    color: #1e293b;
    font-size: 0.9rem;
  }

  .mono {
    font-family: monospace;
    font-size: 0.85rem;
    color: #64748b;
  }

  .badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    font-size: 0.78rem;
    font-weight: 600;
    text-transform: capitalize;
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

  .loading, .info {
    text-align: center;
    color: #64748b;
    padding: 2rem;
    font-size: 0.95rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  th {
    text-align: left;
    padding: 0.65rem 1rem;
    background: #f8fafc;
    color: #64748b;
    font-weight: 600;
    border-bottom: 1px solid #e2e8f0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #f1f5f9;
    color: #1e293b;
    vertical-align: top;
  }

  .key-cell {
    font-weight: 600;
    color: #374151;
    white-space: nowrap;
  }
</style>
