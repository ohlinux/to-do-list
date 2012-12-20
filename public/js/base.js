if(!window.localStorage){
    alert("浏览暂不支持localStorage")
}

$(document).ready(function() { 

    last_update = current_ms_time();
    //Initial loading of tasks
    init();
    //if have cookie , try to get data from server.
    if (getcookie( "username" )){
        //list.get();
        alert( "get data from server" );
    }
    display();

    //get data from server
    // projects.get();	
    //save now 
    $("a.save").live( 'click',function() {
        store_to_server();
        //store_to_server();
    });

    // Add a task
    $("#tasks-form").submit(function(){
        if ($("#task").val() != "" ) {
            entryid=index;
            lSet( "mtask-index",++index );
            var data={
                List    : $("#task").val(),
                Time    : current_ms_time(),
                Id      : entryid,
                Project : "work",
                Gid     : unique_id(10),
                Status  : 0,
                Change  : "ADD",
            };
            lSet("mtask-"+data.Id,data);
            keyList.push("mtask-"+data.Id);
            lSet("mtask-keys",keyList);

            //add change list
            changeFun(changeList,"mtask-"+data.Id);
            //output add list
            var li=$( 'li.template' ).clone();
            $(li).attr( {
                id : 'mtask-'+data.Id,
                style : "",
            });
            $(li).removeClass( 'template');
            $(li).find('span').text( data.List );
            $(li).find('span').editable({
                editBy:"dblclick",
                type:"textarea",
                editClasss:'note_are',
                onSubmit:function(content){
                    //var addid = ;
                    // alert( data.Id);
                    editSave(content,$(this).parent().attr("id"));
                },
            }); 
            $(li).prependTo('#tasks');
            //$("#tasks").prepend($output);
            $(li).hide();
            $(li).slideDown();
            $("#mtask").val("");
        };
        return false;
    });

    //Edit a task 
    $("#tasks li span").editable({
        editBy:"dblclick",
        type:"textarea",
        editClasss:'note_are',
        onSubmit:function(content){
            editSave(content,$(this).parent().attr("id"))
        }
    }); 

    // Remove a task      
    $("#tasks li a").live("click", function() {
        removeKey=$(this).parent().attr("id");
        var delItem=lGet(removeKey);
        if (delItem.Gid){
            delItem.Change="DEL"
        changeFun(changeList,removeKey);
    lSet( removeKey,delItem );
        }else{
            localStorage.removeItem(removeKey);
        }
        keyList.splice(jQuery.inArray(removeKey,keyList),1);
        lSet("mtask-keys",keyList);
        $(this).parent().slideUp('slow', function() { $(this).remove(); } );
    });

    //edit the value
    function editSave(content,id){
        var editData=lGet(id);
        editData.Time=current_ms_time();
        editData.Change = "MODIFY";
        editData.List   = content.current;
        lSet( id,editData );
        changeFun(changeList,"mtask-"+data.Id);
    } 

    function init(){  
        var index = lGet("mtask-index");
        if (!index ){
            lSet("mtask-index",index=1);
        }

        //init status list  ADD DEL MODIFY
        changeList=lGet("mtask-Change");
        if ( !changeList){
            var changeList=new Array();
        }

        var keyList= lGet( "mtask-keys");
        if (!keyList){
            var keyList=new Array();
            for( var i=0;i< localStorage.length ;i++ ){
                var key=localStorage.key(i);
                if ( /mtask-\d+/.test(key)){
                    keyList.push(key);
                }
            }
            lSet( "mtask-keys",keyList);
        }
    }

    function display(){
        var keyList=lGet("mtask-keys");
        for( var i = keyList.length-1; i >= 0 ; i--){  
            key = keyList[i];
            var data=lGet(key);
            //clone the html
            var li=$( 'li.template' ).clone();
            $(li).attr( {
                id : 'mtask-'+data.Id,
                style : "",
            });
            $(li).removeClass( 'template');
            $(li).find('span').text( data.List );
            $(li).appendTo('#tasks');
        }
    }

    function store_to_server(){
        $.ajax({  
            url:'save',
        type:'post',  
        //cache:false,  
        dataType:'json', 
        data:{
            'version':1,
        'data'    : encodeURIComponent(JSON.stringify({
            'Item':get_list(),
        })),
        },  
        success:function(data) {
            if(data.msg =="true" ){  
                // view("修改成功！");  
                alert("修改成功！");  
                // window.location.reload();  
            }else{  
                view(data.msg);  
            }  
        },
        error : function() {  
            // view("异常！");  
            alert("异常！");  
        }  
        });
    }

    function get_list( tag ){
        var tagList=new Array();
        var changeList = lGet( "mtask-Change" );
        for(var i=0;i<changeList.length;i++){
            var item=lGet( changeList[i]);
            if ( ! tag ){
                tagList.push(item);
            }else if (item.Change == tag){
                tagList.push(item);
            }
        }
        return tagList;
    }

    function ajaxstatus (msg)
    {
        msg = msg ? msg : '<span class="loading"></span> talking to server';
        $('#ajaxstatus').html(msg);
    }

    function current_ms_time ()
    {
        var date = new Date();
        return date.getTime();
    }

    function lSet (key, value)
    {
        return localStorage.setItem(key, JSON.stringify(value));
    }

    function lGet (key)
    {
        var val = localStorage.getItem(key);
        if (val) {
            val = JSON.parse(val);
            return val;	  
        }else {
            return val;
        }
    }

    function unique_id (len,charset) {
        var i = 0;
        if (! charset) {charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';}
        if (! len) {len = 32;}
        var id = '', charsetlen = charset.length, charIndex;

        // iterate on the length and get a random character for each position
        for (i = 0; len > i; i += 1) {
            charIndex = Math.random() * charsetlen;
            id += charset.charAt(charIndex);
        }
        return id;
    };

    //get cookie
    function getcookie(objname){//获取指定名称的cookie的值
        var arrstr = document.cookie.split("; ");
        for(var i = 0;i < arrstr.length;i ++){
            var temp = arrstr[i].split("=");
            if(temp[0] == objname) return unescape(temp[1]);
        }
    }

    //change list function 
    function changeFun(list,key){
        var pushit=true;
        for(var i=0;i<list.length;i++){
            if (list[i] == key){
                pushit=false; 
            }
        }
        if (pushit){  
            list.push(key);
            lSet( "mtask-Change",list);
        }
    }


}); 


