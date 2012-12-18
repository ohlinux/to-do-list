package main

import (
    "io"
    "os"
    "log"
    "net/http"
    "io/ioutil"
    "html/template"
    "path"
    "strings"
    "runtime/debug"
    "encoding/json"
    "net/url"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "math/rand"
    "crypto/sha1"
    "time"
    "strconv"
    "fmt"
)

const(
    UPLOAD_DIR = "./uploads"
    TEMPLATE_DIR = "./template/"
    ListDir= 0x0001
)

type GetData struct {
   Item []*ListData 
}

type IndexData struct {
    Index  string
    Num    int64
}

type UserData struct{
    Uid  int64
    Username  string
    Password  string
    Email    string
    Random    string
}

type ListData struct{
    Uid string
    Gid string
    Project string
    Status int
    Change string
    Time int64
    List string
}

type mongoData struct{
    Gid  string
    Project string
    Time    int64
    Status  int
    List    string
}

var templates = make(  map[string]*template.Template)

func init(){
    fileInfoArr,err := ioutil.ReadDir( TEMPLATE_DIR )
    if err != nil {
        panic( err )
        return
    }
    var templateName,templatePath string 
    for _,fileInfo := range fileInfoArr{
        templateName = fileInfo.Name()
        if ext := path.Ext( templateName);ext != ".html" {
            continue
        }
        templatePath= TEMPLATE_DIR + "/" +templateName
        log.Println( "Loading template:",templatePath )
        t := template.Must( template.ParseFiles( templatePath))
        tmpl := strings.Split( templateName,".html" )[0]
        templates[tmpl] = t
    }
}

func check( err error ){
    if err!=nil{
        panic( err )
    }
}

func renderHtml( w http.ResponseWriter,tmpl string,locals map[string]interface{}) ( err error) {
        err = templates[tmpl].Execute( w,locals )
        check( err )
        return
}

func isExists( path string ) bool {
    _,err := os.Stat( path )
    if err == nil {
        return true
    }
    return os.IsExist( err )
}

func registerHandler( w http.ResponseWriter,r *http.Request ){
    //uid := 1
    if r.Method == "GET"{  
     locals := make( map[string]interface{})
     err := renderHtml( w,"register",locals )
     check( err )
     return
    }

    if r.Method == "POST"{
//mongodb opration 
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check( err )
    defer mongoCon.Close()
    mongoCon.SetMode(mgo.Monotonic, true)

   
// 获取数据库,获取集合
   l := mongoCon.DB("test_go").C("user")
   i := mongoCon.DB( "test_go").C( "index" )
   var doc IndexData
   change := mgo.Change{
       Update: bson.M{"$inc": bson.M{"num": 1}},
       ReturnNew: true,
   }
   _, err = i.Find(bson.M{"index":"uid"}).Apply(change, &doc)
   if err !=nil {
       indexUid := IndexData {
           Index : "uid",
           Num   :  1,
       }
       err := i.Insert( indexUid )
       check (err)
       doc.Num = 1
   }
   log.Println(doc.Num)
   users := UserData {
       Uid : doc.Num,
       Username : r.FormValue( "username" ),
       Password : r.FormValue( "password" ),
       Email    : r.FormValue( "email" ),
   }
   err = l.Insert( users )
   check( err )
   http.Redirect(w, r,"login", http.StatusFound)
   return
 }

}

func loginHandler( w http.ResponseWriter,r *http.Request ){
     if r.Method == "GET"{  
     locals := make( map[string]interface{})
     err := renderHtml( w,"login",locals )
     check( err )
     return
    }

    if r.Method == "POST"{
//mongodb opration 
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check( err )
    defer mongoCon.Close()
    mongoCon.SetMode(mgo.Monotonic, true)
    l := mongoCon.DB("test_go").C("user")
    user := r.FormValue( "username" )
    passwd := r.FormValue( "password" )
    result := UserData{}
    err = l.Find(&bson.M{"username":user}).One(&result)
    if err != nil {
       log.Println( "not found the username ",user) 
       panic( err )
    }
    
    if passwd  == result.Password{
      //set cookie
      //// create random string
      randNum := rand.New(rand.NewSource(time.Now().UnixNano()))
      h := sha1.New()
      randString:=result.Password+ strconv.Itoa(randNum.Intn(100))
      log.Println( "rand string ",randString )
      io.WriteString(h,randString)
      hashString :=fmt.Sprintf("%x",h.Sum(nil))
      log.Println(hashString)
      err = l.Update(bson.M{"username":user },bson.M{"$set": bson.M{ "random" :hashString}})
      check( err )
      cookie1 := http.Cookie{Name: "value", Value: hashString}
      cookie2 := http.Cookie{Name: "username", Value: user}
      http.SetCookie(w, &cookie1)
      http.SetCookie(w, &cookie2)
      log.Println("set cookie")
      http.Redirect(w, r,"/", http.StatusFound)
    }else{
        log.Println( "password is not right",passwd )
    }
    return
    }
}

func uploadHandler( w http.ResponseWriter,r *http.Request ){
    if r.Method == "GET" {
        err := renderHtml( w,"upload",nil )
        check( err )
//        io.WriteString( w,"<html><body><form method=\"POST\" action=\"/upload\" enctype=\"multipart/form-data\">"+
//        "Choose an image to upload:<input name=\"image\" type=\"file\" />"+
//        "<input type=\"submit\" value=\"Upload\" />"+
//        "</form></body></html>")
        return
    }

    if r.Method == "POST" {
        f,h,err := r.FormFile( "image" )
        check( err )
        filename := h.Filename
        defer f.Close()
        t,err := os.Create( UPLOAD_DIR + "/" + filename  )
        check( err )
        defer t.Close()
        _,err = io.Copy( t,f )
        check( err ) 
        http.Redirect( w,r,"/view?id="+filename,
        http.StatusFound)
    }
}

func viewHandler( w http.ResponseWriter,r *http.Request ){
    imageId := r.FormValue( "id" )
    imagePath := UPLOAD_DIR + "/" +imageId
    if exists := isExists( imagePath );!exists{
        http.NotFound( w,r )
        return
    }
    w.Header().Set( "Content-Type","image/png" )
    http.ServeFile( w,r,imagePath )
}

func indexHandler( w http.ResponseWriter,r *http.Request){
//    fileInfoArr,err := ioutil.ReadDir( "./uploads" )
//    check( err )
    locals := make( map[string]interface{})
//    images:=[]string{}
//    for _,fileInfo := range fileInfoArr{
//        images = append( images,fileInfo.Name() )
//    }
//
//    locals["images"] = images
    //
//get cookie random string 
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check( err )
    defer mongoCon.Close()
    mongoCon.SetMode(mgo.Monotonic, true)
    l := mongoCon.DB("test_go").C("user")

    // 获取cookie
    cookieUser,err:= r.Cookie("username")
    cookieValue,err:=r.Cookie( "value" )
    result := UserData{}
    err = l.Find(bson.M{"username":cookieUser.Value}).One(&result)
    if err != nil || cookieValue.Value != result.Random{
        http.Redirect(w, r, "/login", http.StatusFound)
    }
     err = renderHtml( w,"todo",locals )
     check( err )
}

func safeHandler( fn http.HandlerFunc ) http.HandlerFunc{
    return func( w http.ResponseWriter,r *http.Request ){
        defer func(){
            e := recover()
            if err,ok:= e.(error);ok{
                http.Error( w,err.Error(),http.StatusInternalServerError )
                log.Println( "WARN: panic in %v. - %v",fn,err )
                log.Println( string( debug.Stack() ) )
            }
        }()
        fn( w,r )
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

func saveHandler( w http.ResponseWriter,r *http.Request ){
    postData,err:=url.QueryUnescape(r.FormValue("data"))
    check( err )
    var data GetData 
    err =json.Unmarshal([]byte(postData),&data)
    check( err )
//mongodb opration 
    mongoCon, err := mgo.Dial("127.0.0.1:27018")
    check( err )

    defer mongoCon.Close()

    mongoCon.SetMode(mgo.Monotonic, true)

// 获取数据库,获取集合
   l := mongoCon.DB("test_go").C("list")

    for _,va := range data.Item{

        switch va.Change {
            case "ADD":
                log.Println( "add item ",va.List )
                // 存储数据
                    result := ListData{}
                    err = l.Find(&bson.M{"gid":va.Gid}).One(&result)
                    if err != nil {
                        err = l.Insert(&va)
                //      check( err )
                    }
            case "MODIFY":
                log.Println( "Modify item ",va.List )
                //  读取数据
                    result := ListData{}
                    err = l.Find(&bson.M{"gid":va.Gid}).One(&result)
                    if err != nil {
                        log.Println( "can't find",va.Gid )
                        err = l.Insert(&va)
                        check( err )
                    }else{
                       if va.Time > result.Time {
                //update mongodb data
                           oldList := bson.M{ "gid":va.Gid }
                           err := l.Update(oldList,&va)
                           check( err )
                       }
                    }
            case "DEL":
                log.Println( "delete item ",va.List )
                result := ListData{}
                err = l.Find(&bson.M{"gid":va.Gid}).One(&result)
                if err == nil {
                    log.Println( "delete it" )
                    err = l.Remove( result )
                    check( err )
                }
           default :
            log.Println( "change content is error ",va.Change )
            return
        }
    }

    output:=make( map[string]interface{})
    output[ "msg" ]="true"
    outputJSON,err := json.Marshal(output)
    check( err )
    w.Write(outputJSON )
}


func main(){
    mux := http.NewServeMux()

    staticDirHandler( mux,"/public/","./public",0 )
    mux.HandleFunc("/",safeHandler(indexHandler))
    mux.HandleFunc("/register",safeHandler(registerHandler))
    mux.HandleFunc("/login",safeHandler(loginHandler))
  //  mux.HandleFunc("/view",safeHandler(  viewHandler ))
  //  mux.HandleFunc("/upload",safeHandler( uploadHandler ))
    mux.HandleFunc("/save",safeHandler( saveHandler ))
    err := http.ListenAndServe( ":8080",mux )
    if err != nil {
        log.Fatal( "ListenAndServe: ",err.Error() )
    }
}
