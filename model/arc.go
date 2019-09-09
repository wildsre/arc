package arc
import (
    "fmt"
    "time"
    "sync"
    "github.com/satori/go.uuid"
    "strings"
    "crypto/md5"
    json "github.com/json-iterator/go"
    "github.com/wildsre/arc/db"
    "github.com/wildsre/arc/arc"
)


var (
    User  sync.Map
//    Group sync.Map
    Token sync.Map
    Data  sync.Map
)


type SpaceData struct {
    App         string `json:"app"`
    Resource    string `json:"resource"`
    Item        string `json:"item"`
    Value       string `json:"value"`
}

type UserData struct {
    UserName string `json:"username"`
    Token string `json:"token"`
    Role  string `json:"role"`
    Group map[string]string `json:"group"`
    DefaultSpace string `json:"default-space"`
    Space map[string]string `json:"space"`
}

//map[string]map[string]map[string]map[string]map[string]string
//name: wangzhou@kingsoft.com
//role: super-admin,admin,user
//sql := fmt.Sprintf("insert into arc_user (`name`, `token`, `role`, `default_space`, `space`) values ('%s','%s','%s','%s','%s');",
//        data.UserName, data.Token, data.Role, data.DefaultSpace, json.Marshal(data. )
func UserAdd(name, role string) (userdata *UserData, err error) {
    var e error
    _data, ok := User.LoadOrStore(name, &UserData{})
    data := _data.(*UserData)
    if ok {
        return nil, fmt.Errorf("User [%s] is existed", name)
    }
    data.UserName = name
    data.Token = TokenAdd(name)
    data.Role = role
    data.DefaultSpace = "default"
    data.Space = make(map[string]string)
    data.Space["default"] = "764"
    data.Group = make(map[string]string)
    //add default space
    SpaceAdd(data.UserName + "+" + data.DefaultSpace)
    j_space, _ := json.Marshal(data.Space)
    j_group, _ := json.Marshal(data.Group)
    _db, ok := arc.Mysql.Load("arc")
    if !ok { return data, e }
    db := _db.(*mysql.DB)
    sql := fmt.Sprintf("insert into arc_user (`name`, `token`, `role`, `default_space`, `space`, `user_group`) values ('%s', '%s', '%s', '%s', '%s', '%s');", 
                data.UserName, data.Token, data.Role, data.DefaultSpace, string(j_space), string(j_group))
    _, e = db.Exec(sql)
    return data, e
}

//func UserDelete() { }
//func UserUpdate() { }

func UserGet(name string) (userdata *UserData) {
    data, ok := User.Load(name)
    if !ok {
        return nil
    }
    return data.(*UserData)
}

func TokenAdd(user string) (token string) {
    var (
        _uuid []byte
        tk string
        ok bool
    )
    for {
        _uuid = uuid.Must(uuid.NewV4()).Bytes()
        tk = fmt.Sprintf("%x", md5.Sum(_uuid))
        _, ok = Token.Load(tk)
        if ok { continue }
        Token.Store(tk, user)
        break
    }
    return tk
}

//func TokenDelete() { }
//name not allow '+' char
func SpaceAdd(name string) error { 
    //sql := fmt.Sprintf("insert into arc_usergroup (`name`) values ('`%s`');", name)
    var e error
    _, ok := Data.Load(name)
    if ok {
        return fmt.Errorf("space is existed")
    } 
    Data.Store(name, &sync.Map{})
    return e
}

//cannot delete default space
//func SpaceDelete() { }
//func SpaceUpdate() { }

func SpaceGet(space string) (data *sync.Map) {
    d, ok := Data.Load(space)
    if !ok {
        return nil    
    }
    return d.(*sync.Map)
}

//func SpaceList() { }
//func SpaceSetDefault() { }
//func GroupAdd() { }
//func GroupDelete() { }
//func GroupUpdate() { }
//func GroupGet() { }

//data: app -> resource -> item -> value -> v
func DataAdd(space string, data []*SpaceData) {
    var  app, resource, item, value *sync.Map
    var storeData []*SpaceData
    var sql string
    _d, _ := Data.Load(space)
    d := _d.(*sync.Map)
    for idx, _ := range(data) {
        app = &sync.Map{}
        resource = &sync.Map{}
        item = &sync.Map{}
        value = &sync.Map{}
        sd := data[idx]
        value.Store(sd.Value, "")
        _app, ok := d.Load(sd.App)
        if !ok {
            item.Store(sd.Item, value) 
            resource.Store(sd.Resource, item)
            d.Store(sd.App, resource)
            storeData = append(storeData, sd)
            continue
        }
        app = _app.(*sync.Map)
        _resource, ok := app.Load(sd.Resource)
        if !ok {
            item.Store(sd.Item, value)
            app.Store(sd.Resource, item)
            storeData = append(storeData, sd)
            continue
        }
        resource = _resource.(*sync.Map)
        _item, ok := resource.Load(sd.Item)
        if !ok {
            resource.Store(sd.Item, value)
            storeData = append(storeData, sd)
            continue
        }
        item = _item.(*sync.Map)
        _, ok = item.Load(sd.Value)
        if !ok {
            item.Store(sd.Value, "")
            storeData = append(storeData, sd)
        }
    }
    _db, ok := arc.Mysql.Load("arc")
    if !ok {
        return
    }
    db := _db.(*mysql.DB)
    for _, sd := range(storeData) {
        sql = fmt.Sprintf("insert into arc_data (`space`, `app`, `resource`, `item`, `value`, `value_tag`) values ( '%s', '%s', '%s', '%s', '%s', '%s')",
                space, sd.App, sd.Resource, sd.Item, sd.Value, "")
        _, e:= db.Exec(sql)
        if e != nil {
            fmt.Println("insert data to arc_data error", sql, e)
        }
    }
}

//data: app, resource, item, value
//return number of delete
func DataDelete(q *SpaceData, data *sync.Map, space string) bool {
    var e error
    var db *mysql.DB
    var sql, ts string
    var deleteResult bool
    ts = time.Now().Format("2006-01-02 15:04:05")
    _db, hasDb := arc.Mysql.Load("arc")
    if hasDb {
        db = _db.(*mysql.DB)
    }
    if q.App == "" { return false }
    _app, ok := data.Load(q.App)
    if !ok { return false }
    if q.Resource == "" { 
        //debug
        //fmt.Println("appdelete")
        deleteResult = AppDelete(q, data)
        if deleteResult && (db != nil) {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s';", ts, space, q.App)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return deleteResult
    }
    app := _app.(*sync.Map)
    //delete all resource
    if q.Resource == "all" {
        //debug
        //fmt.Println("resource delete all")
        data.Delete(q.App)
        if db != nil {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s';", ts, space, q.App)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return true
    }
    _resource, ok := app.Load(q.Resource)
    if !ok { return true }
    if q.Item == "" {
        //debug
        //fmt.Println("resource delete")
        deleteResult = ResourceDelete(q, app)
        if deleteResult && (db != nil) {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s' and `resource`='%s';", 
                ts, q.App, q.Resource)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return deleteResult
    }
    resource := _resource.(*sync.Map)
    //delete all item
    if q.Item == "all" {
        //debug
        //fmt.Println("item delete all")
        app.Delete(q.Resource)
        if db != nil {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s' and `resource`='%s';", 
                ts, space, q.App, q.Resource)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return true
    }
    _item, ok := resource.Load(q.Item)
    if !ok { return true }
    if q.Value == "" {
        //debug
        //fmt.Println("item delete ", db)
        deleteResult = ItemDelete(q, resource)
        if (deleteResult && (db != nil)) {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s' and `resource`='%s' and `item`='%s';", 
                ts, space, q.App, q.Resource, q.Item)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return deleteResult
    }
    item := _item.(*sync.Map)
    //delete all value
    if q.Value == "all" {
        //debug
        //fmt.Println("value delete all")
        resource.Delete(q.Item)
        if db != nil {
            sql = fmt.Sprintf("update arc_data set deleted_at='%s', `deleted`=1 where `deleted`=0 and `space`='%s' and `app`='%s' and `resource`='%s' and `item`='%s';", 
                ts, space, q.App, q.Resource, q.Item)
            _, e = db.Exec(sql)
            //debug
            if e != nil {
            fmt.Println(sql, e)
            }
        }
        return true
    }
    _, ok = item.Load(q.Value)
    if !ok { return true }
    //debug
    //fmt.Println("value delete")
    deleteResult = ValueDelete(q, item)
    if deleteResult && (db != nil) {
        sql = fmt.Sprintf("update arc_data set deleted_at='%s', deleted=1 where `deleted`=0 and `space`='%s' and `app`='%s' and `resource`='%s' and `item`='%s' and `value`='%s';", 
            ts, space, q.App, q.Resource, q.Item, q.Value)
        _, e = db.Exec(sql)
        //debug
        if e != nil {
        fmt.Println(sql, e)
        }
    }
    return deleteResult
}

func ValueDelete(q *SpaceData, item *sync.Map) bool {
    //debug
    //fmt.Println("value-delete:", q, item)
    item.Delete(q.Value)
    return true
}

func ItemDelete(q *SpaceData, resource *sync.Map) bool {
    //debug
    //fmt.Println("item-delete:", q, resource)
    resource.Delete(q.Item)
    return true
}

func SyncMapDeleteAll(data *sync.Map) {
    data.Range(func(k, _ interface{}) bool {
        data.Delete(k)
        return true
    })
}

func ResourceDelete(q *SpaceData, app *sync.Map) bool {
        //debug
        //fmt.Println("res-delete:", q, app)
        //only can delete empty resource
        _res, ok := app.Load(q.Resource)
        if !ok { return true}
        res := _res.(*sync.Map)
        //check item in resource
        if hasItem(res) { return false }
        app.Delete(q.Resource)
    return true
    
}


func AppDelete(q *SpaceData, space *sync.Map) bool {
    //debug
    //fmt.Println("app-delete:", q, space)
    _res, ok := space.Load(q.App)
    if !ok { return true }
    res := _res.(*sync.Map)
    if hasItem(res) { return false }
    space.Delete(q.App)
    return true
}


func hasItem(data *sync.Map) bool {
    var ret bool
    data.Range(func(k, _ interface{}) bool {
        if k != nil {
            ret = true
            //stop range
            return false
        }
        return true
    })
    return ret
}

//func DataUpdate() { }

//qs: [{"App":"", "Resource":"", "Value":"", "Item":""}, ]
//data: spacedata
func DataGet(qs []*SpaceData, data *sync.Map) []*SpaceData {
    var ret, _ret []*SpaceData
    for _, q := range(qs) {
        _ret = _DataGet(q, data)
        ret = append(ret, _ret...)
    }
    return ret
}

func _DataGet(q *SpaceData, data *sync.Map) []*SpaceData {
    var ret []*SpaceData
    if strings.ToUpper(q.App) == "ALL" {
        ret = GetAll("app", q, data)
        return ret
    }
    if strings.ToUpper(q.App) == "LIST" {
        ret = GetList("app", q, data)
        return ret
    }
    _app, ok := data.Load(q.App)
    if !ok { return ret }
    app := _app.(*sync.Map)
    if strings.ToUpper(q.Resource) == "ALL" {
        ret = GetAll("resource", q, app)
        return ret
    }
    if strings.ToUpper(q.Resource) == "LIST" {
        ret = GetList("resource", q, app)
        return ret
    }
    _resource, ok := app.Load(q.Resource)
    if !ok { return ret }
    resource := _resource.(*sync.Map)
    if strings.ToUpper(q.Item) == "ALL" {
        ret = GetAll("item", q, resource)
        return ret
    }
    if strings.ToUpper(q.Item) == "LIST" {
        ret = GetList("item", q, resource)
        return ret
    }
    _item, ok := resource.Load(q.Item)
    if !ok { return ret }
    item := _item.(*sync.Map)
    ret = GetAll("value", q, item)
    return ret
}

func GetAll(pos string, q *SpaceData, data *sync.Map) []*SpaceData {
    var ret []*SpaceData
    if pos == "app" {
        data.Range(func(app, _v1 interface{}) bool {
            v1 := _v1.(*sync.Map)
            v1.Range(func(resource, _v2 interface{}) bool {
                v2 := _v2.(*sync.Map)
                v2.Range(func(item, _v3 interface{}) bool {
                    //fmt.Println(item, _v3)
                    v4 := _v3.(*sync.Map)
                    v4.Range(func(value, _v5 interface{}) bool {
                        sd := SpaceData{App:app.(string), Resource:resource.(string), Item:item.(string), Value:value.(string)}
                        ret = append(ret, &sd)
                        return true
                    })
                    return true
                })
                return true
            })
            return true
        })
    } else if pos == "resource" {
        data.Range(func(resource, _v1 interface{}) bool {
            v1 := _v1.(*sync.Map)
            v1.Range(func(item, _v2 interface{}) bool {
                v2 := _v2.(*sync.Map)
                v2.Range(func(value, _v3 interface{}) bool {
                    sd := SpaceData{App:q.App, Resource:resource.(string), Item:item.(string), Value: value.(string)}
                    ret = append(ret, &sd)
                    return true
                })
                return true
            })
            return true
        })
    } else if pos == "item" {
        data.Range(func(item, _v1 interface{} ) bool {
            v1 := _v1.(*sync.Map)
            v1.Range(func(value, _v2 interface{}) bool {
                sd := SpaceData{App:q.App, Resource:q.Resource, Item:item.(string), Value:value.(string)}
                ret = append(ret, &sd)
                return true
            })
            return true
        })
    } else if pos == "value" {
        //get item values
        data.Range(func(value, _v1 interface{}) bool {
            sd := SpaceData{App:q.App, Resource:q.Resource, Item:q.Item, Value:value.(string)}
            ret = append(ret, &sd)
            return true
        })
    } else {
        return ret
    }
    return ret
}

func GetList(pos string, q *SpaceData, data *sync.Map) []*SpaceData {
    var ret []*SpaceData
    if pos == "app" {
        data.Range(func(app, _ interface{}) bool {
            sd := SpaceData{App:app.(string),}
            ret = append(ret, &sd)
            return true
        })
    } else if pos == "resource" {
        data.Range(func(resource, _ interface{}) bool {
            sd := SpaceData{App:q.App, Resource:resource.(string),}
            ret = append(ret, &sd)
            return true
        })
    } else if pos == "item" {
        data.Range(func(item, _ interface{}) bool {
            sd := SpaceData{App:q.App, Resource:q.Resource, Item:item.(string),}
            ret = append(ret, &sd)
            return true
        })
    } else {
        return ret
    }
    return ret
}
//load data from mysql
func ArcInit() error {
    _db, _ := arc.Mysql.Load("arc")
    db := _db.(*mysql.DB)
    //load user
    sql := "select * from arc_user where deleted=0;"
    users, e := db.Exec(sql)
    if e != nil { return e }
    loadUser(users)
    //load data
    sql = "select * from arc_data where deleted=0;"
    data, e := db.Exec(sql)
    if e != nil { return e }
    loadData(data)
    return e
}


func loadData(data []map[string]string) {
    for idx, _ := range(data) {
        _space, _ := Data.LoadOrStore(data[idx]["space"], &sync.Map{})
        space := _space.(*sync.Map)
        _app, _ := space.LoadOrStore(data[idx]["app"], &sync.Map{})
        app := _app.(*sync.Map)
        _resource, _ := app.LoadOrStore(data[idx]["resource"], &sync.Map{})
        resource := _resource.(*sync.Map)
        _item, _ := resource.LoadOrStore(data[idx]["item"], &sync.Map{})
        item := _item.(*sync.Map)
        item.Store(data[idx]["value"], "")
    }
}


func loadUser(users []map[string]string) {
    var d map[string]string
    for _, user := range(users) {
        d = make(map[string]string)
        _data, _ := User.LoadOrStore(user["name"], &UserData{})
        data := _data.(*UserData)
        data.UserName = user["name"]
        data.Token = user["token"]
        data.Role= user["role"]
        data.DefaultSpace = user["default_space"]
        data.Group = make(map[string]string)
        e := json.Unmarshal([]byte(user["user_group"]), &d)
        if e == nil {
            for k, v := range(d) { data.Group[k] = v }
        }
        data.Space = make(map[string]string)
        e = json.Unmarshal([]byte(user["space"]), &d)
        if e == nil {
            for k, v := range(d) { 
                data.Space[k] = v 
                SpaceAdd(data.UserName + "+" + k)
            }
        }
        Token.Store(data.Token, data.UserName)
    }
}
