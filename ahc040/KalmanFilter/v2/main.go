package main

import (
	"fmt"
	"log"

	"gonum.org/v1/gonum/mat"
)

// ロケットの高度を推定する (加速度を考慮)
// parameters
var (
	dt float64 = 1.0 // 時間間隔
	q  float64 = 0.1 // 加速度ノイズの分散
	r  float64 = 1.0 // 観測ノイズの分散
)

type KalmanFilter struct {
	x *mat.VecDense // 状態ベクトル [高度, 速度]
	P *mat.Dense    // 共分散行列
}

func NewKalmanFilter(initialX, initialV, initialP float64) *KalmanFilter {
	x := mat.NewVecDense(2, []float64{initialX, initialV})
	P := mat.NewDense(2, 2, []float64{
		initialP, 0,
		0, initialP,
	})
	return &KalmanFilter{x: x, P: P}
}

func (kf *KalmanFilter) Update(z float64, a float64) {
	// 行列の定義
	F := mat.NewDense(2, 2, []float64{
		1, dt,
		0, 1,
	})
	B := mat.NewVecDense(2, []float64{(dt * dt) / 2, dt})
	H := mat.NewDense(1, 2, []float64{1, 0})
	Q := mat.NewDense(2, 2, []float64{
		(dt * dt * dt * dt) / 4 * q, (dt * dt * dt) / 2 * q,
		(dt * dt * dt) / 2 * q, dt * dt * q,
	})
	R := mat.NewDense(1, 1, []float64{r})

	// 予測ステップ
	xPred := mat.NewVecDense(2, nil)
	Fmulx := mat.NewVecDense(2, nil)
	Fmulx.MulVec(F, kf.x)
	BmulU := mat.NewVecDense(2, nil)
	BmulU.ScaleVec(a, B)
	xPred.AddVec(Fmulx, BmulU) // xPred = F*x + B*u

	Ppred := mat.NewDense(2, 2, nil)
	FP := mat.NewDense(2, 2, nil)
	FP.Mul(F, kf.P)
	Ppred.Mul(FP, F.T())
	Ppred.Add(Ppred, Q) // Ppred = F*P*F.T + Q

	// 更新ステップ
	y := mat.NewVecDense(1, nil)
	Hx := mat.NewVecDense(1, nil)
	Hx.MulVec(H, xPred)
	y.SubVec(mat.NewVecDense(1, []float64{z}), Hx) // y = z - H*xPred

	S := mat.NewDense(1, 1, nil)
	HPpred := mat.NewDense(1, 2, nil)
	HPpred.Mul(H, Ppred)
	S.Mul(HPpred, H.T())
	S.Add(S, R) // S = H*Ppred*H.T + R

	K := mat.NewDense(2, 1, nil)
	PpredHT := mat.NewDense(2, 1, nil)
	PpredHT.Mul(Ppred, H.T())
	sInv := mat.NewDense(1, 1, nil)
	sInv.Inverse(S)
	K.Mul(PpredHT, sInv) // K = Ppred * H.T * S^-1

	KmulY := mat.NewVecDense(2, nil)
	KmulY.MulVec(K, y)
	kf.x.AddVec(xPred, KmulY) // x = xPred + K*y

	kh := mat.NewDense(2, 2, nil)
	kh.Mul(K, H)
	I := mat.NewDense(2, 2, []float64{
		1, 0,
		0, 1,
	})
	ikH := mat.NewDense(2, 2, nil)
	ikH.Sub(I, kh)
	kf.P.Mul(ikH, Ppred) // P = (I - K*H) * Ppred

	log.Printf("事後分布\n%v", mat.Formatted(kf.x))
	log.Printf("共分散\n%v\n\n", mat.Formatted(kf.P))
}

func main() {
	log.SetFlags(log.Lshortfile)
	kf := NewKalmanFilter(0.0, 0.0, 1.0) // 初期位置0, 初期速度1, 初期共分散1
	measurements := []float64{1.27, 1.58, 3.71, 3.51, 5.07, 5.91, 7.92, 8.02, 8.24, 9.63, 11.13, 11.88, 12.54, 13.32, 15.58, 16.40, 16.63, 18.60, 19.50, 19.69}
	accelerations := []float64{0.1, 0.2, 0.1, 0.05, 0.1, 0.2, 0.1, 0.05, 0.1, 0.2, 0.1, 0.05, 0.1, 0.2, 0.1, 0.05, 0.1, 0.2, 0.1, 0.05}
	for i, z := range measurements {
		kf.Update(z, accelerations[i])
	}
	fmt.Println("Final Estimated State:", mat.Formatted(kf.x))
	fmt.Println("Final Estimated Covariance:", mat.Formatted(kf.P))
}
