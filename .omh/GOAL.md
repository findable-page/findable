# Product Goal: seoscore-cli

## Objective
Deliver a fast, zero-dependency Go CLI and Cloudflare Worker API that provides unified **SEO**, **AEO**, and **GEO** auditing completely offline, replacing proprietary SaaS auditors (`seoscore.tools`) with open-source engineering parity.

---

## 3 Core Pillars

### 1. Zero-Dependency Offline Execution
- **Problem:** SaaS audit tools impose rate limits, require internet access, and cannot inspect local build artifacts or CI pipelines.
- **Solution:** In-process Go engine using `golang.org/x/net/html` that audits both live URLs and local `.html` files in <50ms without headless browsers or external APIs.

### 2. Unified Tri-Family Scoring Formula
- **Taxonomy:**
  - **Traditional SEO (50%):** Metadata length boundaries, H1-H3 hierarchy, OpenGraph, canonicals, security headers, image alt/dimensions.
  - **Answer Engine Optimization / AEO (25%):** FAQ schemas, 30–80 word direct answer blocks, question headings, procedural `<ol>` steps, definition lists `<dl>`, conversational tone.
  - **Generative Engine Optimization / GEO (25%):** Quantitative data density (%, ms, $), expert quotes `<blockquote>`, schema author credentials (`Person`, `Organization`), `sameAs` entity links, citations.
- **Parity:** Exact 0–100 composite score (`round(SEO*0.5 + AEO*0.25 + GEO*0.25)`) and byte-for-byte CSV report generation matching `seoscore.tools`.

### 3. Dual Deployment Targets
- **CLI & Local Daemon:** Run `seoscore-cli <target>` for immediate reports, or `seoscore-cli -serve :8080` for a local REST API with built-in Web UI.
- **Cloudflare Worker:** Standalone edge API in `worker/` (<25 KiB bundle) deploying instantly via `wrangler deploy` to serve `/api/scan` and the web interface globally.

---

## Concrete Acceptance Criteria
- [x] Pass 24 seam-driven test suites covering boundary checks, archetypes, and API endpoints.
- [x] Zero network calls required for local file audits.
- [x] CSV export matches web client format with UTF-8 BOM.
- [x] Standalone Cloudflare Worker bundles under 25 KiB with zero external runtime deps.
- [x] Built-in Web UI accessible on `GET /` for both local server and Cloudflare Worker.
