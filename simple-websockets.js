(function() {
	var
	WS_EVENT_KEY             = "_ws_event",
	WS_TYPE_KEY              = "_ws_type",
	WS_MESSAGE_TYPE_STRING   = "_ws_msg_type_string",
	WS_MESSAGE_TYPE_JSON      = "_ws_msg_type_json",
	WS_MESSAGE_TYPE_ABSTRACT = "_ws_msg_type_abstract";

	function print(msg) {
		console.log('[ws]: ' + msg);
	}

	function connect(addr) {
		if(!addr) {
			addr = 'ws://' + window.location.host + "/ws";
		}
		var ret = {};
		var ws = new WebSocket(addr);
		ret._websocket = ws;
		ws.onopen = function(evt) {
			print("OPEN");
		}
		ws.onclose = function(evt) {
			print("CLOSE");
			ws = null;
		}
		ws.onmessage = function(evt) {
			print("Received message: " + evt.data);
			var json = JSON.parse(evt.data);
			var data;
			if(json[WS_TYPE_KEY]) {
				data = json[json[WS_TYPE_KEY]];
			} else {
				data = json;
			}
			var evt = json[WS_EVENT_KEY];
			if(!evt) {
				// Do nothing
				return;
			}
			print('Message: ' + data);
			$(ret).trigger(evt, data);
		}
		ws.onerror = function(evt) {
			print("ERROR: " + evt.data);
		}

		ret.on = function(evt, cb) {
			$(ret).on(evt, function(e, data) {
				cb(data);
			});
		}

		ret.emit = function(evt, data) {
			if(!data) {
				data = {};
			}
			if(typeof data === 'object') {
				data = JSON.parse(JSON.stringify(data));
			}
			data[WS_EVENT_KEY] = evt;
			data = JSON.stringify(data);
			ret._websocket.send(data);
		}
		return ret;
	}

	window.SimpleWebSocket = {
		connect: connect,
	}
})();
