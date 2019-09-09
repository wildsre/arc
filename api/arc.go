package arcapi
import (
    "fmt"
    "strings"
    "sync"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/wildsre/arc/model"
    arcArc "github.com/wildsre/arc/arc"
    "github.com/wildsre/arc/docs"
)

//TODO: add auth check --done 2019.09.06

type RcvData struct {
    Token string `json:"token"`
    Space string `json:"space"`
    Data  []*arc.SpaceData `json:"data"`
}

func Docs(c *gin.Context) {
    item := c.DefaultQuery("item", "")
    if item == "example" {
        c.String(200, docs.Example())
    } else {
        c.String(404, "not found")
    }
}


//operate: read, write, delete, update
func SpaceAuthCheck(space string, operate string, token string) string {
    op2auth := map[string]uint64{"read":4, "write":2, "delete":1, "update":3 }
    _, ok := arc.Data.Load(space)
    if !ok {
        return fmt.Sprintf("space [%s] not exist", space)
    }
    //get configed space auth
    username := strings.Split(space, "+")[0]
    spacename := strings.Split(space, "+")[1]
    _userdata, _ := arc.User.Load(username)
    userdata := _userdata.(*arc.UserData)
    if RoleAuthCheck(op2auth[operate], userdata.Space[spacename], "other") {
        return ""
    }
    myname, ok := arc.Token.Load(token)
    if !ok {
        return fmt.Sprintf("token [%s] not exist", token)
    }
    if myname == username {
        if RoleAuthCheck(op2auth[operate], userdata.Space[spacename], "owner") {
            return "" 
        } else {
            return fmt.Sprintf("role [%s] not have space [%s] [%s] auth", "owner", space, operate)
        }
    }
    return fmt.Sprintf("auth fail")
    //TODO check if my and user are in same group
}


//want: want auth(read:4, write:2, delete:1)
func RoleAuthCheck(want uint64, space_auth, role string) bool {
    var auth uint64
    var res bool
    if role == "owner" {
        auth, _ = strconv.ParseUint(string(space_auth[0]), 10, 8)
    } else if role == "group" {
        auth, _ = strconv.ParseUint(string(space_auth[1]), 10, 8)
    } else {
        auth, _ = strconv.ParseUint(string(space_auth[2]), 10, 8)
    }
    if (want & auth) > 0 {
        res = true
    } else {
        res = false
    }
    return res
}


//GET: /arc/apply?name=wangzhou@kingsoft.com[&token=123&role=super-admin]
//default role is [user]
func Apply(c *gin.Context) {
    var role string
    name := c.DefaultQuery("name", "")
    token := c.DefaultQuery("token", "")
    if token == arcArc.Config.AdmToken {
        role = strings.ToLower(c.DefaultQuery("role", "user"))
    } else {
        role = "user"
    }
    if name == "" {
        c.JSON(400, gin.H{"status":"fail", "reason":"name is empty"})
        return
    }
    userdata, err := arc.UserAdd(name, role)
    if err != nil {
        c.JSON(400, gin.H{"status":"fail", "reason":fmt.Sprintf("%s", err)})
    } else {
        c.JSON(200, gin.H{"status":"success", "data": userdata})
    }
}


func GenSpace(_space, token string) string {
    var space string
    default_user := "public@local"
    default_space := "default"
    if strings.Contains(_space, "+") {
        space = _space
    } else if (token != "" ) {
        //generate user define space
        _username, ok := arc.Token.Load(token)
        if !ok {
            return ""
        }
        username := _username.(string)
        if _space == "" {
            //get user default space
            _userdata, _ := arc.User.Load(username)
            userdata := _userdata.(*arc.UserData)
            space = username + "+" + userdata.DefaultSpace
        } else {
            space = username + "+" + _space
        }
    } else {
        space = default_user + "+" + default_space
    }
    return space
}


//GET: /arc/get?token=xxx&app=a1&resource=r1&item=i1[&space=public@local+default][&format=json|flatten]
func Get(c *gin.Context) {
    var e, token, space, format string
    token = c.DefaultQuery("token", "")
    format = c.DefaultQuery("format", "json")
    qs := []*arc.SpaceData{ &arc.SpaceData{
            App: c.DefaultQuery("app", ""),
            Resource: c.DefaultQuery("resource", ""),
            Item: c.DefaultQuery("item", ""),
            Value: c.DefaultQuery("value", "") },
    }
    _space := c.DefaultQuery("space", "default")
    _space = strings.Replace(_space, " ", "+", 1)
    //debug
    //fmt.Println(_space)
    space = GenSpace(_space, token)
    //debug
    //fmt.Println(space)
    if space == "" {
        c.JSON(404, gin.H{"status": "fail", "reason": "space not found"})
        return
    }
    //auth check
    e = SpaceAuthCheck(space, "read", token)
    if e != "" {
        c.JSON(401, gin.H{"status": "fail", "reason": e})
        return
    }
    _spacedata, ok:= arc.Data.Load(space)
    if !ok {
        c.JSON(404, gin.H{"status":"fail", "reason": "space not found or unauthorized"})
        return
    }
    spacedata := _spacedata.(*sync.Map)
    res := arc.DataGet(qs, spacedata)
    if len(res) == 0 {
        c.JSON(404, gin.H{"status": "fail", "reason":"empty data set"})
    } else {
        if format == "json" {
            c.JSON(200, gin.H{"status": "success", "code": 20000, "data":res, "size": len(res)})
        } else {
            //flatten result
            txt := FlattenSpaceData(res)
            c.String(200, txt)
        }
    }
}


func FlattenSpaceData(sds []*arc.SpaceData) string {
    var ret []string
    for idx, _ := range(sds) {
        _sd := sds[idx]
        _d := []string{_sd.App, _sd.Resource, _sd.Item, _sd.Value}
        ret = append(ret, strings.Join(_d, ","))
    }
    return strings.Join(ret, "\n") + "\n"
}


//POST: /arc/add
func Add(c *gin.Context) {
    var e, token, space, _space string
    var sds []*arc.SpaceData
    //parse request data
    if c.Request.Method == "GET" {
        token = c.DefaultQuery("token", "")
        _space = c.DefaultQuery("space", "")
        sds = []*arc.SpaceData{ &arc.SpaceData{
                App: c.DefaultQuery("app", ""),
                Resource: c.DefaultQuery("resource", ""),
                Item: c.DefaultQuery("item", ""),
                Value: c.DefaultQuery("value", "") },
        }
    } else if c.Request.Method == "POST" {
        var rcv RcvData
        c.Bind(&rcv)
        token = rcv.Token
        _space = rcv.Space
        sds = rcv.Data
    } else {
        c.String(405, "unsupport mothed")
    }
    for _, sd := range(sds) {
        e = RcvParamCheck(sd, "add")
        if e != "" {
            c.JSON(400, gin.H{"status": "fail", "reason": e})
            return
        }
    }
    //
    space = GenSpace(_space, token)    
    //fmt.Println(space)
    if space == "" {
        c.JSON(404, gin.H{"status": "fail", "reason": "space not found"})
        return
    }
    e = SpaceAuthCheck(space, "write", token)
    if e != "" {
        c.JSON(401, gin.H{"status": "fail", "reason": e})
        return
    }
    _, ok := arc.Data.Load(space)
    if !ok {
        c.JSON(400, gin.H{"status":"fail", "reason": fmt.Sprintf("space [%s] not exist", space)})
        return 
    }
    arc.DataAdd(space, sds)
    c.JSON(200, gin.H{"status":"success", "code":20000})
}


//GET: /arc/delete?token=xxx&app=a1&resource=r1&item=i1&value=v1
func Delete(c *gin.Context) {
    var e, token, space, _space string
    token = c.DefaultQuery("token", "")
    _space = c.DefaultQuery("space", "")
    q := &arc.SpaceData{
        App:      c.DefaultQuery("app", ""),
        Resource: c.DefaultQuery("resource", ""),
        Item:     c.DefaultQuery("item", ""),
        Value:    c.DefaultQuery("value", "") }
    space = GenSpace(_space, token)
    if space == "" {
        c.JSON(404, gin.H{"status": "fail", "reason": "space not found"})
        return
    }
    e = SpaceAuthCheck(space, "delete", token)
    if e != "" {
        c.JSON(401, gin.H{"status": "fail", "reason": e})
        return
    }
    _spacedata, ok := arc.Data.Load(space)
    if !ok {
        c.JSON(400, gin.H{"status":"fail", "reason": fmt.Sprintf("space [%s] not exist", space)})
        return
    }
    spacedata := _spacedata.(*sync.Map)
    ok = arc.DataDelete(q, spacedata, space)
    if ok {
        c.JSON(200, gin.H{"status":"success", "code": 20000})
    } else {
        c.JSON(400, gin.H{"status":"fail", "reason": "unknow"})
    }
}

func RcvParamCheck(sd *arc.SpaceData, op string) (err string) {
    var e string
    //debug
    fmt.Println(sd)
    if op == "update" {
        if (sd.App == "") || (sd.Resource =="") || (sd.Item == "") {
            e = "miss query params"
        }
    } else {
        if (sd.App == "") || (sd.Resource =="") || (sd.Item == "") || (sd.Value=="") {
            e = "miss query params"
        }
    }
    if (sd.App == "all") || (sd.App=="list") || (sd.Resource =="all") || (sd.Resource =="list") || (sd.Item == "all") ||(sd.Item=="list") || (sd.Value=="all") ||(sd.Value=="list") {
            e = "params can not be [all or list]"
        }
    return e
}

//GET: /arc/update?token=xxx&app=a1&resource=r1&item=i1&value=v1
// delete item + add value
func Update(c *gin.Context) {
    var e, space, _space, token string
    del_q := &arc.SpaceData{
        App:  c.DefaultQuery("app", ""),
        Resource: c.DefaultQuery("resource", ""),
        Item: c.DefaultQuery("item", ""), }    
    e = RcvParamCheck(del_q, "update")
    fmt.Println("rcvcheck", e)
    if e != "" {
        c.JSON(400, gin.H{"status": "fail", "reason": e})
        return
    }
    add_q := []*arc.SpaceData{ &arc.SpaceData{
                App: c.DefaultQuery("app", ""),
                Resource: c.DefaultQuery("resource", ""),
                Item: c.DefaultQuery("item", ""),
                Value: c.DefaultQuery("value", "") },
    }
    e = RcvParamCheck(add_q[0], "add")
    if e != "" {
        c.JSON(400, gin.H{"status": "fail", "reason": e})
        return
    }
    token = c.DefaultQuery("token", "")
    _space = c.DefaultQuery("space", "")
    space = GenSpace(_space, token)
    if space == "" {
        c.JSON(404, gin.H{"status": "fail", "reason": "space not found"})
        return
    }
    e = SpaceAuthCheck(space, "update", token)
    if e != "" {
        c.JSON(401, gin.H{"status": "fail", "reason": e})
        return
    }
    _spacedata, _ := arc.Data.Load(space)
    spacedata  := _spacedata.(*sync.Map)
    deleted := arc.DataDelete(del_q, spacedata, space)
    if !deleted {
        c.JSON(400, gin.H{"status": "fail", "reason": "delete data failed"})
        return
    }
    arc.DataAdd(space, add_q)
    c.JSON(200, gin.H{"status": "success", "code":20000})
}


//GET: /arc/userget?token=xxx&name=wangzhou@kingsoft.com
func UserGet(c *gin.Context) {
    var token, name string
    token = c.DefaultQuery("token", "")
    name = c.DefaultQuery("name", "")
    _user, ok := arc.Token.Load(token)
    if !ok {
        c.JSON(400, gin.H{"status":"fail", "reason": "error token"})
        return
    }
    user := _user.(string)
    userinfo := arc.UserGet(user)
    if userinfo == nil {
        c.JSON(400, gin.H{"status":"fail", "reason": "load user info fail"})
        return
    }
    //allow get self info
    if userinfo.UserName == name {
        userinfo = arc.UserGet(name)
        c.JSON(200, gin.H{"status": "success", "data": userinfo})
        return
    }
    if (userinfo.Role != "admin") && (userinfo.Role != "super-admin") {
        c.JSON(400, gin.H{"status":"fail", "reason": "auth fail"})
        return
    }
    userinfo = arc.UserGet(name)
    c.JSON(200, gin.H{"status": "success", "data": userinfo, "code":20000})
}
