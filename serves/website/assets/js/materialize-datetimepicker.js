/*!
 * Fawad Tariq (http://github.com/fawadtariq)
 * Materialize Date Time Picker v0.1.1-beta
 * Based on Materialize (http://materializecss.com)
 */
var MaterialDateTimePicker = {
    control: null,
    dateRange: null,
    pickerOptions: null,
    create: function (element) {
        this.control = element == undefined ? $('#' + localStorage.getItem('element')) : element;
        element = this.control;
        if (this.control.is("input[type='text']")) {
            var defaultDate = new Date();
            element.off('click');
            element.datepicker({
                format: 'yyyy-mm-dd',
                selectMonths: true,
                dismissable: false,
                autoClose: true,
                onClose: function () {
                    element.datepicker('destroy');
                    element.timepicker({
                        dismissable: false,
                        twelveHour: false,
                        onSelect: function (hr, min) {
                            element.attr('selectedTime', (hr + ":" + min + ":00").toString());
                        },
                        onCloseEnd: function () {
                            element.blur();
                        }
                    });
                    $('button.btn-flat.timepicker-close.waves-effect')[0].remove();

                    if (element.val() != "") {
                        element.attr('selectedDate', element.val().toString());
                    } else {
                        element.val(defaultDate.getFullYear().toString() + "-" + (defaultDate.getMonth() + 1).toString() + "-" + defaultDate.getDate().toString())
                        element.attr('selectedDate', element.val().toString());
                    }
                    element.timepicker('open');
                }
            });
            element.unbind('change');
            element.off('change');
            element.on('change', function () {
                if (element.val().indexOf(':') > -1) {
                    element.attr('selectedTime', element.val().toString());
                    element.val(element.attr('selectedDate') + "T" + element.attr('selectedTime'));
                    element.timepicker('destroy');
                    element.unbind('click');
                    element.off('click');
                    element.on('click', function (e) {
                        element.val("");
                        element.removeAttr("selectedDate");
                        element.removeAttr("selectedTime");
                        localStorage.setItem('element', element.attr('id'));
                        MaterialDateTimePicker.create.call(element);
                        element.trigger('click');
                    });
                }
            });
            $('button.btn-flat.datepicker-cancel.waves-effect, button.btn-flat.datepicker-done.waves-effect').remove();
            this.addCSSRules();
            return element;
        } else {
            console.error("The HTML Control provided is not a valid Input Text type.")
        }
    },
    addCSSRules: function () {
        $('html > head').append($('<style>div.modal-overlay { pointer-events:none; }</style>'));
    },
}