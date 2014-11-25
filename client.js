// https://gist.github.com/Xeoncross/7663273

window.mart = function() {
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

	var mart = {};
	var subs = {}
	mart.cid = guid();
	var etag = "";

	mart.sub = function(chan, etag, cb) {
		subs[chan] = cb;
		$("body").append("<pre>Waiting<pre><hr>");
		// console.log(mart, subs);
		bump();
	}

	var bump = function() {
		ajax("/sub?channel=ch1&cid=" + mart.cid + "&b=" + guid(), false, function (text, req) {
			$("body").append("<pre>" + text + "<pre><hr>");
			// read header: alert(req.getAllResponseHeaders("ETag"));
			// TODO: backoff on errors
			window.setTimeout(bump, 0);
		});
	}
	return mart;
}();
