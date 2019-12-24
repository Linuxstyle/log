package watchfiletool

import (
    "fmt"
    "go.uber.org/zap"
    //"go/ast"
    "eagle/product/log"
    "path/filepath"
    "time"
)



type WatchFile struct {
    Wach  chan string
    files []string
    Log  *zap.Logger
}

func NewWatchFile() *WatchFile{
    return &WatchFile{
        Wach: make(chan  string,10),
        files: make([]string,20),
        Log: log.Log(),
    }
}
func(w *WatchFile) Init(dir string) (err error) {
    ticker := time.NewTicker(time.Second * 1)
    for {
        select {
        case <-ticker.C:
            ///opt/hostpath/*/*/*
            matchs, err := filepath.Glob(dir)
            if err != nil {
                w.Log.Error(fmt.Sprintf("filepath Golb failed: %s\n", err))
                return err
            }

            if len(w.files) == 0 {
                w.files = append(w.files, matchs...)
                for _, fileName := range w.files {
                    w.Wach <- fileName
                    w.Log.Info(fileName)
                }
            } else {
                tag := false
                for _, newFilename := range matchs {
                    for _, fileName := range w.files {
                        if newFilename == fileName {
                            tag = true
                            break
                        }
                        tag = false
                    }
                    if !tag {
                        w.files = append(w.files, newFilename)
                        w.Wach <- newFilename
                    }
                }
            }
        default:
            time.Sleep(time.Millisecond * 500)
        }
    }
}


/*
import (
    "github.com/fsnotify/fsnotify"
    "log"
    "os"
    "path/filepath"
    "strings"
)

var (
    watcher *fsnotify.Watcher
    Wach    = make(chan string, 10)
)

func Init() (err error) {
    watcher, err = fsnotify.NewWatcher()
    if err != nil {
        log.Println("init fsnotify  file failed: ", err)
        return
    }
    return
}


//递归目录
func Walk(dir string) (err error) {
    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if info.IsDir() {
            path, err := filepath.Abs(path)
            if err != nil {
                log.Println("watch get Recursive file  path failed: ", err)
                return err
            }
            err = watcher.Add(path)
            if err != nil {
                log.Println("watch Recursive file failed: ", err)
                return err
            }
            log.Println("#########watch path########## :", path)
        }
        return nil
    })
    return
}

func WatchFile(dir string) {
    //defer watcher.Close()
    done := make(chan bool)
    go func() {
        if err := Walk(dir); err != nil {
            log.Println("Walk watch filepath filed: ", err)
            return
        }
        for {
            select {
            case event := <-watcher.Events:
                if event.Op&fsnotify.Create == fsnotify.Create {
                    //如果是新创建的目录，则加入到watch队列中
                    fi, err := os.Stat(event.Name)
                    if err == nil && fi.IsDir() {
                        //watcher.Add(event.Name)
                        if err := Walk(event.Name); err != nil {
                            log.Println("Go  For Walk  filepath failed: ", err)
                            return
                        }
                    }
                }
                if event.Op&fsnotify.Write == fsnotify.Write {
                    //log.Println("Write File:", event.Name)
                    if strings.HasSuffix(event.Name, ".log") {
                        Wach <- event.Name
                    }
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()
    <-done
}
*/
