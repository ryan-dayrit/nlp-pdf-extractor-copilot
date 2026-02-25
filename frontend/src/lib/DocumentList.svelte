<script>
  import { onMount, onDestroy } from 'svelte';
  import { listDocuments } from './api.js';

  export let onViewResults;
  export let onUpload;

  let documents = [];
  let error = '';
  let intervalId;

  const STATUS_STYLES = {
    pending:    { background: '#fefce8', color: '#a16207', border: '#fde68a' },
    processing: { background: '#eff6ff', color: '#1d4ed8', border: '#bfdbfe' },
    completed:  { background: '#f0fdf4', color: '#15803d', border: '#bbf7d0' },
    failed:     { background: '#fef2f2', color: '#dc2626', border: '#fecaca' },
  };

  function badgeStyle(status) {
    const s = STATUS_STYLES[status] || { background: '#f3f4f6', color: '#374151', border: '#d1d5db' };
    return `background:${s.background};color:${s.color};border:1px solid ${s.border};`;
  }

  async function load() {
    try {
      const data = await listDocuments();
      documents = data.documents || [];
      error = '';
    } catch (e) {
      error = e.message;
    }
  }

  onMount(() => {
    load();
    intervalId = setInterval(load, 5000);
  });

  onDestroy(() => clearInterval(intervalId));
</script>

<div class="list-container">
  <div class="list-header">
    <h2>Documents</h2>
    <button class="btn-primary" on:click={onUpload}>+ Upload New Document</button>
  </div>

  {#if error}
    <div class="alert alert-error">{error}</div>
  {/if}

  {#if documents.length === 0 && !error}
    <div class="empty">No documents uploaded yet.</div>
  {:else}
    <table>
      <thead>
        <tr>
          <th>Document ID</th>
          <th>Filename</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {#each documents as doc}
          <tr class="clickable" on:click={() => onViewResults && onViewResults(doc)}>
            <td class="mono">{doc.document_id ? doc.document_id.substring(0, 12) + '…' : '—'}</td>
            <td>{doc.filename || '—'}</td>
            <td>
              <span class="badge" style={badgeStyle(doc.status)}>
                {doc.status || 'unknown'}
              </span>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .list-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  }

  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  h2 {
    margin: 0;
    color: #1e293b;
    font-size: 1.4rem;
  }

  .alert {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.9rem;
    background: #fef2f2;
    color: #dc2626;
    border: 1px solid #fecaca;
  }

  .empty {
    color: #9ca3af;
    text-align: center;
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
  }

  .clickable {
    cursor: pointer;
    transition: background 0.1s;
  }

  .clickable:hover {
    background: #f8fafc;
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

  .btn-primary {
    background: #4f46e5;
    color: #fff;
    border: none;
    padding: 0.55rem 1.1rem;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }

  .btn-primary:hover {
    background: #4338ca;
  }
</style>
