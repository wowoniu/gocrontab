;window.CRON=function(){
    this.apiRoot="/job/";
    this.api=null;
    this.totalCount=0;
    this.pageSize=10;
    this.totalPage=1;
    this.currentPage=1;
    this.requestLock=false;
    this.pageInited=false;
    this.pageLastKey=0;
    this.labelMap={
        type:{
            1:"shell命令",
            2:"WEB触发"
        }
    }
    this.init()
}

CRON.prototype={
    init:function(){
        this.api=new Api(this.apiRoot)
        this.onAdd()
    },
    renderList:function(){
        var context=this;
        this.api.list(function(data){
            context.totalCount=data['count']
            context.totalPage=Math.ceil(context.totalCount/context.pageSize)
            var records=data['data']
            var html="";
            for (var i=0;i<records.length;i++){
                context.pageLastKey=records[i]['id']
                html+="<tr data-name='"+records[i]["name"]+"'><td>"+records[i]['name']+"</td>"+
                    "<td data-type='"+records[i]['type']+"'>"+context.label('type',records[i]['type'])+"</td>"+
                    "<td>"+records[i]['command']+"</td>"+
                    "<td>"+records[i]['cron_expr']+"</td>"+
                    "<td>"+records[i]['desc']+"</td>"+
                    "<td>"+
                    "<div class='btn-toolbar'>"+
                    "    <button class=\"btn btn-info JS-job-edit \">编辑</button>"+
                    "    <a class=\"btn btn-info \" href='/log.html?name="+records[i]['name']+"' target='_blank'>日志</a>"+
                    "    <button class=\"btn btn-warning JS-job-kill \">强杀</button>"+
                    "    <button class=\"btn btn-danger JS-job-del \">删除</button>"+
                    "   </div>"+
                    "</td></tr>";
            }
            //context.renderPage(context.totalPage,context.currentPage,context.renderList)
            $('.JS-job-list').html(html)
            context.onEdit()
            context.onDelete()
            context.onKill()
        })
    },

    renderLog:function(){
        var context=this;
        var jobName=context.getUrlParam("name")
        this.api.log(jobName,context.currentPage,context.pageSize,function(data){
            context.totalCount=data['count']
            context.totalPage=Math.ceil(context.totalCount/context.pageSize)
            var records=data['data']
            var html="";
            for (var i=0;i<records.length;i++){
                html+="<tr data-name='"+records[i]["JobName"]+"'><td>"+records[i]['JobName']+"</td>"+
                    "<td>"+records[i]['Command']+"</td>"+
                    "<td>"+records[i]['Output']+"</td>"+
                    "<td>"+records[i]['Err']+"</td>"+
                    "<td>"+context.timeFormat(records[i]['PlanTime'],"yyyy-MM-dd hh:mm:ss")+"</td>"+
                    "<td>"+context.timeFormat(records[i]['ScheduleTime'],"yyyy-MM-dd hh:mm:ss")+"</td>"+
                    "<td>"+context.timeFormat(records[i]['StartTime'],"yyyy-MM-dd hh:mm:ss")+"</td>"+
                    "<td>"+context.timeFormat(records[i]['EndTime'],"yyyy-MM-dd hh:mm:ss")+"</td>"+
                    "</tr>";
            }
            $('.JS-job-list').html(html)
            context.renderPage(context.totalPage,context.currentPage,context.renderLog)
            context.onEdit()
            context.onDelete()
            context.onKill()
        })
    },

    renderPage:function(totalPage,currentPage,callback){
        var context=this;
        if(totalPage==0)return
        if(context.pageObj)$('.JS-page-box').jqPaginator("destroy")
        context.pageObj=$('.JS-page-box').jqPaginator({
            totalPages: totalPage,
            visiblePages: 10,
            currentPage: currentPage,

            first: '<li class="first"><a href="javascript:void(0);">首页</a></li>',
            prev: '<li class="prev"><a href="javascript:void(0);">上一页</a></li>',
            next: '<li class="next"><a href="javascript:void(0);">下一页</a></li>',
            last: '<li class="last"><a href="javascript:void(0);">最后一页</a></li>',
            page: '<li class="page"><a href="javascript:void(0);">{{page}}</a></li>',
            onPageChange: function (num) {
                if(callback&&num!=context.currentPage){
                    context.currentPage=num;
                    callback.call(context)
                }
            }
        });
    },


    onAdd:function(){
        var context=this;
        $('.JS-job-add').click(function(){
            $('.JS-job-title').html("新增任务");
            $('#edit-name').removeAttr('readonly');

            $('#edit-modal').find('input').val("")
            $('#edit-modal').find('input').val("")
            $('#edit-modal').find('textarea').val("")
            $('#edit-modal').modal("show")
            $('#JS-edit-save').off('click').click(function () {
                    //保存
                    var job={
                        "type":parseInt($('#edit-type').val()),
                        "name":$('#edit-name').val(),
                        "command":$('#edit-command').val(),
                        "cron_expr":$('#edit-cronexpr').val(),
                        "desc":$('#edit-desc').val()
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
            $('#edit-type').val($(this).parents('tr').find('td').eq(1).data('type'))
            $('#edit-command').val($(this).parents('tr').find('td').eq(2).html())
            $('#edit-cronexpr').val($(this).parents('tr').find('td').eq(3).html())
            $('#edit-desc').val($(this).parents('tr').find('td').eq(4).html())
            $('.JS-job-title').html("编辑任务");
            $('#edit-modal').modal("show")
            $('#JS-edit-save').off('click').click(function () {
                //保存
                var job={
                    "type":parseInt($('#edit-type').val()),
                    "name":$('#edit-name').val(),
                    "command":$('#edit-command').val(),
                    "cron_expr":$('#edit-cronexpr').val(),
                    "desc":$('#edit-desc').val()
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
            if(!window.confirm("您确定要删除吗")){
                return false;
            }
            var jobName=$(this).parents('tr').data('name')
            context.api.delete(jobName,function(response){
                console.log(response)
                context.showTips("删除成功",function () {
                    //window.location.reload()
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
        $('.JS-alert').addClass(level).html(msg).fadeIn(1100,function(){
            if(callback)callback()
            $('.JS-alert').fadeOut(1200).removeClass(level);
        })

    },
    timeFormat:function(timestamp,fmt){
            var dateObj=new Date(timestamp*1000);
            var o = {
                "M+" : dateObj.getMonth()+1,                 //月份
                "d+" : dateObj.getDate(),                    //日
                "h+" : dateObj.getHours(),                   //小时
                "m+" : dateObj.getMinutes(),                 //分
                "s+" : dateObj.getSeconds(),                 //秒
                "q+" : Math.floor((dateObj.getMonth()+3)/3), //季度
                "S"  : dateObj.getMilliseconds()             //毫秒
            };
            if(/(y+)/.test(fmt))
                fmt=fmt.replace(RegExp.$1, (dateObj.getFullYear()+"").substr(4 - RegExp.$1.length));
            for(var k in o)
                if(new RegExp("("+ k +")").test(fmt))
                    fmt = fmt.replace(RegExp.$1, (RegExp.$1.length==1) ? (o[k]) : (("00"+ o[k]).substr((""+ o[k]).length)));
            return fmt;
    },
    getUrlParam:function(name) {
        var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)"); //构造一个含有目标参数的正则表达式对象
        var r = window.location.search.substr(1).match(reg);  //匹配目标参数
        if (r != null) return decodeURIComponent(r[2]);
        return null; //返回参数值
    },
    label:function(name,labelKey){
        if(this.labelMap[name]&&this.labelMap[name][labelKey]){
            return this.labelMap[name][labelKey]
        }
        return "--"
    }

}

var Api=function (apiRoot) {
    this.apiRoot=apiRoot;
}

Api.prototype={
    list:function(successCallback,errCallback){
        this._request("list",{},successCallback,errCallback)
    },
    log:function(jobName,page,pageSize,successCallback,errCallback){
        this._request("log",{"name":jobName,"page":page,"page_size":pageSize},successCallback,errCallback)
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
        if(context.requestLock)return
        context.requestLock=true;
        $.ajax({
            url:context.apiRoot+api,
            data:data,
            type:"post",
            success:function(response){
                context.requestLock=false;
                res=JSON.parse(response)
                if(res.errno==0){
                    if(successCallback)successCallback(res['data'])
                }else{
                    if(errCallback)errCallback(res['msg'])
                }
            },
            error:function(err){
                context.requestLock=true;
            }
        });
    }
}