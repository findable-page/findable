=== findable ===
Contributors: nadimtuhin
Tags: seo, aeo, geo, audit, ai search
Requires at least: 5.8
Tested up to: 7.1
Requires PHP: 7.4
Stable tag: 0.1.0
License: GPLv2 or later
License URI: https://www.gnu.org/licenses/gpl-2.0.html

Free SEO, AEO and GEO scans for your posts and pages, powered by the findable.page API.

== Description ==

Scan the published URL of any post or page and see SEO, AEO (answer engines) and GEO (generative engines) scores plus the failing checks.

* Meta box in the editor with an optional keyphrase
* Score column in the Posts and Pages lists
* Tools > findable: bulk scan of your latest posts (up to 20)

= External service =
This plugin sends the public permalink of the post you scan (and the optional keyphrase) to https://findable.page/api/scan. findable.page fetches that public URL to analyse it, and the request comes from your site server's IP address. Results are not stored by findable.page beyond normal caching. Nothing is sent until you click a scan button. Rate limit: 10 scans per minute per IP.

Service: https://findable.page - Source: https://github.com/nadimtuhin/findable

== Installation ==

1. Upload the `findable` folder to `/wp-content/plugins/`.
2. Activate the plugin.
3. Open a published post and click "Scan with findable".

== Changelog ==

= 0.1.0 =
* Initial release.
