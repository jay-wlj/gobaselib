package base

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jie123108/glog"
)

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index+1]
}

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// 去掉文件名后面的后缀。
func TrimExt(filename string) string {
	ext := path.Ext(filename)
	return filename[0 : len(filename)-len(ext)]
}

func Md5SumFile(filename string, bufsize uint) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if bufsize == 0 {
		bufsize = 1024 * 1024
	}

	// calculate the file size
	info, _ := file.Stat()
	filesize := info.Size()
	blocks := uint(math.Ceil(float64(filesize) / float64(bufsize)))

	hash := md5.New()
	for i := uint(0); i < blocks; i++ {
		blocksize := int(math.Min(float64(bufsize), float64(filesize-int64(i*bufsize))))
		buf := make([]byte, blocksize)
		file.Read(buf)
		io.WriteString(hash, string(buf)) // append into the hash
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func MkdirAll(dir string) error {
	if IsExist(dir) {
		return nil
	}
	return os.MkdirAll(dir, os.ModePerm)
}

func Copy(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func SafeCopy(src, dest string) error {
	tmpfile := dest + ".tmp"
	err := Copy(src, tmpfile)
	if err != nil {
		return err
	}
	return os.Rename(tmpfile, dest)
}

func Rename(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err != nil && strings.HasSuffix(err.Error(), "invalid cross-device link") {
		err = SafeCopy(oldpath, newpath)
		if err == nil { //成功了，删除旧文件。
			err = os.Remove(oldpath)
		}
	}
	return err
}

func ReadLines(filename string) ([]string, error) {
	if !IsExist(filename) {
		return nil, os.ErrNotExist
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func ListFilesEx(inputDir string, exts map[string]bool, grep string) ([]string, error) {
	fileList := []string{}
	if !IsExist(inputDir) {
		return fileList, nil
	}

	err := filepath.Walk(inputDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			ext := filepath.Ext(path) //取后缀.mkv,...
			if len(ext) > 1 {
				ext = ext[1:]
			}
			if exts[ext] { //存在该后缀
				if grep == "" { //正则表达式匹配条件，为空，不需要条件匹配
					fileList = append(fileList, path)
				} else {
					matched, _ := regexp.MatchString(grep, path) //正则表达式匹配条件
					if matched {                                 //匹配成功，添加文件
						fileList = append(fileList, path)
					}
				}
			} else {
				//glog.Errorf("unproc file: %s ext: %s\n", path, ext)
			}
		}
		return nil
	})
	if err != nil {
		return fileList, err
	}
	return fileList, nil
}

//add by zwu
func ListDirName(inputDir, grep string) ([]string, error) {
	fileList := []string{}
	if !IsExist(inputDir) {
		return fileList, nil
	}

	glog.Infof("ListDir inputDir:%s, grep:%s\n", inputDir, grep)
	err := filepath.Walk(inputDir, func(path string, f os.FileInfo, err error) error {
		//glog.Infof("ListDir path:%s\n", path)
		if f.IsDir() {
			//offset := strings.LastIndex(path, "\\")
			//if offset >= 0 {
			dirname := path
			glog.Infof("ListDir path:%s, dirname:%s\n", path, dirname)
			if dirname != "" {
				if grep == "" { //正则表达式匹配条件，为空，不需要条件匹配
					fileList = append(fileList, path)
				} else {
					matched, _ := regexp.MatchString(grep, path) //正则表达式匹配条件
					if matched {                                 //匹配成功，添加文件
						fileList = append(fileList, path)
					}
				}
			} else {
				//glog.Errorf("unproc file: %s ext: %s\n", path, ext)
			}
		}

		//}

		return nil
	})
	if err != nil {
		return fileList, err
	}
	return fileList, nil
}

//add end

func ListFiles(inputDir string, exts map[string]bool) ([]string, error) {
	return ListFilesEx(inputDir, exts, "")
}

func IsInExtList(fileurl string, exts map[string]bool) bool {
	ext := filepath.Ext(fileurl) //取后缀.mkv,...
	if len(ext) > 1 {
		ext = ext[1:]
	}
	if exts[ext] {
		return true
	}

	return false
}
