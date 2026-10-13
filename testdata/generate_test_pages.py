import os

os.makedirs("testdata/site", exist_ok=True)

pages = {
"01-thin-poor.html": """<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<p>Just a basic page with almost no text and no metadata.</p>
</body>
</html>""",

"02-standard-blog.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>How to Master Modern Web Performance Optimization in 2026</title>
  <meta name="description" content="Discover actionable web performance strategies, caching architectures, and frontend tuning tips to accelerate your website in 2026.">
  <link rel="canonical" href="https://example.com/02-standard-blog.html">
  <meta property="og:title" content="How to Master Modern Web Performance Optimization in 2026">
  <meta property="og:description" content="Discover actionable web performance strategies, caching architectures, and frontend tuning tips to accelerate your website in 2026.">
  <meta property="og:type" content="article">
  <meta property="og:url" content="https://example.com/02-standard-blog.html">
  <meta name="robots" content="index, follow">
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <h1>How to Master Modern Web Performance Optimization in 2026</h1>
    <p>By Jane Doe | Updated October 2026</p>
    <h2>Why Website Speed Directly Drives Conversions</h2>
    <p>Every millisecond saved translates into tangible business retention. Modern web browsers expect sub-second Time to First Byte and instant rendering. Slow load times frustrate prospective customers and increase bounce rates across desktop and mobile devices.</p>
    <p>Engineering teams must monitor real-user metrics alongside synthetic benchmarks to identify bottlenecks before users encounter degraded workflows.</p>
    <h2>Leveraging Edge Networks and Modern Image Formats</h2>
    <p>Serving WebP and AVIF assets drastically cuts network payloads. Edge caching brings computed assets directly to regional points of presence, bypassing database roundtrips for repeated queries.</p>
    <p>Adopt aggressive HTTP caching headers such as immutable cache directives for versioned bundles to guarantee seamless subsequent page views.</p>
    <p>Read more details on <a href="https://developer.mozilla.org/en-US/docs/Web/Performance" target="_blank" rel="noopener">MDN Web Performance Documentation</a>.</p>
  </main>
  <footer><p>&copy; 2026 Performance Blog</p></footer>
</body>
</html>""",

"03-faq-aeo.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Frequently Asked Questions: Cloud Cost Optimization Explained</title>
  <meta name="description" content="Get direct answers to common cloud cost optimization questions. Understand egress fees, reserved instances, and serverless pricing models.">
  <link rel="canonical" href="https://example.com/03-faq-aeo.html">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    "mainEntity": [
      {
        "@type": "Question",
        "name": "What is cloud cost optimization?",
        "acceptedAnswer": {
          "@type": "Answer",
          "text": "Cloud cost optimization is the strategic process of eliminating waste, rightsizing provisioned instances, and taking advantage of committed use discounts to minimize cloud infrastructure expenditure."
        }
      },
      {
        "@type": "Question",
        "name": "How can teams reduce AWS egress charges?",
        "acceptedAnswer": {
          "@type": "Answer",
          "text": "Teams can reduce AWS egress fees by caching assets through Cloudflare CDN, co-locating interrelated compute workloads within the same availability zone, and using VPC endpoints for S3 storage."
        }
      }
    ]
  }
  </script>
</head>
<body>
  <header role="banner"><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main role="main">
    <h1>Frequently Asked Questions: Cloud Cost Optimization Explained</h1>
    <p>By Cloud Architecture Team | <time datetime="2026-10-10">October 10, 2026</time></p>
    <section>
      <h2>What is cloud cost optimization?</h2>
      <p>Cloud cost optimization is the practice of aligning cloud expenditure with actual resource consumption through continuous rightsizing, architecture tuning, and automated shutdown of idle dev environments.</p>
    </section>
    <section>
      <h2>How can teams reduce AWS egress charges?</h2>
      <p>You can reduce egress fees significantly by placing Cloudflare in front of origin servers, routing internal communication through private VPC endpoints, and compressing JSON payloads with Brotli.</p>
    </section>
    <section>
      <h2>What are reserved instances versus spot instances?</h2>
      <dl>
        <dt>Reserved Instances</dt>
        <dd>A long-term 1 or 3 year commitment offering discounts up to 72% for predictable, steady-state production workloads.</dd>
        <dt>Spot Instances</dt>
        <dd>Surplus compute capacity purchased at discounts up to 90% suitable for stateless, interruptible batch jobs.</dd>
      </dl>
    </section>
  </main>
  <footer><p>&copy; 2026 FinOps Central</p></footer>
</body>
</html>""",

"04-geo-research.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AI Search Engines vs Traditional Google: 2026 Empirical Study</title>
  <meta name="description" content="In-depth research analyzing citation rates, source attribution, and visibility across Perplexity, ChatGPT, and Google AI Overviews.">
  <link rel="canonical" href="https://example.com/04-geo-research.html">
  <meta property="og:title" content="AI Search Engines vs Traditional Google: 2026 Empirical Study">
  <meta property="og:description" content="In-depth research analyzing citation rates, source attribution, and visibility across Perplexity, ChatGPT, and Google AI Overviews.">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "Article",
    "headline": "AI Search Engines vs Traditional Google: 2026 Empirical Study",
    "author": { "@type": "Person", "name": "Dr. Marcus Vance", "jobTitle": "Principal AI Researcher" },
    "datePublished": "2026-09-15",
    "dateModified": "2026-10-10"
  }
  </script>
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <article>
      <h1>AI Search Engines vs Traditional Google: 2026 Empirical Study</h1>
      <p>By Dr. Marcus Vance | Updated <time datetime="2026-10-10">October 10, 2026</time></p>
      <section>
        <h2>Key Takeaways</h2>
        <p>Our empirical benchmark of 1,200 search queries across 4 generative models revealed that content with structured tables and numerical facts was cited 3.2 times more frequently than generic prose.</p>
        <ul>
          <li><strong>Statistical Precision:</strong> Answers citing exact percentages achieved an 84% citation retention rate.</li>
          <li><strong>Direct Definitions:</strong> Concise 50-word summaries positioned under H2 headings were extracted in 73% of AI Overviews.</li>
          <li><strong>Domain Authority:</strong> Pages linking to primary studies in <a href="https://arxiv.org" target="_blank" rel="noopener">arXiv</a> or standard bodies showed higher authority weighting.</li>
        </ul>
      </section>
      <section>
        <h2>Comparative Citation Metrics</h2>
        <p>We measured source attribution rates across the top 4 AI query engines:</p>
        <table>
          <caption>AI Generative Citation Rates by Engine (Q3 2026)</caption>
          <thead>
            <tr><th>Platform</th><th>Queries Tested</th><th>Direct Citation Rate</th><th>Average Sources Cited</th></tr>
          </thead>
          <tbody>
            <tr><td>Perplexity Pro</td><td>1,200</td><td>89.4%</td><td>4.2</td></tr>
            <tr><td>ChatGPT Search</td><td>1,200</td><td>78.1%</td><td>3.1</td></tr>
            <tr><td>Google AI Overview</td><td>1,200</td><td>64.5%</td><td>2.4</td></tr>
            <tr><td>Claude with Search</td><td>1,200</td><td>72.0%</td><td>3.6</td></tr>
          </tbody>
        </table>
      </section>
      <section>
        <h2>Expert Opinions on Generative Engine Optimization</h2>
        <blockquote>
          <p>"Generative engines fundamentally shift search behavior from link navigation to synthesized answer consumption. If your content lacks structured empirical facts, AI systems will bypass you." - Dr. Sarah Chen, Stanford NLP Lab.</p>
        </blockquote>
      </section>
    </article>
  </main>
  <footer><p>&copy; 2026 Generative Search Institute</p></footer>
</body>
</html>""",

"05-ecommerce-product.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>UltraSound Pro ANC Wireless Headphones - TechStore</title>
  <meta name="description" content="Shop UltraSound Pro Active Noise Cancelling wireless headphones. 40-hour battery life, lossless spatial audio, and USB-C fast charging. Free shipping.">
  <link rel="canonical" href="https://example.com/05-ecommerce-product.html">
  <meta property="og:title" content="UltraSound Pro ANC Wireless Headphones - TechStore">
  <meta property="og:description" content="Shop UltraSound Pro Active Noise Cancelling wireless headphones with 40-hour battery.">
  <meta property="og:type" content="product">
  <meta property="og:image" content="https://example.com/headphones.jpg">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "Product",
    "name": "UltraSound Pro ANC Wireless Headphones",
    "image": "https://example.com/headphones.jpg",
    "description": "Premium active noise-cancelling wireless headphones with 40-hour battery life.",
    "brand": { "@type": "Brand", "name": "UltraSound" },
    "offers": {
      "@type": "Offer",
      "price": "249.99",
      "priceCurrency": "USD",
      "availability": "https://schema.org/InStock"
    }
  }
  </script>
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <h1>UltraSound Pro ANC Wireless Headphones</h1>
    <p>Rated 4.9/5 by 340+ verified buyers. In Stock and Ready to Ship.</p>
    <img src="https://example.com/headphones.jpg" alt="Matte black UltraSound Pro ANC headphones shown from 45 degree angle" width="600" height="400">
    <p>Price: <strong>$249.99 USD</strong></p>
    <h2>Product Specifications</h2>
    <table>
      <thead><tr><th>Feature</th><th>Specification</th></tr></thead>
      <tbody>
        <tr><td>Battery Life</td><td>40 hours with ANC enabled (60 hours passive)</td></tr>
        <tr><td>Weight</td><td>245 grams lightweight aluminum construction</td></tr>
        <tr><td>Connectivity</td><td>Bluetooth 5.4, 3.5mm analog, USB-C lossless</td></tr>
        <tr><td>Warranty</td><td>2-year comprehensive manufacturer coverage</td></tr>
      </tbody>
    </table>
    <h2>Key Benefits</h2>
    <ul>
      <li>Adaptive hybrid noise cancellation filters out 98% of ambient cabin and transit noise.</li>
      <li>Custom 40mm graphene drivers reproduce ultra-low distortion audio across 10Hz-45kHz.</li>
      <li>Fast charge delivers 5 hours of playback from just 10 minutes plugged in.</li>
    </ul>
  </main>
  <footer><p>&copy; 2026 TechStore Global</p></footer>
</body>
</html>""",

"06-tech-docs.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>REST API Authentication Guide - Acme Developer Platform</title>
  <meta name="description" content="Step-by-step developer documentation on authenticating requests to Acme API using Bearer tokens, HMAC signatures, and OAuth2.">
  <link rel="canonical" href="https://example.com/06-tech-docs.html">
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <h1>REST API Authentication Guide</h1>
    <p>Version 2.4 | Last updated <time datetime="2026-10-01">October 1, 2026</time></p>
    <h2>Overview</h2>
    <p>All requests made to the Acme Developer Platform API must include an authorized token in the HTTP Authorization header. Unauthorized calls return an HTTP 401 response status code.</p>
    <h2>Generating an API Key</h2>
    <ol>
      <li>Log into your Acme Developer Dashboard.</li>
      <li>Navigate to Settings &gt; API Keys &gt; Create New Secret.</li>
      <li>Store the displayed API secret key in a secure secrets manager.</li>
    </ol>
    <h2>Making an Authenticated Request</h2>
    <p>Pass your token inside the Authorization header with the Bearer prefix:</p>
    <pre><code>curl -X GET "https://api.acme.example.com/v1/projects" \\
  -H "Authorization: Bearer YOUR_API_TOKEN" \\
  -H "Accept: application/json"</code></pre>
    <h2>Handling Token Expiration</h2>
    <p>Access tokens expire every 3600 seconds (1 hour). Use the refresh token endpoint to obtain a new access token without requiring manual re-authentication.</p>
  </main>
  <footer><p>&copy; 2026 Acme Corp Developers</p></footer>
</body>
</html>""",

"07-local-business.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Green Valley Organic Bakery - Fresh Artisan Breads in Seattle</title>
  <meta name="description" content="Green Valley Bakery bakes fresh organic sourdough, pastries, and artisanal breads daily in Seattle, WA. Visit our neighborhood cafe or order online.">
  <link rel="canonical" href="https://example.com/07-local-business.html">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "Bakery",
    "name": "Green Valley Organic Bakery",
    "address": {
      "@type": "PostalAddress",
      "streetAddress": "123 Pike Street",
      "addressLocality": "Seattle",
      "addressRegion": "WA",
      "postalCode": "98101",
      "addressCountry": "US"
    },
    "telephone": "+1-206-555-0199",
    "openingHours": "Mo-Sa 07:00-18:00, Su 08:00-15:00"
  }
  </script>
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <h1>Green Valley Organic Bakery</h1>
    <p>Seattle neighborhood handcrafted artisan bakery since 2014.</p>
    <h2>Our Story and Philosophy</h2>
    <p>We source 100% stone-milled regional heritage grains from local Washington organic farms. Every loaf undergoes a slow 36-hour cold fermentation process that develops complex rustic flavor profiles and natural digestibility.</p>
    <h2>Location and Business Hours</h2>
    <p>Address: 123 Pike Street, Seattle, WA 98101</p>
    <p>Phone: (206) 555-0199</p>
    <ul>
      <li>Monday to Saturday: 7:00 AM to 6:00 PM</li>
      <li>Sunday: 8:00 AM to 3:00 PM</li>
    </ul>
    <h2>Daily Artisan Specialties</h2>
    <dl>
      <dt>Country Sourdough</dt>
      <dd>Crisp blistered crust with open airy crumb, baked in stone hearth ovens.</dd>
      <dt>Cardamom Morning Buns</dt>
      <dd>Flaky laminated dough sprinkled with freshly ground cardamom and raw cane sugar.</dd>
    </dl>
  </main>
  <footer><p>&copy; 2026 Green Valley Organic Bakery</p></footer>
</body>
</html>""",

"08-news-press.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>CleanPower Group Announces $500M Offshore Wind Expansion in Atlantic</title>
  <meta name="description" content="CleanPower Group announces a $500 million investment to deploy 1.2 gigawatts of offshore wind capacity off the Atlantic coast by 2028.">
  <link rel="canonical" href="https://example.com/08-news-press.html">
  <meta property="og:type" content="article">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "NewsArticle",
    "headline": "CleanPower Group Announces $500M Offshore Wind Expansion in Atlantic",
    "datePublished": "2026-10-11T09:00:00Z",
    "dateModified": "2026-10-11T09:30:00Z",
    "author": { "@type": "Person", "name": "Elena Rostova" },
    "publisher": { "@type": "Organization", "name": "Atlantic Energy Tribune" }
  }
  </script>
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <article>
      <h1>CleanPower Group Announces $500M Offshore Wind Expansion in Atlantic</h1>
      <p>BOSTON - By Elena Rostova | Published October 11, 2026</p>
      <h2>Historic Capital Commitment to Renewable Infrastructure</h2>
      <p>CleanPower Group today finalized a landmark $500 million financing package to construct 80 next-generation offshore wind turbines. The facility is expected to supply clean electricity to more than 450,000 coastal households upon completion in late 2028.</p>
      <p>"This milestone represents our largest renewable infrastructure project to date," stated CEO David Kim during a press briefing in Boston harbor. "We are proving that zero-carbon generation can deliver dependable grid baseload while creating thousands of union manufacturing jobs."</p>
      <h2>Environmental Review and Economic Impact</h2>
      <p>Independent marine biologists conducted a two-year ecological study confirming the deep-water installation minimizes disruption to migratory whales and benthic habitats. Regional economic agencies estimate the facility will generate $1.4 billion in economic activity over its thirty-year operating life.</p>
    </article>
  </main>
  <footer><p>&copy; 2026 Atlantic Energy Tribune</p></footer>
</body>
</html>""",

"09-voice-conversational.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>How to Start Running When You Are Out of Shape: Beginner Guide</title>
  <meta name="description" content="You want to start running but do not know where to begin? Here is a simple, injury-free guide that will have you jogging comfortably in 4 weeks.">
  <link rel="canonical" href="https://example.com/09-voice-conversational.html">
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a></nav></header>
  <main>
    <h1>How to Start Running When You Are Out of Shape: Beginner Guide</h1>
    <p>By Coach Emily Parker | Updated October 2026</p>
    <h2>What is the best way for you to start running?</h2>
    <p>The best way you can start running is using the run-walk method. You alternate between one minute of gentle jogging and two minutes of brisk walking. Doing this three times a week builds your stamina without stressing your joints.</p>
    <h2>How fast should you run on your first week?</h2>
    <p>You should run at a conversational pace where you can easily speak full sentences out loud without gasping for breath. If you cannot speak comfortably, you should slow down immediately.</p>
    <h2>What gear do you really need?</h2>
    <ul>
      <li>You need properly fitted running shoes from a specialty running store.</li>
      <li>You need moisture-wicking socks to prevent painful blisters.</li>
      <li>You need a basic water bottle for hydration on warm days.</li>
    </ul>
    <h2>Your 4-Week Beginner Plan</h2>
    <ol>
      <li><strong>Week 1:</strong> Walk 4 minutes, jog 1 minute, repeat 5 times.</li>
      <li><strong>Week 2:</strong> Walk 3 minutes, jog 2 minutes, repeat 5 times.</li>
      <li><strong>Week 3:</strong> Walk 2 minutes, jog 3 minutes, repeat 5 times.</li>
      <li><strong>Week 4:</strong> Walk 1 minute, jog 4 minutes, repeat 5 times.</li>
    </ol>
  </main>
  <footer><p>&copy; 2026 Everyday Runner</p></footer>
</body>
</html>""",

"10-landing-rich.html": """<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>DevMetrics - Automated CI/CD Performance and Observability</title>
  <meta name="description" content="DevMetrics automatically tracks build durations, test flake rates, and deployment latency across your entire GitHub organization. Try free for 14 days.">
  <link rel="canonical" href="https://example.com/10-landing-rich.html">
  <meta property="og:title" content="DevMetrics - Automated CI/CD Performance and Observability">
  <meta property="og:description" content="DevMetrics automatically tracks build durations, test flake rates, and deployment latency.">
  <meta property="og:type" content="website">
  <meta property="og:image" content="https://example.com/banner.jpg">
  <meta property="og:site_name" content="DevMetrics Platform">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="DevMetrics - CI/CD Observability">
  <meta name="twitter:description" content="Track build durations and test flake rates across your GitHub organization.">
  <meta name="robots" content="index, follow, max-image-preview:large">
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    "name": "DevMetrics",
    "applicationCategory": "DeveloperApplication",
    "operatingSystem": "All",
    "offers": { "@type": "Offer", "price": "49.00", "priceCurrency": "USD" }
  }
  </script>
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a> | <a href="/contact">Contact</a> | <a href="/privacy">Privacy</a></nav></header>
  <main>
    <h1>DevMetrics - Automated CI/CD Performance and Observability</h1>
    <p>Eliminate pipeline bottlenecks and cut your team GitHub Actions bills by up to 40% with automated flake detection.</p>
    <h2>Why 500+ Engineering Teams Choose DevMetrics</h2>
    <p>Modern engineering organizations push dozens of commits every hour. When continuous integration pipelines slow down or flaky tests fail unpredictably, developer velocity drops immediately.</p>
    <ul>
      <li><strong>Sub-Minute Detection:</strong> Pinpoint flaky test suites automatically using historical statistical regression.</li>
      <li><strong>Cost Analytics:</strong> Track runner compute consumption per developer and per repository in real time.</li>
      <li><strong>One-Click GitHub App:</strong> Installs in under 60 seconds with zero YAML pipeline modifications required.</li>
    </ul>
    <h2>Customer Testimonials</h2>
    <blockquote>
      <p>"DevMetrics reduced our pull request queue wait time from 28 minutes to 9 minutes within our first sprint." - Tech Lead at FinTech Scaleup.</p>
    </blockquote>
    <img src="https://example.com/dashboard-screenshot.png" alt="DevMetrics live analytics dashboard tracking CI pipeline speed" width="800" height="450">
  </main>
  <footer><p>&copy; 2026 DevMetrics Inc.</p></footer>
</body>
</html>"""
}

# Create index.html linking to all pages
index_links = "\n".join([f'<li><a href="/{name}">{name}</a></li>' for name in sorted(pages.keys())])
index_html = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>SEO Benchmark Test Site - 10 Archetypes</title>
  <meta name="description" content="Directory of 10 distinct HTML page types designed to benchmark SEO, AEO, and GEO scoring systems.">
</head>
<body>
  <header><nav><a href="/">Home</a> | <a href="/about">About</a></nav></header>
  <main>
    <h1>SEO Benchmark Test Suite</h1>
    <p>Select any of the 10 distinct HTML test archetypes below:</p>
    <ul>
      {index_links}
    </ul>
  </main>
  <footer><p>&copy; 2026 Benchmark Suite</p></footer>
</body>
</html>"""

pages["index.html"] = index_html

for filename, content in pages.items():
    path = os.path.join("testdata/site", filename)
    with open(path, "w") as f:
        f.write(content.strip() + "\n")
    print(f"Created {path} ({len(content)} bytes)")

print(f"Successfully generated {len(pages)} test HTML pages.")
