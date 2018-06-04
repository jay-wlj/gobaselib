package gobaselib

import (
	"bytes"
	"fmt"
	"github.com/fatih/structs"
	"github.com/jie123108/glog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 字幕相关文档：http://ffmpeg.org/ffmpeg-filters.html#subtitles-1
// [V4+ Styles] http://activearchives.org/wiki/cookbook
// Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
// Style: Myriam, DejaVu Sans Bold,14,&H00B4FCFC,&H00FF0000,&H00000008,&H80000008,-1,0,0,0,100,100,0.00,0.00,1,1.00,2.00,1,30,30,100,0
// Style: Hahn, DejaVu Serif Bold,14,&H00B4FCFC,&H00B4FCFC,&H00000008,&H80000008,-1,0,1,0,100,100,0.00,45.00,1,1.00,2.00,2,30,30,100,0
// Style: David, DejaVu Sans Bold,28,&H00B4FCFC,&H00FF0000,&H00000008,&H80000008,-1,0,0,0,100,100,0.00,0.00,1,1.00,2.00,1,30,30,100,0
// Note that it is in the opposite order of HTML colors ffmpeg subtitle中设置得颜色与html中设置得相反，比如：0000FF表示红色。
func CommandFmt(args ...string) string {
	return strings.Join(args, " ")
}

func F(filename string) string {
	return `"` + filename + `"`
}

type VideoInfo struct {
	Duration string  `json:"duration"` //时长，字符串格式。
	Second   float64 `json:"second"`   // 时长，秒
	Start    float64 `json:"start"`    //开始时间。
	BitRate  int     `json:"bit_rate"` //比特率 kb/s
	//Video
	VCodec     string  `json:"v_codec"`    // 编码格式
	VBitRate   int     `json:"v_bitrate"`  //视频比特率 kb/s 可能为0,此时需要使用BitRate-ABitRate计算得出。
	Resolution string  `json:"resolution"` // 分辨率
	Width      int     `json:"width"`      //宽。
	Height     int     `json:"height"`     //高。
	DAR        string  `json:"dar"`        // DAR, Display Aspect Ratio 显示宽高比。即最终播放出来的画面的宽与高之比。
	DarF       float32 `json:"dar_f"`      // dar float32
	//Audio
	ACodec      string `json:"a_codec"`       // 音频编码
	ASampleRate int    `json:"a_sample_rate"` // 音频采样频率
	ABitRate    int    `json:"a_bitrate"`     //音频比特率 kb/s
	//Subtitles
	SubtitlesCnt int      `json:"subtitles_cnt"` //字幕数量(不包含硬编码字幕)
	Subtitles    []string `json:"subtitles"`     //字幕列表。

	Resolutions []string `json:"resolutions"` //可以转码的规格列表
}

func (this *VideoInfo) String() string {
	if this == nil {
		return "nil"
	}
	fields := structs.Fields(this)
	var buf bytes.Buffer
	for _, field := range fields {
		buf.WriteString(fmt.Sprintf("%s: %v\n", field.Name(), field.Value()))
	}
	return buf.String()
}

// mkv output:
// Duration: 02:49:04.01, start: 0.000000, bitrate: 15007 kb/s
// Stream #0:0(eng): Video: h264 (High), yuv420p, 1920x1080 [SAR 1:1 DAR 16:9], 23.98 fps, 23.98 tbr, 1k tbn, 47.95 tbc
// Stream #0:1(eng): Audio: dts (DTS-HD MA), 48000 Hz, 5.1(side), s32p (24 bit) (default)
// Stream #0:1: Audio: ac3, 48000 Hz, 5.1(side), fltp, 448 kb/s (default)

// mp4 output:
// Duration: 00:05:00.02, start: 0.000000, bitrate: 850 kb/s
// Stream #0:0(und): Video: h264 (High) (avc1 / 0x31637661), yuv420p, 720x576 [SAR 16:15 DAR 4:3], 714 kb/s, 25 fps, 25 tbr, 12800 tbn, 50 tbc (default)
// Stream #0:1(und): Audio: aac (LC) (mp4a / 0x6134706D), 48000 Hz, stereo, fltp, 130 kb/s (default)

// mpeg output:
// Duration: 01:24:40.08, start: 0.229267, bitrate: 10331 kb/s
// Stream #0:0[0x1e0]: Video: mpeg2video (Main), yuv420p(tv), 720x576 [SAR 16:15 DAR 4:3], 25 fps, 25 tbr, 90k tbn, 50 tbc
// Stream #0:1[0x1c0]: Audio: mp2, 48000 Hz, stereo, s16p, 128 kb/s

// rmvb output:
// Duration: 01:08:10.30, start: 0.000000, bitrate: 657 kb/s
// Stream #0:0: Audio: aac (LC) (raac / 0x63616172), 32000 Hz, stereo, fltp, 64 kb/s
// Stream #0:1: Video: rv40 (RV40 / 0x30345652), yuv420p, 720x404, 579 kb/s, 23.98 fps, 23.98 tbr, 1k tbn, 1k tbc

// 特例 output:
// Duration: 01:59:16.27, start: 0.000000, bitrate: 984 kb/s
// Stream #0:0(jpn): Video: h264 (High), yuv420p, 688x528, SAR 45:43 DAR 15:11, 23.98 fps, 23.98 tbr, 1k tbn, 47.95 tbc (default)
// Stream #0:1(jpn): Audio: aac (HE-AAC), 48000 Hz, mono, fltp (default)
// Stream #0:2(jpn): Audio: aac (HE-AAC), 48000 Hz, mono, fltp

// Duration: 02:30:16.38, start: 0.000000, bitrate: 684 kb/s
// Stream #0:0(jpn): Video: h264 (High), yuv420p(tv, bt709/unknown/unknown), 720x368 [SAR 5760:4739 DAR 259200:108997], SAR 175:144 DAR 875:368, 23.98 fps, 23.98 tbr, 1k tbn, 47.95 tbc (default)
// Stream #0:1(jpn): Audio: aac (HE-AAC), 48000 Hz, stereo, fltp (default)
// Stream #0:2(chi): Subtitle: subrip (default)

// Stream #0:0(jpn): Video: h264 (High), yuv420p, 688x528, SAR 45:43 DAR 15:11, 23.98 fps, 23.98 tbr, 1k tbn, 47.95 tbc (default)

//分割符: [\s,]+
var re_duration = regexp.MustCompile(`Duration: (.*?)[\s,]+start: (.*?), bitrate: (\d*) kb/s`)
var re_video = regexp.MustCompile(`Video: (.*?)[\s,]+(\w*).*[\s,]+(\d*x\d*)[\s,]+\[?SAR \d*:\d* DAR ([\d:]*)\]?`)
var re_video_rmvb = regexp.MustCompile(`Video: (.*?)[\s,]+(\w*).*[\s,]+(\d*x\d*)`)

var re_video_bitrate = regexp.MustCompile(`Video: .* \[?SAR .* DAR .*\]?[\s,]+(\d*) kb/s`)
var re_audio = regexp.MustCompile(`Audio: (.*?), (\d*) Hz[\s,]+.*[\s,]+.*[\s,]+(\d*) kb/s`)
var re_audio_mkv = regexp.MustCompile(`Audio: (.*?), (.*?) Hz`)
var def_subtitle_style = "Fontsize=16,PrimaryColour=&H00FFFFFF,OutlineColour=&H00512906,BackColour=&H32371C06,Shadow=1,BorderStyle=1,Outline=2,Encoding=134"

// Stream #0:2(chi): Subtitle: mov_text (tx3g / 0x67337874), 0 kb/s (default)
//    Metadata:
//      handler_name    : SubtitleHandler

// Stream #0:2(chi): Subtitle: subrip (default) (forced)
//     Metadata:
//       title           : 简体中文 Chs
var re_subtitle = regexp.MustCompile(`Stream .*: Subtitle: .*\n\s*Metadata:.*\n\s*.*\s*:\s(.*)\n`)

func Atoi(str string) (i int) {
	i, err := strconv.Atoi(str)
	if err != nil {
		glog.Errorf("Atoi(%s) failed! err: %v", str, err)
		i = 0
	}

	return
}

func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		glog.Errorf("ParseFloat(%s) failed! err: %v", s, err)
		f = 0.0
	}
	return f
}

func ProcDAR(dar_src string) (dar string, dar_float float32) {
	arr := strings.SplitN(dar_src, ":", 2)
	if len(arr) != 2 {
		dar, dar_float = dar_src, float32(16)/float32(9)
	} else {
		w, h := Atoi(arr[0]), Atoi(arr[1])
		if h > 9 { //dar中的高大于9，刚向9对齐。
			x := float64(h) / 9.0
			w = int(float64(w)/x + 0.5)
			h = 9
			if w%3 == 0 {
				w = w / 3
				h = h / 3
			}
		}
		dar = fmt.Sprintf("%d:%d", w, h)
		dar_float = float32(w) / float32(h)
	}
	glog.Infof("dar_src　%s ==> %s (%.2f)", dar_src, dar, dar_float)

	return
}

func GetVideoInfo(ffmpeg_bin, video_file string) (info *VideoInfo, err error) {
	glog.Infof("GetVideoInfo begin\n")
	if !IsExist(video_file) {
		glog.Errorf("video file [%s] not exist!", video_file)
		err = os.ErrNotExist
		return
	}

	debug_str := CommandFmt(F(ffmpeg_bin), "-i", F(video_file))
	glog.Infof("get video info: [%s]...", debug_str)
	cmd := exec.Command(ffmpeg_bin, "-i", video_file)
	out, err2 := cmd.CombinedOutput()
	if err2 != nil && err2.Error() != "exit status 1" {
		err = err2
		glog.Errorf("get video info (%s) failed! err: [%v]", debug_str, err)
		glog.Errorf("command output [%v]", string(out))
		return
	}
	outstr := string(out)

	var arr []string
	var video_info []string
	var audio_info []string
	var vbitrate_info []string
	var subtitle_infos [][]string

	//duration 解析
	duration_info := re_duration.FindStringSubmatch(outstr)
	glog.Infof("duration_info:: %v", duration_info)
	if len(duration_info) != 4 {
		glog.Errorf("matched duration_info(%v) length != 4", duration_info)
		goto FORMAT_ERR
	}
	info = &VideoInfo{}
	info.Duration = duration_info[1]
	arr = strings.SplitN(info.Duration, ":", 3)
	info.Second = float64(Atoi(arr[0])*3600.0) + float64(Atoi(arr[1])*60.0) + ParseFloat(arr[2])
	info.Start = ParseFloat(duration_info[2])
	info.BitRate = Atoi(duration_info[3])

	//Video Info解析
	video_info = re_video.FindStringSubmatch(outstr)
	if len(video_info) == 0 {
		video_info = re_video_rmvb.FindStringSubmatch(outstr)
	}
	glog.Infof("video_info:: %v", video_info)
	if len(video_info) != 5 && len(video_info) != 4 {
		glog.Errorf("matched video_info(%v) length != 4|5", video_info)
		goto FORMAT_ERR
	}
	info.VCodec = video_info[1]
	// info.VFormat = video_info[2]
	info.Resolution = video_info[3]
	arr = strings.SplitN(info.Resolution, "x", 2)
	info.Width = Atoi(arr[0])
	info.Height = Atoi(arr[1])
	if len(video_info) == 5 {
		info.DAR, info.DarF = ProcDAR(video_info[4])
	} else {
		info.DAR, info.DarF = ProcDAR(arr[0] + ":" + arr[1])
	}
	glog.Infof("info.DAR:%s, info.DarF:%s, info.Width:%d, info.Height:%d\n", info.DAR, info.DarF, info.Width, info.Height)

	//Audio Info 解析
	audio_info = re_audio.FindStringSubmatch(outstr)
	glog.Infof("audio_info:: %v", audio_info)
	if len(audio_info) != 4 {
		audio_info = re_audio_mkv.FindStringSubmatch(outstr)
		glog.Infof("audio_info:: %v", audio_info)
		if len(audio_info) != 3 {
			glog.Errorf("matched audio_info(%v) length != 4", audio_info)
			goto FORMAT_ERR
		}

		info.ACodec = audio_info[1]
		info.ASampleRate = Atoi(audio_info[2])
	} else {
		info.ACodec = audio_info[1]
		info.ASampleRate = Atoi(audio_info[2])
		info.ABitRate = Atoi(audio_info[3])
	}

	//Video BitRate 解析
	vbitrate_info = re_video_bitrate.FindStringSubmatch(outstr)
	if len(vbitrate_info) == 2 {
		info.VBitRate = Atoi(vbitrate_info[1])
	} else {
		info.VBitRate = info.BitRate - info.ABitRate
	}

	//Subtitles 解析
	subtitle_infos = re_subtitle.FindAllStringSubmatch(outstr, -1)
	for _, subtitle_info := range subtitle_infos {
		info.SubtitlesCnt += 1
		info.Subtitles = append(info.Subtitles, strings.TrimSpace(subtitle_info[1]))
	}
	glog.Infof("GetVideoInfo end\n")
	return
FORMAT_ERR:
	glog.Errorf("get video info (%s) failed! output invalid: [%s]", debug_str, outstr)
	err = fmt.Errorf("Invalid-ffmpeg-output")
	info = nil
	return
}

func SetSubtitleStyle(style string) {
	if style != "" {
		def_subtitle_style = style
	}
}

/***
 * 字幕格式校对：
 * 字幕格式如：Fontsize=20,PrimaryColour=xxx 中的字体大小，
 * 我们一般是相对于16:9的视频进行调试的，但遇到视频比例大于16:9，
 * 比如21:9的视频时，在同样的屏幕上，显示得字体会比16:9的视频要小。
 */
func SubtitleAdjust(style string, dar float32) string {
	dar_16_9 := float32(16) / float32(9)
	if dar > (dar_16_9 + float32(0.01)) {
		re_fontsize := regexp.MustCompile(`(?i)Fontsize=\d+`)
		fontsize_info := re_fontsize.FindStringSubmatch(style)
		if len(fontsize_info) == 1 {
			fontsize_raw := fontsize_info[0]
			sizestr := fontsize_raw[len("Fontsize="):]
			size, _ := strconv.Atoi(sizestr)
			if size > 0 {
				size = int(float32(size)/dar_16_9*dar + 0.5)
				fontsize_new := fmt.Sprintf("Fontsize=%d", size)
				style = strings.Replace(style, fontsize_raw, fontsize_new, 1)
			}
		}
	}
	return style
}

type ConvertArgs struct {
	Size          string
	VideoBitrate  int //kbps
	AudioBitrate  int //kbps
	AudioIndex	  int //audioindex
	OverlaySubtitle int//picture subtitle
	FFMpegArgs    string
	SubtitleFile  string //字幕文件，可为空。或者指定一个视频文件。
	SubtitleStyle string
}

/**
 *
 */
func VideoConvert(ffmpeg_bin, video_file string, localfile string, args *ConvertArgs, timeout time.Duration) error {

	if IsExist(localfile) {
		glog.Infof("small file [%s] is exist!", localfile)
		return nil
	}
	glog.Infof("Video Convert [%s] ==> [%s] timeout: %v args.OverlaySubtitle:%d, args.AudioIndex:%d, args.SubtitleFile:%s\n", video_file, localfile, timeout, args.OverlaySubtitle, args.AudioIndex, args.SubtitleFile)
	ext := filepath.Ext(localfile)

	localfile_tmp := localfile[0:len(localfile)-len(ext)] + "_tmp" + ext

	dir := path.Dir(localfile_tmp)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		glog.Errorf("MkdirAll(%s) failed! err: %v", dir, err)
		return err
	}

	AccEncoder := "aac"
	VBitRate := fmt.Sprintf("%dk", args.VideoBitrate)
	ABitRate := fmt.Sprintf("%dk", args.AudioBitrate)

	var vf_args, fc_args, audio_args string
	if args.SubtitleFile != "" {
		subtitle_style_used := def_subtitle_style
		if args.SubtitleStyle != "" {
			subtitle_style_used = args.SubtitleStyle
		}
		vf_args = fmt.Sprintf("subtitles='%s':force_style='%s',scale=%s", args.SubtitleFile, subtitle_style_used, args.Size)
		glog.Infof("af_args:%s\n", vf_args)
	} else if args.OverlaySubtitle >= 0 {
		//var overlaystr string
		//overlaystr = fmt.Sprintf("[0:v]scale=%s[scale],[scale][0:s]overlay[ov]", args.Size)
		fc_args = fmt.Sprintf("[0:v][0:s]overlay[ov]")
		//fc_args = fmt.Sprintf("'[0:v]scale=%s[scale],[scale][0:s]overlay[ov]'", args.Size)
		glog.Infof("fc_args:%s\n", fc_args)
	}else {
		vf_args = "scale=" + args.Size
	}

	if args.AudioIndex >= 0 {
		//args.AudioIndex = args.AudioIndex + 1;
		audio_args = fmt.Sprintf("0:%d", args.AudioIndex)
	}
	var debug_str string
	cmd_args := strings.Fields(args.FFMpegArgs)
	if args.OverlaySubtitle >= 0 {
		cmd_args = append(cmd_args, "-i", video_file, "-y", "-c:v", "libx264", "-filter_complex", fc_args, "-map", "[ov]", "-map", audio_args, /*"-t", "300",*/ "-b:v", VBitRate,
		"-bufsize", VBitRate, "-c:a", AccEncoder, "-ac", "2", "-b:a", ABitRate, localfile_tmp)
		debug_str = CommandFmt(F(ffmpeg_bin), args.FFMpegArgs, "-i", F(video_file),
		"-y", "-c:v", "libx264", "-filter_complex", fc_args, "-map", "[ov]", "-map", audio_args, /*"-t", "300",*/"-b:v", VBitRate,
		"-bufsize", VBitRate, "-c:a", AccEncoder, "-ac", "2", "-b:a", ABitRate, F(localfile_tmp))
		
	}else {
		cmd_args = append(cmd_args, "-i", video_file, "-y", "-c:v", "libx264", "-vf", vf_args, "-map", "0:v", "-map", audio_args, /*"-t", "300",*/ "-b:v", VBitRate,
		"-bufsize", VBitRate, "-c:a", AccEncoder, "-ac", "2", "-b:a", ABitRate, localfile_tmp)
		debug_str = CommandFmt(F(ffmpeg_bin), args.FFMpegArgs, "-i", F(video_file),
		"-y", "-c:v", "libx264", "-vf", vf_args, "-map", "0:v", "-map", audio_args, /*"-t", "300",*/ "-b:v", VBitRate,
		"-bufsize", VBitRate, "-c:a", AccEncoder, "-ac", "2", "-b:a", ABitRate, F(localfile_tmp))
		//glog.Infof("ffmpeg cmd: [%s]", cmd_args)
		//cmd = exec.Command(ffmpeg_bin, cmd_args...)
	}
	glog.Infof("ffmpeg cmd: [%s]", cmd_args)
	cmd := exec.Command(ffmpeg_bin, cmd_args...)

	glog.Infof("cmd_args:%s, debug_str:%s, cmd:%s\n", cmd_args, debug_str, cmd)

	// "-ss", strconv.Itoa(begin), "-t", strconv.Itoa(duration),
	// ./bin/linux/ffmpeg/ffmpeg -y -ss 0 -t 20 -i 偷心.mpg -c:v libx264 -vf scale=720:540 -b:v 512k -bufsize 512k -c:a aac -b:a 64k /data/VideoConvertCache/电影001/偷心512k-64k-720x540.mp4
	// "-ss", "300", "-t", "20",
	// -ss 300 -t 20
	//glog.Infof("ffmpeg cmd: [%s]", debug_str)
	

	// cmd := exec.Command(ffmpeg_bin, "-i", video_file, "-y", "-c:v", "libx264", "-vf", "scale="+size,
	// 	"-b:v", VBitRate, "-bufsize", VBitRate, "-c:a", AccEncoder, "-ac", "2", "-b:a", ABitRate, localfile_tmp)

	// err := cmd.Run()
	var out []byte
	if timeout > 0 {
		out, err = CombinedOutput(cmd, timeout)
	} else {
		out, err = cmd.CombinedOutput()
	}
	if err != nil {
		glog.Errorf("video convert(%s) failed! err: %v", debug_str, err)
		glog.Errorf("command output [%v]", string(out))
		return err
	} else {
		err = os.Rename(localfile_tmp, localfile)
		if err != nil {
			glog.Errorf("Rename [%s] to [%s] failed!", localfile_tmp, localfile)
			return err
		}
		glog.Infof("video convert src [%s] to [%s] success!", video_file, localfile)
	}
	return nil
}
