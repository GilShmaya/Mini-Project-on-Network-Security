package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// CreateTmpFolder check if temp folder exists, if not, creates it
func CreateTmpFolder(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}
}

// FileCopy copy file to destination if exist
func FileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("A error occured while copying the files: %s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func ListDirRecursively(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false && info.Name() != ".DS_Store" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

// GetFileNameFromPath extract file name from path
func GetFileNameFromPath(path string, isWin bool) string {
	separator := "/"
	if isWin {
		separator = "\\"
	}
	s := strings.Split(path, separator)
	return s[len(s)-1]
}

// GetCurrentUser extract current user
func GetCurrentUser() (*user.User, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return usr, err
}

// ZipFiles compress files into zip
func ZipFiles(filesToZip []string, outFilePath string, isWin bool) {
	// Get a Buffer to Write To
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)
	defer w.Close()

	for _, filePath := range filesToZip {
		// read file data
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Add some files to the archive.
		filename := GetFileNameFromPath(filePath, isWin)
		f, err := w.Create(filename)
		if err != nil {
			continue
		}
		_, err = f.Write(data)
		if err != nil {
			continue
		}
	}
}

// GenerateRandomString generate a random string in length n from letters
func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// CreateReadme create a readme file
func CreateReadme(readmeFile string, content string) {
	f, err := os.Create(readmeFile)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(content)
}

func SendFiles(url string, fileToSend string) {
	file, err := os.Open(fileToSend)
	if err != nil {
		panic(err)
	}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"my_data": file,
	}
	err = Upload(url, values)
	if err != nil {
		// fmt.Println("Could not connect to server")
		os.Exit(0)
		// panic(err)
	}
}

func Upload(url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}

	w.Close()

	// submit the form to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

// PostJson send a Post request
func PostJson(url string, jsonString string, username string) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonString)))
	req.Header.Set("Username", username)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}