if(!window.localStorage){
    alert("浏览暂不支持localStorage")
}

$(document).ready(function() { 
    var index,entryid,keyList,changeList;
    var date = new Date();

    last_update = current_ms_time();
    //Initial loading of tasks
    init();
    //if have cookie , try to get data from server.
    var username=getcookie( "username" );
    if (username){
//        alert( username );
        //first send the local data
        //second get the response data
    }else{
        username="";
    }
    display();

    // projects.get();	
    //save now 
    $("a.save").live( 'click',function() {
        store_to_server();
        //store_to_server();
    });

    // Add a task
    $("#tasks-form").submit(function(){
        if ($("#task").val() != "" ) {
            entryid=lGet("mtask-index");
            var data={
                List    : $("#task").val(),
                Time    : current_ms_time(),
                Username :username,
                Project : "work",
                Status  : 0,
                Change  : "ADD",
                Id      :entryid,
            };

            lSet( "mtask-index",++index );
            lSet("mtask-"+entryid,data);
            keyList.push("mtask-"+entryid);
            lSet("mtask-keys",keyList);

            //add change list
            changeFun(changeList,"mtask-"+entryid);
            //output add list
            var li=$( 'li.template' ).clone();
            $(li).attr( {
                id : 'mtask-'+entryid,
                style : "",
            });
            $(li).removeClass( 'template');
            $(li).find('span').text( data.List );
            $(li).find('span').editable({
                editBy:"dblclick",
                type:"textarea",
                editClasss:'note_are',
                onSubmit:function(content){
                    editSave(content,$(this).parent().attr("id"));
                },
            }); 
            $(li).prependTo('#tasks');
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
            editSave(content,$(this).parent().attr("id"));
        }
    }); 

    // Remove a task      
    $("#tasks li a").live("click", function() {
        removeKey=$(this).parent().attr("id");
        var delItem=lGet(removeKey);
        //如果没有gid直接进行删除
        alert( delItem.Gid);
        if (delItem.Gid){
            delItem.Change="DEL";
            changeFun(changeList,removeKey);
            lSet( removeKey,delItem );
        }else{
            changeList.splice(jQuery.inArray(removeKey,changeList),1);
            lSet( "mtask-Change",changeList );
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
        editData.List   = content.current;
        //如果saved过才进行modify设置，并且有gid才进行veriosn加1
        if (editData.Change == "SAVED"){
            editData.Change = "MODIFY";
            if (editData.Gid){
                editData.Version=editData.Version+1;
            }
            changeFun(changeList,id);
        }
        lSet( id,editData );
    } 

    function init(){  
        index = lGet("mtask-index");
        if (!index ){
            lSet("mtask-index",index=0);
        }

        //init status list  ADD DEL MODIFY
        changeList=lGet("mtask-Change");
        if ( !changeList){
            changeList=new Array();
            lSet( "mtask-Change",changeList )
        }

        keyList= lGet( "mtask-keys");
        if (!keyList){
            for( var i=0;i< localStorage.length ;i++ ){
                var key=localStorage.key(i);
                if ( /mtask-\d+/.test(key)){
                    keyList.push(key);
                }
            }
            keyList=new Array();
            lSet( "mtask-keys",keyList);
        }
    }

    function display(){
        keyList=lGet("mtask-keys");
            if( keyList ){  
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
    }

    function store_to_server(){
        $.ajax({  
            url:'save',
        type:'post',  
        //cache:false,  
        dataType:'json', 
        data:{
            'version': 1,
            'username':getcookie("username"),
            'data'   : encodeURIComponent(JSON.stringify({
                          'Item'   : get_list(),
                       })),
        },  
        success:function(data) {
            if(data.msg =="true" ){  
                // view("修改成功！");  
                alert("修改成功！");  
                 restoreData(data.data); 
                 window.location.reload();  
            }else{  
                alert(data.msg);  
            }  
        },
        error : function() {  
            alert("异常！");  
        }  
        });
    }

    //将得到的数据重新进行设置
    function restoreData(data){
        keysList=new Array();
        changeList=new Array();
        lSet( "mtask-keys", keysList);
        if ( data ){  
            for(var i=0;i<data.length;i++){
                data[i].Id=i;
                lSet( "mtask-"+i,data[i]);
                keysList.push( "mtask-"+i);
            }
            lSet("mtask-keys", keysList);
        }
        lSet("mtask-Change",changeList);
    }

    function get_list( tag ){
        var tagList=new Array();
        changeList = lGet( "mtask-Change" );
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


