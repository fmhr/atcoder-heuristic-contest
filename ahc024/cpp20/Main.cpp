#include <bits/stdc++.h>
using namespace std;

template <class T> ostream& operator<<(ostream& os, const vector<T>& v){ os << "{"; for(size_t i=0; i<v.size(); i++) os << v[i] << (i+1==v.size() ? "" : ", "); os << "}"; return os; }
template <class T, class U> ostream& operator<<(ostream& os, const pair<T, U>& p){ return os << "{" << p.first << ", " << p.second << "}"; }

const int N = 50;
const int M = 100;
const int DY[4] = {1, -1, 0, 0};
const int DX[4] = {0, 0, 1, -1};

inline double get_time() {
	using namespace std::chrono;
	return duration_cast<milliseconds>(system_clock::now().time_since_epoch()).count();
}


class Xorshift{
    unsigned x, y, z, w;
public:
    Xorshift(unsigned seed = 123456789) : x(seed), y(362436069), z(521288629), w(88675123) {}

    unsigned xor128(){
        unsigned t;
        t = x ^ (x << 11);
        x = y; y = z; z = w;
        return w = (w ^ (w >> 19)) ^ (t ^ (t >> 8)); 
    }
    
    // [a, b) の int 乱数を生成
    int irand(int a, int b){
        return a + (xor128() % (b - a));
    }
    
    // [0.0, 1.0) の double 乱数を生成
    double drand(){
        return xor128() / 4294967296.0; // UINT_MAX+1 = 4294967296
    }
    
    // [a, b) の double 乱数を生成
    inline double drand(double a, double b){
        return a + drand() * (b - a);
    }
};

bool is_connected(array<array<int, N+2>, N+2> &c, int y0, int x0, int y1, int x1){
    int c0 = c[y0][x0];
    queue<pair<int, int>> que;
    set<pair<int, int>> st;
    que.push(make_pair(y0, x0));
    while(!que.empty()){
        auto p = que.front();
        que.pop();
        for(int d=0; d<4; d++){
            int y2 = p.first + DY[d];
            int x2 = p.second + DX[d];
            if(y2 < 0 || x2 < 0 || y2 >= N+2 || x2 >= N+2) continue;
            if(c[y2][x2] != c0) continue;
            if(y2 == y1 && x2 == x1) return true;
            if(st.find(make_pair(y2, x2)) == st.end()){
                st.insert(make_pair(y2, x2));
                que.push(make_pair(y2, x2));
            }
        }
    }
    return false;
}

const double BETA_FACTOR = 0.6;

bool move(Xorshift &xorshift, double time_ratio, array<array<int, N+2>, N+2> &c, array<array<int, M+1>, M+1> &con, array<array<int, M+1>, M+1> &cnt, int y0, int x0, int y1, int x1){
    int c0 = c[y0][x0];
    int c1 = c[y1][x1];
    if(c0 == c1) return 0;

    double de = 0.0;
    if(c0 == 0) de = 1;
    else if(c1 == 0) de = -1;
    double kbt = max(1.0 - time_ratio, 1e-6);
    double beta = BETA_FACTOR * 1.0 / kbt;

    if(c1 != 0 && exp(-de * beta) < xorshift.drand(0.0, 1.0)) return false;

    vector<pair<int, int>> cps;

    for(int d=0; d<4; d++){
        int y2 = y0 + DY[d];
        int x2 = x0 + DX[d];
        int c2 = c[y2][x2];
        cnt[c0][c2]--;
        cnt[c2][c0]--;
        cnt[c1][c2]++;
        cnt[c2][c1]++;
        if(c0 != c2){
            cps.push_back(make_pair(c0, c2));
            cps.push_back(make_pair(c2, c0));
        }
        if(c1 != c2){
            cps.push_back(make_pair(c1, c2));
            cps.push_back(make_pair(c2, c1));
        }
    }

    // rollback
    bool ok = true;
    for(auto cp : cps){
        if(con[cp.first][cp.second] == 0 && cnt[cp.first][cp.second] != 0){
            ok = false;
            break;
        }
        if(con[cp.first][cp.second] == 1 && cnt[cp.first][cp.second] == 0){
            ok = false;
            break;
        }
    }

    if(!ok){
        for(int d=0; d<4; d++){
            int y2 = y0 + DY[d];
            int x2 = x0 + DX[d];
            int c2 = c[y2][x2];
            cnt[c0][c2]++;
            cnt[c2][c0]++;
            cnt[c1][c2]--;
            cnt[c2][c1]--;
        }
        return false;
    }

    c[y0][x0] = c1;

    // check c0 connectivity
    vector<pair<int, int>> ps;
    for(int d=0; d<4; d++){
        int y2 = y0 + DY[d];
        int x2 = x0 + DX[d];
        if(c[y2][x2] == c0) ps.push_back(make_pair(y2, x2));
    }
    if(ps.size() <= 1) return true;
    pair<int, int> rep = ps[0];
    for(int i=1; i<ps.size(); i++){
        if(!is_connected(c, rep.first, rep.second, ps[i].first, ps[i].second)){
            c[y0][x0] = c0;

            for(int d=0; d<4; d++){
                int y2 = y0 + DY[d];
                int x2 = x0 + DX[d];
                int c2 = c[y2][x2];
                cnt[c0][c2]++;
                cnt[c2][c0]++;
                cnt[c1][c2]--;
                cnt[c2][c1]--;
            }

            return false;
        }
    }

    return true;
}

bool move2(Xorshift &xorshift, double time_ratio, array<array<int, N+2>, N+2> &c, array<array<int, M+1>, M+1> &con, array<array<int, M+1>, M+1> &cnt, int y0a, int x0a, int y0b, int x0b, int y1a, int x1a, int y1b, int x1b){
    int c0a = c[y0a][x0a];
    int c0b = c[y0b][x0b];
    int c1a = c[y1a][x1a];
    int c1b = c[y1b][x1b];
    if(c0a == c1a && c0b == c1b) return 0;

    double de = (c0a == 0) + (c0b == 0) - (c1a == 0) - (c1b == 0);
    double kbt = max(1.0 - time_ratio, 1e-6);
    double beta = BETA_FACTOR * 1.0 / kbt;

    if(de >= 0 && exp(-de * beta) < xorshift.drand(0.0, 1.0)) return false;

    vector<pair<int, int>> cps;

    for(int d=0; d<4; d++){
        int y2 = y0a + DY[d];
        int x2 = x0a + DX[d];
        int c2 = c[y2][x2];
        cnt[c0a][c2]--;
        cnt[c2][c0a]--;
        if(c0a != c2){
            cps.push_back(make_pair(c0a, c2));
            cps.push_back(make_pair(c2, c0a));
        }
    }
    for(int d=0; d<4; d++){
        int y2 = y0b + DY[d];
        int x2 = x0b + DX[d];
        int c2 = c[y2][x2];
        cnt[c0b][c2]--;
        cnt[c2][c0b]--;
        if(c0b != c2){
            cps.push_back(make_pair(c0b, c2));
            cps.push_back(make_pair(c2, c0b));
        }
    }
    cnt[c0a][c0b]++;
    cnt[c0b][c0a]++;

    c[y0a][x0a] = c1a;
    c[y0b][x0b] = c1b;

    for(int d=0; d<4; d++){
        int y2 = y0a + DY[d];
        int x2 = x0a + DX[d];
        int c2 = c[y2][x2];
        cnt[c1a][c2]++;
        cnt[c2][c1a]++;
        if(c1a != c2){
            cps.push_back(make_pair(c1a, c2));
            cps.push_back(make_pair(c2, c1a));
        }
    }
    for(int d=0; d<4; d++){
        int y2 = y0b + DY[d];
        int x2 = x0b + DX[d];
        int c2 = c[y2][x2];
        cnt[c1b][c2]++;
        cnt[c2][c1b]++;
        if(c1b != c2){
            cps.push_back(make_pair(c1b, c2));
            cps.push_back(make_pair(c2, c1b));
        }
    }
    cnt[c1a][c1b]--;
    cnt[c1b][c1a]--;

    // rollback
    bool ok = true;
    for(auto cp : cps){
        if(con[cp.first][cp.second] == 0 && cnt[cp.first][cp.second] != 0){
            ok = false;
            break;
        }
        if(con[cp.first][cp.second] == 1 && cnt[cp.first][cp.second] == 0){
            ok = false;
            break;
        }
    }

    if(!ok){
        cnt[c1a][c1b]++;
        cnt[c1b][c1a]++;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c1a][c2]--;
            cnt[c2][c1a]--;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c1b][c2]--;
            cnt[c2][c1b]--;
        }

        c[y0a][x0a] = c0a;
        c[y0b][x0b] = c0b;

        cnt[c0a][c0b]--;
        cnt[c0b][c0a]--;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c0a][c2]++;
            cnt[c2][c0a]++;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c0b][c2]++;
            cnt[c2][c0b]++;
        }

        return false;
    }

    // check c0 connectivity
    if(c0a == c0b){
        vector<pair<int, int>> ps;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            if(c[y2][x2] == c0a) ps.push_back(make_pair(y2, x2));
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            if(c[y2][x2] == c0b) ps.push_back(make_pair(y2, x2));
        }
        if(ps.size() <= 1) return true;
        pair<int, int> rep = ps[0];
        for(int i=1; i<ps.size(); i++){
            if(!is_connected(c, rep.first, rep.second, ps[i].first, ps[i].second)){
                // rollback
        cnt[c1a][c1b]++;
        cnt[c1b][c1a]++;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c1a][c2]--;
            cnt[c2][c1a]--;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c1b][c2]--;
            cnt[c2][c1b]--;
        }

        c[y0a][x0a] = c0a;
        c[y0b][x0b] = c0b;

        cnt[c0a][c0b]--;
        cnt[c0b][c0a]--;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c0a][c2]++;
            cnt[c2][c0a]++;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c0b][c2]++;
            cnt[c2][c0b]++;
        }
                return false;
            }
        }
    }else{
        {
            vector<pair<int, int>> ps;
            for(int d=0; d<4; d++){
                int y2 = y0a + DY[d];
                int x2 = x0a + DX[d];
                if(c[y2][x2] == c0a) ps.push_back(make_pair(y2, x2));
            }
            for(int d=0; d<4; d++){
                int y2 = y0b + DY[d];
                int x2 = x0b + DX[d];
                if(c[y2][x2] == c0a) ps.push_back(make_pair(y2, x2));
            }
            if(ps.size() >= 2) {
                pair<int, int> rep = ps[0];
                for(int i=1; i<ps.size(); i++){
                    if(!is_connected(c, rep.first, rep.second, ps[i].first, ps[i].second)){
                        // rollback
        cnt[c1a][c1b]++;
        cnt[c1b][c1a]++;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c1a][c2]--;
            cnt[c2][c1a]--;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c1b][c2]--;
            cnt[c2][c1b]--;
        }

        c[y0a][x0a] = c0a;
        c[y0b][x0b] = c0b;

        cnt[c0a][c0b]--;
        cnt[c0b][c0a]--;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c0a][c2]++;
            cnt[c2][c0a]++;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c0b][c2]++;
            cnt[c2][c0b]++;
        }
                        return false;
                    }
                }
            }
        }
        {
            vector<pair<int, int>> ps;
            for(int d=0; d<4; d++){
                int y2 = y0a + DY[d];
                int x2 = x0a + DX[d];
                if(c[y2][x2] == c0b) ps.push_back(make_pair(y2, x2));
            }
            for(int d=0; d<4; d++){
                int y2 = y0b + DY[d];
                int x2 = x0b + DX[d];
                if(c[y2][x2] == c0b) ps.push_back(make_pair(y2, x2));
            }
            if(ps.size() >= 2) {
                pair<int, int> rep = ps[0];
                for(int i=1; i<ps.size(); i++){
                    if(!is_connected(c, rep.first, rep.second, ps[i].first, ps[i].second)){
                        // rollback
        cnt[c1a][c1b]++;
        cnt[c1b][c1a]++;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c1a][c2]--;
            cnt[c2][c1a]--;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c1b][c2]--;
            cnt[c2][c1b]--;
        }

        c[y0a][x0a] = c0a;
        c[y0b][x0b] = c0b;

        cnt[c0a][c0b]--;
        cnt[c0b][c0a]--;
        for(int d=0; d<4; d++){
            int y2 = y0a + DY[d];
            int x2 = x0a + DX[d];
            int c2 = c[y2][x2];
            cnt[c0a][c2]++;
            cnt[c2][c0a]++;
        }
        for(int d=0; d<4; d++){
            int y2 = y0b + DY[d];
            int x2 = x0b + DX[d];
            int c2 = c[y2][x2];
            cnt[c0b][c2]++;
            cnt[c2][c0b]++;
        }
                        return false;
                    }
                }
            }
        }
    }

    return true;
}

int main(){
    double start_time = get_time();

    Xorshift xorshift;

    int n, m;
    array<array<int, N+2>, N+2> c;
    for(int i=0; i<N+2; i++){
        for(int j=0; j<N+2; j++){
            c[i][j] = 0;
        }
    }

    cin >> n >> m;
    for(int i=1; i<=N; i++){
        for(int j=1; j<=N; j++){
            cin >> c[i][j];
        }
    }

    array<array<int, M+1>, M+1> con;
    for(int i=0; i<M+1; i++){
        for(int j=0; j<M+1; j++){
            con[i][j] = 0;
        }
    }

    array<array<int, M+1>, M+1> cnt;
    for(int i=0; i<M+1; i++){
        for(int j=0; j<M+1; j++){
            cnt[i][j] = 0;
        }
    }

    for(int i=0; i<N+2; i++){
        for(int j=0; j<N+2; j++){
            int y0 = i;
            int x0 = j;
            for(int d=0; d<4; d++){
                int y1 = y0 + DY[d];
                int x1 = x0 + DX[d];
                if(y1 < 0 || y1 >= N+2 || x1 < 0 || x1 >= N+2) continue;
                int c0 = c[y0][x0];
                int c1 = c[y1][x1];
                cnt[c0][c1]++;
                con[c0][c1] = con[c1][c0] = 1;
            }
        }
    }

    while(true){
        double elapsed_ms = get_time() - start_time;
        if(elapsed_ms > 1950) break;
        double time_ratio = elapsed_ms / 2000.0;

        for(int t=0; t<1000; t++){
            int y0 = xorshift.irand(1, N+1);
            int x0 = xorshift.irand(1, N+1);
            int d = xorshift.irand(0, 4);

            int y1 = y0 + DY[d];
            int x1 = x0 + DX[d];

            if(y1 < 0 || y1 >= N+2 || x1 < 0 || x1 >= N+2) continue;

            move(xorshift, time_ratio, c, con, cnt, y0, x0, y1, x1);

            int y0a = xorshift.irand(1, N+1);
            int x0a = xorshift.irand(1, N+1);
            int d0 = xorshift.irand(0, 4);
            int y0b = y0a + DY[d0];
            int x0b = x0a + DX[d0];
            if(y0b <= 0 || y0b >= N+1 || x0b <= 0 || x0b >= N+1) continue;

            int d1 = xorshift.irand(0, 2) * 2 - 1;
            int y1a, y1b, x1a, x1b;
            if(d0 < 2){
                y1a = y0a;
                y1b = y0b;
                x1a = x0a + d1;
                x1b = x0a + d1;
                if(x1a < 0 || x1a >= N+2) continue;
                if(x1b < 0 || x1b >= N+2) continue;
            }else{
                x1a = x0a;
                x1b = x0b;
                y1a = y0a + d1;
                y1b = y0a + d1;
                if(y1a < 0 || y1a >= N+2) continue;
                if(y1b < 0 || y1b >= N+2) continue;
            }

            move2(xorshift, time_ratio, c, con, cnt, y0a, x0a, y0b, x0b, y1a, x1a, y1b, x1b);
        }
    }

    for(int i=1; i<=N; i++){
        for(int j=1; j<=N; j++){
            cout << c[i][j] << " ";
        }
        cout << endl;
    }

    double end_time = get_time();
    cerr << "elapsed: " << end_time - start_time << "ms" << endl;
    return 0;
}


