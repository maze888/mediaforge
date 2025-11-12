#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>

#include <yyjson.h>

#include "mp4.h"
#include "rabbitmq.h"
#include "parse.h"
#include "http.h"

static volatile int running = 1;

void handle_signal(int sig) {
    if (sig == SIGTERM || sig == SIGINT) {
        running = 0;
    }
}

int main(int argc, char **argv) {
    amqp_connection_state_t conn = rabbitmq_connect();
    if (!conn) {
        fprintf(stderr, "rabbitmq_connect() is failed\n");
        goto out;
    }

    fprintf(stdout, "Waiting for rabbitmq messages...\n");
        
    struct convert_request req;
    while (running) {
        amqp_envelope_t envelope;
        
        amqp_maybe_release_buffers(conn);
        
        amqp_rpc_reply_t reply = amqp_consume_message(conn, &envelope, NULL, 0);
        if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
            // TODO: error message processing
            fprintf(stderr, "error receive consumer message: (%s)\n", amqp_error_string_ex(reply.reply_type));
            goto out;
        }

        // printf("Received: %.*s\n",
        //         (int)envelope.message.body.len,
        //         (char *)envelope.message.body.bytes);

        memset(&req, 0x00, sizeof(req));
        if (json_parse(envelope.message.body.bytes, envelope.message.body.len, &req) < 0) {
            goto fail;
        }

        char local_download_path[1024] = {0};
        snprintf(local_download_path, sizeof(local_download_path) - 1, "./%s.%s", req.file_name, req.input_format);
        if (file_download(req.download_path, local_download_path) < 0) {
            fprintf(stderr, "file_download() is failed");
            goto fail;
        }
        
        char output_path[1024] = {0};
        snprintf(output_path, sizeof(output_path) - 1, "./%s.%s", req.file_name, req.output_format);
        if (extract_mp3(local_download_path, output_path) < 0) {
            goto fail;
        }

        // printf("file_name:     %s\n", req.file_name);
        // printf("input_format:  %s\n", req.input_format);
        // printf("output_format: %s\n", req.output_format);
        // printf("upload_path:   %s\n", req.upload_path);
        // printf("download_path: %s\n", req.download_path);


        // TODO: 업로드
        // TODO: 완료 메시지 퍼블리시

fail:
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
