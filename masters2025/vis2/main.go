package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// GenerateSliderSVG は指定した cx の位置にスライダーを配置した SVG を生成する関数
func GenerateSliderSVG(cx int) string {
	// SVGを生成
	return fmt.Sprintf(`
	<svg width="400" height="100">
		<!-- 背景のバー -->
		<rect x="50" y="40" width="300" height="20" fill="#ccc" stroke="black" stroke-width="2"></rect>
		
		<!-- スライダーのつまみ -->
		<g id="sliderKnob">
			<circle id="knob" cx="%d" cy="50" r="10" fill="blue" stroke="black" stroke-width="2"></circle>
			<text x="%d" y="30" font-size="14" text-anchor="middle">Position: %d</text>
		</g>
	</svg>
	`, cx, cx, cx)
}

// HTMLテンプレート
const tmpl = `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SVGスライダー</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            text-align: center;
        }
        svg {
            border: 1px solid black;
        }
    </style>
</head>
<body>

    <h2>SVGスライダー</h2>
    <input type="range" id="slider" min="50" max="350" value="200" oninput="updateSlider(this.value)">
    <p>位置: <span id="position">200</span></p>

    <!-- SVGを表示 -->
    <div id="svgContainer">
        {{.SVG}}
    </div>

    <script>
        function updateSlider(value) {
            document.getElementById("position").innerText = value;
            document.getElementById("svgContainer").innerHTML = '{{.SVG}}'.replace('200', value).replace('200', value).replace('Position: 200', 'Position: ' + value);
        }
    </script>

</body>
</html>
`

// ハンドラー関数
func handler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("webpage").Parse(tmpl))

	// 初期値を 200 に設定
	data := struct {
		SVG string
	}{
		SVG: GenerateSliderSVG(200), // 初期値としてcx=200を渡す
	}

	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("サーバーが http://localhost:8080 で起動しました")
	http.ListenAndServe(":8080", nil)
}
