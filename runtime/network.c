/**
 * Cortex networking runtime: TCP, UDP, HTTP client/server, RPC, message helpers.
 * Cross-platform: Windows (Winsock2) and POSIX (BSD sockets).
 */
#include "network.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#if defined(_WIN32) || defined(_WIN64)
#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
#include <ws2tcpip.h>
#pragma comment(lib, "Ws2_32.lib")
typedef int socklen_t;
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <unistd.h>
#define SOCKET int
#define INVALID_SOCKET (-1)
#define SOCKET_ERROR (-1)
#define closesocket(s) close(s)
#endif

static int g_network_init = 0;

static void network_init(void) {
    if (g_network_init) return;
#if defined(_WIN32) || defined(_WIN64)
    WSADATA wsa;
    if (WSAStartup(MAKEWORD(2, 2), &wsa) == 0)
        g_network_init = 1;
#else
    g_network_init = 1;
#endif
}

static cortex_socket_t sock_from(SOCKET s) {
    if (s == INVALID_SOCKET) return CORTEX_SOCKET_INVALID;
    return (cortex_socket_t)(intptr_t)s;
}

static SOCKET sock_to(cortex_socket_t h) {
    if (h == CORTEX_SOCKET_INVALID) return INVALID_SOCKET;
    return (SOCKET)(intptr_t)h;
}

/* --- TCP --- */
cortex_socket_t tcp_listen(int port) {
    network_init();
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (s == INVALID_SOCKET) return CORTEX_SOCKET_INVALID;
    int one = 1;
    setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (const char*)&one, sizeof(one));
    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons((unsigned short)port);
    if (bind(s, (struct sockaddr*)&addr, sizeof(addr)) != 0) {
        closesocket(s);
        return CORTEX_SOCKET_INVALID;
    }
    if (listen(s, 5) != 0) {
        closesocket(s);
        return CORTEX_SOCKET_INVALID;
    }
    return sock_from(s);
}

cortex_socket_t tcp_accept(cortex_socket_t server) {
    SOCKET s = sock_to(server);
    if (s == INVALID_SOCKET) return CORTEX_SOCKET_INVALID;
    struct sockaddr_in client_addr;
    socklen_t len = sizeof(client_addr);
    SOCKET client = accept(s, (struct sockaddr*)&client_addr, &len);
    if (client == INVALID_SOCKET) return CORTEX_SOCKET_INVALID;
    return sock_from(client);
}

cortex_socket_t tcp_connect(const char* host, int port) {
    network_init();
    struct addrinfo hints, *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;
    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);
    if (getaddrinfo(host, port_str, &hints, &res) != 0) return CORTEX_SOCKET_INVALID;
    SOCKET s = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
    if (s == INVALID_SOCKET) { freeaddrinfo(res); return CORTEX_SOCKET_INVALID; }
    int ok = connect(s, res->ai_addr, (socklen_t)res->ai_addrlen);
    freeaddrinfo(res);
    if (ok != 0) { closesocket(s); return CORTEX_SOCKET_INVALID; }
    return sock_from(s);
}

int tcp_send(cortex_socket_t sock, const char* data, int len) {
    if (!data || len <= 0) return -1;
    int sent = (int)send(sock_to(sock), data, (size_t)len, 0);
    return (sent >= 0) ? sent : -1;
}

int tcp_recv(cortex_socket_t sock, char* buf, int len) {
    if (!buf || len <= 0) return -1;
    int n = (int)recv(sock_to(sock), buf, (size_t)len, 0);
    if (n > 0) return n;
    if (n == 0) return 0; /* closed */
    return -1;
}

char* tcp_recv_string(cortex_socket_t sock, int max_len) {
    if (max_len <= 0 || max_len > 1024 * 1024) return NULL;
    char* buf = (char*)malloc((size_t)max_len + 1);
    if (!buf) return NULL;
    int n = tcp_recv(sock, buf, max_len);
    if (n < 0) { free(buf); return NULL; }
    buf[n] = '\0';
    return buf;
}

void tcp_close(cortex_socket_t sock) {
    SOCKET s = sock_to(sock);
    if (s != INVALID_SOCKET) {
        shutdown(s, 2);
        closesocket(s);
    }
}

/* --- UDP --- */
cortex_socket_t udp_socket(void) {
    network_init();
    SOCKET s = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (s == INVALID_SOCKET) return CORTEX_SOCKET_INVALID;
    return sock_from(s);
}

int udp_send_to(cortex_socket_t sock, const char* host, int port, const char* data, int len) {
    if (!host || !data || len < 0) return -1;
    struct addrinfo hints, *res = NULL;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_DGRAM;
    char port_str[16];
    snprintf(port_str, sizeof(port_str), "%d", port);
    if (getaddrinfo(host, port_str, &hints, &res) != 0) return -1;
    int n = (int)sendto(sock_to(sock), data, (size_t)len, 0, res->ai_addr, (socklen_t)res->ai_addrlen);
    freeaddrinfo(res);
    return (n >= 0) ? n : -1;
}

int udp_recv_from(cortex_socket_t sock, char* buf, int len, char* from_host, int host_cap, int* from_port) {
    if (!buf || len <= 0) return -1;
    struct sockaddr_in from;
    socklen_t fromlen = sizeof(from);
    int n = (int)recvfrom(sock_to(sock), buf, (size_t)len, 0, (struct sockaddr*)&from, &fromlen);
    if (n < 0) return -1;
    if (from_host && host_cap > 0) {
        inet_ntop(AF_INET, &from.sin_addr, from_host, (size_t)host_cap);
        from_host[host_cap - 1] = '\0';
    }
    if (from_port) *from_port = (int)ntohs(from.sin_port);
    return n;
}

void udp_close(cortex_socket_t sock) {
    SOCKET s = sock_to(sock);
    if (s != INVALID_SOCKET) closesocket(s);
}

/* --- HTTP client: parse URL and do GET/POST --- */
static int parse_url(const char* url, char* host, int host_len, int* port, char* path, int path_len) {
    if (!url || !host || !port || !path) return -1;
    *port = 80;
    if (strncmp(url, "https://", 8) == 0) { url += 8; *port = 443; }
    else if (strncmp(url, "http://", 7) == 0) url += 7;
    const char* slash = strchr(url, '/');
    const char* colon = strchr(url, ':');
    if (colon && (!slash || colon < slash)) {
        *port = atoi(colon + 1);
        size_t host_part = (size_t)(colon - url);
        if (host_part >= (size_t)host_len) return -1;
        memcpy(host, url, host_part);
        host[host_part] = '\0';
        url = slash ? slash : url + strlen(url);
    } else if (slash) {
        size_t host_part = (size_t)(slash - url);
        if (host_part >= (size_t)host_len) return -1;
        memcpy(host, url, host_part);
        host[host_part] = '\0';
        url = slash;
    } else {
        size_t L = strlen(url);
        if (L >= (size_t)host_len) return -1;
        memcpy(host, url, L + 1);
        url = "";
    }
    if (path_len > 0) {
        if (*url == '\0') { path[0] = '/'; path[1] = '\0'; }
        else {
            size_t pl = strlen(url);
            if (pl >= (size_t)path_len) return -1;
            memcpy(path, url, pl + 1);
        }
    }
    return 0;
}

static char* http_request(const char* method, const char* url, const char* body, const char* extra_header) {
    char host[256], path[1024];
    int port = 80;
    if (parse_url(url, host, sizeof(host), &port, path, sizeof(path)) != 0) return NULL;
    cortex_socket_t sock = tcp_connect(host, port);
    if (sock == CORTEX_SOCKET_INVALID) return NULL;
    char req[4096];
    int n = snprintf(req, sizeof(req),
        "%s %s HTTP/1.1\r\n"
        "Host: %s\r\n"
        "Connection: close\r\n",
        method, path[0] ? path : "/", host);
    if (extra_header) n += snprintf(req + n, sizeof(req) - (size_t)n, "%s\r\n", extra_header);
    if (body) n += snprintf(req + n, sizeof(req) - (size_t)n, "Content-Length: %zu\r\n", strlen(body));
    n += snprintf(req + n, sizeof(req) - (size_t)n, "\r\n");
    if (tcp_send(sock, req, (int)strlen(req)) < 0) { tcp_close(sock); return NULL; }
    if (body && tcp_send(sock, body, (int)strlen(body)) < 0) { tcp_close(sock); return NULL; }
    size_t cap = 4096, len = 0;
    char* out = (char*)malloc(cap);
    if (!out) { tcp_close(sock); return NULL; }
    for (;;) {
        char buf[1024];
        int r = tcp_recv(sock, buf, sizeof(buf));
        if (r <= 0) break;
        while (len + (size_t)r >= cap) { cap *= 2; char* t = (char*)realloc(out, cap); if (!t) { free(out); tcp_close(sock); return NULL; } out = t; }
        memcpy(out + len, buf, (size_t)r);
        len += (size_t)r;
    }
    tcp_close(sock);
    out[len] = '\0';
    /* Skip headers to return body only */
    char* body_start = strstr(out, "\r\n\r\n");
    if (body_start) {
        body_start += 4;
        size_t body_len = len - (size_t)(body_start - out);
        char* body_only = (char*)malloc(body_len + 1);
        if (body_only) {
            memcpy(body_only, body_start, body_len + 1);
            free(out);
            return body_only;
        }
    }
    return out;
}

char* http_get(const char* url) {
    return http_request("GET", url, NULL, NULL);
}

char* http_post(const char* url, const char* body) {
    return http_request("POST", url, body, "Content-Type: application/x-www-form-urlencoded\r\n");
}

char* http_get_with_header(const char* url, const char* user_agent) {
    char hdr[512];
    if (user_agent) snprintf(hdr, sizeof(hdr), "User-Agent: %s\r\n", user_agent);
    else hdr[0] = '\0';
    return http_request("GET", url, NULL, hdr[0] ? hdr : NULL);
}

/* --- HTTP server --- */
cortex_socket_t http_server_listen(int port) {
    return tcp_listen(port);
}

char* http_server_read_request(cortex_socket_t client) {
    size_t cap = 4096, len = 0;
    char* out = (char*)malloc(cap);
    if (!out) return NULL;
    for (;;) {
        if (len + 5 > cap) { cap *= 2; char* t = (char*)realloc(out, cap); if (!t) { free(out); return NULL; } out = t; }
        int r = tcp_recv(client, out + len, (int)(cap - len - 1));
        if (r <= 0) { free(out); return NULL; }
        len += (size_t)r;
        out[len] = '\0';
        if (strstr(out, "\r\n\r\n")) break;
    }
    return out;
}

void http_server_send_response(cortex_socket_t client, const char* status, const char* body) {
    char hdr[256];
    int body_len = body ? (int)strlen(body) : 0;
    snprintf(hdr, sizeof(hdr), "HTTP/1.1 %s\r\nConnection: close\r\nContent-Length: %d\r\n\r\n", status ? status : "200 OK", body_len);
    tcp_send(client, hdr, (int)strlen(hdr));
    if (body && body_len > 0) tcp_send(client, body, body_len);
    tcp_close(client);
}

/* --- RPC --- */
char* rpc_call(const char* url, const char* json_request) {
    char hdr[128];
    snprintf(hdr, sizeof(hdr), "Content-Type: application/json\r\n");
    char* req = (char*)malloc(strlen("POST ") + strlen(url) + 256 + (json_request ? strlen(json_request) : 0) + 64);
    if (!req) return NULL;
    (void)req;
    return http_request("POST", url, json_request, "Content-Type: application/json\r\n");
}

/* --- Message framing (4-byte big-endian length + payload) --- */
bool net_send_message(cortex_socket_t sock, const char* data, int len) {
    if (!data || len < 0) return false;
    unsigned char len_buf[4];
    len_buf[0] = (unsigned char)(len >> 24);
    len_buf[1] = (unsigned char)(len >> 16);
    len_buf[2] = (unsigned char)(len >> 8);
    len_buf[3] = (unsigned char)len;
    if (tcp_send(sock, (const char*)len_buf, 4) != 4) return false;
    if (len > 0 && tcp_send(sock, data, len) != len) return false;
    return true;
}

char* net_recv_message(cortex_socket_t sock, int* out_len) {
    if (!out_len) return NULL;
    *out_len = 0;
    unsigned char len_buf[4];
    if (tcp_recv(sock, (char*)len_buf, 4) != 4) return NULL;
    int len = (int)((len_buf[0] << 24) | (len_buf[1] << 16) | (len_buf[2] << 8) | len_buf[3]);
    if (len < 0 || len > 1024 * 1024) return NULL;
    char* buf = (char*)malloc(len + 1);
    if (!buf) return NULL;
    int got = 0;
    while (got < len) {
        int r = tcp_recv(sock, buf + got, len - got);
        if (r <= 0) { free(buf); return NULL; }
        got += r;
    }
    buf[len] = '\0';
    *out_len = len;
    return buf;
}

char* net_recv_message_str(cortex_socket_t sock) {
    int dummy;
    return net_recv_message(sock, &dummy);
}
