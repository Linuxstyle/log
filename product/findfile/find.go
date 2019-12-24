package findfile

import (
    "io/ioutil"
    "log"
    "os"
    "strings"
)

type FileInfo struct {
    Files []string
}

//查找文件
func (f *FileInfo)FindFile(fileStr string) (err error) {
    if strings.Contains(fileStr, "*") {
        files:= strings.Split(fileStr, "*")[0]
        files = strings.TrimSuffix(files, "/")
        err = f.FindFileQuery(files)
        if err != nil {
            return
        }
    } else {
        if strings.HasSuffix(fileStr, "/") {
            files := strings.TrimSuffix(fileStr, "/")
            err = f.FindFileQuery(files)
            return
        } else {
            err = f.FindFileQuery(fileStr)
            return
        }
    }
    return
}

//递归查找
func (f *FileInfo)FindFileQuery(fileStr string) (err error) {
    dirs, err := ioutil.ReadDir(fileStr)
    if err != nil {
        return
    }
    PthSep := string(os.PathSeparator)
    for _, dir := range dirs {
        if dir.IsDir() {
            newDir := fileStr + PthSep + dir.Name()
            err = f.FindFileQuery(newDir)
            if err != nil {
                log.Println("find file failed: ",err)
                return
            }
        } else {
            if strings.HasSuffix(dir.Name(), "log") {
                //log.Println(dir.Name())
                f.Files = append(f.Files, fileStr+PthSep+dir.Name())
            }
        }
    }
    return
}
