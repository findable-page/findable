import os
import sys
import time
import json
import subprocess
import urllib.request
import urllib.error

TUNNEL_HOST = "declined-parcel-saver-employment.trycloudflare.com"
BASE_URL = f"https://{TUNNEL_HOST}"

PAGES = [
    "01-thin-poor.html",
    "02-standard-blog.html",
    "03-faq-aeo.html",
    "04-geo-research.html",
    "05-ecommerce-product.html",
    "06-tech-docs.html",
    "07-local-business.html",
    "08-news-press.html",
    "09-voice-conversational.html",
    "10-landing-rich.html",
]

os.makedirs("testdata/reports/cli", exist_ok=True)
os.makedirs("testdata/reports/web", exist_ok=True)
os.makedirs("testdata/reports/raw", exist_ok=True)

results_summary = []

print(f"Starting test run for 10 HTML archetypes against {BASE_URL}...")
print("-" * 75)

for idx, page in enumerate(PAGES, start=1):
    page_url = f"{BASE_URL}/{page}"
    page_slug = page.replace(".html", "")
    print(f"[{idx}/10] Testing {page}...")

    # Step 1: Run our Go CLI
    cli_out_dir = f"testdata/reports/cli/{page_slug}"
    os.makedirs(cli_out_dir, exist_ok=True)
    cli_cmd = ["go", "run", ".", "-q", "-o", cli_out_dir, page_url]
    cli_proc = subprocess.run(cli_cmd, capture_output=True, text=True)
    if cli_proc.returncode != 0:
        print(f"  CLI failed: {cli_proc.stderr}")
        continue
    cli_stdout = cli_proc.stdout.strip()
    print(f"  CLI Output: {cli_stdout.splitlines()[0]}")

    # Locate generated CLI CSV
    cli_csv_file = f"{cli_out_dir}/seoscore-report-{TUNNEL_HOST}-{time.strftime('%Y-%m-%d')}.csv"

    # Step 2: Query API directly to get JSON scan
    req = urllib.request.Request(
        "https://seoscore.tools/api/scan",
        data=json.dumps({"url": page_url}).encode("utf-8"),
        headers={"Content-Type": "application/json", "User-Agent": "Mozilla/5.0"}
    )
    raw_json_path = f"testdata/reports/raw/{page_slug}.json"
    
    # Retry on 429
    scan_json = None
    for attempt in range(4):
        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                scan_json = json.loads(resp.read().decode("utf-8"))
                break
        except urllib.error.HTTPError as e:
            if e.code == 429:
                print("  Rate limited on direct API call, waiting 10s...")
                time.sleep(10)
            else:
                print(f"  API HTTP error: {e}")
                break
        except Exception as e:
            print(f"  API error: {e}")
            break

    if scan_json:
        with open(raw_json_path, "w") as f:
            json.dump(scan_json, f, indent=2)

        # Step 3: Run official website JS to generate website CSV
        web_csv_file = f"testdata/reports/web/{page_slug}.csv"
        subprocess.run(["node", "testdata/web_csv.js", raw_json_path, page_url, web_csv_file], check=True)

        seo = scan_json.get("seo", {}).get("score", 0)
        aeo = scan_json.get("aeo", {}).get("score", 0)
        geo = scan_json.get("geo", {}).get("score", 0)
        overall = round(seo * 0.5 + aeo * 0.25 + geo * 0.25)

        results_summary.append({
            "page": page,
            "overall": overall,
            "seo": seo,
            "aeo": aeo,
            "geo": geo,
            "cli_csv": cli_csv_file,
            "web_csv": web_csv_file
        })

    # Avoid hitting rate limits
    time.sleep(3)

print("-" * 75)
print("COMPLETED 10 ARCHETYPES TEST RUN.")
print("\nResults Table:")
print(f"{'Page':<32} {'Overall':<8} {'SEO':<8} {'AEO':<8} {'GEO':<8}")
print("-" * 64)
for r in results_summary:
    print(f"{r['page']:<32} {r['overall']:<8} {r['seo']:<8} {r['aeo']:<8} {r['geo']:<8}")

with open("testdata/reports/benchmark_summary.json", "w") as f:
    json.dump(results_summary, f, indent=2)
