;window.CRON=function(){
    this.apiRoot="http://localhost:8080/job/";
    this.api=null;
    this.init()
}

CRON.prototype={
    init:function(){
        this.api=new Api(this.apiRoot)
        this.onAdd()
    },
    renderList:function(){
        var context=this;
        this.api.list(function(records){
            var html="";
            for (var i=0;i<records.length;i++){
                html+="<tr data-name='"+records[i]["name"]+"'><td>"+records[i]['name']+"</td>"+
                    "<td>"+records[i]['command']+"</td>"+
                    "<td>"+records[i]['cron_expr']+"</td>"+
                    "<td>"+
                    "<div class='btn-toolbar'>"+
                    "    <button class=\"btn btn-info JS-job-edit \">编辑</button>"+
                    "    <button class=\"btn btn-warning JS-job-kill \">强杀</button>"+
                    "    <button class=\"btn btn-danger JS-job-del \">删除</button>"+
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
        var context=this;
        $('.JS-job-add').click(function(){
            $('.JS-job-title').html("新增任务");
            $('#edit-modal').modal("show")
            $('#JS-edit-save').off('click').click(function () {
                    //保存
                    var job={
                        "name":$('#edit-name').val(),
                        "command":$('#edit-command').val(),
                        "cron_expr":$('#edit-cronexpr').val()
                    }
                    context.api.save({"job":JSON.stringify(job)},function(data){
                        context.showTips("新增成功",function(){
                            //window.location.reload()
                            $('#edit-modal').modal("hide")
                            setTimeout(function(){
                                window.location.reload()
                            },1500)
                        })
                    },function(errMsg){
                        context.showTips(errMsg)
                    })

                });
        })
    },
    onEdit:function(){
        var context=this;
        $('.JS-job-container').on("click",'.JS-job-edit',function(event){
            $('#edit-name').val($(this).parents('tr').find('td').eq(0).html()).attr('readonly',true)
            $('#edit-old-name').val($(this).parents('tr').find('td').eq(0).html())
            $('#edit-command').val($(this).parents('tr').find('td').eq(1).html())
            $('#edit-cronexpr').val($(this).parents('tr').find('td').eq(2).html())
            $('.JS-job-title').html("编辑任务");
            $('#edit-modal').modal("show")
            $('#JS-edit-save').off('click').click(function () {
                //保存
                var job={
                    "name":$('#edit-name').val(),
                    "command":$('#edit-command').val(),
                    "cron_expr":$('#edit-cronexpr').val()
                }
                context.api.save({"job":JSON.stringify(job)},function(data){
                    context.showTips("编辑成功",function(){
                        //window.location.reload()
                        $('#edit-modal').modal("hide")
                        setTimeout(function(){
                            window.location.reload()
                        },1500)
                    })
                },function(errMsg){
                    context.showTips(errMsg,"","warning")
                })

            });
        })
    },
    onDelete:function(){
        var context=this;
        $('.JS-job-container').on("click",'.JS-job-del',function(event){
            var jobName=$(this).parents('tr').data('name')
            context.api.delete(jobName,function(response){
                context.showTips("删除成功",function () {
                    window.location.reload()
                })
            },function(errMsg){
                context.showTips(errMsg,"","danger")
            })
        })
    },
    onKill:function(){
        var context=this;
        $('.JS-job-container').on("click",'.JS-job-kill',function(event){
            var jobName=$(this).parents('tr').data('name')
            context.api.kill(jobName,function(response){
                context.showTips("强杀成功")
            },function(errMsg){
                context.showTips(errMsg)
            })
        })
    },
    showTips:function(msg,callback,level){
        level=(level===undefined||level===null||level=="")?"alert-success":("alert-"+level);
        $('.JS-alert').addClass(level).html(msg).fadeIn(500,function(){
            if(callback)callback()
            $('.alert-danger').fadeOut(500).hide().removeClass(level);
        })

    }
}

var Api=function (apiRoot) {
    this.apiRoot=apiRoot;
}

Api.prototype={
    list:function(successCallback,errCallback){
        this._request("list",{},successCallback,errCallback)
    },
    save:function(data,successCallback,errCallback){
        this._request("save",data,successCallback,errCallback)
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