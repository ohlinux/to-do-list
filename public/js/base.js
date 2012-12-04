if(!window.localStorage){
   alert("浏览暂不支持localStorage")
}
//var status_count = 2;
//var column_count = 1;
//var last_update = 0;
//var update_every = 8000; // never update more than every x milliseconds


$(document).ready(function() { 
  
    last_update = current_ms_time();
	

  // Initial loading of tasks
  //  project.init();
    var index = lGet("mtask:index");
    if (!index ){
     lSet("mtask:index",index=1);
    }
    
  //init status list  ADD DEL MODIFY
    ADDList=lGet("mtask:ADD");
    DELList=lGet("mtask:DEL");
    MODIFYList=lGet("mtask:MODIFY");
    if ( !ADDList ){
        var ADDList=new Array();
    }
    if ( !DELList){
        var DELList=new Array();
    }
    if (!MODIFYList){
        var MODIFYList=new Array();
    }
    
    var keyRange = lGet( "mtask:keys");
    if (!keyRange){
        var keyRange=new Array();
        for( var i=0;i< localStorage.length ;i++ ){
            var key=localStorage.key(i);
           if ( /mtask:\d+/.test(key)){
            keyRange.push(key);
           }
        }
        lSet( "mtask:keys",keyRange);
    }
    
   
   //Initial tasks
    for( i = keyRange.length-1; i >= 0 ; i--){  
            key = keyRange[i];
            var data=lGet(key);
            $("#tasks").append("<li id='mtask:"+ data.id +"'> <input type='checkbox' /> <span>"+ data.content + "</span> <a data-id="+data.id+" href='#'>x</a></li>");
    }
    
   
    //get data from server
 //   projects.get();	
  
  
  // Add a task
  $("#tasks-form").submit(function() {
    if (  $("#task").val() != "" ) {
      var d=new Date();
      entryid=index;
      lSet( "mtask:index",++index );
      var data={
	      content : $("#task").val(),
              Time : d.getTime(),
              id :entryid,
      }
      lSet("mtask:"+data.id,data);
      keyRange.push("mtask:"+data.id);
      lSet("mtask:keys",keyRange);
      
      //add change list
      changeList(ADDList,"mtask:"+data.id,"ADD");

      var $output=$("<li id='mtask:"+ data.id +"'> <input type='checkbox' /><span>"+data.content+"</span><a data-id="+data.id+" href='#'>x</a></li>");
      $output.editable({editBy:"dblclick",type:"textarea",editClasss:'note_are',onSubmit:function(content){editSave(content,$(this).parent().attr("id"))}}); 
      $("#tasks").prepend($output);
      $("#mtask:" + data.id).css('display', 'none');
      $("#mtask:" + data.id).slideDown();
      $("#mtask").val("");
    }
    return false;
  });
  
  //Edit a task 
  $("#tasks li span").editable({
      editBy:"dblclick",type:"textarea",editClasss:'note_are',onSubmit:function(content){editSave(content,$(this).parent().attr("id"))}
  }); 

  // Remove a task      
  $("#tasks li a").live("click", function() {
    removeKey=$(this).parent().attr("id");
    
    var delItem=lGet(removeKey);
    if (delItem.gid){
       changeList(DELList,delItem.gid,"DEL");
    }
    localStorage.removeItem(removeKey);
//    var newkeyRange=new Array();
//    for ( i=0;i<keyRange.length;i++ ){
//        if (keyRange[i] != removeKey){
//            newkeyRange.push( keyRange[i]);
//        }
//    }
    keyRange.splice(jQuery.inArray(removeKey,keyRange),1);
    lSet("mtask:keys",keyRange);
    $(this).parent().slideUp('slow', function() { $(this).remove(); } );
  });

  //edit the value
  function editSave(content,id){
      var data=lGet(id);
      var d=new Date();
      data={
	   Time : d.getTime(),
           content :content.current,
      }
      lSet(id,data);   
  }

 function changeList(list,key,tag ){
     list.push(key);
     lSet( "mtask:"+tag,list);
}


//var timeOutID = 0;
//$.ajax({
//  url: 'save/todo.html',
//  success: function(data) {     
//
//    clearTimeOut(timeOutID);
//    // Remove the abort button if it exists.
//}
//});
//timeOutID = setTimeout(function() {
//                 // Add the abort button here.
//               }, 5000); 
}); 

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

