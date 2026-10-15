(function () {
	var d = window.findableData;
	if (!d) { return; }

	function post(params) {
		var body = new URLSearchParams(params);
		body.append('action', 'findable_scan');
		body.append('nonce', d.nonce);
		return fetch(d.ajaxUrl, { method: 'POST', credentials: 'same-origin', body: body })
			.then(function (r) { return r.json(); });
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
			for (var i = 0; i < rows.length; i++) {
				var row = rows[i];
				status.textContent = d.i18n.scanning + ' ' + (i + 1) + '/' + rows.length;
				try {
					var res = await post({ post_id: row.dataset.post });
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
				if (i < rows.length - 1) { await sleep(d.delayMs); }
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
