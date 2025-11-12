#pragma once

#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <rabbitmq-c/amqp.h>
#include <rabbitmq-c/tcp_socket.h>

amqp_connection_state_t rabbitmq_connect();

const char * amqp_error_string_ex(int err);
