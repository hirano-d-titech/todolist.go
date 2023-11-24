const confirm_delete_user = () => {
    if(window.confirm(`この登録ユーザーは削除され，本IDでログインできなくなります．よろしいですか？`)) {
        location.href = `/user/delete`;
    }
}

const confirm_delete_task = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}

const confirm_update_task = (id) => {
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) {
        location.href = `/task/update/${id}`;
    }
}
