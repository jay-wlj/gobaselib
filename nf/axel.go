package base

import (
	"github.com/jay-wlj/gobaselib/log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
)

func url_encode(orig_url string) string {
	u, err := url.Parse(orig_url)
	if err != nil {
		return orig_url
	}
	return u.String()
}

func DownloadFileByAxel(axel_bin, download_uri string, localfile string, connections int) error {
	log.Infof("begin download file [%s] ==> [%s]", download_uri, localfile)

	// 文件已经存在。
	if IsExist(localfile) {
		log.Infof("file [%s] is exist!", localfile)
		return nil
	}

	localfile_tmp := localfile + ".tmp"
	dir := path.Dir(localfile_tmp)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Errorf("MkdirAll(%s) failed! err: %v", dir, err)
		return err
	}

	download_uri = url_encode(download_uri)
	debug_str := CommandFmt(F(axel_bin), "--num-connections="+strconv.Itoa(connections), "--output="+F(localfile_tmp), F(download_uri))
	log.Infof("axel cmd: [%s]", debug_str)

	cmd := exec.Command(axel_bin, "--num-connections="+strconv.Itoa(connections), "--output="+localfile_tmp, download_uri)
	// 如果用Run，执行到该步则会阻塞等待5秒
	// err := cmd.Run()
	err = cmd.Run()
	if err != nil {
		log.Errorf("download [%s] failed! err: %v", download_uri, err)
		return err
	} else {
		err = os.Rename(localfile_tmp, localfile)
		if err != nil {
			log.Errorf("Rename [%s] to [%s] failed!", localfile_tmp, localfile)
			return err
		}
		log.Infof("File [%s] download ok, write to [%s] ", download_uri, localfile)
	}
	return nil
}
