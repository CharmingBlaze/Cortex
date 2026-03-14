#ifndef CORTEX_RUNTIME_NETWORK_H
#define CORTEX_RUNTIME_NETWORK_H

#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Socket handle: use with tcp_* and udp_* functions. Invalid = -1. */
typedef int cortex_socket_t;

#define CORTEX_SOCKET_INVALID (-1)

/* --- TCP (stream) --- */
/* Listen on port; returns server socket or CORTEX_SOCKET_INVALID. */
cortex_socket_t tcp_listen(int port);
/* Accept one client; returns client socket or CORTEX_SOCKET_INVALID. Blocks. */
cortex_socket_t tcp_accept(cortex_socket_t server);
/* Connect to host:port; returns socket or CORTEX_SOCKET_INVALID. */
cortex_socket_t tcp_connect(const char* host, int port);
/* Send bytes; returns bytes sent or -1 on error. */
int tcp_send(cortex_socket_t sock, const char* data, int len);
/* Receive up to len bytes into buf; returns bytes read, 0 on close, -1 on error. */
int tcp_recv(cortex_socket_t sock, char* buf, int len);
/* Receive up to max_len bytes; returns allocated string (caller frees) or NULL. */
char* tcp_recv_string(cortex_socket_t sock, int max_len);
/* Close socket. */
void tcp_close(cortex_socket_t sock);

/* --- UDP (datagram) --- */
/* Create UDP socket; returns socket or CORTEX_SOCKET_INVALID. */
cortex_socket_t udp_socket(void);
/* Send packet to host:port. Returns bytes sent or -1. */
int udp_send_to(cortex_socket_t sock, const char* host, int port, const char* data, int len);
/* Receive into buf (max len bytes). Optional from_host/from_port (buffers, capacity 256/8). Returns bytes read or -1. */
int udp_recv_from(cortex_socket_t sock, char* buf, int len, char* from_host, int host_cap, int* from_port);
/* Close UDP socket. */
void udp_close(cortex_socket_t sock);

/* --- HTTP client (high-level) --- */
/* GET url; returns response body (caller frees) or NULL on error. */
char* http_get(const char* url);
/* POST url with body; returns response body (caller frees) or NULL on error. */
char* http_post(const char* url, const char* body);
/* GET with custom User-Agent. */
char* http_get_with_header(const char* url, const char* user_agent);

/* --- HTTP server (minimal) --- */
/* Start HTTP server on port. Returns server socket or CORTEX_SOCKET_INVALID. */
cortex_socket_t http_server_listen(int port);
/* Read one HTTP request from client socket; returns request line + headers (caller frees) or NULL. Body read separately. */
char* http_server_read_request(cortex_socket_t client);
/* Send HTTP response (status e.g. "200 OK", body). */
void http_server_send_response(cortex_socket_t client, const char* status, const char* body);

/* --- RPC (JSON over HTTP) --- */
/* POST url with JSON body; returns JSON response (caller frees) or NULL. */
char* rpc_call(const char* url, const char* json_request);

/* --- Multiplayer / real-time helpers --- */
/* Send a length-prefixed message (4-byte big-endian len then payload). Returns true on success. */
bool net_send_message(cortex_socket_t sock, const char* data, int len);
/* Receive length-prefixed message; returns malloc'd buffer (caller frees) and sets *out_len. NULL on error. */
char* net_recv_message(cortex_socket_t sock, int* out_len);
/* Same but ignores length; for Cortex. */
char* net_recv_message_str(cortex_socket_t sock);

#ifdef __cplusplus
}
#endif

#endif /* CORTEX_RUNTIME_NETWORK_H */
