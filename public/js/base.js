if(!window.localStorage){
   alert("浏览暂不支持localStorage")
}

$(document).ready(function() { 
    var index,i,key,changeList;

  // Initial loading of tasks
    index = localStorage.getItem("mtask:index");
    if ( ! index ){
     localStorage.setItem("mtask:index",index=1);
    }
   var keyRange=new Array();
   keyRange = JSON.parse(localStorage.getItem( "mtask:keys" ));
    if ( !keyRange){
        var keyRange=new Array();
        for( i=0;i< localStorage.length ;i++ ){
            key=localStorage.key(i);
           if ( /mtask:\d+/.test(key)){
            keyRange.push(key);
           }
        }
        localStorage.setItem( "mtask:keys",JSON.stringify(keyRange));
    }
  //Initial tasks
    for( i = keyRange.length-1; i >= 0 ; i--){  
            key = keyRange[i];
            var str=localStorage.getItem(key);
            var data=JSON.parse(str);
            $("#tasks").append("<li id='mtask:"+ data.id +"'> <input type='checkbox' /> <span>"+ data.content + "</span> <a data-id="+data.id+" href='#'>x</a></li>");
    }
    
   //change status list  ADD DEL MODIFY
    ADDList=JSON.parse(localStorage.getItem("mtask:ADD"));
    DELList=JSON.parse(localStorage.getItem("mtask:DEL"));
    MODIFYList=JSON.parse(localStorage.getItem("mtask:MODIFY"));
    if ( !ADDList ){
        var ADDList=new Array();
    }
    if ( !DELList){
        var DELList=new Array();
    }
    if (!MODIFYList){
        var MODIFYList=new Array();
    }

  // Add a task
  $("#tasks-form").submit(function() {
    if (  $("#task").val() != "" ) {
      var data=new Object;
      var d=new Date();

      entryid=index;
      localStorage.setItem( "mtask:index",++index );
      data.content=$("#task").val();
      data.Time=d.getTime();
      data.id=entryid;
      var str=JSON.stringify(data);
      localStorage.setItem("mtask:"+data.id,str);
      keyRange.push("mtask:"+data.id);
      localStorage.setItem("mtask:keys",JSON.stringify(keyRange));
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
    
    var delItem=JSON.parse(localStorage.getItem(removeKey));
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
    localStorage.setItem("mtask:keys",JSON.stringify(keyRange));
    $(this).parent().slideUp('slow', function() { $(this).remove(); } );
  });

  //edit the value
  function editSave(content,id){
      var str=localStorage.getItem(id);
      var data=JSON.parse(str);
      var d=new Date();
      data.Time=d.getTime();
      data.content=content.current;
      var save=JSON.stringify(data);
      localStorage.setItem(id,save);   
  }

 function changeList(list,key,tag ){
     list.push(key);
     localStorage.setItem( "mtask:"+tag,JSON.stringify(list));
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
