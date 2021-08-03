function loadTimezone() {
    let userTimezone = moment.tz.guess();
    window.location.replace(window.location.href + "?timezone=" + userTimezone);
}


$(document).ready(function () {
    $('.alert .close').click(function () {
        $(this).parent().fadeOut(300, function () {
            $(this).closest();
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
            return;
        }
        let d = Date.parse($(this).attr("datetime"));
        let prefix = "";
        if ($(this).data("prefix")) {
            prefix = $(this).data("prefix");
        }
        if ($(this).hasClass("tooltipped")) {
            $(this).attr("data-tooltip", prefix + moment(d).format('MMM DD, HH:mm'));
        } else {
            $(this).html(prefix + moment(d).format('MMM DD, YYYY'));
        }

    });

    $('#subscribe-email button[type="submit"]').click(function (e) {
        e.preventDefault();
        formData = new FormData(document.getElementById("subscribe-email"));
        path = "/v1/subscribe?email=" + formData.get("email");
        let btn = $(this);
        btn.append($('.preloader-box').html());
        btn.addClass("disabled");
        $.ajax({
            url: path,
            type: "PUT",
            cache: false,
            timeout: 30000,
            error: function (err) {
                btn.removeClass("disabled");
                $('.preload-btn', btn).remove();
                $(btn).before('<span style="color: red;">Code ' + err.responseJSON.status + ' ' + err.responseJSON.description + ': ' + err.responseJSON.detail + '</span>     ');
            },
            success: function (msg) {
                $('#subscribe-email').html("Successfully registered your email");
            }
        });
    });

    $('.scheduled.incident').each(function () {
        if ($('.markdown', this).height() < 100) {
            return;
        }
        $('.markdown', this).data('real-height', $('.markdown', this).height());
        $('.markdown', this).css('max-height', '100px');
        $('.btn', this).show();
        $('.btn', this).data('action', 'show-more');
        $('.fade', this).show();
    });

    $('.show-more-button-wrapper .btn').click(function (e) {
        e.preventDefault();
        const parent = $(this).parent().parent();
        if ($('.btn', parent).data('action') === 'show-more') {
            $('.markdown', parent).css("max-height", $('.markdown', parent).data('real-height') + 'px');
            $('.fade', parent).hide();
            $('.btn', parent).data('action', 'show-less');
            $('.btn', parent).html('<i class="material-icons">arrow_drop_up</i> Show less');
        } else {
            $('.markdown', parent).css("max-height", "100px");
            $('.fade', parent).show();
            $('.btn', parent).data('action', 'show-more');
            $('.btn', parent).html('<i class="material-icons">arrow_drop_down</i> Show more');
        }
    });

});
