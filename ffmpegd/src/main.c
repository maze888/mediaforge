#include "mp4.h"

#include <stdio.h>

int main(int argc, char **argv) {
    if (argc < 3) {
        fprintf(stderr, "사용법: %s <입력_MP4_파일> <출력_MP3_파일>\n", argv[0]);
        return 1;
    }

    return extract_mp3(argv[1], argv[2]);
}
