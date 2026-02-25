<script>
  import Upload from './lib/Upload.svelte';
  import DocumentList from './lib/DocumentList.svelte';
  import Results from './lib/Results.svelte';

  // view: 'upload' | 'documents' | 'results'
  let view = 'upload';
  let selectedDocument = null;

  function goToDocuments() {
    selectedDocument = null;
    view = 'documents';
  }

  function goToUpload() {
    view = 'upload';
  }

  function viewResults(doc) {
    selectedDocument = doc;
    view = 'results';
  }
</script>

<div class="app">
  <nav>
    <div class="nav-brand">📄 NLP PDF Extractor</div>
    <div class="nav-tabs">
      <button
        class="nav-tab"
        class:active={view === 'upload'}
        on:click={goToUpload}
      >Upload Document</button>
      <button
        class="nav-tab"
        class:active={view === 'documents' || view === 'results'}
        on:click={goToDocuments}
      >View Documents</button>
    </div>
  </nav>

  <main>
    {#if view === 'upload'}
      <Upload onSuccess={goToDocuments} />
    {:else if view === 'documents'}
      <DocumentList onViewResults={viewResults} onUpload={goToUpload} />
    {:else if view === 'results' && selectedDocument}
      <Results document={selectedDocument} onBack={goToDocuments} />
    {/if}
  </main>
</div>

<style>
  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #f1f5f9;
    color: #1e293b;
  }

  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  nav {
    background: #fff;
    border-bottom: 1px solid #e2e8f0;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    gap: 2rem;
    height: 56px;
  }

  .nav-brand {
    font-weight: 700;
    font-size: 1rem;
    color: #1e293b;
    white-space: nowrap;
  }

  .nav-tabs {
    display: flex;
    gap: 0.25rem;
  }

  .nav-tab {
    background: none;
    border: none;
    padding: 0.45rem 1rem;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: 500;
    color: #64748b;
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
  }

  .nav-tab:hover {
    background: #f1f5f9;
    color: #1e293b;
  }

  .nav-tab.active {
    background: #ede9fe;
    color: #4f46e5;
    font-weight: 600;
  }

  main {
    flex: 1;
    padding: 2rem 1rem;
  }
</style>
