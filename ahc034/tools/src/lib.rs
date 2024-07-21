#![allow(non_snake_case, unused_macros)]

use itertools::Itertools;
use noise::{NoiseFn, Perlin};
use proconio::input;
use rand::prelude::*;
use std::ops::RangeBounds;
use svg::node::element::{Group, Line, Rectangle, Style, Symbol, Title, Use};

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
	($($e:expr),*) => { Vec::from(vec![$($e),*]) };
	($($e:expr,)*) => { Vec::from(vec![$($e),*]) };
	($e:expr; $d:expr) => { Vec::from(vec![$e; $d]) };
	($e:expr; $d:expr $(; $ds:expr)+) => { Vec::from(vec![mat![$e $(; $ds)*]; $d]) };
}

#[derive(Clone, Debug)]
pub struct Input {
    N: usize,
    h: Vec<Vec<i64>>,
}

impl std::fmt::Display for Input {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        writeln!(f, "{}", self.N)?;
        for i in 0..self.N {
            writeln!(f, "{}", self.h[i].iter().join(" "))?;
        }
        Ok(())
    }
}

pub fn parse_input(f: &str) -> Input {
    let f = proconio::source::once::OnceSource::from(f);
    input! {
        from f,
        N: usize,
        h: [[i64; N]; N],
    }
    Input { N, h }
}

pub fn read<T: Copy + PartialOrd + std::fmt::Display + std::str::FromStr, R: RangeBounds<T>>(
    token: Option<&str>,
    range: R,
) -> Result<T, String> {
    if let Some(v) = token {
        if let Ok(v) = v.parse::<T>() {
            if !range.contains(&v) {
                Err(format!("Out of range: {}", v))
            } else {
                Ok(v)
            }
        } else {
            Err(format!("Parse error: {}", v))
        }
    } else {
        Err("Unexpected EOF".to_owned())
    }
}

const DIJ: [(usize, usize); 4] = [(0, 1), (1, 0), (0, !0), (!0, 0)];
const DIR: [char; 4] = ['R', 'D', 'L', 'U'];

#[derive(Clone, Debug, Copy)]
pub enum Action {
    Move(usize),
    Load(i64),
}

pub struct Output {
    pub out: Vec<Action>,
}

pub fn parse_output(_input: &Input, f: &str) -> Result<Output, String> {
    let mut out = vec![];
    for line in f.trim().lines() {
        let line = line.trim();
        if line.len() > 0 {
            if line.starts_with('+') {
                out.push(Action::Load(read(Some(&line[1..]), 1..=1000000)?));
            } else if line.starts_with('-') {
                out.push(Action::Load(-read(Some(&line[1..]), 1..=1000000)?));
            } else if line.len() == 1 {
                let dir = line.chars().next().unwrap();
                if let Some(dir) = DIR.iter().position(|&d| d == dir) {
                    out.push(Action::Move(dir));
                } else {
                    return Err(format!("Invalid action: {}", line));
                }
            } else {
                return Err(format!("Invalid action: {}", line));
            }
        }
    }
    if out.len() > 100000 {
        return Err("Too many actions".to_owned());
    }
    Ok(Output { out })
}

pub fn gen(seed: u64) -> Input {
    let mut rng = rand_chacha::ChaCha20Rng::seed_from_u64(seed);
    let N = 20;
    loop {
        let mut h = mat![0; N; N];
        let perlin = Perlin::new(rng.gen());
        let D = 10.0;
        let mut sum = 0;
        let mut ok = false;
        for i in 0..N {
            for j in 0..N {
                let x = i as f64 / D;
                let y = j as f64 / D;
                h[i][j] = ((perlin.get([x, y])) * 50.0).round() as i64;
                sum += h[i][j];
                if h[i][j] != 0 {
                    ok = true;
                }
            }
        }
        if !ok {
            continue;
        }
        let mut order = (0..N * N).collect_vec();
        order.shuffle(&mut rng);
        while sum > 0 {
            for &p in &order {
                h[p / N][p % N] -= 1;
                sum -= 1;
                if sum == 0 {
                    break;
                }
            }
        }
        while sum < 0 {
            for &p in &order {
                h[p / N][p % N] += 1;
                sum += 1;
                if sum == 0 {
                    break;
                }
            }
        }
        return Input { N, h };
    }
}

pub fn compute_score(input: &Input, out: &Output) -> (i64, String) {
    let (mut score, err, _) = compute_score_details(input, &out.out);
    if err.len() > 0 {
        score = 0;
    }
    (score, err)
}

fn compute_diff(h: &Vec<Vec<i64>>) -> i64 {
    let mut diff = 0;
    for i in 0..h.len() {
        for j in 0..h[i].len() {
            if h[i][j] != 0 {
                diff += 10000 + h[i][j].abs() * 100;
            }
        }
    }
    diff
}

pub fn compute_score_details(
    input: &Input,
    out: &[Action],
) -> (
    i64,
    String,
    (Vec<Vec<i64>>, usize, usize, i64, i64, i64, usize),
) {
    let mut h = input.h.clone();
    let mut i = 0;
    let mut j = 0;
    let mut v = 0;
    let mut cost = 0;
    let mut last_dir = 0;
    for turn in 0..out.len() {
        match out[turn] {
            Action::Move(dir) => {
                i += DIJ[dir].0;
                j += DIJ[dir].1;
                last_dir = dir;
                if i >= input.N || j >= input.N {
                    return (
                        0,
                        format!("Out of the board. (turn: {})", turn),
                        (h.clone(), i, j, v, cost, compute_diff(&h), last_dir),
                    );
                }
                cost += v.abs() + 100;
            }
            Action::Load(d) => {
                if d < 0 {
                    if v + d < 0 {
                        return (
                            0,
                            format!(
                                "The unloading amount exceeds the carrying amount. (turn: {})",
                                turn
                            ),
                            (h.clone(), i, j, v, cost, compute_diff(&h), last_dir),
                        );
                    }
                    v += d;
                    h[i][j] -= d;
                } else {
                    v += d;
                    h[i][j] -= d;
                }
                cost += d.abs();
            }
        }
    }
    let diff = compute_diff(&h);
    let mut base = 0;
    for i in 0..input.N {
        for j in 0..input.N {
            base += input.h[i][j].abs();
        }
    }
    let score = (1e9 * base as f64 / (cost + diff) as f64).round() as i64;
    (score, String::new(), (h, i, j, v, cost, diff, last_dir))
}

/// 0 <= val <= 1
pub fn color(mut val: f64) -> String {
    val.setmin(1.0);
    val.setmax(0.0);
    let (r, g, b) = if val < 0.5 {
        let x = val * 2.0;
        (
            30. * (1.0 - x) + 144. * x,
            144. * (1.0 - x) + 255. * x,
            255. * (1.0 - x) + 30. * x,
        )
    } else {
        let x = val * 2.0 - 1.0;
        (
            144. * (1.0 - x) + 255. * x,
            255. * (1.0 - x) + 30. * x,
            30. * (1.0 - x) + 70. * x,
        )
    };
    format!(
        "#{:02x}{:02x}{:02x}",
        r.round() as i32,
        g.round() as i32,
        b.round() as i32
    )
}

pub fn rect(x: usize, y: usize, w: usize, h: usize, fill: &str) -> Rectangle {
    Rectangle::new()
        .set("x", x)
        .set("y", y)
        .set("width", w)
        .set("height", h)
        .set("fill", fill)
}

pub fn group(title: String) -> Group {
    Group::new().add(Title::new(title))
}

pub fn vis_default(input: &Input, out: &Output) -> (i64, String, String) {
    let VisResult {
        mut score,
        err,
        svg,
        ..
    } = vis(input, &out.out);
    if err.len() > 0 {
        score = 0;
    }
    (score, err, svg)
}

pub struct VisResult {
    pub score: i64,
    pub cost: i64,
    pub diff: i64,
    pub v: i64,
    pub err: String,
    pub svg: String,
}

// https://www.svgrepo.com/svg/183941/dump-truck-truck
const DUMP: &str = r#"<g>
<path style="fill:#6B4425;" d="M224.975,126.109l13.365-33.421h52.974l-44.138-52.966l-61.793-8.828l-26.483,26.483l-26.483,17.655
    l-52.966-8.828L52.97,136.826h156.178C216.113,136.826,222.389,132.58,224.975,126.109"/>
<path style="fill:#378CD5;" d="M506.605,219.17l-76.712-84.383c-3.928-4.317-9.498-6.788-15.342-6.788h-81.311
    c-8.527,0-15.448,6.921-15.448,15.448v196.414H26.482v75.767c0,6.912,5.597,12.509,12.509,12.509h13.974
    c0-29.246,23.711-52.966,52.966-52.966s52.966,23.72,52.966,52.966h158.897h44.138c0-29.246,23.711-52.966,52.966-52.966
    c29.255,0,52.966,23.72,52.966,52.966h31.629c6.912,0,12.509-5.597,12.509-12.509V233.127
    C511.999,227.963,510.075,222.993,506.605,219.17"/>
<path style="fill:#FEC24B;" d="M256,339.864H26.483L0,225.105l35.31-88.276h173.833c6.974,0,13.241-4.246,15.828-10.717
    l13.374-33.421h52.965L256,339.864z"/>
<path style="fill:#E0E0E0;" d="M368.275,269.243h143.722v-36.122c0-5.155-1.924-10.134-5.394-13.948l-50.776-55.861h-87.552
    c-3.505,0-6.347,2.842-6.347,6.347v93.237C361.928,266.401,364.771,269.243,368.275,269.243"/>
<g>
    <path style="fill:#494949;" d="M467.862,428.14c0,29.255-23.711,52.966-52.966,52.966s-52.966-23.711-52.966-52.966
        c0-29.255,23.711-52.966,52.966-52.966S467.862,398.885,467.862,428.14"/>
    <path style="fill:#494949;" d="M158.897,428.14c0,29.255-23.711,52.966-52.966,52.966s-52.966-23.711-52.966-52.966
        c0-29.255,23.711-52.966,52.966-52.966S158.897,398.885,158.897,428.14"/>
</g>
<path style="fill:#83BEEB;" d="M459.034,339.864H512v-17.655h-52.966V339.864z"/>
<g>
    <path style="fill:#A2A2A2;" d="M123.586,428.14c0,9.754-7.901,17.655-17.655,17.655s-17.655-7.901-17.655-17.655
        c0-9.754,7.901-17.655,17.655-17.655S123.586,418.385,123.586,428.14"/>
    <path style="fill:#A2A2A2;" d="M432.552,428.14c0,9.754-7.901,17.655-17.655,17.655s-17.655-7.901-17.655-17.655
        c0-9.754,7.901-17.655,17.655-17.655S432.552,418.385,432.552,428.14"/>
</g>
<path style="fill:#FEC24B;" d="M353.103,110.347h-70.621c-4.882,0-8.828-3.946-8.828-8.828c0-4.882,3.946-8.828,8.828-8.828h70.621
    c4.882,0,8.828,3.946,8.828,8.828C361.931,106.401,357.985,110.347,353.103,110.347"/>
<g>
    <path style="fill:#D1942A;" d="M3.531,216.278h102.4c4.882,0,8.828-3.946,8.828-8.828s-3.946-8.828-8.828-8.828H10.593
        L3.531,216.278z"/>
    <path style="fill:#D1942A;" d="M8.149,260.416H150.07c4.882,0,8.828-3.946,8.828-8.828c0-4.882-3.946-8.828-8.828-8.828H4.07
        L8.149,260.416z"/>
</g>
</g>"#;
pub fn vis(input: &Input, out: &[Action]) -> VisResult {
    let D = 600 / input.N;
    let W = D * input.N;
    let H = D * input.N;
    let (score, err, (h, pi, pj, v, cost, diff, last_dir)) = compute_score_details(input, &out);
    let mut doc = svg::Document::new()
        .set("id", "vis")
        .set("viewBox", (-5, -5, W + 10, H + 10))
        .set("width", W + 10)
        .set("height", H + 10)
        .set("style", "background-color:white");
    doc = doc.add(Style::new(format!(
        "text {{text-anchor: middle;dominant-baseline: central;}}"
    )));
    doc = doc.add(
        Symbol::new()
            .set("id", "dump")
            .set("viewBox", (0, 0, 512, 512))
            .add(svg::node::Blob::new(DUMP)),
    );
    for i in 0..input.N {
        for j in 0..input.N {
            let c = if h[i][j] == 0 {
                0.5
            } else if h[i][j] < 0 {
                0.5 - ((-h[i][j] as f64).ln() * 0.1).min(0.5)
            } else {
                0.5 + ((h[i][j] as f64).ln() * 0.1).min(0.5)
            };
            doc = doc.add(group(format!("h[{},{}]={}", i, j, h[i][j])).add(rect(
                j * D,
                i * D,
                D,
                D,
                &color(c),
            )));
        }
    }
    let dump = Use::new()
        .set("x", pj * D + D / 8)
        .set("y", pi * D + D / 8)
        .set("width", D * 3 / 4)
        .set("height", D * 3 / 4)
        .set("href", "#dump");
    if last_dir == 0 {
        doc = doc.add(dump);
    } else if last_dir == 1 {
        doc = doc.add(dump.set(
            "transform",
            format!(
                "rotate(90, {}, {}) translate({}, {}) scale(1, -1) translate({}, {})",
                pj * D + D / 2,
                pi * D + D / 2,
                pj * D + D / 2,
                pi * D + D / 2,
                -((pj * D + D / 2) as i32),
                -((pi * D + D / 2) as i32)
            ),
        ));
    } else if last_dir == 2 {
        doc = doc.add(dump.set(
            "transform",
            format!(
                "translate({}, {}) scale(-1, 1) translate({}, {})",
                pj * D + D / 2,
                pi * D + D / 2,
                -((pj * D + D / 2) as i32),
                -((pi * D + D / 2) as i32)
            ),
        ));
    } else {
        doc = doc.add(dump.set(
            "transform",
            format!("rotate(270, {}, {})", pj * D + D / 2, pi * D + D / 2),
        ));
    }
    for i in 0..=input.N {
        doc = doc.add(
            Line::new()
                .set("x1", 0)
                .set("y1", i * D)
                .set("x2", W)
                .set("y2", i * D)
                .set("stroke", "gray")
                .set("stroke-width", 1),
        );
        doc = doc.add(
            Line::new()
                .set("x1", i * D)
                .set("y1", 0)
                .set("x2", i * D)
                .set("y2", H)
                .set("stroke", "gray")
                .set("stroke-width", 1),
        );
    }
    VisResult {
        score,
        err,
        svg: doc.to_string(),
        cost,
        v,
        diff,
    }
}
