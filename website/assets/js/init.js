function loadTimezone() {
    let userTimezone = moment.tz.guess();
    window.location.replace(window.location.href + "?timezone=" + userTimezone);
}


$(document).ready(function () {
    $('.alert .close').click(function () {
        $(this).parent().fadeOut(300, function () {
            $(this).closest()
            $(this).hide();
        });
    });

    $('select').formSelect();
    $('.tooltipped').tooltip();
    $('.tooltipped-btn').tooltip({
        enterDelay: 1000,
    });
    $('.dropdown-trigger').dropdown();
    // $('.modal-tabs').tabs();
    $('#subscribe').modal({
        onOpenEnd: function () {
            $('.modal-tabs').tabs();
        }
    });

    $('.modal').modal({
        onOpenEnd: function () {
            $('.modal-tabs').tabs();
        }
    });

    $('#preview-incident').modal({
        onOpenStart: function () {
            $.ajax({
                url: '/v1/markdown/preview',
                type: 'POST',
                async: false,
                cache: false,
                timeout: 30000,
                data: $('form input[name="title"]').val(),
                error: function () {
                    return true;
                },
                success: function (msg) {
                    $('#preview-incident .title').html(msg);
                }
            });
            $.ajax({
                url: '/v1/markdown/preview',
                type: 'POST',
                async: false,
                cache: false,
                timeout: 30000,
                data: $('form textarea[name="content"]').val(),
                error: function () {
                    return true;
                },
                success: function (msg) {
                    $('#preview-incident .content').html(msg);
                }
            });
        },
    });
    $('.datepicker').datepicker({
        format: 'yyyy-mm-dd',
        selectMonths: true,
        autoClose: true,
    });
    $('.timepicker').timepicker({
        twelveHour: false,
        autoClose: true,
    });

    $("time").each(function (index) {
        if (!($(this).attr("datetime"))) {
            return
        }
        let d = Date.parse($(this).attr("datetime"));
        let prefix = ""
        if ($(this).data("prefix")) {
            prefix = $(this).data("prefix");
        }
        if ($(this).hasClass("tooltipped")) {
            $(this).attr("data-tooltip", prefix + moment(d).format('MMM DD, HH:mm'));
        } else {
            $(this).html(prefix + moment(d).format('MMM DD, YYYY'));
        }

    });

});