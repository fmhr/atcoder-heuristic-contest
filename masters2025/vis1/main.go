package main

import (
	"fmt"
	"html/template"
	"os"

	svg "github.com/ajstarks/svgo"
)

func main() {
	file, _ := os.Create("slider.svg")
	defer file.Close()

	canvas := svg.New(file)
	canvas.Start(400, 100)

	// 背景のバー
	canvas.Rect(50, 40, 300, 20, "fill:#ccc;stroke:black;stroke-width:2")

	// スライダーのつまみ（IDを設定、title属性を追加）
	cx := 200 // つまみの初期位置
	titleText := fmt.Sprintf("Position: %d", cx)

	// HTMLエスケープを適用
	escapedTitleText := template.HTMLEscapeString(titleText)

	// <g>グループを作成し、circle と title をまとめる
	canvas.Gid("sliderKnob")
	canvas.Circle(cx, 50, 10, "fill:blue;stroke:black;stroke-width:2")
	canvas.Title(escapedTitleText) // title を circle の中に配置
	canvas.Gend()

	canvas.End()
}
