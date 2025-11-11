#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>

#include "mp4.h"
#include "rabbitmq.h"

static volatile int running = 1;

void handle_signal(int sig) {
    if (sig == SIGTERM || sig == SIGINT) {
        running = 0;
    }
}

// AMQP_RESPONSE_NONE = 0, /**< the library got an EOF from the socket */
//  538   AMQP_RESPONSE_NORMAL, /**< response normal, the RPC completed successfully */
//  539   AMQP_RESPONSE_LIBRARY_EXCEPTION, /**< library error, an error occurred in the
//  540                                       library, examine the library_error */
//  541   AMQP_RESPONSE_SERVER_EXCEPTION   /**< server exception, the broker returned an
//  542                                       error, check replay */

static const char * amqp_error_string_ex(int err) {
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

int main(int argc, char **argv) {
    // TODO: not yet
    // if (deamon(1, 0) == -1) {
    //     perror("daemon() is failed");
    //     goto out;
    // }
    //
    // signal(SIGPIPE, SIG_IGN);
    // signal(SIGTSTP, SIG_IGN);
    // signal(SIGQUIT, SIG_IGN);
    //

    amqp_connection_state_t conn = rabbitmq_connect();
    if (!conn) {
        fprintf(stderr, "rabbitmq_connect() is failed\n");
        goto out;
    }

    fprintf(stdout, "Waiting for rabbitmq messages...\n");
    while (running) {
        amqp_envelope_t envelope;
        
        amqp_maybe_release_buffers(conn);
        
        amqp_rpc_reply_t reply = amqp_consume_message(conn, &envelope, NULL, 0);
        if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
            // TODO: error message processing
            fprintf(stderr, "error receive consumer message: (%s)\n", amqp_error_string_ex(reply.reply_type));
            goto out;
        }

        printf("Received: %.*s\n",
                (int)envelope.message.body.len,
                (char *)envelope.message.body.bytes);

        amqp_basic_ack(conn, 1, envelope.delivery_tag, 0);
        amqp_destroy_envelope(&envelope);
    }

    fprintf(stdout, "ffmpegd exiting..\n");
    fflush(stdout);
    
    amqp_channel_close(conn, 1, AMQP_REPLY_SUCCESS);
    amqp_connection_close(conn, AMQP_REPLY_SUCCESS);
    amqp_destroy_connection(conn);

    return 0;

out:
    return 1;
}
