var events = require('events')
var msgpack = require('msgpack-lite')

function print (msg) {
  console.log('[ws]: ' + msg)
}

class SimpleWebSocket extends events.EventEmitter {
  constructor (url) {
    super()
  }

  connect (addr) {
    if (!addr) {
      addr = 'ws://' + window.location.host + '/ws'
    }

    var socket = this
    var ws = new WebSocket(addr)
    // Set up binary type for msgpack ease-of-use
    ws.binaryType = 'arraybuffer'

    ws.onopen = function (evt) {
      print('OPEN')
      if (socket.onopen) {
        socket.onopen(evt)
      }
    }
    ws.onclose = function (evt) {
      print('CLOSE')
      if (socket.onclose) {
        socket.onclose(evt)
      }
    }
    ws.onmessage = function (evt) {
      var rawBinary = new Uint8Array(evt.data)
      var msg = msgpack.decode(rawBinary)

      print('Message: ' + msg.data)
      socket.emit(msg.event, msg.data)
      if (socket.onmessage) {
        socket.onmessage(evt)
      }
    }
    ws.onerror = function (evt) {
      print('ERROR: ' + evt.data)
      if (socket.onerror) {
        socket.onerror(evt)
      }
    }
    socket._websocket = ws
  }
}

window.SimpleWebSocket = SimpleWebSocket
module.exports = SimpleWebSocket
