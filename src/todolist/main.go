package main

import (
    "crypto/sha1"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "io/ioutil"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "os/signal"
    "os"
    "syscall"
//    "log"
    "math/rand"
    "net/http"
    "net/url"
    "path"
    "runtime/debug"
    "strconv"
    "strings"
    "time"
    "net"
//    "log4go"
)

import log "log4go"

const (
//    UPLOAD_DIR   = "./uploads"
    TEMPLATE_DIR = "./template/"
    ListDir      = 0x0001
    filename = "./log/flw.log"
)

type GetData struct {
    Item []*ListData
}

type IndexData struct {
    Index string
    Num   int64
}

type UserData struct {
    Uid      int64
    Username string
    Password string
    Email    string
    Random   string
}

type ListData struct {
    Username    string 
    Gid         int64 
    Project     string
    Status      int
    Change      string
    Time        int64
    List        string
    Version     int
    Id          int
}

//type mongoData struct {
//    Gid     int64
//    Project string
//    Time    int64
//    Status  int
//    List    string
//}

var templates = make(map[string]*template.Template)

func init() {

    fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
    if err != nil {
        panic(err)
        return
    }
    var templateName, templatePath string
    for _, fileInfo := range fileInfoArr {
        templateName = fileInfo.Name()
        if ext := path.Ext(templateName); ext != ".html" {
            continue
        }
        templatePath = TEMPLATE_DIR + "/" + templateName
        log.Info("Loading template:"+templatePath)
        t := template.Must(template.ParseFiles(templatePath))
        tmpl := strings.Split(templateName, ".html")[0]
        templates[tmpl] = t
    }
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}

func renderHtml(w http.ResponseWriter, tmpl string, locals map[string]interface{}) (err error) {
    err = templates[tmpl].Execute(w, locals)
    check(err)
    return
}

func isExists(path string) bool {
    _, err := os.Stat(path)
    if err == nil {
        return true
    }
    return os.IsExist(err)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        locals := make(map[string]interface{})
        err := renderHtml(w, "register", locals)
        check(err)
        return
    }

    if r.Method == "POST" {
        //mongodb opration 
        mongoCon, err := mgo.Dial("127.0.0.1:27018")
        check(err)
        defer mongoCon.Close()
        mongoCon.SetMode(mgo.Monotonic, true)

        // 获取数据库,获取集合
        l := mongoCon.DB("test_go").C("user")
        i := mongoCon.DB("test_go").C("index")
        var doc IndexData
        change := mgo.Change{
            Update:    bson.M{"$inc": bson.M{"num": 1}},
            ReturnNew: true,
        }
        _, err = i.Find(bson.M{"index": "uid"}).Apply(change, &doc)
        if err != nil {
            indexUid := IndexData{
                Index: "uid",
                Num:   1,
            }
            err := i.Insert(indexUid)
            check(err)
            doc.Num = 1
        }
        log.Info(doc.Num)
        users := UserData{
            Uid:      doc.Num,
            Username: r.FormValue("username"),
            Password: r.FormValue("password"),
            Email:    r.FormValue("email"),
        }
        err = l.Insert(users)
        check(err)
        http.Redirect(w, r, "login", http.StatusFound)
        return
    }

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    milliseconds:=time.Now().UnixNano()/10E5
    if r.Method == "GET" {
        locals := make(map[string]interface{})
        err := renderHtml(w, "login", locals)
        check(err)
        return
    }

    if r.Method == "POST" {
        //mongodb opration 
        mongoCon, err := mgo.Dial("127.0.0.1:27018")
        check(err)
        defer mongoCon.Close()
        mongoCon.SetMode(mgo.Monotonic, true)
        l := mongoCon.DB("test_go").C("user")
        user := r.FormValue("username")
        passwd := r.FormValue("password")
        result := UserData{}
        err = l.Find(&bson.M{"username": user}).One(&result)
        if err != nil {
            log.Info("not found the username "+ user)
            panic(err)
        }

        if passwd == result.Password {
            //set cookie
            // create random string
            randNum := rand.New(rand.NewSource(milliseconds))
            h := sha1.New()
            randString := result.Password + strconv.Itoa(randNum.Intn(100))
            log.Info("rand string "+ randString)
            io.WriteString(h, randString)
            hashString := fmt.Sprintf("%x", h.Sum(nil))
            log.Info(hashString)
            err = l.Update(bson.M{"username": user}, bson.M{"$set": bson.M{"random": hashString}})
            check(err)
            cookie1 := http.Cookie{Name: "value", Value: hashString}
            cookie2 := http.Cookie{Name: "username", Value: user}
            http.SetCookie(w, &cookie1)
            http.SetCookie(w, &cookie2)
            log.Info("set cookie!")
            http.Redirect(w, r, "/", http.StatusFound)
        } else {
            log.Info("password is not right "+ passwd)
        }
        return
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    locals := make(map[string]interface{})
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check(err)
    defer mongoCon.Close()
    mongoCon.SetMode(mgo.Monotonic, true)
    l := mongoCon.DB("test_go").C("user")

    // 获取cookie
    cookieUser, err := r.Cookie("username")
    if err == nil {  
        cookieValue, err := r.Cookie("value")
        result := UserData{}
        err = l.Find(bson.M{"username": cookieUser.Value}).One(&result)
        if err != nil || cookieValue.Value != result.Random {
            http.Redirect(w, r, "/login", http.StatusFound)
        }
    }
    err = renderHtml(w, "todo", locals)
    check(err)
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            e := recover()
            if err, ok := e.(error); ok {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                log.Info("WARN: panic in %v. - %v", fn, err)
                log.Info(string(debug.Stack()))
            }
        }()
        fn(w, r)
    }
}

func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int) {
    mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
        file := staticDir + r.URL.Path[len(prefix)-1:]
        if (flags & ListDir) == 0 {
            fi, err := os.Stat(file)
            if err != nil || fi.IsDir() {
                http.NotFound(w, r)
                return
            }
        }
        http.ServeFile(w, r, file)
    })
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    postData, err := url.QueryUnescape(r.FormValue("data"))
    check(err)
    username, err := url.QueryUnescape(r.FormValue("username"))
    check(err)

    var data GetData
    err = json.Unmarshal([]byte(postData), &data)
    check(err)
    //mongodb opration 
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check(err)

    defer mongoCon.Close()

    mongoCon.SetMode(mgo.Monotonic, true)

    // 获取数据库,获取集合
    l := mongoCon.DB("test_go").C("list")

    for _, va := range data.Item {
        va.Username=username
        switch va.Change {
        case "ADD":
            log.Info("add item "+ va.List)
            // 存储数据
            //result := ListData{}
            //err = l.Find(&bson.M{"gid": va.Gid}).One(&result)
            i := mongoCon.DB("test_go").C("index")
            var doc IndexData
            change := mgo.Change{
                Update:    bson.M{"$inc": bson.M{"num": 1}},
                ReturnNew: true,
            }
            _, err = i.Find(bson.M{"index": "gid"}).Apply(change, &doc)
            if err != nil {
                indexUid := IndexData{
                    Index: "gid",
                    Num:   1000,
                }
                err := i.Insert(indexUid)
                check(err)
                doc.Num = 1000
            }
            va.Gid=doc.Num 
            va.Time=time.Now().UnixNano()/10E5
            va.Change="SAVED"
            va.Version=1
            err = l.Insert(&va)
            if err != nil {
               log.Info( "find the same list "+va.List )
            }
        case "MODIFY":
            log.Info("Modify item "+ va.List)
            //  读取数据
            result := ListData{}
            err = l.Find(&bson.M{"gid": va.Gid}).One(&result)
            if err != nil {
                log.Info("can't find ", va.Gid)
                err = l.Insert(&va)
                check(err)
            } else {
                if va.Version > result.Version || va.Time > result.Time {
                    //update mongodb data
                    oldList := bson.M{"gid": va.Gid}
                    va.Time=time.Now().UnixNano()/10E5
                    va.Version=result.Version+1
                    va.Change="SAVED"
                    err := l.Update(oldList, &va)
                    check(err)
                   }
                }
        case "DEL":
            log.Info("delete item "+ va.List)
            result := ListData{}
            err = l.Find(&bson.M{"gid": va.Gid}).One(&result)
            if err == nil {
                err = l.Remove(result)
                check(err)
                log.Info("have deleted it")
            }
        default:
            log.Info("change content is error "+ va.Change)
            return
        }
    }
    
    var outputResult []ListData
    err = l.Find(&bson.M{"username": username}).All(&outputResult)
    output := make(map[string]interface{})
    output["msg"] = "true"
    log.Info( "output data" ,outputResult )
    output["data"]=outputResult
    outputJSON, err := json.Marshal(output)
    check(err)
    w.Write(outputJSON)
}

func main() {

        // Get a new logger instance
        //log := l4g.NewLogger()

        // Create a default logger that is logging messages of FINE or higher
        log.AddFilter("file", log.FINE, log.NewFileLogWriter(filename, false))
        //log.Close()

        /* Can also specify manually via the following: (these are the defaults) */
        flw := log.NewFileLogWriter(filename, false)
        //flw.SetFormat("[%D %T] [%L] (%S) %M")
        //flw.SetRotate(false)
        //flw.SetRotateSize(0)
        //flw.SetRotateLines(0)
        //flw.SetRotateDaily(false)
        log.AddFilter("file", log.FINE, flw)

    mux := http.NewServeMux()
    staticDirHandler(mux, "/public/", "./public", 0)
    mux.HandleFunc("/", safeHandler(indexHandler))
    mux.HandleFunc("/register", safeHandler(registerHandler))
    mux.HandleFunc("/login", safeHandler(loginHandler))
    mux.HandleFunc("/save", safeHandler(saveHandler))

    lis,err := net.Listen( "tcp",":8080" )
    check( err )

    go func(){
        http.Serve( lis,mux )
        //err := http.ListenAndServe(":8080", mux)
        if err != nil {
            log.Critical("ListenAndServe: ", err.Error())
        }
    }()

    ch := make(chan os.Signal)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2,syscall.SIGHUP)
    //#WORKER is a new process tag.
    //newArgs := append(os.Args, "#WORKER")
    attr := syscall.ProcAttr{
        Env: os.Environ(),
    }
    for {
        sig := <-ch
        //log.Info("Signal received:", sig)
        switch sig {
        case syscall.SIGHUP:
            log.Info("get sighup sighup")
        case syscall.SIGINT:
            log.Info("get SIGINT ,exit!")
            os.Exit(1)
        case syscall.SIGUSR1:
            log.Info("usr1")
            //close the net
            lis.Close()
            log.Info( "close connect" )
            if _,_,err:=syscall.StartProcess(os.Args[0],os.Args,&attr);err !=nil{
                check(err)
            }
            //exit current process.
            return
        case syscall.SIGUSR2:
            log.Info("usr2 ")
        }
    }
}
