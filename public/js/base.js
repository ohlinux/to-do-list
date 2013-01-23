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
        retrieve_from_server();
        $('#welcome').html('Hi '+username);
        $('a.login').replaceWith('<a class="logout" href="#">Logout</a> | <a class="count" href="/myaccount">My Account</a> | <a class="save" href="#">Save Now</a> ');
    }else{
        username="";
    }
    display();

    $('a.logout').live('click',function(){
        delCookie();
    });
    // projects.get();	
    //save now 
    $("a.save").live( 'click',function() {
        store_to_server();
        //store_to_server();
    });

    $('ul.list li').live('mouseover', function(){
            $(this).find('span.todo-remove-icon').css('visibility', 'visible');
    });
    $('ul.list li').live('mouseout', function(){
            $(this).find('span.todo-remove-icon').css('visibility', 'hidden');
    });

    $('input[type=checkbox]').live("change",function(){ 
        var liE=$(this).parent();
        var changeKey=$(liE).attr("id");
        var changeData=lGet(changeKey);
        if(changeData.Change == "SAVED"){
             changeData.Version=changeData.Version+1;
             changeData.Change = 'MODIFY';
        }
        changeData.Time=current_ms_time(); 

        if($(this).attr("checked")==true){ 
            $(liE).addClass('finished');
            $(liE).appendTo('#done');
            changeData.Status=true;
        }else{ 
            $(liE).removeClass('finished');
            $(liE).appendTo('#tasks');
            changeData.Status=false;
        } 
            lSet(changeKey,changeData);
            changeFun(changeList,changeKey);
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
                Status  : false,
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
            $(li).find('span:first').text( data.List );
            $(li).find('span:first').editable({
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
//login action
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
        $('#tasks').empty();
        $('#done').empty();
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
                    $(li).find('span:first').text( data.List );
                    if (data.Status) {
                        $(li).find('input[type=checkbox]').attr("checked","checked");
                        $(li).addClass('finished');
                        $(li).appendTo('#done');
                    }else{
                        $(li).appendTo('#tasks');
                    }
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
                //alert("修改成功！");  
                 restoreData(data.data); 
                 display();
                 //window.location.reload();  
            }else{  
                alert(data.msg);  
            }  
        },
        error : function() {  
            alert("异常！");  
        }  
        });
    }

     function retrieve_from_server(){
        $.ajax({  
            url:'user_list',
            type:'post',  
            dataType:'json', 
            data:{
            'version': 1,
            'username':username,
            },  
            success:function(data) {
                if(data.msg =="true" ){  
                // view("修改成功！");  
                 restoreData(data.data); 
				 ajaxstatus('Data loaded.');
                 display();
                 //window.location.reload();  
                }else{  
                    alert(data.msg);  
            }  
        },
        error : function() {  
		    ajaxstatus('Data load fail.');
            alert("异常！");  
        }  
        });
    }

    //将得到的数据重新进行设置
    function restoreData(data){
        localStorage.clear();
        init();
        if ( data ){  
            for(var i=0;i<data.length;i++){
                data[i].Id=i;
                lSet( "mtask-"+i,data[i]);
                keyList.push( "mtask-"+i);
            }
            lSet("mtask-keys", keyList);
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

    //del all cookie
    function delCookie() {
        //获取 Cookie 字符串  
        var strCookie = document.cookie;  
        //将多 Cookie 切割为多个名、值对  
        var arrCookie = strCookie.split(";");  
        //遍历 Cookie 数据，处理每个 Cookie 对  
        var thisCookie;  
        for (var i=0;i<arrCookie.length;i++ ){  
            //将每个 Cookie 对切割分为名和值  
            thisCookie = arrCookie[i];  
            var arrThisCookie = thisCookie.split("=");  
            //获取每个 Cookie 的变量名  
            var thisCookieName;thisCookieName=arrThisCookie[0];    
            document.cookie = thisCookieName + " =" +";expires=Thu, 01-Jan-1970 00:00:01 GMT";  
        }  
       location.reload(); 
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

    //clear current all data
    function clearAll(){
        localStorage.clear();
        init();
    }
    //delete all to-do list
    function deleteAll(){
    }
}); 


