#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include "rabbitmq.h"
#include "util.h"


// TODO: yaml config
amqp_connection_state_t rabbitmq_connect(const char *req_qname) {
    int rv;
    amqp_connection_state_t conn = NULL;
    amqp_socket_t *socket = NULL;
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

    amqp_channel_open(conn, CONSUME_CHANNEL);
    reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error opening CONSUME channel: (errmsg: %s errno: %d)\n", amqp_error_string_ex(reply.reply_type), reply.reply_type);
        goto out;
    }
    
    // consume queue
    amqp_queue_declare(
            conn,
            CONSUME_CHANNEL, // 채널 번호
            amqp_cstring_bytes(req_qname),
            0,               // passive     : 큐가 없는 경우 (0 = 새로 생성, 1 = 오류 발생)
            1,               // durable     : 큐의 영속성 여부 (0 = 비영속, 1 = 영속)
            0,               // exlcusive   : 0 = 모든 연결 접근 가능, 1 = 선언한 연결만 접근 가능
            0,               // auto_delete : 마지막 컨슈머 떠날시 큐 자동 삭제 (0 = 유지, 1 = 삭제)
            amqp_empty_table
            );
    reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error declare consume queue: (errmsg: %s errno: %d)\n", amqp_error_string_ex(reply.reply_type), reply.reply_type);
        goto out;
    }

    // consume setting
    amqp_basic_consume(
            conn,
            CONSUME_CHANNEL, // 채널 번호
            amqp_cstring_bytes(req_qname),
            amqp_empty_bytes,
            0,
            0, // auto ack
            0,
            amqp_empty_table);
    reply = amqp_get_rpc_reply(conn);
    if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
        fprintf(stderr, "error consumer setting: (errmsg: %s errno: %d)\n", amqp_error_string_ex(reply.reply_type), reply.reply_type);
        goto out;
    }
    
    return conn;

out:
    if (conn) {
        amqp_destroy_connection(conn);
    }

    return NULL;
}

int amqp_publish_json(amqp_connection_state_t conn, const char *queue_name, amqp_bytes_t correlation_id, const char *json) {
    if (!queue_name || !json) {
        fprintf(stderr, "invalid argument (queue_name: %p, json: %p)\n", CKNUL(queue_name), CKNUL(json));
        return -1;
    }
    
    amqp_basic_properties_t props;

    props._flags = 
        AMQP_BASIC_CONTENT_TYPE_FLAG |
        AMQP_BASIC_MESSAGE_ID_FLAG |
        AMQP_BASIC_CORRELATION_ID_FLAG |
        AMQP_BASIC_TIMESTAMP_FLAG |
        AMQP_BASIC_DELIVERY_MODE_FLAG;
    props.content_type = amqp_cstring_bytes("application/json");
    props.delivery_mode = 2; // 2 = persistent
    props.correlation_id = correlation_id;
    props.timestamp = time(NULL);
    
    char uuid[64] = {0};
    generate_uuid(uuid);
    props.message_id = amqp_cstring_bytes(uuid);

    int rv = amqp_basic_publish(
            conn,
            1,
            amqp_cstring_bytes(""), // default exchange (direct)
            amqp_cstring_bytes(queue_name),
            0,
            0,
            &props,
            amqp_cstring_bytes(json)
            );
    if (rv != AMQP_STATUS_OK) {
        fprintf(stderr, "amqp_basic_publish() is failed: (errmsg: %s errno: %d)\n", amqp_error_string2(rv), rv);
        return -1;
    }

    return 0;
}

char * amqp_bytes_cstring(amqp_bytes_t amqp_byte) {
    char *p = calloc(1, amqp_byte.len + 1);
    if (!p) {
        fprintf(stderr, "calloc() is failed: (errmsg: %s, errno: %d)\n", strerror(errno), errno);
        exit(1);
    }

    memcpy(p, amqp_byte.bytes, amqp_byte.len);

    return p;
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
