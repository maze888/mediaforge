#pragma once

#include <stdio.h>
#include <yyjson.h>

#define PATH_LEN 1024

struct convert_request {
    char job_id[64];
    char file_name[512];
    char input_format[32];
    char output_format[32];
    char upload_path[PATH_LEN];
    char download_path[PATH_LEN];
};

struct convert_response {
    char job_id[64];
    char upload_file_name[PATH_LEN]; // MinIO 에 업로드 된 파일명
    char errmsg[256];
    bool status;
};

int json_unmarshal(char *json, size_t json_len, struct convert_request *req);
yyjson_mut_doc * json_marshal(struct convert_response *res);
