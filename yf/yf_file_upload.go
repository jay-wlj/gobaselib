package yf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gobaselib/log"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	base "github.com/jay-wlj/gobaselib"
)

// [Content-Type] = {文件扩展名，路径前缀}
var g_contentType map[string][]string

func init() {
	g_contentType = make(map[string][]string)
	// 图片
	g_contentType["image/gif"] = []string{"gif", "img"}
	g_contentType["image/jpeg"] = []string{"jpg", "img"}
	g_contentType["image/png"] = []string{"png", "img"}

	// 视频
	g_contentType["video/mp4"] = []string{"mp4", "video"}
	g_contentType["video/mkv"] = []string{"mkv", "video"}
	g_contentType["application/octet-stream"] = []string{"bin", "file"}
}

func get_content_type(filename string) string {
	contentType := ""
	if strings.HasSuffix(filename, ".mp4") {
		contentType = "video/mp4"
	} else if strings.HasSuffix(filename, ".mkv") {
		contentType = "video/mkv"
	} else if strings.HasSuffix(filename, ".jpg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(filename, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(filename, ".gif") {
		contentType = "image/gif"
	} else {
		contentType = "application/octet-stream"
	}

	return contentType
}

func CheckFileExist(host, filename, hash, app_id, app_key, id string) *base.OkJson {
	uri := host + "/upload/check_exist"

	res := &base.OkJson{Ok: false, Reason: base.ERR_SERVER_ERROR}
	contentType := get_content_type(filename)
	if contentType == "" {
		//fmt.Println("不支持的文件类型：", filename)
		log.Errorf("un support file type: %s", filename)
		res.ReqDebug = uri
		res.Reason = ERR_CONTENT_TYPE_INVALID
		res.StatusCode = 400
		return res
	}

	headers := make(map[string]string, 10)
	headers["Content-Type"] = contentType
	headers["X-YF-Platform"] = "test"
	headers["X-YF-hash"] = hash
	headers["X-YF-AppId"] = app_id
	if id != "" {
		headers["X-YF-Id"] = id
	}

	if app_key == "439081c403882a0c86fbed7ce2b4932cfcad47e1" {
		headers["X-YF-SIGN"] = app_key
		app_key = ""
	}

	res = YfHttpGet(uri, headers, 10, app_key)

	return res
}

func PostFile2YfUpload(host, filename, app_id, app_key, id, resize, target string, timeout time.Duration, is_test bool) *base.OkJson {
	uri := host + "/upload/simple"
	res := &base.OkJson{Ok: false, Reason: base.ERR_SERVER_ERROR}

	contentType := get_content_type(filename)
	if contentType == "" {
		log.Errorf("un support file type: %s", filename)
		res.Reason = ERR_CONTENT_TYPE_INVALID
		res.StatusCode = 400
		return res
	}

	info := g_contentType[contentType]
	size := &base.Size{0, 0}
	if len(info) == 2 && info[1] == "img" {
		var err error
		size, err = base.GetImageSizeF(filename)
		if err != nil {
			fmt.Println("GetImageSizeF failed! err:", err)
			size = &base.Size{0, 0}
		}
	}

	bodyfile, err := os.Open(filename)
	if err != nil {
		log.Errorf("Open file failed! err: %v", err)
		res.Reason = ERR_OPEN_INPUT_FILE
		res.StatusCode = 400
		return res
	}
	defer bodyfile.Close()

	content, _ := ioutil.ReadFile(filename)
	hash := Sha1hex(content)

	headers := make(map[string]string, 10)
	headers["Content-Type"] = contentType
	headers["X-YF-Platform"] = "test"
	headers["X-YF-hash"] = hash
	headers["X-YF-AppId"] = app_id
	headers["X-YF-width"] = strconv.Itoa(size.Width)
	headers["X-YF-height"] = strconv.Itoa(size.Height)

	if id != "" {
		headers["X-YF-Id"] = id
	}
	if resize != "" {
		headers["X-YF-resize"] = resize
	}
	if target != "" {
		headers["X-YF-target"] = target
	}
	if is_test {
		headers["X-YF-Test"] = "1"
	}

	if app_key == "439081c403882a0c86fbed7ce2b4932cfcad47e1" {
		headers["X-YF-SIGN"] = app_key
		app_key = ""
	}

	res = YfHttpPost(uri, content, headers, timeout, app_key)

	return res
}

type ChunkResponse struct {
	CompletedChunks    []int          `json:"completed_chunks"`    //已经上传成功的块ID列表
	NotCompletedChunks []int          `json:"notcompleted_chunks"` //未上传的块ID列表
	Url                string         `json:"url"`                 //资源URL
	Rid                string         `json:"rid"`                 //资源ID
	ChunkSize          int            `json:"chunksize"`           //块大小（后面按该大小分块上传，不同的文件，块大小可能不同）
	VideoInfo          base.VideoInfo `json:"videoinfo"`           //上传视频时，返回视频信息。
}

func chunk_response_parse(body map[string]interface{}) (resp *ChunkResponse, err error) {
	resp = &ChunkResponse{}
	buf, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Marshal(%v) failed! err: %v", body, err)
		return nil, err
	}

	err = json.Unmarshal(buf, resp)
	if err != nil {
		resp = nil
		return
	}
	return
}

func web_chunk_init(host string, headers map[string]string, timeout time.Duration) (*ChunkResponse, error) {
	uri := host + "/upload/web/chunk/init"

	res := base.HttpPostJson(uri, []byte(""), headers, timeout)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! err: %v", res.ReqDebug, res.Error)
		var err error
		err = res.Error
		if err == nil {
			err = fmt.Errorf("http-error: %d", res.StatusCode)
		}
		return nil, err
	}

	if !res.Ok {
		log.Errorf("request [%s] failed! reason: %v", res.ReqDebug, res.Reason)
		return nil, fmt.Errorf(res.Reason)
	}

	resp, err := chunk_response_parse(res.Data)
	if err != nil {
		log.Errorf("request [%s] parse response [%s] failed! err: %v", res.ReqDebug, string(res.RawBody), err)
		return nil, err
	}

	return resp, nil
}

func get_offset_and_chunksize(chunksize, chunkindex int, filesize int64) (int64, int) {
	offset := int64(0)
	offset = int64(chunksize) * int64(chunkindex)
	chunks := int(math.Ceil(float64(filesize) / float64(chunksize)))
	if chunkindex == chunks-1 { //最后一块。
		chunksize_new := filesize % int64(chunksize)
		if chunksize_new > 0 {
			chunksize = int(chunksize_new)
		}
	}

	return offset, chunksize
}

func web_chunk_upload(host string, bodyfile *os.File, headers map[string]string, filesize int64, chunksize, chunkindex int, timeout time.Duration) (*ChunkResponse, error) {
	uri := host + "/upload/web/chunk/upload"
	filename := headers["X-YF-filename"]
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("chunk", filepath.Base(filename))
	if err != nil {
		writer.Close()
		log.Errorf("CreateFormFile('chunk', '%s') failed! err: %v", filepath.Base(filename), err)
		return nil, err
	}

	offset, chunksize_real := get_offset_and_chunksize(chunksize, chunkindex, filesize)
	if offset >= filesize {
		writer.Close()
		log.Errorf("chunksize(%d) x chunkindex(%d) : offset(%d) >= filesize(%d)", chunksize, chunkindex, offset, filesize)
		return nil, fmt.Errorf("chunk invalid")
	}

	buf := make([]byte, chunksize_real)
	_, err = bodyfile.ReadAt(buf, offset)
	if err != nil {
		writer.Close()
		log.Errorf("ReadAt(%s, offset: %d, len: %d) failed! err: %v", filename, offset, chunksize_real, err)
		return nil, err
	}
	chunkhash := Md5hex(buf)
	part.Write(buf)
	writer.Close()

	headers["X-YF-chunksize"] = strconv.Itoa(chunksize_real)
	headers["X-YF-chunkindex"] = strconv.Itoa(chunkindex)
	headers["X-YF-chunkhash"] = chunkhash
	headers["Content-Type"] = writer.FormDataContentType()

	buf = body.Bytes()
	res := base.HttpPostJson(uri, buf, headers, timeout)
	if res.StatusCode != 200 {
		log.Errorf("request [%s] failed! status: %d, err: %v", res.ReqDebug, res.StatusCode, res.Error)
		err = res.Error
		if err == nil {
			err = fmt.Errorf("http-error: %d", res.StatusCode)
		}
		return nil, err
	}

	if !res.Ok {
		log.Errorf("request [%s] failed! reason: %v", res.ReqDebug, res.Reason)
		return nil, fmt.Errorf(res.Reason)
	}

	resp, err := chunk_response_parse(res.Data)
	if err != nil {
		log.Errorf("request [%s] parse response [%s] failed! err: %v", res.ReqDebug, string(res.RawBody), err)
		return nil, err
	}
	return resp, nil
}

func YfChunkUploadWeb(host, filename, app_id, id, token string, timeout time.Duration, chunksize int, is_test bool) (*ChunkResponse, error) {
	// hash, filesize, err := Md5File(filename)
	hash, filesize, err := NFSimpleMd5File(filename)
	if err != nil {
		log.Errorf("Md5File(%s) failed! err: %v", filename, err)
		return nil, err
	}

	headers := make(map[string]string, 10)
	headers["Content-Type"] = "application/json"
	headers["X-YF-Platform"] = "web"
	headers["X-YF-hash"] = hash
	headers["X-YF-filename"] = url.QueryEscape(filename)
	headers["X-YF-filesize"] = strconv.FormatInt(filesize, 10)
	if chunksize > 0 {
		headers["X-YF-chunksize"] = strconv.Itoa(chunksize)
	}
	headers["X-YF-AppId"] = app_id
	headers["X-YF-Token"] = token
	if id != "" {
		headers["X-YF-Id"] = id
	}
	if is_test {
		headers["X-YF-Test"] = "1"
	}
	if strings.HasSuffix(filename, ".txt") {
		headers["X-Body-Is-Text"] = "1"
	}

	resp, err := web_chunk_init(host, headers, timeout)
	if err != nil {
		log.Errorf("web_chunk_init(headers:%v) failed! err: %v", headers, err)
		return nil, err
	}
	if len(resp.NotCompletedChunks) == 0 && resp.Url != "" {
		return resp, nil
	}

	if resp.ChunkSize > 0 {
		chunksize = resp.ChunkSize
	}

	bodyfile, err := os.Open(filename)
	if err != nil {
		log.Errorf("Open file failed! err:%v", err)
		return nil, err
	}
	defer bodyfile.Close()

	var chunk_resp *ChunkResponse
	var chunk_err error

	NotCompletedChunks := resp.NotCompletedChunks
	for i := 0; i < 3; i++ {
		for _, chunkindex := range NotCompletedChunks {
			chunk_resp, chunk_err = web_chunk_upload(host, bodyfile, headers, filesize, chunksize, chunkindex, timeout)
			if chunk_err != nil {
				log.Errorf("web_chunk_upload(headers: %v) failed! err: %v", headers, chunk_err)
				//TODO: 出错重试。
				return nil, chunk_err
			}
		}
		if chunk_resp == nil {
			return nil, fmt.Errorf("chunk_resp is nil")
		}

		NotCompletedChunks = chunk_resp.NotCompletedChunks
		if len(NotCompletedChunks) < 1 {
			break
		}
	}
	if len(NotCompletedChunks) > 0 {
		chunk_err = fmt.Errorf("Upload-Failed: Too many errors")
	}
	return chunk_resp, chunk_err
}
