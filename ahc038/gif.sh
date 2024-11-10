#!/bin/bash

# 圧縮対象のファイル
INPUT_FILE="input.gif"
OUTPUT_FILE="output.gif"

# 初期設定
TARGET_SIZE_MB=5        # 目標サイズ (MB)
TARGET_SIZE=$((TARGET_SIZE_MB * 1024 * 1024))  # 目標サイズ (バイト)
CURRENT_SIZE=$(stat -c %s "$INPUT_FILE")  # 初期ファイルサイズ (バイト)
TOTAL_FRAMES=$(ffmpeg -i "$INPUT_FILE" 2>&1 | grep 'frame=' | awk '{print $2}')  # 総フレーム数を取得
FRAME_RATE=$(ffmpeg -i "$INPUT_FILE" 2>&1 | grep 'Stream' | grep -oP ', \K[0-9]+(?= fps)')  # フレームレート取得
CUT_FRAMES=0        # 削るフレーム数の初期値

# 圧縮ループ
while [ "$CURRENT_SIZE" -gt "$TARGET_SIZE" ]; do
    echo "現在のファイルサイズ: $((CURRENT_SIZE / 1024 / 1024)) MB"
    echo "削除するフレーム数: $CUT_FRAMES フレーム"

    # 削るフレーム数を反映してGIFを生成
    ffmpeg -i "$INPUT_FILE" -vf "select='not(mod(n\,$CUT_FRAMES))',setpts=N/($FRAME_RATE*TB)" "$OUTPUT_FILE"

    # 新しいファイルサイズを確認
    CURRENT_SIZE=$(stat -c %s "$OUTPUT_FILE")

    # 次の圧縮のために削るフレーム数を増加
    CUT_FRAMES=$((CUT_FRAMES + 10))

    # 5MB以下になったら終了
    if [ "$CURRENT_SIZE" -le "$TARGET_SIZE" ]; then
        echo "圧縮成功: ファイルサイズが $((CURRENT_SIZE / 1024 / 1024)) MB になりました"
        break
    fi

    # 全フレームを削り切った場合もループを終了
    if [ "$CUT_FRAMES" -ge "$TOTAL_FRAMES" ]; then
        echo "全フレームを削除しました。これ以上削ることはできません。"
        break
    fi
done

echo "最終ファイルサイズ: $((CURRENT_SIZE / 1024 / 1024)) MB"
