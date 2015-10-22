// https://gist.github.com/Xeoncross/7663273

window.martd = function() {
	var ajax = function (url, data, callback, etag) {
		var x = new(this.XMLHttpRequest || ActiveXObject)('MSXML2.XMLHTTP.3.0');
		x.open(data ? 'POST' : 'GET', url, 1);
		if (etag) {
			x.setRequestHeader('If-None-Match', etag);
		}
		x.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
		x.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
		x.onreadystatechange = function () {
			x.readyState > 3 && callback && callback(x.responseText, x);
		};
		x.send(data);
		return x;
	};

	var guid = (function() {
		function s4() {
			return Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
		}

		return function() {
			return (
				s4() + s4() + '-' + s4() + '-' + s4() + '-' +
				s4() + '-' + s4() + s4() + s4()
			);
		};
	})();

	var martd = {};
	martd.request = null;
	martd.channels = {};
	martd.cid = guid();

	martd.sub = function(chan, etag, cb) {
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
		bump();

		return {
			gid: gid,
			cancel: function() {
				delete channel.callbacks[gid]
			}
		}
	}

	var bump = function() {
		if (martd.request) {
			martd.request.abort();
		}
		var url = "/sub?cid=" + martd.cid;
		for (chan in martd.channels) {
			url += ("&" + chan + "=" + martd.channels[chan]["etag"]);
		}
		martd.request = ajax(url, false, function (text) {
			martd.request = null;
			try{
				var resp = JSON.parse(text);
				for (chan in resp.channels) {
					martd.channels[chan].etag = resp.channels[chan].etag;
					for (i in resp.channels[chan].payload) {
						var payload = resp.channels[chan].payload[i];
						for (j in martd.channels[chan].callbacks) {
							martd.channels[chan].callbacks[j](payload);
						}
					}
				}
				window.setTimeout(bump, 0);
			} catch (err) {
				console.log("Error: ", err, text)
				window.setTimeout(bump, 1000);
			}
		});
	}
	return martd;
}();
