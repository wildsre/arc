package main
import (
    "fmt"
    "sync"
    "github.com/wildsre/arc/arc"
    arcModel "github.com/wildsre/arc/model"
    "github.com/wildsre/arc/router"
)

func main() {
    var addr string
    arc.ParseCfg("./conf/arc.yml")
    arc.Mysql = new(sync.Map)
    arcdb, e := arc.NewDB("arc")
    if e != nil {
        fmt.Println("connect db failed, pass")
    } else {
        arc.Mysql.Store("arc", arcdb)
        //fmt.Println("load data from db")
        e = arcModel.ArcInit()
        if e != nil {
            fmt.Println("[error] init db:", e)
        }
        
    }
    fmt.Println("\n=== ARC(Applicaion Resource Center) ===")
    fmt.Println("Running at:", addr)
    fmt.Printf("Example: http://%s/arc/docs?item=example\n\n", arc.Config.SvcAddr)
    r := router.SetRouter()
    r.Run(arc.Config.SvcAddr)
}
