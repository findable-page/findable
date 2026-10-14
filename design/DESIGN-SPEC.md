# Light redesign spec (Fable + Opus + Codex fusion; mockup: design/mockup-fusion.png)
Do this AFTER the review fixes, as its own commits. Only edit internal/server/web/index.html: re-skin the :root tokens, keep every id/JS hook, CSP, FAQ, author line and question H2s.
- Theme: light is the brand. Drop the OS-dark auto-pick; dark stays as the opt-in #themeToggle.
- Tokens: --bg #FFFFFF, --surface-hover #F6F7FB, --text #0F1222, --text-2 #3A3F5C, --muted #5D6382,
  --accent #4F2BFF, --accent-light #EEEBFF, --amber #FFB020, --ok #0E7C4A, --warn #8A4B00, --bad #C81E2B,
  --fam-seo #0E7C4A, --fam-aeo #4F2BFF, --fam-geo #D1480D, --border #E6E8F0, --shadow: 4px 4px 0 #0F1222 (hard, no blur).
- Background: a 24px dot grid in #E6E8F0, with blurred blobs (amber top-left, violet top-right) behind the hero.
- Type: system stack only (also remove Inter from the logo SVG). h1 clamp(2.4rem,6vw,4.25rem)/1.02, weight 800, -0.03em. Scores: ui-monospace, tabular-nums.
- Hero: h1 left-aligned; .hl = amber marker highlight `linear-gradient(transparent 55%,#FFB020 55%)` on black text. Badge pill in --accent-light: "<real check count> checks · 0 signups · 100% free". The count must be true for the engine that serves the page (Opus HIGH #1).
- Scan card: 2px --text border, hard shadow, radius 16px; :focus-within shadow grows to 6px. Mode = segmented control with a sliding violet pill. #scanBtn violet with a hard shadow; :active translate(4px,4px) and shadow 0.
- Results: .score-hero bordered with a hard shadow; ring stroke 10; grade = amber stamp rotated -3deg. Family cards get a 6px top border in --fam-*. Top fixes = a numbered checklist with mono numbers and green +pts chips. Failing rows get a 4px --bad left rail. Status always shows colour plus text.
- Motion (only under prefers-reduced-motion: no-preference): rings fill over 900ms; #agentSteps items slide in with a 60ms stagger; the marker highlight sweeps once on load.
- Cut: gradients, glow, glassmorphism, blur shadows, and the "beta" label on crawl.
- Gates: AA contrast; gzip budget passes; self-scan SEO>=90 AEO>=85 GEO>=80; worker CSP hashes rebuilt; bun test and go test green. Screenshot before/after into ui-shots/.
