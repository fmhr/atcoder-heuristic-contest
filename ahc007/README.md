# AHC007-4hour
AtCoder Heuristic Contest 007

https://atcoder.jp/contests/ahc007

# 問題　Online MST
１つずつ辺の真の長さが与えられるので、それをつかうかを決める。

点の位置から他の辺の長さを推測して最小木をつくる

# コンテスト
座標から点間の距離を2dとしてMSTをつくる（50分）

残り時間はバグでつぶす

順位 338位

# 上位解法

# コンテスト中にやったバグ
sliceを引数にしたとき、関数の中でsortしたら、そとのsliceもsortされる
- #golang ではarray引数の場合は配列全体の値コピーを伴い、slice引数の場合はarray参照が渡されます。
つまり、arrayは値、sliceは参照型のように振る舞います。
（sliceは値型のメタ情報とarrayポインタをもつハイブリッドな値）https://twitter.com/nobonobo/status/1470722646228992000

state（union find)で不必要なコピーをしていた
- 解決策　しない

そういえばStateはよくコピーするのでスライスを使わないようにしてたね
