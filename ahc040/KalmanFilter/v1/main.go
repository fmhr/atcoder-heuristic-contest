package main

import "log"

// ロケットの高度を推定する
// parameters
var (
	H float64 = 1.0 // 観測行列
	F float64 = 1.0 // 状態遷移行列
	Q float64 = 1.0 // プロセスノイズ共分散
	u float64 = 1.0 // 制御入力 (一定速度)
)

type KalmanFilter struct {
	x float64 // 状態 (高度)
	P float64 // 共分散
}

func NewKalmanFilter(initialX float64, initialP float64) *KalmanFilter {
	return &KalmanFilter{
		x: initialX,
		P: initialP,
	}
}

func (kf *KalmanFilter) Update(z float64) {
	// 予測ステップ (事前分布)
	xPred := F*kf.x + u
	Ppred := F*F*kf.P + Q
	//log.Printf("事前分布 %.4f  %.4f\n", xPred, Ppred)

	// カルマンゲイン
	K := H * Ppred / (H*H*Ppred + Q)
	log.Printf("カルマンゲイン %.4f\n", K)

	// 更新ステップ (事後分布)
	kf.x = xPred + K*(z-H*xPred)
	kf.P = (1 - K*H) * Ppred
	log.Printf("事後分布 %.4f  %.4f\n", kf.x, kf.P)
}

func main() {
	kf := NewKalmanFilter(0.0, 1.0)
	measurements := []float64{1.27, 1.58, 3.71, 3.51, 5.07, 5.91, 7.92, 8.02, 8.24, 9.63, 11.13, 11.88, 12.54, 13.32, 15.58, 16.40, 16.63, 18.60, 19.50, 19.69}
	for _, z := range measurements {
		kf.Update(z)
	}
}
