// table row highlight
document.querySelectorAll('table tr').forEach(e => e.addEventListener('click', function () {
    if (e.classList.contains('highlight'))
        return;
    document.querySelectorAll('table tr').forEach(e => e.classList.remove('highlight'));
    e.classList.add('highlight');
}));

// get selected row
function selectedRow() {
    var row = document.querySelector('tr.highlight');
    return row;
}

// toast (notify)
const {
    Toast
} = bootstrap;

function toast(title, body) {
    var htmlMarkup = `
  <div aria-atomic="true" aria-live="assertive" class="toast bg-primary text-white position-absolute end-0 top-0 m-3" role="alert" id="myAlert">
      <div class="toast-header">
            <strong class="me-auto">` + title + `</strong>
            <small></small>
            <button aria-label="Close" class="btn-close" 
                    data-bs-dismiss="toast" type="button">
            </button>
      </div>
      <div class="toast-body">` + body + ` </div>
  </div>
`;

    var template = document.createElement('template')
    html = htmlMarkup.trim()
    template.innerHTML = html
    return template.content.firstChild
}

function notify(title, body) {
    var toastEl = toast(title, body);
    document.body.appendChild(toastEl)
    const myToast = new Toast(toastEl);
    myToast.show();
}

// method: PUT|DELETE
// url: endpoint
// data: the data send to server
// redirect: where to go after success
function ajax(method, url, data, redirect) {
    var xhr = new XMLHttpRequest();
    xhr.open(method, url, true);
    xhr.send(data);
    xhr.onload = function (e) {
        if ((xhr.status >= 200 && xhr.status < 300) || xhr.status == 304) {
            console.log("OK");
            window.location = redirect;
        } else {
            console.log("What:", e)
        }
    };
    xhr.onerror = function (e) {
        console.log(e);
    };
}

function selected(title, content) {
    var selected_row = selectedRow();
    if (selected_row == null) {
        notify(title, content);
        return null;
    }
    return selected_row;
}

function delete_confirm() {
    if (!window.confirm("确定要删除吗?")) {
        return false;
    }
    return true;
}