# Server Side Events
Server side events is a unidirectionally way to stream notifications/events to clients. It sits on top of TCP/IP and works by keeping a long running connection between client and server through which the server is able to stream event data.

<img src="https://github.com/lucasbflopes/server-side-events-example/blob/master/example.png" />

# Run server

To start the server run:
```sh
go run server/main.go
```

# Run clients

## Web

Open the file `client/index.html` on your browser.

## Others

One can subscribe to the server through http as follows:
```sh
curl -X GET http://localhost:8080/sse/subscribe
```

# Dispatch events

To send events to all clients subscribed to the same server, run:
```sh
curl -X POST http://localhost:8080/sse/dispatch?data=${event-content}
```

# References

- [Mozilla](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
