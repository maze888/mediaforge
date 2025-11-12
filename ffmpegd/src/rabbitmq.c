#include "rabbitmq.h"

// TODO: yaml config
amqp_connection_state_t rabbitmq_connect() {
    int rv;
    amqp_connection_state_t conn = NULL;
    amqp_socket_t *socket = NULL;
    amqp_bytes_t queue;
    amqp_rpc_reply_t reply;

    conn = amqp_new_connection();
    if (!conn) {
        fprintf(stderr, "error new connection: %s\n", strerror(errno));
        goto out;
    }

    socket = amqp_tcp_socket_new(conn);
    if (!socket) {
        fprintf(stderr, "error opening TCP socket: %s\n", strerror(errno));
        goto out;
    }

    rv = amqp_socket_open(socket, "localhost", 5672);
    if (rv != AMQP_STATUS_OK) {
        fprintf(stderr, "error opening AMQP socket: %s\n", amqp_error_string2(rv));
        goto out;
    }

    reply = amqp_login(
            conn,
            "/",                    // 기본 가상 호스트
            0,                      // 채널 제한 0 = 서버 기본
            131072,                 // 최대 프레임 크기 128kb
            0,                      // heartbeat 0 = 사용 안함
            AMQP_SASL_METHOD_PLAIN, // username/password 인증
            "admin",
            "admin123");

    amqp_channel_open(conn, 1);
        reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error opening channel\n");
        goto out;
    }

    queue = amqp_cstring_bytes("convert_test");
    amqp_queue_declare(
            conn,
            1, // 채널 번호
            queue,
            0, // passive     : 큐가 없는 경우 (0 = 새로 생성, 1 = 오류 발생)
            1, // durable     : 큐의 영속성 여부 (0 = 비영속, 1 = 영속)
            0, // exlcusive   : 0 = 모든 연결 접근 가능, 1 = 선언한 연결만 접근 가능
            0, // auto_delete : 마지막 컨슈머 떠날시 큐 자동 삭제 (0 = 유지, 1 = 삭제)
            amqp_empty_table
            );
    reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error declare queue: %d\n", reply.reply_type);
        goto out;
    }

    amqp_basic_consume(
            conn,
            1, // 채널명
            queue,
            amqp_empty_bytes,
            0,
            0, // auto ack
            0,
            amqp_empty_table);
    reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error consumer setting\n");
        goto out;
    }

    return conn;

out:
    if (conn) {
        amqp_destroy_connection(conn);
    }

    return NULL;
}

const char * amqp_error_string_ex(int err) {
    switch (err) {
        case AMQP_RESPONSE_NONE:
            return "the library got an EOF from the socket";
        case AMQP_RESPONSE_NORMAL:
            return "response normal, the RPC completed successfully";
        case AMQP_RESPONSE_LIBRARY_EXCEPTION:
            return "library error, an error occurred in the library, examine the library_error";
        case AMQP_RESPONSE_SERVER_EXCEPTION:
            return "server exception, the broker returned an error, check replay";
    }
    return "not applicable(N/A)";
}
