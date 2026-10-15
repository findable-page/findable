(function () {
	var d = window.findableData;
	if (!d) { return; }

	function post(params) {
		var body = new URLSearchParams(params);
		body.append('action', 'findable_scan');
		body.append('nonce', d.nonce);
		return fetch(d.ajaxUrl, { method: 'POST', credentials: 'same-origin', body: body })
			.then(function (r) { return r.json().then(function (j) { j.status = r.status; return j; }); });
	}

	var btn = document.getElementById('findable-scan');
	if (btn) {
		btn.addEventListener('click', function () {
			var out = document.getElementById('findable-result');
			var spin = document.getElementById('findable-spinner');
			var kp = document.getElementById('findable-keyphrase');
			btn.disabled = true;
			spin.classList.add('is-active');
			post({ post_id: btn.dataset.post, keyphrase: kp ? kp.value : '' })
				.then(function (res) {
					if (res.success) {
						out.innerHTML = res.data.html; // server-escaped HTML
					} else {
						out.textContent = (res.data && res.data.message) || d.i18n.failed;
					}
				})
				.catch(function () { out.textContent = d.i18n.failed; })
				.then(function () { btn.disabled = false; spin.classList.remove('is-active'); });
		});
	}

	var bulk = document.getElementById('findable-bulk');
	if (bulk) {
		var status = document.getElementById('findable-bulk-status');
		var sleep = function (ms) { return new Promise(function (r) { setTimeout(r, ms); }); };
		bulk.addEventListener('click', async function () {
			bulk.disabled = true;
			var rows = document.querySelectorAll('#findable-bulk-body tr');
			var tbl = document.getElementById('findable-bulk-body').parentNode;
			var offset = parseInt(tbl.dataset.offset, 10) || 0;
			var total = parseInt(tbl.dataset.total, 10) || rows.length;
			var retried = {};
			for (var i = 0; i < rows.length; i++) {
				var row = rows[i];
				var started = Date.now();
				status.textContent = d.i18n.scanning + ' ' + (offset + i + 1) + ' ' + d.i18n.of + ' ' + total;
				try {
					var res = await post({ post_id: row.dataset.post });
					if (res.status === 429 && !retried[i]) {
						retried[i] = true;
						status.textContent = (res.data && res.data.message) || d.i18n.failed;
						await sleep(60000);
						i--;
						continue;
					}
					if (res.success) {
						['overall', 'seo', 'aeo', 'geo', 'failing'].forEach(function (k) {
							row.querySelector('[data-f="' + k + '"]').textContent = res.data[k] == null ? '-' : res.data[k];
						});
						var cell = row.querySelector('[data-f="report"]');
						var a = document.createElement('a');
						a.href = res.data.report;
						a.target = '_blank';
						a.rel = 'noopener';
						a.textContent = d.i18n.report;
						cell.textContent = '';
						cell.appendChild(a);
					} else {
						row.querySelector('[data-f="overall"]').textContent = (res.data && res.data.message) || d.i18n.failed;
					}
				} catch (e) {
					row.querySelector('[data-f="overall"]').textContent = d.i18n.failed;
				}
				if (i < rows.length - 1) { await sleep(Math.max(0, d.delayMs - (Date.now() - started))); }
			}
			status.textContent = d.i18n.done;
			bulk.disabled = false;
		});
	}

	function fix(postId, mode) {
		var body = new URLSearchParams({ action: 'findable_fix', nonce: d.fixNonce, post_id: postId, mode: mode });
		return fetch(d.ajaxUrl, { method: 'POST', credentials: 'same-origin', body: body })
			.then(function (r) { return r.json(); });
	}

	document.querySelectorAll('.findable-fix').forEach(function (b) {
		b.addEventListener('click', function () {
			var st = document.getElementById('findable-fix-status');
			fix(b.dataset.post, b.dataset.mode)
				.then(function (res) { st.textContent = res.success ? res.data.status : ((res.data && res.data.message) || d.i18n.failed); })
				.catch(function () { st.textContent = d.i18n.failed; });
		});
	});

	var fixAll = document.getElementById('findable-fixall');
	if (fixAll) {
		var fixStatus = document.getElementById('findable-fixall-status');
		fixAll.addEventListener('click', async function () {
			fixAll.disabled = true;
			var rows = document.querySelectorAll('#findable-bulk-body tr');
			for (var i = 0; i < rows.length; i++) {
				if (!rows[i].querySelector('[data-f="report"]').textContent.trim()) { continue; }
				fixStatus.textContent = d.i18n.fixing + ' ' + (i + 1) + '/' + rows.length;
				try {
					var res = await fix(rows[i].dataset.post, 'apply');
					rows[i].querySelector('[data-f="fixes"]').textContent = res.success ? res.data.status : d.i18n.failed;
				} catch (e) {
					rows[i].querySelector('[data-f="fixes"]').textContent = d.i18n.failed;
				}
			}
			fixStatus.textContent = d.i18n.done;
			fixAll.disabled = false;
		});
	}
})();
