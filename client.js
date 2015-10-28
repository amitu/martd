// https://gist.github.com/Xeoncross/7663273

window.martd = function() {
	var ajax = function (url, data, callback, etag) {
		var x = new(this.XMLHttpRequest || ActiveXObject)('MSXML2.XMLHTTP.3.0');
		x.open(data ? 'POST' : 'GET', url, 1);
		if (etag) {
			x.setRequestHeader('If-None-Match', etag);
		}
		x.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
		x.onreadystatechange = function () {
			x.readyState > 3 && callback && callback(x.responseText, x);
		};
		x.send(data);
		return x;
	};

	function s4() {
		return Math.floor(
			(1 + Math.random()) * 0x10000
		).toString(16).substring(1);
	}

	var guid = function() {
		return (
			s4() + s4() + '-' + s4() + '-' + s4() + '-' +
			s4() + '-' + s4() + s4() + s4()
		);
	};

	var martd = {};
	martd.SERVER = "http://localhost:54321";
	martd.request = null;
	martd.channels = {};
	martd.cid = guid();
	martd.ever_bumped = false;
	martd.forcing_close = false;

	martd.sub = function(chan, cb, etag) {
		if (!etag) etag = 0;

		var channel = martd.channels[chan]
		if (!channel) {
			channel = {
				name: chan,
				etag: etag,
				callbacks: {}
			}
			martd.channels[chan] = channel;
		}

		var gid = guid();
		channel.callbacks[gid] = cb;

		if (martd.ever_bumped) {
			if (martd.request) {
				martd.request.abort();
				martd.forcing_close = true;
			}
		} else {
			martd.ever_bumped = true;
			bump();
		}

		return {
			gid: gid,
			cancel: function() {
				delete channel.callbacks[gid]
			}
		}
	}

	var bump = function() {
		var url = "/sub?cid=" + martd.cid;
		for (chan in martd.channels) {
			if (Object.keys(martd.channels[chan].callbacks).length > 0) {
				url += ("&" + chan + "=" + martd.channels[chan].etag);
			}
		}
		martd.request = ajax(martd.SERVER + url, false, function (text) {
			martd.request = null;
			try {
				var resp = JSON.parse(text);
				for (chan in resp.channels) {
					if (!martd.channels[chan]) {
						console.log("Unknown channel: ", chan)
						continue;
					}
					for (i in resp.channels[chan].payload) {
						var payload = resp.channels[chan].payload[i];
						for (j in martd.channels[chan].callbacks) {
							try {
								martd.channels[chan].callbacks[j](payload);
							} catch (err) {
								console.log(
									"Callback Error: ", err, payload, chan, j
								);
							}
						}
					}
					martd.channels[chan].etag = resp.channels[chan].etag;
				}
				window.setTimeout(bump, 0);
			} catch (err) {
				if (martd.forcing_close) {
					martd.forcing_close = false;
					window.setTimeout(bump, 0);
				} else {
					if (text) {
						console.log("Error: ", err, text);
					}
					window.setTimeout(bump, 1000);
				}
			}
		});
	}
	return martd;
}();
