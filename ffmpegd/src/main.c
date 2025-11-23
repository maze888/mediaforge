#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>

#include <yyjson.h>

#include "mp4.h"
#include "rabbitmq.h"
#include "marshal.h"
#include "http.h"
#include "util.h"

static volatile int running = 1;

void handle_signal(int sig) {
    if (sig == SIGTERM || sig == SIGINT) {
        running = 0;
    }
}

int main(int argc, char **argv) {
    curl_global_init(CURL_GLOBAL_DEFAULT);

    int rv;
    const char *req_qname = "convert_request";

    amqp_connection_state_t conn = rabbitmq_connect(req_qname);
    if (!conn) {
        fprintf(stderr, "rabbitmq_connect() is failed\n");
        goto out;
    }

    fprintf(stdout, "Waiting for rabbitmq messages...\n");

    struct convert_request req;
    struct convert_response res;
    amqp_envelope_t envelope;
    while (running) {
        amqp_maybe_release_buffers(conn);

        memset(&req, 0x00, sizeof(req));
        memset(&res, 0x00, sizeof(res));
        memset(&envelope, 0x00, sizeof(envelope));

        // 작업 요청 수신
        amqp_rpc_reply_t reply = amqp_consume_message(conn, &envelope, NULL, 0);
        if (reply.reply_type != AMQP_RESPONSE_NORMAL) {
            fprintf(stderr, "error receive consumer message: (%s)\n", amqp_error_string_ex(reply.reply_type));
            goto fail;
        }
        
        // printf("Received: %.*s\n",
        //         (int)envelope.message.body.len,
        //         (char *)envelope.message.body.bytes);
        
        // check ReplyToQueue
        if (envelope.message.properties.reply_to.len == 0) {
            fprintf(stderr, "no setting reply queue\n");
            goto fail;
        }
        
        // 작업 요청 메시지(JSON) 구조체에 바인딩
        memset(&req, 0x00, sizeof(req));
        if (json_unmarshal(envelope.message.body.bytes, envelope.message.body.len, &req) < 0) {
            goto fail;
        }

        // printf("job_id:        %s\n", req.job_id);
        // printf("file_name:     %s\n", req.file_name);
        // printf("input_format:  %s\n", req.input_format);
        // printf("output_format: %s\n", req.output_format);
        // printf("upload_path:   %s\n", req.upload_path);
        // printf("download_path: %s\n", req.download_path);
        
        // MinIO 에서 작업 파일 다운로드
        char local_download_path[1024] = {0};
        snprintf(local_download_path, sizeof(local_download_path) - 1, "./%s", req.file_name);
        if (http_file_download(req.download_path, local_download_path) < 0) {
            fprintf(stderr, "http_file_download() is failed\n");
            goto fail;
        }

        // mp4->mp3 추출 및 로컬 저장
        char output_path[1024] = {0};
        snprintf(output_path, sizeof(output_path) - 1, "./%s.%s", req.file_name, req.output_format);
        if (extract_mp3(local_download_path, output_path) < 0) {
            unlink(local_download_path);
            goto fail;
        }
        unlink(local_download_path);

        // MinIO 에 작업된 파일 업로드
        if (http_file_upload(req.upload_path, output_path) < 0) {
            fprintf(stderr, "http_file_upload() is failed\n");
            goto fail;
        }
        unlink(output_path);

        // 작업 완료 메시지 작성
        snprintf(res.job_id, sizeof(res.job_id), "%s", req.job_id);
        res.status = true;

        // 작업 완료 메시지 JSON 으로 마샬링
        yyjson_mut_doc *json_doc = json_marshal(&res);
        if (!json_doc) {
            fprintf(stderr, "json_marshal() is failed");
            goto fail;
        }

        // 작업 완료 메시지 전송
        char *reply_to_queue = amqp_bytes_cstring(envelope.message.properties.reply_to);
        if (amqp_publish_json(conn, reply_to_queue, envelope.message.properties.correlation_id, yyjson_mut_write(json_doc, 0, NULL)) < 0) {
            fprintf(stderr, "amqp_basic_publish() is failed\n");
            safe_free(reply_to_queue);
            goto fail;
        }
        safe_free(reply_to_queue);

fail:
        // TODO: 작업 fail 메시지 전송 처리
        // TODO: ack 에러 처리
        amqp_basic_ack(conn, CONSUME_CHANNEL, envelope.delivery_tag, 0);
        amqp_destroy_envelope(&envelope);
    }

    fprintf(stdout, "ffmpegd exiting..\n");
    fflush(stdout);

    amqp_channel_close(conn, CONSUME_CHANNEL, AMQP_REPLY_SUCCESS);
    amqp_connection_close(conn, AMQP_REPLY_SUCCESS);
    amqp_destroy_connection(conn);

    curl_global_cleanup();

    return 0;

out:
    if (conn) amqp_destroy_connection(conn);
    curl_global_cleanup();

    return 1;
}
