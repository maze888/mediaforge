#pragma once

#include <rabbitmq-c/amqp.h>
#include <rabbitmq-c/tcp_socket.h>

#define CONSUME_CHANNEL 1

amqp_connection_state_t rabbitmq_connect(const char *req_qname);

int amqp_publish_json(amqp_connection_state_t conn, const char *queue_name, amqp_bytes_t correlation_id, const char *json);

char * amqp_bytes_cstring(amqp_bytes_t amqp_byte);
const char * amqp_error_string_ex(int err);
