package utils

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Unzip(zipFile, destDir string) (err error) {
	var zipReader *zip.ReadCloser
	if zipReader, err = zip.OpenReader(zipFile); err != nil {
		return
	}
	defer zipReader.Close()
	for _, f := range zipReader.File {
		var fPath = filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(fPath, os.ModePerm); err != nil {
				return
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
				return
			}
			var inFile io.ReadCloser
			if inFile, err = f.Open(); err == nil {
				var outFile *os.File
				if outFile, err = os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()); err == nil {
					_, err = io.Copy(outFile, inFile)
					_ = inFile.Close()
					_ = outFile.Close()
					if err != nil {
						return
					}
				} else {
					_ = inFile.Close()
					return
				}
			} else {
				return
			}
		}
	}
	return
}

func ExtractGzip(gzipFile, destDir string) (err error) {
	var fp *os.File
	fp, err = os.Open(gzipFile)
	if err != nil {
		return
	}
	var gzipReader *gzip.Reader
	gzipReader, err = gzip.NewReader(fp)
	if err != nil {
		return
	}
	var tarReader = tar.NewReader(gzipReader)
	for true {
		var header *tar.Header
		header, err = tarReader.Next()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
		switch header.Typeflag {
		case tar.TypeDir:
			var dir = filepath.Join(destDir, header.Name)
			if _, err = os.Stat(dir); err != nil && os.IsNotExist(err) {
				if err = os.MkdirAll(dir, 0755); err != nil {
					return
				}
			}
		case tar.TypeReg:
			var destFile = filepath.Join(destDir, header.Name)
			var dir = filepath.Dir(destFile)
			if _, err = os.Stat(dir); err != nil && os.IsNotExist(err) {
				if err = os.MkdirAll(dir, 0755); err != nil {
					return
				}
			}
			var outFile *os.File
			outFile, err = os.Create(destFile)
			if err != nil {
				return
			}
			if _, err = io.Copy(outFile, tarReader); err != nil {
				return
			}
			_ = outFile.Close()
		default:
			err = errors.New(fmt.Sprintf("unknown type: %d in %s", header.Typeflag, header.Name))
		}
	}
	return
}

func CopyFiles(src, dest string) (err error) {
	var srcInfo os.FileInfo
	srcInfo, err = os.Stat(src)
	if err != nil {
		return
	}
	if srcInfo.IsDir() {
		var list []fs.FileInfo
		if list, err = ioutil.ReadDir(src); err == nil {
			for _, item := range list {
				if err = CopyFiles(filepath.Join(src, item.Name()), filepath.Join(dest, item.Name())); err != nil {
					return
				}
			}
		}
	} else {
		var destDir = filepath.Dir(dest)
		if _, err = os.Stat(destDir); err != nil {
			if err = os.MkdirAll(destDir, 0755); err != nil {
				return
			}
		}
		var srcFile *os.File
		srcFile, err = os.Open(src)
		if err != nil {
			return
		}
		defer srcFile.Close()
		var bufReader = bufio.NewReader(srcFile)
		var destFile *os.File
		destFile, err = os.Create(dest)
		if err != nil {
			return
		}
		defer destFile.Close()
		_, err = io.Copy(destFile, bufReader)
	}
	return
}

func ReadJsonFile(filename string, content interface{}) (ok bool) {
	if info, err := os.Stat(filename); err == nil && !info.IsDir() && info.Size() > 0 {
		var buffer []byte
		if buffer, err = ioutil.ReadFile(filename); err == nil {
			err = json.Unmarshal(buffer, &content)
			ok = err == nil
		}
	}
	return
}

func WriteJsonFile(filename string, v interface{}) (err error) {
	var info os.FileInfo
	if info, err = os.Stat(filename); err == nil || os.IsNotExist(err) || !info.IsDir() {
		var perm os.FileMode
		if info != nil {
			perm = info.Mode()
		}
		if perm < 1 {
			perm = 0644
		}
		var buffer []byte
		if buffer, err = json.Marshal(v); err == nil {
			err = ioutil.WriteFile(filename, buffer, perm)
		}
	}
	return
}

func ReadTextFile(filename string) (content string, ok bool) {
	if info, err := os.Stat(filename); err == nil && !info.IsDir() && info.Size() > 0 {
		var buffer []byte
		if buffer, err = ioutil.ReadFile(filename); err == nil {
			content = string(buffer)
			ok = true
		}
	}
	return
}

func WriteTextFile(filename string, content string) (err error) {
	var info os.FileInfo
	if info, err = os.Stat(filename); err == nil || os.IsNotExist(err) || !info.IsDir() {
		var perm os.FileMode
		if info != nil {
			perm = info.Mode()
		}
		if perm < 1 {
			perm = 0644
		}
		err = ioutil.WriteFile(filename, []byte(content), info.Mode())
	}
	return
}
