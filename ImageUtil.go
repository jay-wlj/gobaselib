package base

import (
	"bufio"
	"bytes"

	"github.com/jie123108/glog"

	// "fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/jie123108/imaging"
)

type Size struct {
	Width, Height int
}

func _getImageSize(reader io.Reader) (*Size, error) {
	var img, _, err = image.Decode(reader)
	if err != nil {
		return nil, err
	}
	size := &Size{img.Bounds().Dx(), img.Bounds().Dy()}
	return size, nil
}

func GetImageSizeF(path string) (*Size, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return _getImageSize(file)
}

func GetImageSize(imgContent []byte) (*Size, error) {
	return _getImageSize(bytes.NewReader(imgContent))
}

/**
src size: 160x90
input size: 200x300, output: 60x90
input size: 200x100, output: 160x80
**/
func ResizeSmaller(srcWidth, srcHeight, width, height int) (new_width int, new_height int) {
	if width < srcWidth && height < srcHeight {
		new_width, new_height = width, height
		return
	}
	ratioW := float64(width) / float64(srcWidth)
	ratioH := float64(height) / float64(srcHeight)
	if ratioW >= ratioH {
		new_width = srcWidth
		new_height = srcWidth * height / width
	} else {
		new_height = srcHeight
		new_width = srcHeight * width / height
	}
	return
}

func ResizeImgToBytes(srcImg image.Image, filename string, width, height int, enlarge_smaller bool, quality int) (img_bytes []byte, err error) {
	srcBound := srcImg.Bounds()
	var new_img image.Image
	if width > 0 && height == 0 { // 按宽度调整图片尺寸。
		//如果不拉伸小图片，并且原图比目标尺寸小，则尺寸保持不变。
		if !enlarge_smaller && srcBound.Dx() < width {
			width = srcBound.Dx()
		}
		new_img = imaging.Resize(srcImg, width, height, imaging.Lanczos)
	} else if width == 0 && height > 0 { // 按高度调整图片尺寸。
		//如果不拉伸小图片，并且原图比目标尺寸小，则尺寸保持不变。
		if !enlarge_smaller && srcBound.Dy() < height {
			height = srcBound.Dy()
		}
		new_img = imaging.Resize(srcImg, width, height, imaging.Lanczos)
	} else { //调整尺寸并裁剪图片
		if !enlarge_smaller {
			width, height = ResizeSmaller(srcBound.Dx(), srcBound.Dy(), width, height)
		}
		new_img = imaging.Fill(srcImg, width, height, imaging.Center, imaging.Lanczos)
	}

	var buf bytes.Buffer
	buf_writer := bufio.NewWriter(&buf)
	//err = imaging.Encode(buf_writer, new_img, imaging.JPEG, quality)
	err = imaging.Encode(buf_writer, new_img, imaging.JPEG)
	if err != nil {
		glog.Errorf("imaging.Encode(src: %s, width: %d, height: %d) failed! err: %v",
			filename, width, height, err)
		return
	}
	buf_writer.Flush()
	img_bytes = buf.Bytes()
	return
}

func ResizeBytesImgToBytes(srcBytes []byte, filename string, width, height int, enlarge_smaller bool, quality int) (img_bytes []byte, err error) {
	reader := bytes.NewReader(srcBytes)
	var srcImg image.Image
	srcImg, err = imaging.Decode(reader)
	if err != nil {
		glog.Errorf("imaging.Decode(%s) failed! err: %v", filename, err)
		return
	}
	img_bytes, err = ResizeImgToBytes(srcImg, filename, width, height, enlarge_smaller, quality)
	return
}

func ResizeFileToFile(src string, dst string, width int, height int, enlarge_smaller bool, quality int) error {
	srcImg, err := imaging.Open(src)
	if err != nil {
		glog.Errorf("Open Img File [%s] failed! err: %v", src, err)
		return err
	}

	dstImg, err := ResizeImgToBytes(srcImg, src, width, height, enlarge_smaller, quality)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, dstImg, os.ModePerm)
	if err != nil {
		glog.Errorf("Save Img To File [%s] failed! err: %v", dst, err)
		return err
	}
	return nil
}
