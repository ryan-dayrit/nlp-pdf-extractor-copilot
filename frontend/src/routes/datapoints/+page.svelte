<script>
  const API_BASE = 'http://localhost:8080/api';
  let documentId = $state('');
  let dataPoints = $state([{ name: '', description: '' }]);
  let submitting = $state(false);
  let result = $state(null);
  let error = $state(null);

  function addDataPoint() {
    dataPoints = [...dataPoints, { name: '', description: '' }];
  }
  function removeDataPoint(i) {
    dataPoints = dataPoints.filter((_, idx) => idx !== i);
  }

  async function submit() {
    submitting = true;
    error = null;
    try {
      const res = await fetch(`${API_BASE}/documents/${documentId}/datapoints`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ data_points: dataPoints })
      });
      result = await res.json();
    } catch (e) {
      error = e.message;
    } finally {
      submitting = false;
    }
  }
</script>

<h1>Submit Data Points</h1>
<div>
  <label>Document ID: <input bind:value={documentId} placeholder="Enter document ID" /></label>
</div>
<h2>Data Points to Extract</h2>
{#each dataPoints as dp, i}
  <div style="display:flex; gap:0.5rem; margin:0.5rem 0;">
    <input bind:value={dp.name} placeholder="Field name (e.g. invoice_number)" />
    <input bind:value={dp.description} placeholder="Description (e.g. The invoice number)" />
    <button onclick={() => removeDataPoint(i)}>Remove</button>
  </div>
{/each}
<button onclick={addDataPoint}>+ Add Data Point</button>
<br/><br/>
<button onclick={submit} disabled={submitting || !documentId}>
  {submitting ? 'Submitting...' : 'Submit for Extraction'}
</button>
{#if result}
  <p>✅ Submitted! Status: {result.status}</p>
  <p>Go to <a href="/review">Review</a> to see results.</p>
{/if}
{#if error}
  <p style="color:red">Error: {error}</p>
{/if}
