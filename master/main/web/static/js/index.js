//编辑任务
function newJob(event) {
    $('#edit-name').val("")
    $('#edit-cmd').val("")
    $('#edit-cronExpr').val("")
    $('#edit-name').removeAttr("readonly")
}
function saveJob(event) {
    var jobInfo = {name:$('#edit-name').val(),command:$('#edit-cmd').val(),cronExpr:$('#edit-cronExpr').val()}
    $.ajax({
        url:'/jobs/save',
        dataType:'json',
        method:'post',
        data:{job:JSON.stringify(jobInfo)},
        complete:function (msg) {
            window.location.reload()
        },
    })
}
function editJob(event){
   var jobName = $(this).parents('tr').children('.job-name').text()
   var jobCommand = $(this).parents('tr').children('.job-cmd').text()
   var jobCronExpr = $(this).parents('tr').children('.job-cronExpr').text()
   $('#edit-name').val(jobName)
   $('#edit-cmd').val(jobCommand)
   $('#edit-cronExpr').val(jobCronExpr)

}

//删除任务
function deleteJob(event){
    var jobName = $(this).parents("tr").children(".job-name").text()
    $.ajax({
        url:'/jobs/delete',
        dataType:'json',
        method:'post',
        data:{'name':jobName},
        complete:function () {
            window.location.reload()
        }
    })
}

//杀死任务
function killJob(event){
    var jobName = $(this).parents("tr").children(".job-name").text()
    $.ajax({
        url:'/jobs/kill',
        dataType:'json',
        method:'post',
        data:{'name':jobName},
        complete:function () {
            window.location.reload()
        }
    })
}
$(document).ready(function () {
    //绑定编辑按钮的click事件
    $("#job-list").on("click",".edit-job",editJob)
    $("#job-list").on("click",".delete-job",deleteJob)
    $("#job-list").on("click",".kill-job",killJob)
    $('#saveJobBtn').on("click",saveJob)
    $('#saveJobForm').on("click",newJob)

    function rebuildJobList() {
        $.ajax({
            url:'/jobs/list',
            dataType:'json',
            success:function (msg) {
                if(msg.errno){
                    return
                }

                var jobList = msg.data
                $('#job-list tbody').empty()
                for(var i=0;i<jobList.length;i++){
                    var job = jobList[i]
                    var tr=$("<tr>")
                    tr.append($('<td class="job-name">').html(job.name))
                    tr.append($('<td class="job-cmd">').html(job.command))
                    tr.append($('<td class="job-cronExpr">').html(job.cronExpr))
                    var toobar = $('<div class="btn-toolbar">')
                        .append('<button class="btn btn-info edit-job" data-toggle="modal" data-target="#editJobModal">编辑</button>')
                        .append('<button class="btn btn-danger delete-job">删除</button>')
                        .append('<button class="btn btn-warning kill-job">强杀</button>')
                    tr.append($('<td>').append(toobar))
                    $('#job-list').append(tr)
                }
            }
        })
    }
    rebuildJobList()
})