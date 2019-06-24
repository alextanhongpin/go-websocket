# golang websocket

Designing the structure for scalable websocket in golang. This is actually a modification of the [chat example](https://github.com/gorilla/websocket/tree/master/examples/chat) from the `gorilla/websocket` repo.

There are some improvements that I want to add to the example:

- authentication/authorization using jwt token
- access control list for users (users can only send to their friends, for example)
- the same users can connect to multiple websockets
- handling for different message type


## Project

- [go-chat](https://github.com/alextanhongpin/go-chat.git)
