package arc

import (
    "io/ioutil"
    "fmt"
    "time"
    "sync"
    "gopkg.in/yaml.v2"
    "github.com/wildsre/arc/db"
)

var (
    Config *ConfigYaml
    Mysql  *sync.Map
)


type DB struct {
    User   string `yaml:"user"`
    Passwd string `yaml:"passwd"`
    Addr   string `yaml:"addr"`
    Port   string `yaml:"port"`
    Db     string `yaml:"db"`
    Timeout int   `yaml:"timeout"`
    ConnMaxOpen     int `yaml:"conn_max_open"`
    ConnMaxIdle     int `yaml:"conn_max_idle"`
    ConnMaxLifeTime int `yaml:"conn_max_lifetime"`
    Params map[string]string `yaml:"params"`
}


type ConfigYaml struct {
    SvcAddr  string     `yaml:"svc_addr"`
    AdmToken string     `yaml:"admin_token"`
    LogPath string      `yaml:"logpath"`
    MyDB map[string]DB  `yaml:"mysql"`
}


func ParseCfg(configFile string) {
    f, err := ioutil.ReadFile(configFile)
    if err != nil {
        panic(err.Error())
    }
    yaml.Unmarshal(f, &Config)
}

func NewDB(dbname string) (*mysql.DB, error) {
    var mydb *mysql.DB
    mydb = new(mysql.DB)
    cfg, ok := Config.MyDB[dbname]
    if !ok {
        return nil, fmt.Errorf("not have [%s] config, passed", dbname)
    }
    mydb.User = cfg.User
    mydb.Passwd = cfg.Passwd
    mydb.Addr = cfg.Addr + ":" + cfg.Port
    mydb.DBName = cfg.Db
    mydb.Timeout = time.Duration(cfg.Timeout) * time.Second
    mydb.PoolMaxOpenConn = cfg.ConnMaxOpen
    mydb.PoolMaxIdleConn = cfg.ConnMaxIdle
    mydb.PoolMaxLifeTime = cfg.ConnMaxLifeTime
    mydb.Params = cfg.Params
    err := mydb.Connect()
    return mydb, err
}
