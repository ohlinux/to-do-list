package resignal

import (
    "syscall"
    "log"
    "net/http"
    "net"
    "os"
    "os/signal"
)

func StartHttpServe(lis net.Listener ,mux http.Handler){
    // start http serve
     go func(){
        err:=http.Serve( lis,mux )
        if err != nil {
            log.Fatal("ListenAndServe: ", err.Error())
        }
    }()

    //received Signal
    ch := make(chan os.Signal)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2,syscall.SIGHUP)
    //#WORKER is a new process tag.
    //get the current process args ...
    newArgs := append(os.Args, "#WORKER")
    for {
        sig := <-ch
        //log.Println("Signal received:", sig)
        switch sig {
        case syscall.SIGHUP:
            log.Println("get sighup sighup")
        case syscall.SIGINT:
            log.Println("get SIGINT ,exit!")
            os.Exit(1)
        case syscall.SIGUSR1:
            log.Println("get usr1 signal")
            //close the net
            if err:=lis.Close();err!=nil{
                log.Println( "Close ERROR ",err )
            }
            log.Println( "Close connect!" )

            attr := syscall.ProcAttr{
                Env: os.Environ(),
            }
            //start a new same process
            if _,_,err:=syscall.StartProcess(os.Args[0],newArgs,&attr);err !=nil{
                log.Println(err)
            }
            //exit current process.
            return
        case syscall.SIGUSR2:
            log.Println("usr2 ")
        }
    }
}
