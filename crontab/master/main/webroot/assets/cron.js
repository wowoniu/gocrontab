;window.CRON=function(){
    this.api=null
    this.init()
}

CRON.prototype={
    init:function(){
        this.api=new Api()
        this.onAdd()
    },
    renderList:function(){
        var context=this;
        this.api.list(function(records){
            var html="";
            for (var i=0;i<records.length;i++){
                html+="<tr data-name='"+records[i]["name"]+"'><td>"+records[i]['name']+"</td>"+
                    "<td>echo 1</td>"+
                    "<td>* * * * * * *</td>"+
                    "<td>"+
                    "<div class='btn-toolbar'>"+
                    "    <button class=\"btn btn-info JS-job-edit \">编辑</button>"+
                    "    <button class=\"btn btn-danger JS-job-del \">删除</button>"+
                    "    <button class=\"btn btn-warning JS-job-kill \">强杀</button>"+
                    "   </div>"+
                    "    </td></tr>";
            }
            $('.JS-job-list').html(html)
            context.onEdit()
            context.onDelete()
            context.onKill()
        })
    },

    onAdd:function(){
        $('.JS-job-add').click(function(event){
            console.log('add')
        })
    },
    onEdit:function(){
        $('.JS-job-container').on("click",'.JS-job-edit',function(event){
            console.log('edit')
        })
    },
    onDelete:function(){
        var context=this;
        $('.JS-job-container').on("click",'.JS-job-del',function(event){
            var jobName=$(this).parents('tr').data('name')
            context.api.delete(jobName,function(response){
                alert("删除成功")
                window.reload()
            },function(errMsg){
                alert(errMsg)
            })
        })
    },
    onKill:function(){
        var context=this;
        $('.JS-job-container').on("click",'.JS-job-kill',function(event){
            var jobName=$(this).parents('tr').data('name')
            context.api.kill(jobName,function(response){
                alert("强杀成功")
            },function(errMsg){
                alert(errMsg)
            })
        })
    }


}

var Api=function () {
    this.apiRoot="http://localhost:8080/job/";
}

Api.prototype={
    list:function(successCallback,errCallback){
        this._request("list",{},successCallback,errCallback)
    },
    delete:function(jobName,successCallback,errCallback){
        this._request("delete",{"name":jobName},successCallback,errCallback)
    },
    kill:function(jobName,successCallback,errCallback){
        this._request("kill",{"name":jobName},successCallback,errCallback)
    },

    _request:function(api,data,successCallback,errCallback){
        var context=this;
        $.ajax({
            url:context.apiRoot+api,
            data:data,
            type:"post",
            success:function(response){
                res=JSON.parse(response)
                if(res.errno==0){
                    if(successCallback)successCallback(res['data'])
                }else{
                    if(errCallback)errCallback(res['msg'])
                }
            }
        });
    }


}