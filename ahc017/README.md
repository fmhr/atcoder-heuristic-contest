# ahc017

---
1/28/2023
問題文を読む
最初の提出(毎日上限いっぱいに工事を行う)
2点間が繋がれていないときのペナルティが大きいので全域木が崩れないのが必要条件
    *これが避けられないケースの存在を探す
    実装方針
        Deleteができるunion find
            online dynamic connectivity
            offline dynamic connectivity
                消す順番を決めていればundoで対応できる
        ある辺を消したときの最短代替ルートを保持しておいて　同時に消さないようにする

初日に消した辺は翌日以降消さなくていいのでビームサーチを探索に使える
試行数では厳密なスコア計算ができる可能性がある
日数が比較的小さく

1/29/2023
いじってない状態での最短距離を理想とする

問題文:「このとき、k 日目の工事に対する不満度は、以下のようにランダムな異なる二点間の最短距離の増加量の期待値として定義される。」
ソースコードから　各頂点とほかの全ての頂点との距離の増加量
    矛盾してなかったけど、問題文が無駄では？
→使用回数の多い辺は削除されたときの増加量が回数に比例する
    毎回全経路計算　or ある辺の使用回数をメモしておく
    すべての辺が工事されるのでいつ工事しても変わらないのでは？
BFSで迂回路を検索できる

1/30/2023
使えないエッジの更新を日毎に更新するのは難しい
    全点間の距離と使用辺を記録しておく
    辺から使用された区間を逆参照できるようにしておく
    更新される2点間の距離だけ更新する

どうして点の座標が与えられてるのか？
    使って欲しい？
    方向性のA*探索？

1/31/2023
ランダムで道を選んでその日に工事する橋を決めるビームサーチ　
工事する橋で再探索するルートをだす
A*　それぞれ探索で検索　or 枝切り全探索
    wfでkチタンがs-t間にあるかを平面で判断
    ijkのループなのでijが再構築に必要　kが中間地点のすべて