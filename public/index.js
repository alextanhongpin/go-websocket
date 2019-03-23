(function () {
  const SECOND = 1000
  const MINUTE = 60 * SECOND
  const pingMessage = JSON.stringify({ type: 'ping' })
  let missedPing = 0
  let pingInterval = null

  function connect (retry = 0, maxRetry = 10) {
    if (retry) {
      console.log(`[websocket] retry $ { retry } times`)
    }
    // TODO: Fetch authorization token for websocket.
    // Enter the test JWT token here.
    const token = ''
    const websocket = new window.WebSocket(`ws://localhost:8080/ws?token=${token}`)
    websocket.onopen = function () {
      console.log('opened')
      // Start ping
      pingInterval = window.setInterval(function () {
        try {
          console.log('pinging server')
          missedPing++
          if (missedPing >= 3) {
            throw new Error('too many missed ping', missedPing)
          }
          websocket.send(pingMessage)
        } catch (error) {
          clearInterval(pingInterval)
          pingInterval = null
          console.warn('closing websocket connection', error.message)
          websocket.close()
        }
      }, jitter(5 * SECOND, 7 * SECOND))
    }

    websocket.onclose = function (evt) {
      window.clearInterval(pingInterval)
      // 1006 means connection is closed abnormally,
      // and this is because we terminated the connection early.
      console.log('[websocket.onclose]:', evt.code, evt.reason)
      if (evt.code === 1002) {
        // Protocol Error.
        console.log('terminating due to invalid request')
        return
      }
      // Perform retry here.
      if (retry < maxRetry) {
        // Attempt to reconnect after 1 second.
        window.setTimeout(() => {
          connect(retry + 1)
        }, jitter())
      }
    }

    websocket.onmessage = function (evt) {
      try {
        const msg = JSON.parse(evt.data)
        console.log('received message:', msg)
        switch (msg.type) {
          case 'ping':
            // Reset the counter.
            missedPing = 0
            break
          case 'authorized':
            console.log('successfully authorized')
            break
          default:
            console.log('unhandled message', msg)
            break
        }
      } catch (error) {
        console.log(error)
      }
    }

    websocket.onerror = function (err) {
      console.log('error', err)
    }

    window.addEventListener('beforeunload', function (e) {
      e.preventDefault()
      e.returnValue = ''
      // Close the websocket connection gracefully.
      websocket.close()
    })
  }

  connect()

  function jitter (min = 10 * SECOND, max = 1 * MINUTE) {
    return min + Math.round(Math.random() * max)
  }
})()
