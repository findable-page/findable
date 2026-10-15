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
})();
