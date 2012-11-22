if(!window.localStorage){
   alert("浏览暂不支持localStorage")
}

$(document).ready(function() { 
  var i = 0;
  // Initial loading of tasks
  for( i = 0; i < localStorage.length; i++){  
    var str=localStorage.getItem("mtask-"+i);
    var data=JSON.parse(str);
    $("#tasks").append("<li id='mtask-"+ i +"'> <input type='checkbox' /> <span>"+ data.content + "</span> <a href='#'>x</a></li>");
    }
  // Add a task
  $("#tasks-form").submit(function() {
    if (  $("#task").val() != "" ) {
      var data=new Object;
      var d=new Date();

      data.content=$("#task").val();
      data.createTime=d.getTime();
      data.globalId=i;
      var str=JSON.stringify(data);
      localStorage.setItem("mtask-"+i,str);
      var $output="<li id='mtask-"+ i +"'> <input type='checkbox' /><span>"+data.content+"</span><a href='#'>x</a></li>"
      $output.editable( { editBy:"dblclick",type:"textarea",editClasss:'note_are',onSubmit:function(content){editSave(content,$(this).parent().attr("id"))}}); 

      $("#tasks").append($output)
      $("#mtask-" + i).css('display', 'none');
      $("#mtask-" + i).slideDown();
      $("#mtask").val("");
      i++;
    }
    return false;
  });
  
  //Edit a task 
  $("#tasks li span").editable({ editBy:"dblclick",type:"textarea",editClasss:'note_are',onSubmit:function(content){editSave(content,$(this).parent().attr("id"))}}); 

  // Remove a task      
  $("#tasks li a").live("click", function() {
    localStorage.removeItem($(this).parent().attr("id"));
    $(this).parent().slideUp('slow', function() { $(this).remove(); } );

    for(i=0; i<localStorage.length; i++) {
      if( !localStorage.getItem("mtask-"+i)) {
        localStorage.setItem("mtask-"+i, localStorage.getItem('mtask-' + (i+1) ) );
        localStorage.removeItem('mtask-'+ (i+1) );
      }
    }
  });

  function editSave(content,id){
      var str=localStorage.getItem(id);
      var data=JSON.parse(str);
      data.content=content.current;
      var save=JSON.stringify(data);
      localStorage.setItem(id,save);   
  }
}); 