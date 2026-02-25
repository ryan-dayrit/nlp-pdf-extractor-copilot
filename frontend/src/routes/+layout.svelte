<nav style="padding: 1rem; background: #1a1a2e; color: white; display: flex; gap: 1rem; align-items: center;">
  <strong>PDF Extractor</strong>
  <a href="/" style="color: #e94560;">Upload Document</a>
  <a href="/datapoints" style="color: #e94560;">Data Points</a>
  <a href="/review" style="color: #e94560;">Review</a>
</nav>
<main style="padding: 2rem;">
  {@render children()}
</main>

<script>
  let { children } = $props();
</script>
