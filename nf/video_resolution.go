package base

import (
	"fmt"
	"strings"
	"github.com/jie123108/glog"
)

/**
{"4:3": []VInfo}

**/

var FIX_HEIGHT int = 0
var FIX_WIDTH int = 1

type VInfo struct {
	Width, Height, ABitRate, VBitRate int
	FixType                           int    // 0 固定高度，1：固定宽度
	Name                              string //分辨率类型
}

func (this VInfo) FixHeight() bool {
	return this.FixType == FIX_HEIGHT
}

var video_prop = 900
var audio_prop = 1024 * 8
var RESOLUTIONS = make(map[string][]VInfo)

var AUDIO_PROPS = make(map[string]int)
var VIDEO_PROPS = make(map[string]int)

func AudioProp(name string) int {
	prop := AUDIO_PROPS[name]
	if prop == 0 {
		prop = audio_prop
	}
	return prop
}

func VideoProp(name string) int {
	prop := VIDEO_PROPS[name]
	if prop == 0 {
		prop = video_prop
	}
	return prop
}

//比例大于16:9的影片，使用固定宽度。
var widths = []int{1920, 1280, 960}

//比例小于等于16:9的，使用固定高度。
var heights = []int{1080, 720, 540, 360} //不同分辨率的视频的高(依次是:1080p,720p,高清,标清)

var SpecNames = []string{"1080p", "720p", "540p", "360p"}

func ConsVInfoByHeight(Height, w, h int, name string) VInfo {
	Width := Height * w / h
	ABitRate := Width * Height / AudioProp(name)
	VBitRate := Width * Height / VideoProp(name)
	return VInfo{Width, Height, ABitRate, VBitRate, FIX_HEIGHT, name}
}
func ConsVInfoByWidth(Width, w, h int, name string) VInfo {
	Height := Width * h / w
	ABitRate := Width * Height / AudioProp(name)
	VBitRate := Width * Height / VideoProp(name)
	return VInfo{Width, Height, ABitRate, VBitRate, FIX_WIDTH, name}
}

func get_vinfo_by_dar_ex(aspect_ratio string, width, height int) []VInfo {
	//glog.Infof("get_vinfo_by_dar_ex aspect_ratio:%s, width:%d, height:%d\n", aspect_ratio, width, height)
	allinfos, ok := RESOLUTIONS[aspect_ratio]
	if ok {
		glog.Infof("get_vinfo_by_dar_ex len(allinfos):%d\n", len(allinfos))
		return allinfos
	}

	arr := strings.SplitN(aspect_ratio, ":", 2)
	if len(arr) != 2 {
		glog.Infof("get_vinfo_by_dar_ex strings.SplitN(aspect_ratio:%s, :, 2) failed\n", aspect_ratio)
		return nil
	}
	w, h := Atoi(arr[0]), Atoi(arr[1])
	var infos = make([]VInfo, len(heights))

	if float32(float32(w)/float32(h)) <= 16.0/9.0+0.001 { //比例小于等于16:9
		for i, Height := range heights {
			infos[i] = ConsVInfoByHeight(Height, w, h, SpecNames[i])
		}
	} else { //比例大于16:9的影片，使用固定宽度。
		iw := len(widths)
		for i := 0; i < len(heights); i++ {
			if i < iw {
				Width := widths[i]
				infos[i] = ConsVInfoByWidth(Width, w, h, SpecNames[i])
			} else {
				Height := heights[i]
				infos[i] = ConsVInfoByHeight(Height, w, h, SpecNames[i])
			}
		}
	}
	RESOLUTIONS[aspect_ratio] = infos
	//glog.Infof("get_vinfo_by_dar_ex len(infos):%d\n", len(infos))
	return infos
}

func get_vinfo_by_dar(aspect_ratio string) []VInfo {
	return get_vinfo_by_dar_ex(aspect_ratio, 400, 300)
}

func init() {
	aspect_ratios := []string{"4:3", "16:9", "17:9", "21:9", "22:9"}
	// "1080p", "720p", "540p", "360p"
	AUDIO_PROPS["1080p"] = 1024 * 8
	AUDIO_PROPS["720p"] = 1024 * 7
	AUDIO_PROPS["540p"] = 1024 * 6
	AUDIO_PROPS["360p"] = 1024 * 6
	VIDEO_PROPS["720p"] = 700

	for _, dar := range aspect_ratios {
		get_vinfo_by_dar(dar)
	}
}

func ShowResolutions() {
	for aspect_ratio, infos := range RESOLUTIONS {
		fmt.Printf("%s W x H, VBitRate, ABitRate ------\n", aspect_ratio)
		for _, vinfo := range infos {
			fmt.Printf("    %s: %dx%d,%dkb/s,%dkb/s\n", vinfo.Name, vinfo.Width, vinfo.Height, vinfo.VBitRate, vinfo.ABitRate)
		}
	}
}

//获取当前视频，需要转码的分辨率,码率(只获取比输入视频小的分辨率)
func GetResolutions(aspect_ratio string, width, height int) (infos []VInfo) {
	allinfos := get_vinfo_by_dar_ex(aspect_ratio, width, height)
	glog.Infof("GetResolutions len(allinfos):%d\n", len(allinfos))
	infos = make([]VInfo, 0)
	for _, info := range allinfos {
		// fmt.Printf("width(%d) >= info.Width(%d) && height(%d) >= info.Height(%d): %v\n",
		// 	width, info.Width, height, info.Height, width >= info.Width && height >= info.Height)
		if info.FixHeight() && ((height >= info.Height) || (info.Height == 540 && height >= 495)) {
			infos = append(infos, info)
		} else if !info.FixHeight() && ((width >= info.Width) || (info.Width == 960 && width >= 880)) {
			infos = append(infos, info)
		}
	}
	glog.Infof(" len(infos):%d == 0 && len(allinfos):%d\n",  len(infos), len(allinfos))
	if len(infos) == 0 && len(allinfos) >= 1{
		infos = append(infos, allinfos[len(allinfos)-1])
	}
	return infos
}

func GetResolutionByName(aspect_ratio string, name string) *VInfo {
	allinfos := get_vinfo_by_dar(aspect_ratio)
	for _, info := range allinfos {
		if info.Name == name {
			return &info
		}
	}
	return nil
}
