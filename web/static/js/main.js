// Example POST method implementation:
async function postData(url = '', data = {}) {
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });

    return response.json();
}

function generateShortUrl() {
    const longUrl = $('#real-url').val().trim();
    if (longUrl === "") {
        alert("Please enter a url!")
        return
    }

    const shortUrl = $('#short-slug').val().trim();
    const expirationDate = $('#expiration-date').val().trim();
    const expirationTime = $('#expiration-time').val().trim();

    let expires;
    if (expirationDate === "" && expirationTime === "") {
        expires = ""
    } else {
        if (expirationDate === "") {
            alert("Please enter expiration date or remove the expiration time!");
            return;
        } else if (expirationTime === "") {
            alert("Please enter expiration time or remove the expiration date!");
            return;
        } else {
            expires = expirationDate + " " + expirationTime;
        }
    }

    postData('http://localhost:8080/api/create', {
        "real-url": longUrl,
        "short-slug": shortUrl,
        "expires": expires
    })
        .then(data => {
            if (data["error-message"] === "") {
                var $short_url = $('#short-url');
                if (!$short_url.length) {
                    $short_url = $('<div id="short-url"></div>').appendTo('.container');
                }

                $short_url.empty().append(
                    '<div class="input-group">' +
                    '  <input type="url" value=\'' + data["short-url"] + '\' id="copy" class="form-control" readonly>' +
                    '  <div class="input-group-append">' +
                    '    <button class="btn btn-primary copy" data-clipboard-target="#copy" type="button">Copy</button>' +
                    '  </div>' +
                    '</div>'
                )
                new ClipboardJS('.copy');
            } else {
                alert(data["error-message"])
            }
        });
}

$('#expiration-date').datepicker({
    orientation: 'bottom left',
    weekStart: 1,
    daysOfWeekHighlighted: "0,6",
    todayHighlight: true,
    startDate: 'now',
    format: "dd/mm/yyyy",
    autoclose: true
});

$('#expiration-time').clockpicker({
    placement: 'bottom',
    default: 'now',
    autoclose: 'true'
});

$('textarea').each(function () {
    this.setAttribute('style', 'height:' + (this.scrollHeight) + 'px;overflow-y:hidden;');
}).on('input', function () {
    this.style.height = 'auto';
    this.style.height = (this.scrollHeight) + 'px';
});
