package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func GetSoundByURL(url string, collection string, name string) (file string, err error) {
	ext := filepath.Ext(url)
	//TODO check if it is a sound file (extention)
	err, download := DownloadFile(url, collection+"-"+name, ext)
	if err != nil {
		fmt.Println("error", err)
	}
	file = BASEPATH + "sounds/" + collection + "-" + name + ".dca"
	if ext != ".dca" {
		err = convertToDCA(download, file)
		os.Remove(download)
		if err != nil {
			return
		}
	} else {
		err = os.Rename(download, file)
		if err != nil {
			return
		}
	}

	return
}

func convertToDCA(inputFile string, outputFile string) (err error) {
	output, err := os.Create(outputFile)
	defer output.Close()
	if err != nil {
		fmt.Println("Error creating file", err)
		return
	}

	var out bytes.Buffer
	dca := exec.Command("dca", "-i", inputFile, "-raw", "true")
	dca.Stdout = &out
	dca.Start()
	if err != nil {
		fmt.Println("StartDCA Error:", err)
		return
	}

	err = dca.Wait()
	if err != nil {
		fmt.Println("DCA Error:", err)
		return
	}
	_, err = output.Write(out.Bytes())
	if err != nil {
		fmt.Println("Error writing", err)
	}
	return
}

func DownloadFile(url string, name string, ext string) (err error, file string) {
	//ddMMyyhhmmssff
	file = BASEPATH + "download/"
	os.MkdirAll(file, os.FileMode(int(0777)))
	file += name + ext
	//time.Now().Format("020106150405.00") + ext
	out, err := os.Create(file)
	defer out.Close()

	resp, err := http.Get(url)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return
}
