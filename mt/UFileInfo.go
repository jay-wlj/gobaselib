package mt

import (
	"fmt"
	"io/ioutil"
	// "net/http"
	"encoding/json"
	base "gobaselib"
	"os"
	"sort"
	"strings"
)

type FileInfo struct {
	Filename string `json:"filename"` //本地文件名。
	Url      string `json:"url"`      //上传后的URL
	Times    int    `json:"times"`    //使用次数.
}

type FileInfos struct {
	Infos    []FileInfo
	metafile string
}

func (this *FileInfos) Len() int {
	return len(this.Infos)
}

func (this *FileInfos) Less(i, j int) bool {
	if this.Infos[i].Times == this.Infos[j].Times {
		return this.Infos[i].Filename < this.Infos[j].Filename
	}
	return this.Infos[i].Times < this.Infos[j].Times
}

func (this *FileInfos) Swap(i, j int) {
	this.Infos[i], this.Infos[j] = this.Infos[j], this.Infos[i]
}

func NewFileInfos(metafile string) *FileInfos {
	fileinfos := &FileInfos{metafile: metafile}
	return fileinfos
}

func (this *FileInfos) ReadFileInfo() error {
	lines, err := base.ReadLines(this.metafile)
	if err == os.ErrNotExist {
		this.Infos = make([]FileInfo, 0)
		return nil
	} else if err != nil {
		return err
	}

	this.Infos = make([]FileInfo, 0)
	for _, line := range lines {
		info := FileInfo{}
		err = json.Unmarshal([]byte(line), &info)
		if err != nil {
			fmt.Printf("Unmarshal(%s) failed! err: %v\n", line, err)
			return err
		}
		this.Infos = append(this.Infos, info)
	}

	return nil
}

func (this *FileInfos) WriteFileInfo() error {
	lines := make([]string, 0)
	for _, info := range this.Infos {
		line, _ := json.Marshal(info)
		if line != nil && len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	content := strings.Join(lines, "\n")
	err := ioutil.WriteFile(this.metafile, []byte(content), os.ModePerm)
	return err
}

func (this *FileInfos) FindFileInfo(filename string) *FileInfo {
	for i, info := range this.Infos {
		if info.Filename == filename {
			return &this.Infos[i]
		}
	}
	return nil
}

func (this *FileInfos) RandomGet() *FileInfo {
	if this.Len() > 0 {
		sort.Sort(this)
		return &this.Infos[0]
	}
	return nil
}
