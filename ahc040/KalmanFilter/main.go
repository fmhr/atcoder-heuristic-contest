package main

import "log"

// ロケットの高度を推定する
// parameters
var (
	a  float64 = 1.0 // xの拡大係数 (観測行列) H
	b  float64 = 1.0 // 拡大係数 (状態遷移行列) F
	c  float64 = 0.5 // ノイズの拡大係数 (プロセスノイズ共分散) Q
	q2 float64 = 1.0 // ノイズの分散 (プロセスノイズ共分散)
	m  float64 = 1.0 // ロケットのスピード 一定　(1.0m/s) u
)

type KalmanFilter struct {
	mean   float64 // 状態(高度) x
	sigma2 float64 // 分散 P
}

func NewKalmanFilter(mean float64, sigma2 float64) *KalmanFilter {
	return &KalmanFilter{
		mean:   mean,
		sigma2: sigma2,
	}
}

func (k *KalmanFilter) Update(measured float64) {
	// 事前分布
	mu1 := b*k.mean + m
	variance := b*b*k.sigma2 + c*c*q2
	//log.Printf("事前分布 %.4f  %.4f\n", mu1, variance)
	// 事後分布
	k.mean = (mu1*q2 + a*variance*measured) / (q2 + a*a*variance)
	k.sigma2 = variance * q2 / (q2 + a*a*variance)
	log.Printf("事後分布 %.4f  %.4f\n", k.mean, k.sigma2)
}

func main() {
	kf := NewKalmanFilter(0.0, 1.0)
	kf.Update(1.27)
	kf.Update(1.58)
	kf.Update(3.71)
	kf.Update(3.51)
	kf.Update(5.07)
	kf.Update(5.91)
	kf.Update(7.92)
	kf.Update(8.02)
	kf.Update(8.24)
	kf.Update(9.63)
	kf.Update(11.13)
	kf.Update(11.88)
	kf.Update(12.54)
	kf.Update(13.32)
	kf.Update(15.58)
	kf.Update(16.40)
	kf.Update(16.63)
	kf.Update(18.60)
	kf.Update(19.50)
	kf.Update(19.69)
}
