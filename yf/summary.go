package yf

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func Sha256hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Sha1hex(data []byte) string {
	h := sha1.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Md5hex(data []byte) string {
	h := md5.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
func Md5Hex(data []byte) string {
	h := md5.New()
	h.Write(data)
	bs := h.Sum(nil)
	return fmt.Sprintf("%X", bs)
}

func Md5Reader(reader io.Reader) (string, int64, error) {
	h := md5.New()
	written, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), written, nil
}

func Sha1Reader(reader io.Reader) (string, int64, error) {
	h := sha1.New()
	written, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), written, nil
}

// 耐飞HASH算法: 小于40K的文件，计算整个文件的hash,大于40K的文件，计算文件头20K及文件尾部20K的数据。
func NFSimpleMd5(reader io.ReadSeeker, filesize int64) (string, int64, error) {
	var simple_md5_block = int64(1024 * 20)
	if filesize <= simple_md5_block*2 {
		return Md5Reader(reader)
	}

	h := md5.New()
	_, err := io.CopyN(h, reader, simple_md5_block)
	if err != nil {
		return "", 0, err
	}
	_, err = reader.Seek(simple_md5_block*-1, io.SeekEnd)
	if err != nil {
		return "", 0, err
	}

	_, err = io.CopyN(h, reader, simple_md5_block)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), filesize, nil
}

func NFSimpleMd5File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	return NFSimpleMd5(file, fi.Size())
}
func NFStrongSimpleMd5File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	return NFStrongSimpleMd5(file, fi.Size())
}

// 耐飞增强HASH算法: 小于60K的文件，计算整个文件的hash,大于60K的文件，计算文件头20K及文件中部20K和文件尾部20K的数据。
func NFStrongSimpleMd5(reader io.ReadSeeker, filesize int64) (string, int64, error) {
	var simple_md5_block = int64(1024 * 20)
	if filesize <= simple_md5_block*3 {
		return Md5Reader(reader)
	}

	h := md5.New()
	_, err := io.CopyN(h, reader, simple_md5_block)
	if err != nil {
		return "", 0, err
	}
	//中部前端
	halfSize := int64(filesize / 2)
	halfBlockSize := int64(simple_md5_block / 2)
	_, err = reader.Seek(halfSize-halfBlockSize, io.SeekStart)
	if err != nil {
		return "", 0, err
	}
	_, err = io.CopyN(h, reader, simple_md5_block)
	if err != nil {
		return "", 0, err
	}
	//尾部
	_, err = reader.Seek(simple_md5_block*-1, io.SeekEnd)
	if err != nil {
		return "", 0, err
	}

	_, err = io.CopyN(h, reader, simple_md5_block)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), filesize, nil
}

func Md5File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	return Md5Reader(file)
}

func Sha1File(path string) (string, int64, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", 0, err
	}

	return Sha1Reader(file)
}
