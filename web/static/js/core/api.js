window.API = {
    get: function (url, success) {
        $.ajax({
            url,
            method: 'GET',
            success,
            error: API._handleError
        });
    },
    post: function (url, data, success) {
        const ajaxOptions = {
            url,
            method: 'POST',
            success,
            error: API._handleError
        };

        if (data !== null && data !== undefined) {
            ajaxOptions.contentType = 'application/json';
            ajaxOptions.data = JSON.stringify(data);
        }

        $.ajax(ajaxOptions);
    },
    delete: function (url, data, success) {
        const ajaxOptions = {
            url,
            method: 'DELETE',
            success,
            error: API._handleError
        };

        if (data !== null && data !== undefined) {
            ajaxOptions.contentType = 'application/json';
            ajaxOptions.data = JSON.stringify(data);
        }

        $.ajax(ajaxOptions);
    },
    upload: function (url, formData, success) {
        $.ajax({
            url: url,
            method: 'POST',
            data: formData,
            contentType: false,
            processData: false,
            success: success,
            error: API._handleError
        });
    },
    _handleError: function (xhr) {
        if (xhr.status === 401) {
            showErrorAndRedirect('Login expired, redirecting to login page...', 3000, '/login');
        } else {
            showError(xhr.responseText || 'Request failed');
        }
    }
};

