# Networking in Cortex

Cortex’s networking runtime gives you a simple, high-level API for sockets, HTTP, RPC, and real-time multiplayer. It’s implemented in C (`runtime/network.c`) and is linked only when your code uses any of the network builtins.

## Overview

- **TCP** — Reliable streams: servers (`tcp_listen`, `tcp_accept`), clients (`tcp_connect`), send/receive with `tcp_send` / `tcp_recv` or `tcp_recv_string`.
- **UDP** — Datagrams: `udp_socket`, `udp_send_to`, `udp_recv_from` for low-latency or P2P.
- **HTTP** — Client: `http_get`, `http_post`, `http_get_with_header`. Server: `http_server_listen`, `http_server_read_request`, `http_server_send_response`.
- **RPC** — `rpc_call(url, json_request)` for JSON-RPC style calls over HTTP.
- **Multiplayer** — `net_send_message` / `net_recv_message` for length-prefixed messages over TCP.

Socket handles are `int`; use `-1` (or `CORTEX_SOCKET_INVALID`) to check for errors. The compiler emits `#include "runtime/network.h"` only when you use at least one of these functions, and links `network.c` (and on Windows `ws2_32`) only then.

## TCP client example

```c
void main() {
    int sock = tcp_connect("example.com", 80);
    if (sock == -1) {
        writeline("connect failed");
        return;
    }
    tcp_send(sock, "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n", 60);
    string body = tcp_recv_string(sock, 4096);
    tcp_close(sock);
    if (body != "") writeline(body);
}
```

## TCP server example

```c
void main() {
    int server = tcp_listen(8080);
    if (server == -1) {
        writeline("listen failed");
        return;
    }
    writeline("Listening on 8080");
    int client = tcp_accept(server);
    if (client != -1) {
        string msg = tcp_recv_string(client, 1024);
        writeline("Received: " + msg);
        tcp_close(client);
    }
    tcp_close(server);
}
```

## HTTP client example

```c
void main() {
    string body = http_get("https://httpbin.org/get");
    if (body != "") {
        writeline(body);
    }
}
```

## RPC example

```c
void main() {
    string req = "{\"jsonrpc\":\"2.0\",\"method\":\"echo\",\"params\":[\"hi\"],\"id\":1}";
    string res = rpc_call("https://httpbin.org/post", req);
    if (res != "") writeline(res);
}
```

## Multiplayer message framing

For real-time games or apps, use length-prefixed messages so the receiver knows how much to read:

```c
// Sender
net_send_message(client_sock, "Hello", 5);

// Receiver
string msg = net_recv_message(client_sock);
if (msg != "") {
    writeline("Got: " + msg);
}
```

## Best practices

- **Error handling** — Check return values: socket -1, NULL strings, negative send/recv.
- **Cleanup** — Always `tcp_close` / `udp_close` when done; free strings returned by `http_get`, `http_post`, `tcp_recv_string`, `net_recv_message`, `rpc_call` if your runtime expects it (Cortex strings from these are owned by the runtime in the current model).
- **Dedicated servers** — Use `tcp_listen` + loop with `tcp_accept`; spawn or queue work per client.
- **P2P / low latency** — Prefer UDP with `udp_send_to` / `udp_recv_from` and your own protocol.
- **Cloud / HTTP APIs** — Use `http_get` / `http_post` or `rpc_call` for REST/JSON-RPC.

The runtime is cross-platform (Windows Winsock2, POSIX sockets elsewhere). For WebSockets you can use an external C library and declare it with `extern` and `#include`.
