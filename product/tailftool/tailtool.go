package tailftool

import (
    "fmt"
    "github.com/hpcloud/tail"
    "go.uber.org/zap"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "eagle/product/log"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

var (
   wg sync.WaitGroup
)

type FileTail struct {
    PositionFile string
    Offset       int64
    File
    OpenFile []*tail.Tail
    Log  *zap.Logger
}
type File struct {
    Positions map[string]int64 `yaml:"positions"`
}

func NewFileTail(positionFile string) *FileTail {
    var f FileTail
    files, err := f.PositionRead(positionFile)
    if err != nil {
        fmt.Println("read posistion file failed: ", err)
    }
    return &FileTail{
        PositionFile: positionFile,
        Offset:       0,
        File:         files,
        OpenFile:     make([]*tail.Tail, 0),
        Log:  log.Log(),
    }
}

//初始化tailf需要的信息
func (f *FileTail) Init(fileName string) (err error) {
    var cli *tail.Tail
    var tag bool
    if len(f.File.Positions) > 0 {
        for file, offset := range f.File.Positions {
            if file == fileName {
                tag = true
                config := tail.Config{
                    ReOpen:    true,
                    Follow:    true,
                    Location:  &tail.SeekInfo{Offset: offset, Whence: 0},
                    MustExist: false,
                    Poll:      true,
                }
                cli, err = tail.TailFile(file, config)
                if err != nil {
                    f.Log.Error(fmt.Sprintf("find file name failed: %s\n", err))
                    return
                }
            }
        }
        if !tag {
            config := tail.Config{
                ReOpen:    true,
                Follow:    true,
                Location:  &tail.SeekInfo{Offset: f.Offset, Whence: 0},
                MustExist: false,
                Poll:      true,
            }
            cli, err = tail.TailFile(fileName, config)
            if err != nil {
                f.Log.Error(fmt.Sprintf("find file name failed: %s\n", err))
                return
            }
        }
    } else {
        config := tail.Config{
            ReOpen:    true,
            Follow:    true,
            Location:  &tail.SeekInfo{Offset: f.Offset, Whence: 0},
            MustExist: false,
            Poll:      true,
        }
        cli, err = tail.TailFile(fileName, config)
        if err != nil {
            f.Log.Error(fmt.Sprintf("find file name failed: %s\n", err))
            return
        }
    }
    f.OpenFile = append(f.OpenFile, cli)
    return
}

//读取文件内容发送到kafka
func (f *FileTail) ReadFileLine(kafka func(message string, topic string) (err error)) {
    for _, client := range f.OpenFile {
        //log.Println("###############",client.Filename,client.Lines)
        wg.Add(1)
        go func() {
            path := strings.Split(client.Filename, "/")
            topic := path[3] + "_" + path[4] + "_" + path[5]
            for line := range client.Lines {
                err := kafka(line.Text+topic, topic)
                //log.Printf("message: %s topic: %s  filename: %s ", line.Text, topic, fileName)
                if err != nil {
                    f.Log.Error(fmt.Sprintf("send kafka message failed: %s\n", err))
                    return
                }
            }
        }()
    }
}

func (f *FileTail) Stop() {
    for _, cli := range f.OpenFile {
        offset, err := cli.Tell()
        if err != nil {
            f.Log.Error(fmt.Sprintf("Get File Offset Failed: %s\n", err))
        }
        //pos := map[string]int64{file:offset}
        f.Positions[cli.Filename] = offset
        cli.Cleanup()
        err = cli.Stop()
        if err != nil {
            f.Log.Error(fmt.Sprintf("stop tail file failed: %s\n" ,err))
        }
    }
    //保存读取文件的offet位置
    err := f.PositionSave()
    if err != nil {
        f.Log.Error(fmt.Sprintf("save file offset failed: %s\n", err))
        return
    }
}

func (f *FileTail) PositionRead(positionFile string) (pos File, err error) {
    if _, err = os.Stat(positionFile); os.IsNotExist(err) {
        return
    }
    buf, err := ioutil.ReadFile(filepath.Clean(positionFile))
    if err != nil {
        f.Log.Error(fmt.Sprintf("Read Position File Failed: %s\n ", err))
        return
    }
    err = yaml.UnmarshalStrict(buf, &pos)
    if err != nil {
        f.Log.Error(fmt.Sprintf("yaml UnmarshalStrict Failed: %s\n", err))
        return
    }
    return pos, nil
}

func (f *FileTail) PositionSave() (err error) {

    buf, err := yaml.Marshal(File{
        Positions: f.Positions,
    })
    if err != nil {
        f.Log.Error(fmt.Sprintf("Marshal Position File Failed: %s\n", err))
        return
    }
    err = ioutil.WriteFile(f.PositionFile, buf, os.FileMode(0600))
    if err != nil {
        f.Log.Error(fmt.Sprintf("Write Position TO File Failed: %s\n", err))
        return
    }
    return
}
