package main

import (
	"io"
	"net/http"
	"os"
)

// downloadFile は指定されたURLからファイルをダウンロードし、指定されたパスに保存します。
func downloadFile(URL, filepath string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
