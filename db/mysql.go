package mysql

import (
    "encoding/json"
    "database/sql"
    driver "github.com/go-sql-driver/mysql"
    "time"
    //"fmt"
)

//dsn: root:123@w321@tcp(127.0.0.1:3307)/test


type DB struct {
    Conn             *sql.DB
    User             string            // Username
    Passwd           string            // Password (requires User)
    Net              string            // Network type
    Addr             string            // Network address (requires Net)
    DBName           string            // Database name
    Params           map[string]string // Connection parameters
    Collation        string            // Connection collation
    Loc              *time.Location    // Location for time.Time values
    MaxAllowedPacket int               // Max packet size allowed
    ServerPubKey     string            // Server public key name

    PoolMaxOpenConn  int
    PoolMaxIdleConn  int
    PoolMaxLifeTime  int

    TLSConfig string // TLS configuration name

    Timeout      time.Duration // Dial timeout
    ReadTimeout  time.Duration // I/O read timeout
    WriteTimeout time.Duration // I/O write timeout

    AllowAllFiles           bool // Allow all files to be used with LOAD DATA LOCAL INFILE
    AllowCleartextPasswords bool // Allows the cleartext client side plugin
    AllowNativePasswords    bool // Allows the native password authentication method
    AllowOldPasswords       bool // Allows the old insecure password method
    ClientFoundRows         bool // Return number of matching rows instead of rows changed
    ColumnsWithAlias        bool // Prepend table alias to column names
    InterpolateParams       bool // Interpolate placeholders into query string
    MultiStatements         bool // Allow multiple statements in one query
    ParseTime               bool // Parse time values to time.Time
    RejectReadOnly          bool // Reject read-only connections
    // contains filtered or unexported fields
}


//db.conn.SetMaxOpenConns(n int) 设置打开数据库的最大连接数
//db.SetMaxIdleConns(n int) 最大空闲连接数
//db.SetConnMaxLifetime(time.Second * 20) 每个连接的最长生命周期, 解决的问题: 在客户端连接池中的一条空闲链接，可能是一条已经被MySQL服务端关闭掉的链接
func (db *DB) Connect() error {
    // add default params
    //db.Params = make(map[string]string)
    //db.Params["charset"]="utf8"
    db.Net="tcp"
    db.AllowNativePasswords=true
    //copy Db -> driver.Config
    tmp, _ := json.Marshal(db)
    cfg := new(driver.Config)
    _ = json.Unmarshal(tmp, cfg)
    //connect db
    var err error
    dsn := cfg.FormatDSN()
    //fmt.Println(dsn)
    db.Conn, err = sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    if db.PoolMaxLifeTime != 0 {
        db.Conn.SetConnMaxLifetime(time.Second * time.Duration(db.PoolMaxLifeTime))        
    }
    if db.PoolMaxIdleConn != 0 {
        db.Conn.SetMaxIdleConns(db.PoolMaxIdleConn)
    }
    if db.PoolMaxOpenConn !=0 {
        db.Conn.SetMaxOpenConns(db.PoolMaxOpenConn)
    }
    err = db.Conn.Ping()
    return err
}

func (db *DB) Exec(query string) (result []map[string]string, err error)  {
    var ret []map[string]string
    var e error
    stmt, e := db.Conn.Prepare(query)
    if e != nil {
        return ret, e
    }
    defer stmt.Close()
    rows, e := stmt.Query()
    if e != nil {
        return ret, e
    } 
    defer rows.Close()
    cols, e := rows.Columns()
    vals := make([]sql.RawBytes, len(cols))
    scanArgs := make([]interface{}, len(cols))
    for i:=0; i<len(cols); i++ {
        scanArgs[i] = &vals[i]
    }
    for rows.Next() {
        err  = rows.Scan(scanArgs...)
        d := make(map[string]string)
        for j:=0; j<len(cols); j++ {
            if vals[j] == nil {
                d[cols[j]] = "null"
            } else {
                d[cols[j]] = string(vals[j])
            }
        }
        ret = append(ret, d)
    }
    return ret, e
}


func (db *DB) ExecOnce(query string) (result []map[string]string, err error)  {
    var ret []map[string]string
    var e error
    stmt, e := db.Conn.Prepare(query)
    if e != nil {
        return ret, e
    }
    defer stmt.Close()
    defer db.Conn.Close()
    rows, e := stmt.Query()
    if e != nil {
        return ret, e
    } 
    defer rows.Close()
    cols, e := rows.Columns()
    vals := make([]sql.RawBytes, len(cols))
    scanArgs := make([]interface{}, len(cols))
    for i:=0; i<len(cols); i++ {
        scanArgs[i] = &vals[i]
    }
    for rows.Next() {
        err  = rows.Scan(scanArgs...)
        d := make(map[string]string)
        for j:=0; j<len(cols); j++ {
            if vals[j] == nil {
                d[cols[j]] = "null"
            } else {
                d[cols[j]] = string(vals[j])
            }
        }
        ret = append(ret, d)
    }
    return ret, e
}

