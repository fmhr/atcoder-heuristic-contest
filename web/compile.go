package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/fmhr/fj"
)

// sendToCompileServer はフォームデータを外部のアプリサーバーに送信します。
func sendToCompileServer(username, language, source string, conf *fj.Config) error {
	// マルチパートフォームデータ用のバッファを用意
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	// "source" フィールドをファイルとして追加
	fw, err := mw.CreateFormFile("file", filepath.Base(conf.SourceFilePath))
	if err != nil {
		log.Println("Error creating form file:", err)
		return err
	}
	_, err = io.Copy(fw, bytes.NewBufferString(source))
	if err != nil {
		log.Println("Error copying source to form file:", err)
		return err
	}

	// "username" フィールドを追加
	if err := mw.WriteField("username", username); err != nil {
		log.Println("Error writing username field:", err)
		return err
	}

	// "language" フィールドを追加
	if err := mw.WriteField("language", language); err != nil {
		log.Println("Error writing language field:", err)
		return err
	}

	// "bucket" フィールドを追加
	if err := mw.WriteField("bucket", conf.Bucket); err != nil {
		log.Println("Error writing bucket field:", err)
		return err
	}

	// マルチパートライターを閉じてバッファを完了
	if err := mw.Close(); err != nil {
		log.Println("Error closing multipart writer:", err)
		return err
	}

	// 外部サーバーにPOSTリクエストを送信
	resp, err := http.Post(conf.CompilerURL, mw.FormDataContentType(), &b)
	if err != nil {
		log.Println("Error sending POST request:", err, resp.Body)
		return err
	}
	defer resp.Body.Close()

	// 応答を確認（ここでは単純にステータスコードをチェック）
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println(conf.CompilerURL)
		return fmt.Errorf("server responded with status: %d, body: %s", resp.StatusCode, body)
	}

	// バイナリファイルの名前を所得
	content := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(content)
	if err != nil {
		log.Println("Error parsing media type:", err)
		return err
	}
	filename, err := url.QueryUnescape(params["filename"])
	if err != nil {
		log.Println("Error unescaping filename:", err)
		return err
	}
	conf.TmpBinary = filename
	log.Println("Received binary file:", filename)
	return nil
}
