// +build freebsd netbsd openbsd linux

package gobaselib

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jie123108/glog"
	mediainfo "github.com/jie123108/go_mediainfo"
)

func ms2sec(ms int64) float64 {
	return float64(ms) / 1000.0
}
func byte2kb(b int64) int {
	return int(float64(b)/1000.0 + 0.5)
}

func Atof(str string) (f float64) {

	f, err := strconv.ParseFloat(str, 32)
	if err != nil {
		glog.Errorf("ParseFloat(%s) failed! err: %v", str, err)
		f = 0
	}

	return
}

func DarFormat(dar_src string, width, height int) (dar string, dar_float float32) {
	arr := strings.SplitN(dar_src, ":", 2)
	if len(arr) != 2 {
		arr = []string{strconv.Itoa(width), strconv.Itoa(height)}
	}
	w, h := Atof(arr[0]), Atof(arr[1])
	if h > 9 { //dar中的高大于9，则向9对齐。
		x := h / 9.0
		w = math.Floor(w/x + 0.5)
		h = 9
	} else if h < 3 { //dar中的高小于3，则向9对齐。
		x := 9.0 / h
		w = math.Floor(w*x + 0.5)
		h = 9
	}
	if int(w)%3 == 0 && int(h) == 9 {
		w = w / 3
		h = h / 3
	}
	dar = fmt.Sprintf("%d:%d", int(w), int(h))
	dar_float = float32(w) / float32(h)

	glog.Infof("dar_src　%s ==> %s (%.2f)", dar_src, dar, dar_float)

	return
}

func GetDefaultVideoIdx(vinfo []mediainfo.VideoInfo) int {
	for i := 0; i < len(vinfo); i++ {
		glog.Infof("GetDefaultVideoIdx i:%d, DAR:%s, Default:%s", i, vinfo[i].DAR, vinfo[i].Default)
		if strings.Compare(vinfo[i].Default, "Yes") == 0 {
			return i
		}
	}

	if len(vinfo) > 0 {
		return 0
	} else {
		return -1
	}
}

func GetDefaultAudioIdx(ainfo []mediainfo.AudioInfo) int {
	for i := 0; i < len(ainfo); i++ {
		if strings.Compare(ainfo[i].Default, "Yes") == 0 {
			return i
		}
	}

	if len(ainfo) > 0 {
		return 0
	} else {
		return -1
	}
}

func GetDefaultSubtitlesIdx(subinfo []mediainfo.SubtitlesInfo) int {
	for i := 0; i < len(subinfo); i++ {
		glog.Infof("GetDefaultSubtitlesIdx i:%d, Title:%s, Language:%s, Default:%s\n", i, subinfo[i].Title, subinfo[i].Language, subinfo[i].Default)
		if strings.Index(subinfo[i].Title, "国语") >= 0 { //普通话优先
			return i
		}
		if strings.Compare(subinfo[i].Default, "Yes") == 0 && strings.Compare(subinfo[i].Language, "zh") == 0 {
			return i
		}
	}

	for i := 0; i < len(subinfo); i++ {
		if strings.Index(subinfo[i].Language, "zh") == 0 {
			return i
		}
	}

	if len(subinfo) > 0 {
		return 0
	} else {
		return -1
	}
}

func GetMediaInfo(video_file string) (info *VideoInfo, err error) {
	if !IsExist(video_file) {
		glog.Errorf("video file [%s] not exist!", video_file)
		err = os.ErrNotExist
		return
	}

	mi, merr := GetMediaInfoEx(video_file)
	if merr != nil {
		err = merr
		return
	}

	info = &VideoInfo{}
	glog.Infof("GetMediaInfo mi.General.Duration:%d, mi.General.BitRate:%d\n", mi.General.Duration, mi.General.BitRate)
	info.Duration = mi.General.DurationStr
	info.Second = ms2sec(mi.General.Duration)
	info.Start = ms2sec(mi.General.Start)
	info.BitRate = byte2kb(mi.General.BitRate)
	//Video
	glog.Infof("GetMediaInfo len(mi.Video):%d\n", len(mi.Video))
	// info.VFormat    string `json:"v_format"`   // 视频格式
	vinfo := GetDefaultVideoIdx(mi.Video)

	if vinfo >= 0 {
		info.VCodec = mi.Video[vinfo].CodecID
		info.VBitRate = int(mi.Video[vinfo].BitRate) // byte2kb(mi.Video.BitRate)
		info.Resolution = mi.Video[vinfo].Resolution //mi.Video.Resolution
		info.Width = int(mi.Video[vinfo].Width)      //int(mi.Video.Width)
		info.Height = int(mi.Video[vinfo].Height)    //int(mi.Video.Height)
		glog.Infof("GetMediaInfo video VCodec:%s, VBitRate:%d, Width:%d, Height:%d\n", info.VCodec, info.VBitRate, info.Width, info.Height)
		info.Duration = mi.General.DurationStr
		info.DAR, info.DarF = DarFormat(mi.Video[vinfo].DAR, int(mi.Video[vinfo].Width), int(mi.Video[vinfo].Height)) //DarFormat(mi.Video.DAR, info.Width, info.Height)
	}
	glog.Infof("GetMediaInfo vinfo:%d, info.Resolution:%s, info.Width:%d, info.Height:%d, info.DAR:%s, info.DarF:%s\n", vinfo, info.Resolution, info.Width, info.Height, info.DAR, info.DarF)
	//Audio
	ainfo := GetDefaultAudioIdx(mi.Audio)
	if ainfo >= 0 {
		info.ACodec = mi.Audio[ainfo].CodecID                //mi.Audio.CodecID
		info.ASampleRate = int(mi.Audio[ainfo].SamplingRate) //byte2kb(mi.Audio.SamplingRate)
		info.ABitRate = int(mi.Audio[ainfo].BitRate)         //byte2kb(mi.Audio.BitRate)
		glog.Infof("GetMediaInfo audio ACodec:%s, info.ASampleRate:%d, info.ABitRate:%d\n", info.ACodec, info.ASampleRate, info.ABitRate)
	}

	//Subtitles
	mi.SubtitlesCnt = mi.General.TextCount
	info.SubtitlesCnt = mi.SubtitlesCnt
	info.Subtitles = make([]string, 0)
	for _, subtitle := range mi.Subtitles {
		glog.Infof("GetMediaInfo subtitle subtitle.Title:%s\n", subtitle.Title)
		info.Subtitles = append(info.Subtitles, subtitle.Title)
	}
	glog.Infof("GetMediaInfo videocount:%d, audiocount:%d, subtext:%d, info.SubtitlesCnt:%d\n", mi.General.VideoCount, mi.General.AudioCount, mi.General.TextCount, info.SubtitlesCnt)
	// info.Resolutions []string `json:"resolutions"` //可以转码的规格列表
	return
}

func GetMediaInfoEx(video_file string) (info *mediainfo.SimpleMediaInfo, err error) {
	glog.Infof("GetMediaInfoEx begin\n")
	if !IsExist(video_file) {
		glog.Errorf("video file [%s] not exist!", video_file)
		err = os.ErrNotExist
		return
	}

	//error merr
	mi, merr := mediainfo.GetMediaInfo(video_file)
	if merr != nil {
		err = merr
		glog.Errorf("mediainfo.GetMediaInfo err:%v\n", err)
		return
	}

	glog.Infof("GetMediaInfoExmi.SubtitlesCnt:%d\n", mi.SubtitlesCnt)
	if mi.SubtitlesCnt > 0 {
		glog.Infof("GetMediaInfoEx CodecID:%s\n", mi.Subtitles[0].CodecID)
	}

	info = mi
	return
	/*
			info = &VideoInfo{}

			info.Duration = mi.General.DurationStr
			info.Second = ms2sec(mi.General.Duration)
			info.Start = ms2sec(mi.General.Start)
			info.BitRate = byte2kb(mi.General.BitRate)
			//Video
			info.VCodec = mi.Video.CodecID
			// info.VFormat    string `json:"v_format"`   // 视频格式
			info.VBitRate = byte2kb(mi.Video.BitRate)
			info.Resolution = mi.Video.Resolution
			info.Width = int(mi.Video.Width)
			info.Height = int(mi.Video.Height)
			info.DAR, info.DarF = DarFormat(mi.Video.DAR, info.Width, info.Height)

			//Audio
			info.ACodec = mi.Audio.CodecID
			info.ASampleRate = byte2kb(mi.Audio.SamplingRate)
			info.ABitRate = byte2kb(mi.Audio.BitRate)

			//Subtitles
			info.SubtitlesCnt = mi.SubtitlesCnt
			info.Subtitles = make([]string, 0)
			for _, subtitle := range mi.Subtitles {
				info.Subtitles = append(info.Subtitles, subtitle.Title)
			}

			// info.Resolutions []string `json:"resolutions"` //可以转码的规格列表
			return
		}

		func main() {
			info, err := mediainfo.GetMediaInfo(os.Args[1])
			if err != nil {
				fmt.Printf("open failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%v\n", info)*/
}
