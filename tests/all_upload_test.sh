#!/bin/bash

# 현재 디렉토리의 모든 mp4 파일을 반복
for f in *.mp4; do
    [ -e "$f" ] || continue   # mp4 없을 때 에러 방지

    echo "Running upload_test.sh for: $f"

    # 백그라운드 실행
    ./upload_test.sh "$f" &
done

# 모든 백그라운드 작업 대기
wait

echo "All uploads finished."

