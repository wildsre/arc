package router

import (
    "os"
    "io"
    "fmt"
    "path/filepath"
    "github.com/gin-gonic/gin"
    "github.com/wildsre/arc/api"
    "github.com/wildsre/arc/arc"
)


func SetRouter() *gin.Engine {
    gin.DisableConsoleColor()
    var logf string
    proj_root, _ := filepath.Abs(os.Args[0])
    if arc.Config.LogPath != "" {
        logf = arc.Config.LogPath
    } else {
        logf = filepath.Join(filepath.Dir(proj_root), "arc.log")
    }
    f, err := os.OpenFile(logf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err.Error())
    }
    gin.DefaultWriter = io.MultiWriter(f)
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        // your custom format
        return fmt.Sprintf("%s %s %s %s \"%s\" %s %d\n",
                param.TimeStamp.Format("2006-01-02T15:04:05.000"),
                param.ClientIP,
                param.Method,
                param.Path,
                param.Request.UserAgent(),
                param.Latency,
                param.StatusCode,
        )
    }))
    router.Use(gin.Recovery())
    router.GET("/arc/docs", arcapi.Docs)
    router.GET("/arc/apply", arcapi.Apply)
    router.GET("/arc/add", arcapi.Add)
    router.GET("/arc/delete", arcapi.Delete)
    router.GET("/arc/get", arcapi.Get)
    router.GET("/arc/update", arcapi.Update)
    router.GET("/arc/userget", arcapi.UserGet)
    return router
}
