#set -xe

make build

total=0

for RAW_SEED in {0..10}
do
  SEED=$(printf "%04d" $RAW_SEED)
  ./bin/a.out < "tools/in/${SEED}.txt" > "out/${SEED}.txt" 2>/dev/null
  output=$(cd ./tools && cargo run -r --bin vis "in/${SEED}.txt" "../out/${SEED}.txt")

 # 結果から数字のみを抽出
  result=$(echo "$output" | grep -o 'Score = [0-9]*' | grep -o '[0-9]*')
  # resultを数値として合計に加算
  if [[ $result =~ ^[0-9]+$ ]]; then
    total=$((total + result))
  else
    echo "Warning: Non-numeric value encountered in '$output'"
  fi
done

# 最終的な合計を表示
echo "Total sum of all outputs: $total"