# 問題概要
N個の都市とM本の道路がある。道路には信号があり青であれば渡れる。旅行者は指定された順序で都市を訪問したい。
信号の操作と移動の操作を繰り返して全ての都市を訪問せよ。
制約
信号が同時に青になる数は決まっている
信号の操作には特別な制約と方法を用いる
## 信号の操作
信号は配列Aと配列Bによっておこなわれる
配列Aには旅行者は指定した順番で信号のナンバーが重複なしで収められる
配列Bには現在青の信号のナンバーが収められる
旅行者は配列Aから連続するl個（Pa~）の信号を指定して青にすることができる
このとき配列BにはPbからl個に配列Aで指定した信号で更新される
## 配点
信号の操作と移動の操作を最小化する


# Diary
## 1日目
問題文をよむ。
最短経路探索がベースぽそう。配列Aの作り方を考える。
移動ルートの重複は多い方がいい。
訪問する都市間のルートを文字列集合としてみたとき、それが多く含まれる配列Aをつくりたい。
(TODO:過去のAHCでみたことがあるので調べる)
配列Aに含まれなかったルートを再構築したときに、コストが下がるかを考える。
### 低スコアをとることを恐れず実装提出する！

## 2日目
どうスコアの経路をもっておけば切り替えも早い

## 3日目
配列Aが最大２＊Nの長さをもつ。

## 4日目
### 配列Bの更新方法
更新しない部分が数ステップ後に使われる可能性
あるステップの更新方法選定の影響はステップが進むほど小さくなる
    ビームサーチが有効(?)
都市間のルート
    x配列AのどのnextCityをとるか (配列Aに含まれるnextCityの数)
    x何文字（nextCityを含む）（どこからどこまで）とるか
    x配列Bのどこにいれるか(上を含んで何文字更新するか)
### 配列Aの決定法
スコアは全ステップでだしたほうがいい
n-グラムで生成(順番を守る)
配列Aの長さでペアの出現確率から生成
全ての都市が出現するように調整
### 都市間をどのルートで回るか
都市間の移動コストが１なので、同コストの複数ルートが生まれやすい
どのルートを選ぶかでスコアもかわる
### 全ての要素が別の要素に絡む
## 順番
配列Aを決める->ルートx 更新方法を考える
今後Nターンにある都市をできるだけのこす
&&今後Nターンにある都市をできるだけ配列Bにいれる
配列Bと今後のルートで評価をつくれる。
配列Bと複数のルートから最適なもの選べる