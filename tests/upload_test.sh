#!/bin/bash
# upload_test.sh
# Usage: ./upload_test.sh /path/to/file.mp4 [server_url]
# Default server_url: http://localhost:5000/upload

# 입력 인자 확인
if [ $# -lt 1 ]; then
    echo "Usage: $0 /path/to/file [server_url]"
    exit 1
fi

FILE_PATH="$1"
SERVER_URL="${2:-http://localhost:5000/convert}"

# 파일 존재 여부 체크
if [ ! -f "$FILE_PATH" ]; then
    echo "Error: File does not exist: $FILE_PATH"
    exit 1
fi

# 변환할 파일 업로드
response=$(curl -v -X POST "$SERVER_URL" \
    -F "JobID=$(uuidgen)" \
    -F "FileName=${FILE_PATH}" \
    -F "InputFormat=mp4" \
    -F "OutputFormat=mp3" \
    -F "files=@${FILE_PATH}" \
    -H "Accept: application/json")

download_url=$(echo "$response" | jq -r ".downloadURL")
download_file_name=$(echo "$response" | jq -r ".downloadFileName")

# echo "downloadURL: $download_url"
# echo "downloadFileName: $download_file_name"

# 변환된 파일 다운로드
curl -L "$download_url" -o "$download_file_name"

