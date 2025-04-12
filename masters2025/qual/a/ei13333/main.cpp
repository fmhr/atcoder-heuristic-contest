#line 1 "template/template.hpp"
#include <bits/stdc++.h>
#if __has_include(<atcoder/all>)
#include <atcoder/all>
#endif

using namespace std;

using int64 = long long;

const int64 infll = (1LL << 62) - 1;
const int inf = (1 << 30) - 1;

struct IoSetup {
  IoSetup() {
    cin.tie(nullptr);
    ios::sync_with_stdio(false);
    cout << fixed << setprecision(10);
    cerr << fixed << setprecision(10);
  }
} iosetup;

template<typename T1, typename T2>
ostream &operator<<(ostream &os, const pair<T1, T2> &p) {
  os << p.first << " " << p.second;
  return os;
}

template<typename T1, typename T2>
istream &operator>>(istream &is, pair<T1, T2> &p) {
  is >> p.first >> p.second;
  return is;
}

template<typename T>
ostream &operator<<(ostream &os, const vector<T> &v) {
  for (int i = 0; i < (int) v.size(); i++) {
    os << v[i] << (i + 1 != v.size() ? " " : "");
  }
  return os;
}

template<typename T>
istream &operator>>(istream &is, vector<T> &v) {
  for (T &in : v) is >> in;
  return is;
}

template<typename T1, typename T2>
inline bool chmax(T1 &a, T2 b) {
  return a < b && (a = b, true);
}

template<typename T1, typename T2>
inline bool chmin(T1 &a, T2 b) {
  return a > b && (a = b, true);
}

template<typename T = int64>
vector<T> make_v(size_t a) {
  return vector<T>(a);
}

template<typename T, typename... Ts>
auto make_v(size_t a, Ts... ts) {
  return vector<decltype(make_v<T>(ts...))>(a, make_v<T>(ts...));
}

template<typename T, typename V>
typename enable_if<is_class<T>::value == 0>::type fill_v(T &t, const V &v) {
  t = v;
}

template<typename T, typename V>
typename enable_if<is_class<T>::value != 0>::type fill_v(T &t, const V &v) {
  for (auto &e : t) fill_v(e, v);
}

template<typename F>
struct FixPoint : F {
  explicit FixPoint(F &&f) : F(std::forward<F>(f)) {}

  template<typename... Args>
  decltype(auto) operator()(Args &&...args) const {
    return F::operator()(*this, std::forward<Args>(args)...);
  }
};

template<typename F>
inline decltype(auto) MFP(F &&f) {
  return FixPoint<F>{std::forward<F>(f)};
}


int N, M;
int ax, ay;

struct State {
  vector<string> C; // 盤面
  int y, x; // あなたの場所
  vector< pair< int, int > > op; // 操作列
  int score; // スコア
  bool operator<(const State& s) const {
    return score < s.score;
  }
};


#line 2 "structure/union-find/union-find.hpp"

struct UnionFind {
  vector<int> data;

  UnionFind() = default;

  explicit UnionFind(size_t sz) : data(sz, -1) {}

  bool unite(int x, int y) {
    x = find(x), y = find(y);
    if (x == y) return false;
    if (data[x] > data[y]) swap(x, y);
    data[x] += data[y];
    data[y] = x;
    return true;
  }

  int find(int k) {
    if (data[k] < 0) return (k);
    return data[k] = find(data[k]);
  }

  int size(int k) { return -data[find(k)]; }

  bool same(int x, int y) { return find(x) == find(y); }

  vector<vector<int> > groups() {
    int n = (int)data.size();
    vector<vector<int> > ret(n);
    for (int i = 0; i < n; i++) {
      ret[find(i)].emplace_back(i);
    }
    ret.erase(remove_if(begin(ret), end(ret),
                        [&](const vector<int> &v) { return v.empty(); }),
              end(ret));
    return ret;
  }
};



constexpr int vy[] = {-1, 0, 1, 0};
constexpr int vx[] = {0, -1, 0, 1};
const string vs = "ULDR";

bool in(int y, int x) {
  return 0 <= y and y < N and 0 <= x and x < N;
};

void set_score(State& s) {
  vector< vector< int > > dp(N, vector< int >(N, inf));
  queue< pair< int, int > > que;
  que.emplace(ay, ax);
  dp[ay][ax] = 0;
  for(int i = ay - 1; i >= 0; i--) {
    if(s.C[i][ax] == '@') break;
    if(chmin(dp[i][ax], 1)) {
      que.emplace(i, ax);
    }
  }
  for(int i = ay + 1; i < N; i++) {
    if(s.C[i][ax] == '@') break;
    if(chmin(dp[i][ax], 1)) {
      que.emplace(i, ax);
    }
  }
  for(int j = ax - 1; j >= 0; j--) {
    if(s.C[ay][j] == '@') break;
    if(chmin(dp[ay][j], 1)) {
      que.emplace(ay, j);
    }
  }
  for(int j = ax + 1; j < N; j++) {
    if(s.C[ay][j] == '@') break;
    if(chmin(dp[ay][j], 1)) {
      que.emplace(ay, j);
    }
  }
  while(not que.empty()) {
    auto [y, x] = que.front();
    que.pop();
    for(int k = 0; k < 4; k++) {
      int ny = y + vy[k], nx = x + vx[k];
      if(in(ny, nx) and s.C[ny][nx] != '@' and dp[ny][nx] == inf) {
        dp[ny][nx] = dp[y][x] + 1;
        que.emplace(ny, nx);
      }
    }
  }
  int score = 0;
  for(int i = 0; i < N; i++) {
    for(int j = 0; j < N; j++) {
      if(s.C[i][j] == 'a') {
        if(dp[i][j] >= inf) {
          s.score = inf;
          return;
        }
        score += dp[i][j];
      }
    }
  }
  s.score = score;
}


vector< vector< int > > build_score_map(State& s) {
  vector< vector< int > > dp(N, vector< int >(N, inf));
  queue< pair< int, int > > que;
  que.emplace(ay, ax);
  dp[ay][ax] = 0;
  for(int i = ay - 1; i >= 0; i--) {
    if(s.C[i][ax] == '@') break;
    if(chmin(dp[i][ax], 1)) {
      que.emplace(i, ax);
    }
  }
  for(int i = ay + 1; i < N; i++) {
    if(s.C[i][ax] == '@') break;
    if(chmin(dp[i][ax], 1)) {
      que.emplace(i, ax);
    }
  }
  for(int j = ax - 1; j >= 0; j--) {
    if(s.C[ay][j] == '@') break;
    if(chmin(dp[ay][j], 1)) {
      que.emplace(ay, j);
    }
  }
  for(int j = ax + 1; j < N; j++) {
    if(s.C[ay][j] == '@') break;
    if(chmin(dp[ay][j], 1)) {
      que.emplace(ay, j);
    }
  }
  while(not que.empty()) {
    auto [y, x] = que.front();
    que.pop();
    for(int k = 0; k < 4; k++) {
      int ny = y + vy[k], nx = x + vx[k];
      if(in(ny, nx) and s.C[ny][nx] != '@' and dp[ny][nx] == inf) {
        dp[ny][nx] = dp[y][x] + 1;
        que.emplace(ny, nx);
      }
    }
  }
  return dp;
}



int main() {
  cin >> N >> M;
  vector<string> C(N);
  cin >> C;

  for(int i = 0; i < N; i++) {
    for(int j = 0; j < N; j++) {
      if(C[i][j] == 'A') {
        ay = i;
        ax = j;
      }
    }
  }


  State init {C};
  int y = ay, x = ax;
  init.y = y;
  init.x = x;

  for(;;) {
    UnionFind uf(N * N);
    for(int i = 0; i < N; i++) {
      for(int j = 0; j < N; j++) {
        if(C[i][j] != '@') {
          if(i and C[i - 1][j] != '@') uf.unite((i - 1) * N + j, i * N + j);
          if(j and C[i][j - 1] != '@') uf.unite(i * N + (j - 1), i * N + j);
        }
      }
    }
    vector< vector< pair< int, int > > > belong(N * N);
    vector< int > roots;
    for(int i = 0; i < N; i++) {
      for (int j = 0; j < N; j++) {
        if (C[i][j] == 'A' or C[i][j] == 'a') {
          roots.emplace_back(uf.find(N * i + j));
          belong[uf.find(N * i + j)].emplace_back(i, j);
        }
      }
    }
    sort(roots.begin(), roots.end());
    roots.erase(unique(roots.begin(), roots.end()), roots.end());
    if(roots.size() <= 1) {
      break;
    }
    int min_cost = inf, a, b, c, d;
    for(int i = 0; i < (int) roots.size(); i++) {
      for(int j = 0; j < i; j++) {
        for(auto& p : belong[roots[i]]) {
          for(auto& q : belong[roots[j]]) {
            if(chmin(min_cost, abs(p.first - q.first) + abs(p.second - q.second))) {
              a = p.first;
              b = p.second;
              c = q.first;
              d = q.second;
            }
          }
        }
      }
    }
    while(a < c) {
      if(C[a][b] == '@') C[a][b] = 'a';
      ++a;
    }
    while(b < d) {
      if(C[a][b] == '@') C[a][b] = 'a';
      ++b;
    }
    while(a > c) {
      if(C[a][b] == '@') C[a][b] = 'a';
      --a;
    }
    while(b > d) {
      if(C[a][b] == '@') C[a][b] = 'a';
      --b;
    }
  }

  init.C = C;
  set_score(init);
  vector< priority_queue< State > > dp(10101);
  dp[0].emplace(init);

  int best = inf;
  State best_s;

  for(int turn = 0; turn < best; turn++) {
    while(not dp[turn].empty()) {
      auto s = dp[turn].top();
      dp[turn].pop();
      auto score_map = build_score_map(s);
      for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
          if (s.C[i][j] == 'a') {
            int nxt_turn = turn + abs(i - s.y) + abs(j - s.x) + 1;
            if (nxt_turn >= (int) dp.size()) {
              continue;
            }
            // 2. 上下左右に隣接するマスに移動する
            for (int k = 0; k < 4; k++) {
              int ny = i + vy[k], nx = j + vx[k];
              if (in(ny, nx) and (s.C[ny][nx] == '.' or s.C[ny][nx] == 'A')) {

                auto nxt_score = s.score - score_map[i][j] + score_map[ny][nx];

                if(dp[nxt_turn].size() <= 2 or nxt_score < dp[nxt_turn].top().score) {
                  State nxt_s = s;
                  nxt_s.op.emplace_back(1, i * N + j);
                  nxt_s.y = ny;
                  nxt_s.x = nx;
                  nxt_s.op.emplace_back(2, vs[k]);
                  if (s.C[ny][nx] != 'A') nxt_s.C[ny][nx] = nxt_s.C[i][j];
                  nxt_s.C[i][j] = '.';
                  nxt_s.score = nxt_score;
                  if (nxt_s.score == 0) {
                    if (chmin(best, nxt_turn)) best_s = nxt_s;
                  } else {
                    dp[nxt_turn].emplace(nxt_s);
                    if (dp[nxt_turn].size() > 2) dp[nxt_turn].pop();
                  }
                }
              }
            }

            // 3. どこかに蹴る
            for (int k = 0; k < 4; k++) {
              int ny = i, nx = j;
              while (s.C[ny][nx] != 'A' and in(ny + vy[k], nx + vx[k])
                  and (s.C[ny + vy[k]][nx + vx[k]] == 'A' or s.C[ny + vy[k]][nx + vx[k]] == '.')) {
                ny += vy[k];
                nx += vx[k];
              }
              if (ny == i and nx == j) continue;

              auto nxt_score = s.score - score_map[i][j] + score_map[ny][nx];

              if(dp[nxt_turn].size() < 5 or nxt_score < dp[nxt_turn].top().score) {
                State nxt_s = s;
                nxt_s.y = i;
                nxt_s.x = j;
                nxt_s.op.emplace_back(1, i * N + j);
                assert(nxt_s.C[i][j] == '@' or nxt_s.C[i][j] == 'a');
                nxt_s.op.emplace_back(3, vs[k]);
                if (s.C[ny][nx] != 'A') nxt_s.C[ny][nx] = nxt_s.C[i][j];
                nxt_s.C[i][j] = '.';
                nxt_s.score = nxt_score;
                if (nxt_s.score == 0) {
                  if (chmin(best, nxt_turn)) best_s = nxt_s;
                } else {
                  dp[nxt_turn].emplace(nxt_s);
                  if(dp[nxt_turn].size() > 5) dp[nxt_turn].pop();
                }
              }
            }
          }
        }
      }

      for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
          if (s.C[i][j] == '@') {
            int nxt_turn = turn + abs(i - s.y) + abs(j - s.x) + 1;

            if (nxt_turn >= (int) dp.size()) {
              continue;
            }

            // 2. 上下左右に隣接するマスに移動する
            for (int k = 0; k < 4; k++) {
              int ny = i + vy[k], nx = j + vx[k];
              if (in(ny, nx) and (s.C[ny][nx] == '.' or s.C[ny][nx] == 'A')) {

                State nxt_s = s;
                nxt_s.op.emplace_back(1, i * N + j);
                nxt_s.y = ny;
                nxt_s.x = nx;
                nxt_s.op.emplace_back(2, vs[k]);
                if (s.C[ny][nx] != 'A') nxt_s.C[ny][nx] = nxt_s.C[i][j];
                nxt_s.C[i][j] = '.';
                set_score(nxt_s);

                if (nxt_s.score == 0) {
                  if (chmin(best, nxt_turn)) best_s = nxt_s;
                } else {
                  dp[nxt_turn].emplace(nxt_s);
                  if (dp[nxt_turn].size() > 5) dp[nxt_turn].pop();
                }
              }
            }
          }
        }
      }
    }
  }

  int yy = ay, xx = ax;
  for(auto& p : best_s.op) {
    if(p.first == 1) {
      int ny = p.second / N, nx = p.second % N;
      while (yy < ny) {
        ++yy;
        cout << 1 << " " << 'D' << endl;
      }
      while (xx < nx) {
        ++xx;
        cout << 1 << " " << 'R' << endl;
      }
      while (yy > ny) {
        --yy;
        cout << 1 << " " << 'U' << endl;
      }
      while (xx > nx) {
        --xx;
        cout << 1 << " " << 'L' << endl;
      }
    } else {
      cout << p.first << " " << (char) p.second << endl;
      int d = (int) vs.find((char)p.second);
      if(p.first == 2) {
        yy += vy[d];
        xx += vx[d];
      }
    }
  }
}