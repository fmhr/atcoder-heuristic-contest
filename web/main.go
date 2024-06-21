package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/fmhr/fj"
)

// Templateをパース
var tmpl = template.Must(template.ParseFiles("form.html"))

func main() {
	st := http.FileServer(http.Dir("standing/"))
	http.Handle("/standing/", http.StripPrefix("/standing/", st))
	http.HandleFunc("/submit", formHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// フォームのデータを処理
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// フォームデータの取得
		username := r.FormValue("username")
		language := r.FormValue("language")
		source := r.FormValue("source")

		fmt.Fprintf(w, "Received: username=%s, language=%s \n", username, language)
		// 言語に応じた設定を取得
		conf, err := setConfig(language)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error setting config", http.StatusInternalServerError)
			return
		}
		// コンパイルサーバーに送信
		fmt.Fprintf(w, "Sending data to compile server\n")
		err = sendToCompileServer(username, language, source, &conf)
		if err != nil {
			// エラーが発生した場合、ユーザーにエラーメッセージを表示
			log.Println(err)
			http.Error(w, "Error sending data to compile server", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Compile Finishded\n")
		fmt.Fprintf(w, "Testing...\n")

		// ワーカーに送信
		log.Println("Sending data to worker")
		go sendWorker(conf, username)
	} else {
		// GETリクエストの場合、フォームを表示
		tmpl.Execute(w, nil)
	}
}

// setConfig は言語に応じた設定を返します
func setConfig(lang string) (fj.Config, error) {
	var conf fj.Config
	switch lang {
	case "Go":
		conf.Language = "Go"
		conf.CompilerURL = "https://compiler-go-k65yicvf4a-an.a.run.app/compiler"
	case "java":
		conf.Language = "java"
		//conf.CompilerURL = "https://compiler-java-k65yicvf4a-an.a.run.app/compiler"
		conf.CompilerURL = "https://compiler-go-k65yicvf4a-an.a.run.app/compiler"
		// javaはコンパイルせずに、ソースを保存してworkerでコンパイルする
	case "C#":
		conf.Language = "C#"
		conf.CompilerURL = "https://compiler-csharp-k65yicvf4a-an.a.run.app/compiler"
	default:
		return conf, fmt.Errorf("invalid language: %s", lang)
	}
	conf.Contest = "ahc032"
	conf.BinaryPath = fj.LanguageSets[conf.Language].BinaryPath
	conf.SourceFilePath = fj.LanguageSets[conf.Language].FileName
	conf.BinaryPath = fj.LanguageSets[conf.Language].BinaryPath
	conf.Reactive = false
	conf.TimeLimitMS = 4000
	conf.TesterPath = "tools/target/release/tester"
	conf.VisPath = "tools/target/release/vis"
	conf.GenPath = "tools/target/release/gen"
	conf.InfilePath = "tools/in/"
	conf.OutfilePath = "out/"
	conf.Jobs = 100
	conf.CloudMode = true
	conf.ConcurrentRequests = 100
	conf.Bucket = "ahc032"
	conf.WorkerURL = "https://ahc032-worker-k65yicvf4a-an.a.run.app/worker"
	return conf, nil
}
