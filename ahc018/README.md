# ahc018

kmykさんのリポジトリにあるコンテスト日記が役にたつ
terryさんの記事＋リポジトリもおすすめ


wam @gmeriaog 暫定3位
モンテカルロ法による探索地点決め
#AHC018  #rcl_procon 暫定3位！
斜めの格子上(間隔17)に代表点をとる。
モンテカルロ法により有力な代表点を決めて、有力な代表点を試し掘り。
開通した代表点のみ経由して家から水源まで到達できるならそこを経路にする。
代表点の間を開通させるとき、期待値DPによって計算したパワーを使う。
https://twitter.com/gmeriaog/status/1629801812902674433
https://zenn.dev/gmeriaog/articles/880af6fb3728b5

Psyho Approach for #AHC018
https://twitter.com/FakePsyho/status/1629892674063876097


マラソンマッチ用パラメータチューニングツール
https://github.com/threecourse/marathon-cloud-run-public:q

@_phocom
大域的な推定は逆距離加重風のお気持ちヒューリスティックでなんとかしたけど、局所的(半径8マス程度)にはx+y+x^2+xy+y^2型の曲面で近似できそうと仮定して最小二乗法で重回帰分析してみた
思ってた以上に高精度で当てるからちょっと感動した
https://twitter.com/_phocom/status/1629797064811986944

Yuki Yoshida@yos1up
#AHC018 お疲れ様でした。暫定9位。「学習済固定カーネル使った近傍数固定のガウス過程回帰で強度を推定→楽観的なコストで家から湿地までの最短パスを求め、それ上の最も不確定なマスを少し掘る」を繰り返しました。不確定さが一定以下になったらパスを確定しました。（左：推定強度，右：不確かさ）
https://twitter.com/yos1up/status/1629827983187013632

bowwowforeach @bowwowforeach
#AHC018 やったこと
・家付近の柔い岩を壊す
・最小全域木で家と水のつなぎ方を大まかに決める
・ダイクストラで最短経路を出してその経路付近の岩を攻撃。攻撃点は間隔をあける
・岩へ攻撃した結果で周りの岩の固さを推定。事前に2000テストケースで統計取っておいて利用
・最後に経路接続させる
https://twitter.com/bowwowforeach/status/1629810644001357825

Psyho
Some quick thoughts regarding automatic parameter optimization in heuristic contests (e.g. OPTUNA).

I saw this mentioned a lot during #AHC018. While some things may have been lost in translation, I got the overall impression that people thought it's extremely useful.

It's not.
https://twitter.com/FakePsyho/status/1631275687058219010?s=20
thunder⚡技術書全国販売中

@thun_c
·
Mar 5
#ahc018
psyhoさんのコードをコメントつけながら読んでるんだけど、1個の関数読むだけでもだいぶ疲れるのでわりとしんどみがある
https://twitter.com/thun_c/status/1632370017206673408?s=20


thunder⚡技術書全国販売中
@thun_c
·
Mar 13
#ahc018 
Qiitaに記事を投稿しました！！
僕ではなく1位のPsyhoさんの解法についての解説です。
本人の記事を読んでもよくわからなかったよーって方にオススメです。

AHC018の1位解法(Psyho氏の解法)解説 https://qiita.com/thun-c/items/11af0980cc938dc28d3b #Qiita 
@thun_c
より

@yosu1up
AHC018 ガウス過程回帰を用いた解法
https://docs.google.com/presentation/d/1JEcyHLw8XrDqL4FHUGYIVQC63KSZ2eaHRjO0E2y1WeU/edit?usp=sharing
