#include "parse.h"

int json_parse(char *json, size_t json_len, struct convert_request *req) {
    if (!json || !req) {
        fprintf(stderr, "invalid argument: (json: %p, req: %p)", json, req);
        goto out;
    }
    
    yyjson_read_err err;
    yyjson_doc *doc = yyjson_read_opts(json, json_len, 0, NULL, &err);
    if (!doc) {
        fprintf(stderr, "JSON parse error (msg: %s , pos: %zu , code: %u)\n%.*s", err.msg, err.pos, err.code, (int)json_len, json);
        goto out;
    }

    yyjson_val *root = yyjson_doc_get_root(doc);
    if (!root) {
        fprintf(stderr, "JSON parse error (root attr)");
        goto out;
    }
    yyjson_val *file_name = yyjson_obj_get(root, "FileName");
    if (!file_name) {
        fprintf(stderr, "JSON parse error (FileName attr)");
        goto out;
    }
    yyjson_val *input_format = yyjson_obj_get(root, "InputFormat");
    if (!input_format) {
        fprintf(stderr, "JSON parse error (InputFormat attr)");
        goto out;
    }
    yyjson_val *output_format = yyjson_obj_get(root, "OutputFormat");
    if (!output_format) {
        fprintf(stderr, "JSON parse error (OutputFormat attr)");
        goto out;
    }
    yyjson_val *presigned_upload_url = yyjson_obj_get(root, "PresignedUploadURL");
    if (!presigned_upload_url) {
        fprintf(stderr, "JSON parse error (PresignedUploadURL attr)");
        goto out;
    }
    yyjson_val *presigned_download_url = yyjson_obj_get(root, "PresignedDownloadURL");
    if (!presigned_download_url) {
        fprintf(stderr, "JSON parse error (PresignedDownloadURL attr)");
        goto out;
    }

    snprintf(req->file_name, sizeof(req->file_name), "%s", yyjson_get_str(file_name));
    snprintf(req->input_format, sizeof(req->input_format), "%s", yyjson_get_str(input_format));
    snprintf(req->output_format, sizeof(req->output_format), "%s", yyjson_get_str(output_format));
    snprintf(req->upload_path, sizeof(req->upload_path), "%s", yyjson_get_str(presigned_upload_url));
    snprintf(req->download_path, sizeof(req->download_path), "%s", yyjson_get_str(presigned_download_url));
    
    yyjson_doc_free(doc);
    
    return 0;
    
out:
    if (doc) yyjson_doc_free(doc);

    return -1;
}
