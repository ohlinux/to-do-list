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
)

const(
    UPLOAD_DIR = "./uploads"
    TEMPLATE_DIR = "./template/"
    ListDir= 0x0001
)

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

func listHandler( w http.ResponseWriter,r *http.Request){
//    fileInfoArr,err := ioutil.ReadDir( "./uploads" )
//    check( err )
    locals := make( map[string]interface{})
//    images:=[]string{}
//    for _,fileInfo := range fileInfoArr{
//        images = append( images,fileInfo.Name() )
//    }
//
//    locals["images"] = images
     err := renderHtml( w,"todo",locals )
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
    data := make( map[string]interface{})
    err =json.Unmarshal([]byte(postData),&data)
    check( err )
    for i , iv := range data {
        log.Println( i )
        switch v2 := iv.(type){
        case []interface{}:
            for j,jv := range v2 {
                log.Println( j,jv )
                switch v3 := jv.(type){
                    case interface{}:
                        d,ok:=v3.(map[string]string)
                        if ok {
                            for x,xv := range d{ 
                              log.Println( x,xv )
                            }
                        }else{
                                log.Println( "not ok" )
                        }
                       // log.Println( v3["List"] )
                       // for x,xv := range v3{
                       //   log.Println( x,xv )
                       // }
                }
            }
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
//    http.Handle("/css/", http.FileServer(http.Dir("public")))
//    http.Handle("/js/", http.FileServer(http.Dir("public")))

    staticDirHandler( mux,"/public/","./public",0 )
    mux.HandleFunc("/",safeHandler(listHandler))
  //  mux.HandleFunc("/view",safeHandler(  viewHandler ))
  //  mux.HandleFunc("/upload",safeHandler( uploadHandler ))
    mux.HandleFunc("/save",safeHandler( saveHandler ))
    err := http.ListenAndServe( ":8080",mux )
    if err != nil {
        log.Fatal( "ListenAndServe: ",err.Error() )
    }
}
