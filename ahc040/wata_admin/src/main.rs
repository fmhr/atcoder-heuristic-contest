#![allow(non_snake_case)]

use proconio::input_interactive;
use rand::prelude::*;
use rustc_hash::FxHashSet;
use std::collections::BinaryHeap;
use std::f64::{
    self,
    consts::{PI, SQRT_2},
};

const TL: f64 = 2.9;

fn main() {
    get_time();
    let mut input = read_input();
    update_input(&mut input);
    beam(&mut input);
}

/// 入力の精度を高める
fn update_input(input: &mut Input) {
    // 入力生成方法からベイズ推定
    let (means, vars) = estimate(
        &input.wv.iter().map(|&(w, _)| w as f64).collect::<Vec<_>>(),
        input.sigma as f64,
    );
    for i in 0..input.N * 2 {
        input.wv[i] = (means[i].round() as i64, vars[i].round() as i64);
    }
    input.estimator = Estimator::new(
        &input.wv.iter().map(|&(w, _)| w as f64).collect::<Vec<_>>(),
        &input.wv.iter().map(|&(_, v)| v as f64).collect::<Vec<_>>(),
        input.sigma as f64,
    );

    for i in 0..input.N*2 {
        eprintln!("{}: {} ± {:.0}", i, input.wv[i].0, (input.wv[i].1 as f64).sqrt());
    }

    // 十字型に並べてクエリすることで和の計測を行う
    // 現在の事後分布における計測結果の分散が大きくなるようにクエリを構築することで情報量を高める
    let mut rng = rand_pcg::Pcg64Mcg::seed_from_u64(432897);
    let stime = get_time();
    let mut iter = 0;
    while input.turn + NUM_WIDTH < input.T {
        let mut best: (Vec<usize>, Vec<usize>, Vec<Cmd>) = (vec![], vec![], vec![]);
        let mut max = 0.0;
        'lp: while best.2.is_empty()
            || get_time() - stime
                < 0.5 * (input.turn + 1) as f64 / (input.T as f64 - NUM_WIDTH as f64)
        {
            iter += 1;
            let mut cmds = vec![Cmd {
                p: 0,
                r: false,
                d: false,
                b: !0,
            }];
            let mut row = vec![0];
            let mut col = vec![0];
            let mut rb = !1;
            let mut cb = !1;
            let mut sim = Sim::new(input);
            sim.put(input, cmds[0], true);
            for p in 1..input.N {
                if sim.W.mean > 1e5 && sim.H.mean > 1e5 && rb != !1 && cb != !1 {
                    // 衝突可能性がなさそうなら分散が大きくなるように貪欲に配置
                    let mut c = Cmd {
                        p,
                        r: false,
                        d: false,
                        b: !0,
                    };
                    let mut max = -1e100;
                    for r in [false, true] {
                        for d in [false, true] {
                            if !d {
                                let id = p * 2 + !r as usize;
                                let mut sum = input.estimator.covariance[(id, id)];
                                for &q in &sim.H.a {
                                    sum += 2.0 * input.estimator.covariance[(id, q)];
                                }
                                if max.setmax(sum) {
                                    c = Cmd { p, r, d, b: cb };
                                }
                            } else {
                                let id = p * 2 + r as usize;
                                let mut sum = input.estimator.covariance[(id, id)];
                                for &q in &sim.W.a {
                                    sum += 2.0 * input.estimator.covariance[(id, q)];
                                }
                                if max.setmax(sum) {
                                    c = Cmd { p, r, d, b: rb };
                                }
                            }
                        }
                    }
                    if sim.put(input, c, true) {
                        cmds.push(c);
                        continue;
                    }
                }
                // 意図しない衝突が発生する可能性があるので最初の方はランダムに置く
                let (r, d) = (rng.gen_bool(0.5), rng.gen_bool(0.5));
                if !d {
                    col.push(p);
                    if cb == !1 {
                        cb = *[!0]
                            .iter()
                            .chain(&row[..row.len() - 1])
                            .choose(&mut rng)
                            .unwrap();
                    }
                    cmds.push(Cmd { p, r, d, b: cb });
                } else {
                    row.push(p);
                    if rb == !1 {
                        rb = *[!0]
                            .iter()
                            .chain(&col[..col.len() - 1])
                            .choose(&mut rng)
                            .unwrap();
                    }
                    cmds.push(Cmd { p, r, d, b: rb });
                }
                if !sim.put(input, cmds[p], true) {
                    continue 'lp;
                }
            }
            let mut var = 0.0;
            for &i in &sim.W.a {
                for &j in &sim.W.a {
                    var += input.estimator.covariance[(i, j)];
                }
            }
            for &i in &sim.H.a {
                for &j in &sim.H.a {
                    var += input.estimator.covariance[(i, j)];
                }
            }
            if max.setmax(var) {
                best = (sim.W.a, sim.H.a, cmds);
            }
        }
        let (row, col, cmds) = best;
        let (W, H) = input.query(&cmds);
        input.estimator.update(&row, W as f64);
        input.estimator.update(&col, H as f64);
    }
    eprintln!("!log iter_update {iter}");
    for i in 0..input.N * 2 {
        let (mean, var) = (input.estimator.mean[i], input.estimator.covariance[(i, i)]);
        input.wv[i] = (mean.round() as i64, var.round() as i64);
    }
    for i in 0..input.N {
        eprintln!(
            "w[{i}]: {} ± {:.0}, ({})",
            input.wv[i * 2].0,
            (input.wv[i * 2].1 as f64).sqrt(),
            input.wv[i * 2].0 - input.true_w[i * 2]
        );
        eprintln!(
            "h[{i}]: {} ± {:.0}, ({})",
            input.wv[i * 2 + 1].0,
            (input.wv[i * 2 + 1].1 as f64).sqrt(),
            input.wv[i * 2 + 1].0 - input.true_w[i * 2 + 1]
        );
    }
    eprintln!("!log diff {:.0}", input.diff());
    eprintln!("!log time_update {:.3}", get_time());
    input.update();
}

#[derive(Clone, Debug)]
struct Val {
    mean: f64,
    sigma: f64,
    a: Vec<usize>,
}

impl Val {
    const ZERO: Self = Val {
        mean: 0.0,
        sigma: 0.0,
        a: vec![],
    };
    fn add(&mut self, i: usize, estimator: &Estimator) {
        let (mean, var) = (estimator.mean[i], estimator.covariance[(i, i)]);
        self.mean += mean;
        let mut var = self.sigma * self.sigma + var;
        for &j in &self.a {
            var += 2.0 * estimator.covariance[(i, j)];
        }
        self.sigma = var.sqrt();
        self.a.push(i);
    }
    fn ub(&self, d: f64) -> f64 {
        self.mean + self.sigma * d
    }
    fn lb(&self, d: f64) -> f64 {
        self.mean - self.sigma * d
    }
}

// 事後分布の分散共分散行列を用いて、実際に配置した場合のスコアを推定する
// 共分散は無視して単純に独立な正規分布として扱うだけで十分かも
struct Sim {
    pos: Vec<(Val, Val, Val, Val)>,
    done: Vec<bool>,
    W: Val,
    H: Val,
}

impl Sim {
    fn new(input: &Input) -> Self {
        Self {
            pos: vec![
                (
                    Val::ZERO.clone(),
                    Val::ZERO.clone(),
                    Val::ZERO.clone(),
                    Val::ZERO.clone()
                );
                input.N
            ],
            done: vec![false; input.N],
            W: Val::ZERO.clone(),
            H: Val::ZERO.clone(),
        }
    }
    fn put(&mut self, input: &Input, c: Cmd, strict: bool) -> bool {
        let d = 1.5;
        let mut w = c.p * 2;
        let mut h = c.p * 2 + 1;
        if c.r {
            std::mem::swap(&mut w, &mut h);
        }
        let mut touch = false;
        if !c.d {
            let x1 = if c.b == !0 {
                Val::ZERO.clone()
            } else {
                self.pos[c.b as usize].1.clone()
            };
            let mut x2 = x1.clone();
            x2.add(w, &input.estimator);
            let mut y1 = Val::ZERO;
            for q in 0..c.p {
                if !self.done[q] || q == c.b as usize {
                    continue;
                }
                if self.pos[q].1.ub(d) > x1.lb(d) && x2.ub(d) > self.pos[q].0.lb(d) {
                    if self.pos[q].3.mean > y1.mean {
                        if strict {
                            touch =
                                self.pos[q].1.lb(d) < x1.ub(d) || x2.lb(d) < self.pos[q].0.ub(d);
                        }
                        y1 = self.pos[q].3.clone();
                    }
                }
            }
            if touch {
                return false;
            }
            let mut y2 = y1.clone();
            y2.add(h, &input.estimator);
            self.pos[c.p] = (x1, x2, y1, y2);
        } else {
            let y1 = if c.b == !0 {
                Val::ZERO.clone()
            } else {
                self.pos[c.b as usize].3.clone()
            };
            let mut y2 = y1.clone();
            y2.add(h, &input.estimator);
            let mut x1 = Val::ZERO;
            for q in 0..c.p {
                if !self.done[q] || q == c.b as usize {
                    continue;
                }
                if self.pos[q].3.ub(d) > y1.lb(d) && y2.ub(d) > self.pos[q].2.lb(d) {
                    if self.pos[q].1.mean > x1.mean {
                        if strict {
                            touch =
                                self.pos[q].3.lb(d) < y1.ub(d) || y2.lb(d) < self.pos[q].2.ub(d);
                        }
                        x1 = self.pos[q].1.clone();
                    }
                }
            }
            if touch {
                return false;
            }
            let mut x2 = x1.clone();
            x2.add(w, &input.estimator);
            self.pos[c.p] = (x1, x2, y1, y2);
        }
        self.done[c.p] = true;
        if self.W.mean < self.pos[c.p].1.mean {
            self.W = self.pos[c.p].1.clone();
        }
        if self.H.mean < self.pos[c.p].3.mean {
            self.H = self.pos[c.p].3.clone();
        }
        true
    }
    fn puts(&mut self, input: &Input, cmds: &[Cmd]) {
        for c in cmds {
            self.put(input, *c, false);
        }
    }
}

/// 箱の横幅を試す候補数
const NUM_WIDTH: usize = 10;

/// ビームサーチで詰め込みを行う
/// 縦方向をレイヤーに分割し、各レイヤーに箱を横方向に詰めていく
/// 箱は下から上へ動かして配置するため、上のレイヤーより先に下のレイヤーに箱を詰めないようにする
fn beam(input: &mut Input) {
    let mut trace = Trace::new();
    let mut beam = vec![];
    // 箱の横幅を sqrt(AREA) ~ 1.2 sqrt(AREA) まで0.2刻みで試す
    for i in 0..NUM_WIDTH {
        beam.push(vec![State {
            id: !0,
            p: 0,
            max_width: (input.total_area as f64 * (1.0 + 0.2 * i as f64 / (NUM_WIDTH - 1) as f64))
                .sqrt()
                .round() as i64,
            layer: vec![],
            score: 0,
        }]);
    }
    let stime = get_time();
    let mut total_width = 0;
    let mut prev_width = 3000000 / input.N;
    for p in 0..input.N {
        let width = if p <= 2 {
            prev_width
        } else {
            // これまでの幅の総和と経過時間から、幅1あたりの期待時間を求め、残りの時間とターン数に応じて幅を自動調整
            let t = (get_time() - stime) / (TL - stime);
            (total_width as f64 / t.max(1e-6) * (1.0 - t) / (input.N - p) as f64)
                .round()
                .min(prev_width as f64 * 1.2)
                .max(NUM_WIDTH as f64) as usize
        };
        prev_width = width;
        eprintln!("{}: {}", p, width);
        for i in 0..NUM_WIDTH {
            let mut cand = BoundedSortedList::new(width / NUM_WIDTH);
            for s in 0..beam[i].len() {
                let state = &beam[i][s];
                for j in 0..=state.layer.len() {
                    for r in [false, true] {
                        if let Some(score) = state.try_put(input, p, j, r) {
                            if cand.can_insert(score) {
                                cand.insert(score, (s, j, r));
                            }
                        }
                    }
                }
            }
            total_width += beam[i].len();
            let mut next = vec![];
            let mut used = FxHashSet::default();
            for (score, (s, j, r)) in cand.list() {
                let mut state = beam[i][s].clone();
                state.put(input, p, j, r, &mut trace);
                state.score = score;
                if used.insert(state.hash()) {
                    next.push(state);
                }
            }
            beam[i] = next;
        }
    }
    let num = input.T - input.turn;
    for i in 0..NUM_WIDTH {
        // 各横幅に対し、ビームサーチで求めた解のうち、実際に配置した際のスコアの高いものを出力
        // T-NUM_WIDTH回先にクエリしているのでk=1で固定
        let k = num * (i + 1) / NUM_WIDTH - num * i / NUM_WIDTH;
        beam[i].truncate(100);
        let mut sorted = (0..beam[i].len())
            .map(|s| {
                let mut sim = Sim::new(input);
                let cmds = trace.get(beam[i][s].id);
                sim.puts(input, &cmds);
                let mut var = sim.W.sigma * sim.W.sigma + sim.H.sigma * sim.H.sigma;
                for &i in &sim.W.a {
                    for &j in &sim.W.a {
                        var += 2.0 * input.estimator.covariance[(i, j)];
                    }
                }
                ((sim.W.mean + sim.H.mean).round() as i64, s, var.sqrt())
            })
            .collect_vec();
        sorted.sort_by_key(|&(score, _, _)| score);
        sorted.truncate(k);
        for (score, s, sigma) in sorted {
            let state = &beam[i][s];
            let cmds = trace.get(state.id);
            println!("# estimate = {}", state.score);
            println!("# estimate2 = {:.0} ± {:.0}", score, sigma);
            for i in 0..state.layer.len() {
                println!("# {}", state.layer[i].w);
            }
            input.query(&cmds);
        }
    }
    eprintln!("!log total_width {}", total_width);
    eprintln!("Time = {:.3}", get_time());
}

#[derive(Clone, Debug)]
struct State {
    id: usize,
    p: usize,
    max_width: i64,
    layer: Vec<Layer>,
    score: i64,
}

// 各レイヤーは、最後に置いた長方形の番号、レイヤーの横幅、縦幅、標準偏差、下のレイヤーに先を越されてもう配置不能か、の情報を持つ
#[derive(Clone, Copy, Debug)]
struct Layer {
    last: usize,
    w: i64,
    h: i64,
    sigma: f64,
    closed: bool,
}

impl Layer {
    fn new() -> Self {
        Self {
            last: !0,
            w: 0,
            h: 0,
            sigma: 0.0,
            closed: false,
        }
    }
}

const DELTA: f64 = 1.5;

impl State {
    /// 長方形 p をレイヤー j に回転 r で配置した場合のスコアを計算
    fn try_put(&self, input: &Input, p: usize, j: usize, r: bool) -> Option<i64> {
        if j < self.layer.len() && self.layer[j].closed {
            return None;
        }
        let (w, h) = if !r {
            (input.wv[p * 2], input.wv[p * 2 + 1])
        } else {
            (input.wv[p * 2 + 1], input.wv[p * 2])
        };
        if j < self.layer.len() && self.layer[j].w + w.0 > self.max_width {
            return None;
        }
        let mut layer_j = if j == self.layer.len() {
            Layer::new()
        } else {
            self.layer[j]
        };
        layer_j.last = p;
        layer_j.w += w.0;
        layer_j.sigma = (layer_j.sigma * layer_j.sigma + w.1 as f64).sqrt();
        layer_j.h.setmax(h.0);
        let mut total_h = 0;
        // 残りの長方形は液体だと思って、各レイヤーの空きスペースに流し込む
        // 本来は隙間が出来るはずなので液体の面積は長方形の面積の総和より少し大きくする
        let mut remaining_area = input.remaining_area[self.p + 1] * 105 / 100;
        let min_width = input.min_width[self.p + 1];
        let max_width = input.max_width[self.p + 1];
        let ub = layer_j.w as f64 + layer_j.sigma * DELTA;
        // 先を越されたレイヤーを閉じる
        for k in 0..j {
            let mut closed = self.layer[k].closed;
            if !closed && self.layer[k].w as f64 - self.layer[k].sigma * DELTA < ub {
                closed = true;
            }
            total_h += self.layer[k].h;
            if !closed && self.layer[k].w + min_width <= self.max_width {
                // レイヤーの横幅に残りの長方形の最小幅以上の空きがある場合、残横幅×min(レイヤーの縦幅,長方形の最大縦幅)だけ液体を流し込む
                remaining_area -=
                    (self.max_width - self.layer[k].w) * self.layer[k].h.min(max_width);
            }
        }
        total_h += layer_j.h;
        if layer_j.w + min_width <= self.max_width {
            remaining_area -= (self.max_width - layer_j.w) * layer_j.h.min(max_width);
        }
        for k in j + 1..self.layer.len() {
            total_h += self.layer[k].h;
            if !self.layer[k].closed && self.layer[k].w + min_width <= self.max_width {
                remaining_area -=
                    (self.max_width - self.layer[k].w) * self.layer[k].h.min(max_width);
            }
        }
        if remaining_area > 0 {
            // 液体が残った場合、追加のレイヤーが必要
            total_h += (remaining_area / self.max_width).max(input.min_width[self.p + 1]);
        }
        Some(self.max_width + total_h)
    }
    /// 長方形 p をレイヤー j に回転 r で配置して状態を更新
    fn put(&mut self, input: &Input, p: usize, j: usize, r: bool, trace: &mut Trace<Cmd>) {
        let (w, h) = if !r {
            (input.wv[p * 2], input.wv[p * 2 + 1])
        } else {
            (input.wv[p * 2 + 1], input.wv[p * 2])
        };
        self.p += 1;
        if j == self.layer.len() {
            self.layer.push(Layer::new());
        }
        let b = self.layer[j].last;
        self.layer[j].last = p;
        self.layer[j].w += w.0;
        self.layer[j].sigma = (self.layer[j].sigma * self.layer[j].sigma + w.1 as f64).sqrt();
        self.layer[j].h.setmax(h.0);
        let ub = self.layer[j].w as f64 + self.layer[j].sigma * DELTA;
        for k in 0..j {
            if !self.layer[k].closed && self.layer[k].w as f64 - self.layer[k].sigma * DELTA < ub {
                self.layer[k].closed = true;
            }
        }
        self.id = trace.add(Cmd { p, r, d: false, b }, self.id);
    }
    /// 高さと幅が同じような状態は同じハッシュ値にすることで多様性を確保
    fn hash(&self) -> u64 {
        const MOD: u64 = 92230953439943;
        const MUL: u64 = 100003;
        let mut hash = 0;
        for l in &self.layer {
            if !l.closed {
                hash = (hash * MUL + l.w as u64 / 10000 + 1) % MOD;
            } else {
                hash = (hash * MUL + 0) % MOD;
            }
            hash = (hash * MUL + l.h as u64 / 10000 + 1) % MOD;
        }
        hash
    }
}

/// 入力生成方法を元にベイズ推定することで、真のサイズの推定を補正する
/// chatGPTに方針の指示を出して作成
fn estimate(ys: &[f64], sigma: f64) -> (Vec<f64>, Vec<f64>) {
    // L の範囲を設定（整数値）
    let l_min = 10000;
    let l_max = 50000;
    let l_step = 10;

    // 事前分布の対数
    let log_p_l_prior = -(((l_max - l_min) / l_step + 1) as f64).ln();

    // 各観測値に対して、L に依存しない値を事前計算
    let mut betas = Vec::with_capacity(ys.len());
    let mut cdf_betas = Vec::with_capacity(ys.len());
    let mut phi_betas = Vec::with_capacity(ys.len());
    for &y_i in ys {
        let beta = (100000.0 - y_i) / sigma;
        betas.push(beta);
        cdf_betas.push(normal_cdf(beta));
        phi_betas.push(normal_pdf(beta));
        //eprintln!("{}: {} {} {}", y_i, beta, cdf_betas.last().unwrap(), phi_betas.last().unwrap());
    }

    // L の事後対数確率を格納するベクター
    let mut l_log_posterior = Vec::with_capacity(((l_max - l_min) / l_step + 1) as usize);

    // 各 L に対して事後分布の対数を計算
    for l in (l_min..=l_max).step_by(l_step) {
        // x の事前分布の対数
        let n_l = (100000 - l + 1) as f64;
        let log_p_x_prior = -n_l.ln();

        // 各観測値に対する対数尤度の計算
        let mut log_likelihood_sum = 0.0;
        let mut valid = true;
        for i in 0..ys.len() {
            let y_i = ys[i];
            let cdf_beta = cdf_betas[i];

            let alpha = (l as f64 - y_i) / sigma;
            let cdf_alpha = normal_cdf(alpha);
            let cdf_diff = cdf_beta - cdf_alpha;

            // cdf_diff が負またはゼロの場合、尤度を無視
            if cdf_diff <= 0.0 {
                valid = false;
                break;
            }

            // 対数尤度の計算
            let log_likelihood_i = log_p_x_prior + cdf_diff.ln();
            log_likelihood_sum += log_likelihood_i;
        }

        if !valid {
            continue;
        }

        // 事後分布の対数を計算
        let log_posterior = log_p_l_prior + log_likelihood_sum;

        // L と対数事後確率を保存
        l_log_posterior.push((l as f64, log_posterior));
    }

    // 分母の対数和を計算
    let log_denominator = log_sum_exp(
        &l_log_posterior
            .iter()
            .map(|&(_, log_p)| log_p)
            .collect::<Vec<_>>(),
    );
    eprintln!("log_denominator: {}", log_denominator);

    // 各 x_i の事後平均と分散を初期化
    let mut x_means = vec![0.0; ys.len()];
    let mut x_vars = vec![0.0; ys.len()];

    // 各 x_i の条件付き平均と分散を保存するベクターを事前に確保
    let num_l = l_log_posterior.len();
    let mut x_cond_means = vec![Vec::with_capacity(num_l); ys.len()];
    let mut x_cond_vars = vec![Vec::with_capacity(num_l); ys.len()];

    // 各 L に対して事後確率を計算し、各 x_i の条件付き事後平均と分散を計算
    let mut l_posterior = Vec::with_capacity(num_l); // (L, P(L | {y_i}))
    for (l_value, log_posterior) in &l_log_posterior {
        let l = *l_value;
        // 事後確率を計算
        let log_p_l = log_posterior - log_denominator;
        let p_l = f64::exp(log_p_l);
        l_posterior.push((*l_value, p_l));

        // 各 x_i の条件付き事後平均と分散を計算
        for i in 0..ys.len() {
            let y_i = ys[i];
            let beta = betas[i];
            let cdf_beta = cdf_betas[i];
            let phi_beta = phi_betas[i];

            let alpha = (l - y_i) / sigma;
            let cdf_alpha = normal_cdf(alpha);
            let phi_alpha = normal_pdf(alpha);

            let cdf_diff = cdf_beta - cdf_alpha;
            if cdf_diff <= 0.0 {
                continue;
            }

            let phi_diff = phi_alpha - phi_beta;

            // 条件付き平均と分散
            let mean = y_i + sigma * (phi_diff / cdf_diff);
            let variance = sigma
                * sigma
                * (1.0 + (alpha * phi_alpha - beta * phi_beta) / cdf_diff
                    - (phi_diff / cdf_diff).powi(2));

            // 値を保存
            x_cond_means[i].push(mean);
            x_cond_vars[i].push(variance);
        }
    }

    // 各 x_i の周辺化された平均を計算
    for i in 0..ys.len() {
        let mut mean_sum = 0.0;
        for (j, &(_, p_l)) in l_posterior.iter().enumerate() {
            mean_sum += x_cond_means[i][j] * p_l;
        }
        x_means[i] = mean_sum;
    }

    // 各 x_i の周辺化された分散を計算
    for i in 0..ys.len() {
        let mean_i = x_means[i];
        let mut var_sum = 0.0;
        for (j, &(_, p_l)) in l_posterior.iter().enumerate() {
            let mean_diff = x_cond_means[i][j] - mean_i;
            var_sum += (x_cond_vars[i][j] + mean_diff * mean_diff) * p_l;
        }
        x_vars[i] = var_sum;
    }
    (x_means, x_vars)
}

// 入出力と得点計算
#[allow(unused)]
#[derive(Clone, Debug)]
struct Input {
    N: usize,
    T: usize,
    sigma: i64,
    /// 幅と分散(i番の箱の幅は2i,2i+1番目の要素)
    wv: Vec<(i64, i64)>,
    /// 使用したクエリ回数
    turn: usize,
    /// 真の幅(デバッグ用)
    true_w: Vec<i64>,
    /// 面積の総和
    total_area: i64,
    /// i番以降の面積の総和
    remaining_area: Vec<i64>,
    /// i番以降の幅の最小値(隙間埋め判定に使う)
    min_width: Vec<i64>,
    /// i番以降の幅の最大値(充填率計算に使う)
    max_width: Vec<i64>,
    estimator: Estimator,
}

fn read_input() -> Input {
    input_interactive! {
        N: usize,
        T: usize,
        sigma: i64,
        wh: [(i64, i64); N],
    }
    #[allow(unused)]
    let mut true_wh = vec![(0, 0); N];
    #[cfg(feature = "local")]
    {
        println!("#!local");
        input_interactive! {
            true_wh_: [(i64, i64); N],
        }
        true_wh = true_wh_;
    }
    Input {
        N,
        T,
        sigma,
        wv: wh
            .iter()
            .flat_map(|&(w, h)| [(w, sigma * sigma), (h, sigma * sigma)])
            .collect(),
        turn: 0,
        true_w: true_wh.iter().flat_map(|&(w, h)| [w, h]).collect(),
        total_area: 0,
        remaining_area: vec![0; N + 1],
        min_width: vec![1000000; N + 1],
        max_width: vec![0; N + 1],
        estimator: Estimator::new(
            &wh.iter().map(|&(w, _)| w as f64).collect::<Vec<_>>(),
            &wh.iter().map(|&(_, v)| v as f64).collect::<Vec<_>>(),
            sigma as f64,
        ),
    }
}

impl Input {
    fn update(&mut self) {
        self.total_area = 0;
        let mut min_width = 1000000;
        let mut max_width = 0;
        for i in (0..self.N).rev() {
            self.total_area += self.wv[i * 2].0 * self.wv[i * 2 + 1].0;
            self.remaining_area[i] = self.total_area;
            min_width.setmin(self.wv[i * 2].0.min(self.wv[i * 2 + 1].0));
            max_width.setmax(self.wv[i * 2].0.max(self.wv[i * 2 + 1].0));
            self.min_width[i] = min_width;
            self.max_width[i] = max_width;
        }
    }
}

#[derive(Copy, Clone, Debug)]
struct Cmd {
    p: usize,
    r: bool,
    d: bool,
    b: usize,
}

impl Input {
    fn query(&mut self, cmds: &[Cmd]) -> (i64, i64) {
        self.turn += 1;
        assert!(self.turn <= self.T);
        println!("{}", cmds.len());
        for cmd in cmds {
            println!(
                "{} {} {} {}",
                cmd.p,
                if cmd.r { 1 } else { 0 },
                if cmd.d { 'L' } else { 'U' },
                if cmd.b == !0 { -1 } else { cmd.b as i32 }
            );
        }
        input_interactive!(W: i64, H: i64);
        (W, H)
    }
    /// 真の幅とのズレを計算
    fn diff(&self) -> f64 {
        let mut sum = 0.0;
        for (t, &(w, _)) in self.true_w.iter().zip(&self.wv) {
            sum += ((w - t) as f64).powi(2);
        }
        (sum / (2 * self.N) as f64).sqrt()
    }
}

// ここからライブラリ

pub trait SetMinMax {
    fn setmin(&mut self, v: Self) -> bool;
    fn setmax(&mut self, v: Self) -> bool;
}
impl<T> SetMinMax for T
where
    T: PartialOrd,
{
    fn setmin(&mut self, v: T) -> bool {
        *self > v && {
            *self = v;
            true
        }
    }
    fn setmax(&mut self, v: T) -> bool {
        *self < v && {
            *self = v;
            true
        }
    }
}

#[macro_export]
macro_rules! mat {
	($($e:expr),*) => { vec![$($e),*] };
	($($e:expr,)*) => { vec![$($e),*] };
	($e:expr; $d:expr) => { vec![$e; $d] };
	($e:expr; $d:expr $(; $ds:expr)+) => { vec![mat![$e $(; $ds)*]; $d] };
}

pub fn get_time() -> f64 {
    static mut STIME: f64 = -1.0;
    let t = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap();
    let ms = t.as_secs() as f64 + t.subsec_nanos() as f64 * 1e-9;
    unsafe {
        if STIME < 0.0 {
            STIME = ms;
        }
        // ローカル環境とジャッジ環境の実行速度差はget_timeで吸収しておくと便利
        #[cfg(feature = "local")]
        {
            (ms - STIME) * 1.0
        }
        #[cfg(not(feature = "local"))]
        {
            ms - STIME
        }
    }
}

use itertools::Itertools;
use nalgebra::{DMatrix, DVector};

type DynamicVector = DVector<f64>;
type DynamicMatrix = DMatrix<f64>;

/// 計測誤差sigmaで確率変数の和を計測した際にカルマンフィルタを用いて真の値を推定する
#[derive(Clone, Debug)]
struct Estimator {
    mean: DynamicVector,
    covariance: DynamicMatrix,
    sigma2: f64,
    n: usize,
}

impl Estimator {
    fn new(means: &[f64], vars: &[f64], sigma: f64) -> Self {
        let n = means.len();
        let sigma2 = sigma * sigma;
        let mean = DynamicVector::from(means.to_vec());
        let covariance = DMatrix::from_diagonal(&DVector::from_column_slice(vars));
        Estimator {
            mean,
            covariance,
            sigma2,
            n,
        }
    }

    /// S の和が t であったという情報をもとに推定値を更新
    fn update(&mut self, S: &[usize], t: f64) {
        let mut h = DynamicVector::zeros(self.n);
        for &index in S {
            h[index] = 1.0;
        }
        // S_i = H * Σ * H^T + σ²
        let s_i_matrix = h.transpose() * &self.covariance * &h;
        let s_i = s_i_matrix[(0, 0)] + self.sigma2;
        // Kalman gain K_i = Σ * H^T / S_i
        let k = (&self.covariance * &h) / s_i;
        // t - H * μ
        let residual = t - h.dot(&self.mean);
        // μ_new = μ_old + K * residual
        self.mean += &k * residual;
        // Σ_new = (I - K * H^T) * Σ_old
        self.covariance = &self.covariance - &k * (h.transpose() * &self.covariance);
    }
}

fn normal_cdf(x: f64) -> f64 {
    0.5 * (1.0 + libm::erf(x / SQRT_2))
}

fn normal_pdf(x: f64) -> f64 {
    (1.0 / (SQRT_2 * (PI).sqrt())) * f64::exp(-0.5 * x * x)
}

fn log_sum_exp(log_probs: &[f64]) -> f64 {
    let max_log = log_probs.iter().cloned().fold(f64::NEG_INFINITY, f64::max);
    if max_log.is_infinite() {
        return max_log;
    }
    let sum_exp = log_probs
        .iter()
        .map(|&x| f64::exp(x - max_log))
        .sum::<f64>();
    max_log + sum_exp.ln()
}

pub struct Trace<T: Copy> {
    log: Vec<(T, usize)>,
}

impl<T: Copy> Trace<T> {
    pub fn new() -> Self {
        Trace { log: vec![] }
    }
    pub fn add(&mut self, c: T, p: usize) -> usize {
        self.log.push((c, p));
        self.log.len() - 1
    }
    pub fn get(&self, mut i: usize) -> Vec<T> {
        let mut out = vec![];
        while i != !0 {
            out.push(self.log[i].0);
            i = self.log[i].1;
        }
        out.reverse();
        out
    }
}

#[derive(Clone, Debug)]
struct Entry<K, V> {
    k: K,
    v: V,
}

impl<K: PartialOrd, V> Ord for Entry<K, V> {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.partial_cmp(other).unwrap()
    }
}

impl<K: PartialOrd, V> PartialOrd for Entry<K, V> {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        self.k.partial_cmp(&other.k)
    }
}

impl<K: PartialEq, V> PartialEq for Entry<K, V> {
    fn eq(&self, other: &Self) -> bool {
        self.k.eq(&other.k)
    }
}

impl<K: PartialEq, V> Eq for Entry<K, V> {}

/// K が小さいトップn個を保持
#[derive(Clone, Debug)]
pub struct BoundedSortedList<K: PartialOrd + Copy, V: Clone> {
    que: BinaryHeap<Entry<K, V>>,
    size: usize,
}

impl<K: PartialOrd + Copy, V: Clone> BoundedSortedList<K, V> {
    pub fn new(size: usize) -> Self {
        Self {
            que: BinaryHeap::with_capacity(size),
            size,
        }
    }
    pub fn can_insert(&self, k: K) -> bool {
        self.que.len() < self.size || self.que.peek().unwrap().k > k
    }
    pub fn insert(&mut self, k: K, v: V) {
        if self.que.len() < self.size {
            self.que.push(Entry { k, v });
        } else if let Some(mut top) = self.que.peek_mut() {
            if top.k > k {
                top.k = k;
                top.v = v;
            }
        }
    }
    pub fn list(self) -> Vec<(K, V)> {
        let v = self.que.into_sorted_vec();
        v.into_iter().map(|e| (e.k, e.v)).collect()
    }
    pub fn len(&self) -> usize {
        self.que.len()
    }
}
