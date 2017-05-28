package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)

func LoadObjectFromJsonFile(path string, object interface{}) (err error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Errorf("Error opening file: &s > %s", path, err)
		return
	}
	err = json.NewDecoder(file).Decode(object)
	if err != nil {
		log.Errorf("Error decoding json file stream: %s > %s", path, err)
		return
	}
	return
}

func SaveObjectAsJsonToFile(path string, object interface{}) (err error) {
	_, err = isFile(path)
	if err != nil {
		if os.IsExist(err) {
			log.Errorf("Error accessing the file: %s > %s", path, err)
			return
		}
		dirPath := filepath.Dir(path)
		os.MkdirAll(dirPath, os.FileMode(OS_ALL_R|OS_USER_RW))
	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		log.Errorf("Error creating/truncating file: %s > %s", path, err)
		return
	}
	err = json.NewEncoder(file).Encode(object)
	if err != nil {
		log.Errorf("Error encoding json to file stream: %s > %s", path, err)
		return
	}
	return

}

func isFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.Mode().IsRegular(), nil
}

func isDirectory(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.Mode().IsDir(), nil
}

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

func GetSoundByURL(url string, path string, fileName string) (file string, err error) {
	//TODO check if it is a sound file (extention)
	download, err := DownloadFile(url, path, fileName)
	if err != nil {
		log.Errorf("Error downloading file: %s", url)
	}
	file = filepath.Join(path, fileName+".dca")
	ext := filepath.Ext(download)
	if ext != ".dca" {
		err = convertToDCA(download, file)
		if err != nil {
			log.Errorf("Error converting sound to .dca > %s", err)
			return
		}
		err = os.Remove(download)
		if err != nil {
			log.Noticef("Couldn't remove file: %s", download)
		}

	} else {
		err = os.Rename(download, file)
		if err != nil {
			log.Errorf("Error renaming file: %s", download)
			return
		}
	}

	return
}

func convertToDCA(inputFile string, outputFile string) (err error) {
	dirPath := filepath.Dir(outputFile)
	os.MkdirAll(dirPath, os.FileMode(OS_ALL_R|OS_USER_RW))
	output, err := os.Create(outputFile)
	defer output.Close()
	if err != nil {
		log.Errorf("Error creating file > %s", err)
		return
	}

	var out bytes.Buffer
	dca := exec.Command("dca", "-i", inputFile, "-raw", "true")
	dca.Stdout = &out
	dca.Start()
	if err != nil {
		log.Errorf("Error starting dca > ", err)
		return
	}

	err = dca.Wait()
	if err != nil {
		log.Errorf("Error from dca > ", err)
		return
	}
	_, err = output.Write(out.Bytes())
	if err != nil {
		log.Errorf("Error writing to file: %s > %s", outputFile, err)
	}
	return
}

func DownloadFile(url string, dirPath string, name string) (file string, err error) {
	ext := filepath.Ext(url)
	os.MkdirAll(dirPath, os.FileMode(OS_ALL_R|OS_USER_RW))
	file = filepath.Join(dirPath, name+ext)
	//ddMMyyhhmmssff
	//time.Now().Format("020106150405.00") + ext
	out, err := os.Create(file)
	defer out.Close()

	resp, err := http.Get(url)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return
}
