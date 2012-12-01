package main 

import (
    "log"
    "net/http"
    "html/template"
    "encoding/json"
    "strings"
    "reflect"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
//    "github.com/ziutek/mymysql/mysql"
//    _ "github.com/ziutek/mymysql/thrsafe"
)

type User struct {
    UserName string
}

type adminController struct {
}

type Result struct{
    Ret int
    Reason string
    Data interface{}
}

type adminUser struct{
    User string
    Password string
}

type ajaxController struct {
}


type loginController struct {
}

type registerController struct {
}

func (this *adminController)IndexAction(w http.ResponseWriter, r *http.Request, user string) {
    t, err := template.ParseFiles("template/admin/index.html")
    if (err != nil) {
        log.Println(err)
    }
    t.Execute(w, &User{user})
}

func (this *ajaxController)LoginAction(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/json")
    err := r.ParseForm()
    if err != nil {
        OutputJson(w, 0, "参数错误", nil)
        return
    }

    admin_name := r.FormValue("admin_name")
    admin_password := r.FormValue("admin_password")

    if admin_name == "" || admin_password == ""{
        OutputJson(w, 0, "参数错误", nil)
        return
    }

//mysql
//    db := mysql.New("tcp", "", "192.168.1.21:3306", "test1", "1234567", "test_login")
//    if err := db.Connect(); err != nil {
//        log.Println(err)
//        OutputJson(w, 0, "数据库操作失败", nil)
//        return
//    }
//    defer db.Close()
//
//    rows, res, err := db.Query("select * from webdemo_admin where admin_name = '%s'", admin_name)
//    if err != nil {
//        log.Println(err)
//        OutputJson(w, 0, "数据库操作失败", nil)
//        return
//    }
//
//    name := res.Map("admin_password")
//    admin_password_db := rows[0].Str(name)
//
//    if admin_password_db != admin_password {
//        OutputJson(w, 0, "密码输入错误", nil)
//        return
//    }

    session, err := mgo.Dial("127.0.0.1:27018")
    if err != nil {
        panic(err)
    }
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)

    // 获取数据库,获取集合
    c := session.DB("test_go").C("user")

//    // 存储数据
//    m1 := adminUser{"user1", "111"}
//    m2 := adminUser{"user2", "222"}
//    err = c.Insert(&m1,&m2)
//    if err != nil {
//        panic(err)
//    }

    // 读取数据
    result := adminUser{}
    err = c.Find(&bson.M{"user":admin_name}).One(&result)
    if err != nil {
        OutputJson(w, 0, "用户名或者密码输入错误", nil)
        return
    }
    // 显示数据
    if admin_password != result.Password {
        OutputJson(w, 0, "密码输入错误", nil)
        return
    }

    // 存入cookie,使用cookie存储
    cookie := http.Cookie{Name: "admin_name", Value: result.Password, Path: "/"}
    http.SetCookie(w, &cookie)

    OutputJson(w, 1, "操作成功", nil)
    return
}

func (this *ajaxController)registerAction(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/json")
    err := r.ParseForm()
    if err != nil {
        OutputJson(w, 0, "参数错误", nil)
        return
    }

    name := r.FormValue("register_name")
    password := r.FormValue("register_password")
//    email := r.FormValue("register_email")
//    sex := r.FormValue( "register_sex" )

    if name == "" || password == ""{
        OutputJson(w, 0, "参数错误", nil)
        return
    }

    session, err := mgo.Dial("127.0.0.1:27018")
    if err != nil {
        panic(err)
    }
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)

    // 获取数据库,获取集合
    c := session.DB("test_go").C("user")

//    // 存储数据
    m1 := adminUser{name, password}
//    m2 := adminUser{"user2", "222"}
    err = c.Insert(&m1)
    if err != nil {
        panic(err)
    }
    OutputJson(w, 1, "注册成功", nil)
    return
}

func (this *loginController)IndexAction(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFiles("template/login/index.html")
    if (err != nil) {
        log.Println(err)
    }
    t.Execute(w, nil)
}

func OutputJson(w http.ResponseWriter, ret int, reason string, i interface{}) {
    out := &Result{ret, reason, i}
    b, err := json.Marshal(out)
    if err != nil {
        return
    }
    w.Write(b)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    // 获取cookie
    cookie, err := r.Cookie("admin_name")
    if err != nil || cookie.Value == ""{
        http.Redirect(w, r, "/login/index", http.StatusFound)
    }

    pathInfo := strings.Trim(r.URL.Path, "/")
    parts := strings.Split(pathInfo, "/")
    var action = ""
    if len(parts) > 1 {
        action = strings.Title(parts[1]) + "Action"
    }

    admin := &adminController{}
    controller := reflect.ValueOf(admin)
    method := controller.MethodByName(action)
    if !method.IsValid() {
        method = controller.MethodByName(strings.Title("index") + "Action")
    }
    requestValue := reflect.ValueOf(r)
    responseValue := reflect.ValueOf(w)
    userValue := reflect.ValueOf(cookie.Value)
    method.Call([]reflect.Value{responseValue, requestValue, userValue})
}

func ajaxHandler(w http.ResponseWriter, r *http.Request) {
    pathInfo := strings.Trim(r.URL.Path, "/")
    parts := strings.Split(pathInfo, "/")
    var action = ""
    if len(parts) > 1 {
        action = strings.Title(parts[1]) + "Action"
    }

    ajax := &ajaxController{}
    controller := reflect.ValueOf(ajax)
    method := controller.MethodByName(action)
    if !method.IsValid() {
        method = controller.MethodByName(strings.Title("index") + "Action")
    }
    requestValue := reflect.ValueOf(r)
    responseValue := reflect.ValueOf(w)
    method.Call([]reflect.Value{responseValue, requestValue})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("loginHandler")
    pathInfo := strings.Trim(r.URL.Path, "/")
    parts := strings.Split(pathInfo, "/")
    var action = ""
    if len(parts) > 1 {
        action = strings.Title(parts[1]) + "Action"
    }

    login := &loginController{}
    controller := reflect.ValueOf(login)
    method := controller.MethodByName(action)
    if !method.IsValid() {
        method = controller.MethodByName(strings.Title("index") + "Action")
    }
    requestValue := reflect.ValueOf(r)
    responseValue := reflect.ValueOf(w)
    method.Call([]reflect.Value{responseValue, requestValue})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("registerHandler")
    pathInfo := strings.Trim(r.URL.Path, "/")
    parts := strings.Split(pathInfo, "/")
    log.Println( parts)
    log.Println( parts[1])
    var action = ""
    if len(parts) > 1 {
        action = strings.Title(parts[1]) + "Action"
    }
    log.Println( "action"+action)
    register := &registerController{}
    controller := reflect.ValueOf(register)
    method := controller.MethodByName(action)
    if !method.IsValid() {
        method = controller.MethodByName(strings.Title("index") + "Action")
    }
    requestValue := reflect.ValueOf(r)
    responseValue := reflect.ValueOf(w)
    method.Call([]reflect.Value{responseValue, requestValue})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.Redirect(w, r, "/login/index", http.StatusFound)
    }

    t, err := template.ParseFiles("template/404.html")
    if (err != nil) {
        log.Println(err)
    }
    t.Execute(w, nil)
}

func main() {
    log.Println("main")
    http.Handle("/css/", http.FileServer(http.Dir("public")))
    http.Handle("/js/", http.FileServer(http.Dir("public")))

    http.HandleFunc("/admin/", adminHandler)
    http.HandleFunc("/login/",loginHandler)
    http.HandleFunc("/register/",registerHandler)
    http.HandleFunc("/ajax/",ajaxHandler)
    http.HandleFunc("/",NotFoundHandler)
    http.ListenAndServe(":8888", nil)
}

