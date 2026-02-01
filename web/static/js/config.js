let initialConfig = {};

function getCurrentConfig() {
    return {
        language: $('#cfg-language').val(),
        first_day_of_week: Number($('#cfg-first_day_of_week').val()),
        location: $('#cfg-location').val().trim(),
        show_lunar: $('#cfg-show_lunar').is(':checked'),
        show_bill: $('#cfg-show_bill').is(':checked'),
        show_interest: $('#cfg-show_interest').is(':checked'),
        show_note: $('#cfg-show_note').is(':checked'),
        show_sport: $('#cfg-show_sport').is(':checked'),
        ui_default: $('#cfg-ui_default').val(),
    };
}


function collectSettings(section='global') {
    const current = getCurrentConfig();
    const items = [];
    Object.keys(current).forEach(key => {
        if (current[key] !== initialConfig[key]) {
            items.push({
                section: section,
                key,
                value: serializeValue(current[key])
            });
        }
    });
    return items;
}

$('#settingsForm').on('submit', function (e) {
    e.preventDefault();
    items = collectSettings()
    if (!items.length) return;
    API.post('/api/config/batch', { items }, () => {
        location.reload();
    });
});


$(function () {
    const cfg = window.APP_CONFIG || {};

    $('#cfg-language').val(cfg.language);
    $('#cfg-first_day_of_week').val(cfg.first_day_of_week);
    $('#cfg-location').val(cfg.location || '');
    $('#cfg-ui_default').val(cfg.ui_default || 'diary');
    $('#cfg-show_lunar').prop('checked', cfg.show_lunar);
    $('#cfg-show_bill').prop('checked', cfg.show_bill);
    $('#cfg-show_interest').prop('checked', cfg.show_interest);
    $('#cfg-show_note').prop('checked', cfg.show_note);
    $('#cfg-show_sport').prop('checked', cfg.show_sport);

    initialConfig = getCurrentConfig();
});

$('#passwordForm').on('submit', async function (e) {
    e.preventDefault();

    const oldPwd = $('#pwd-old').val();
    const newPwd = $('#pwd-new').val();
    const confirmPwd = $('#pwd-confirm').val();

    if (newPwd.length > 16) {
        showError(I18N['Password too long (>16)']);
        return;
    }
    if (newPwd !== confirmPwd) {
        showError(I18N['New passwords do not match']);
        return;
    }
    if (newPwd === oldPwd) {
        ShowError(I18N['Same new and old password']);
        return
    }
    if (newPwd === '') {
        const ok = await showConfirm(
            I18N['New password is empty. Continue?'],
            'warning'
        );
        if (!ok) return;
    }
    API.post('/api/config/change_password', {
        old_password: oldPwd,
        new_password: newPwd
    }, () => {
        showSuccess(I18N['Password updated']);

        $('#pwd-old').val('');
        $('#pwd-new').val('');
        $('#pwd-confirm').val('');
        setTimeout(() => {
            window.location.href = '/login.html';
        }, 800);
    });
});

applyNavConfig();
