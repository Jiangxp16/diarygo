let list = [];
let state = loadAppState("diary");


function applyState() {
    if (!state.date) {
        state.date = new Date().toISOString().split('T')[0];
    }
    $('#date-picker').val(state.date);
}

function loadDiaries() {
    let d = new Date(state.date);
    let month = d.getMonth() + 1;
    let year = d.getFullYear()
    $.getJSON(`/api/diary/list?month=${month}&year=${year}`, function (data) {
        list = data;
        updateView();
    });
}

function updateView() {
    if (state.view === VIEW_DAILY) {
        $('#input-weather, #input-location, #diary-table').show();
        $('#diary-list-table').hide();
        updateDailyView();
    } else if (state.view === VIEW_MONTHLY) {
        $('#input-weather, #input-location, #diary-table').show();
        $('#diary-list-table').hide();
        updateMonthlyView();
    } else if (state.view === VIEW_LIST) {
        $('#input-weather, #input-location, #diary-table').hide();
        $('#diary-list-table').show();
        updateListView();
    }
}

function updateListView() {
    let date = new Date(state.date);
    let year = date.getFullYear();
    let month = date.getMonth();

    let firstDate = new Date(year, month, 1);
    let lastDate = new Date(year, month + 1, 0);

    let html = '';
    let day = new Date(firstDate);
    while (day <= lastDate) {
        let id = dateToInt(day);
        let diary = list.find(d => d.id === id) || { content: "", weather: "", location: ""};
        html += `
            <tr data-id="${id}">
                <td data-field="date" data-type="int" class="td-center">${id}</td>
                <td contenteditable="true" data-field="weather" data-type="string" class="td-left">${diary.weather}</td>
                <td contenteditable="true" data-field="location" data-type="string" class="td-left">${diary.location}</td>
                <td contenteditable="true" data-field="content" data-type="string" class="td-left">${str2contenteditable(diary.content)}</td>
            </tr>
        `;
        day.setDate(day.getDate() + 1);
    }
    $('#diary-list-table tbody').html(html);
    const currentId = dateToInt(state.date);
    $(`#diary-list-table tbody tr[data-id="${currentId}"]`)
        .addClass('table-active')
        .siblings().removeClass('table-active');
}

function updateWeatherAndLocation() {
    const diary = list.find(d => d.id === dateToInt(state.date)) || { content: "", weather: "", location: "" };
    $('#input-weather').val(diary.weather);
    $('#input-location').val(diary.location);
}

function getLunarText(date, withMonth = false) {
    if (typeof solarlunar === 'undefined') {
        console.error('solarlunar-es undefined。');
        return '';
    }
    if (!APP_CONFIG.show_lunar) {
        return '';
    }
    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();
    try {
        const lunarData = solarlunar.solar2lunar(year, month, day);
        let text = ``;
        if (lunarData.dayCn === '初一' || withMonth) {
            text = `${lunarData.monthCn}`;
        }
        text += `${lunarData.dayCn}`;
        if (lunarData.isTerm) {
            text += ` ${lunarData.term}`;
        }
        return text;
    } catch (error) {
        console.error('getLunarText:', error);
        return '';
    }
}

function getHolidayText(date) {
    return ''; // TODO
}

function renderDailyHeader(date) {
    let solar = date.toDateString();
    let lunar = getLunarText(date, true);
    let holiday = getHolidayText(date);
    let html = `<div class="daily-date-solar">${solar}</div>`;
    if (lunar) {
        html += `<div class="daily-date-lunar">${lunar}</div>`;
    }
    if (holiday) {
        html += `<div class="daily-date-holiday">${holiday}</div>`;
    }
    return html;
}

function updateDailyView() {
    const date = new Date($('#date-picker').val());
    const diary = list.find(d => d.id === dateToInt(date)) || { content: "", weather: "", location: "" };
    $('#diary-table thead tr').html(`<th>${renderDailyHeader(date)}</th>`);
    $('#diary-table tbody').html(`<tr><td class="daily-cell"><div id="te-content"
                  class="daily-editor"
                  contenteditable="true"
                >${str2contenteditable(diary.content)}</div></td></tr>`);
    updateWeatherAndLocation();
}

function renderCellDate(date) {
    const dayNum = date.getDate();
    const lunar = getLunarText(date);
    const holiday = getHolidayText(date);
    let html = `<div class="month-date-solar">${dayNum}</div>`;
    if (lunar) {
        html += `<div class="month-date-lunar">${lunar}</div>`;
    }
    if (holiday) {
        html += `<div class="month-date-holiday">${holiday}</div>`;
    }
    return html;
}
function updateMonthlyView() {
    const date = new Date($('#date-picker').val());
    updateWeatherAndLocation()

    const year = date.getFullYear();
    const month = date.getMonth();

    const firstDate = new Date(year, month, 1);
    const lastDate = new Date(year, month + 1, 0);

    const firstDayOfWeek = APP_CONFIG.first_day_of_week % 7;
    const headers = WEEK_DAYS.slice(firstDayOfWeek - 1).concat(WEEK_DAYS.slice(0, firstDayOfWeek - 1));
    $('#diary-table thead tr').html(headers.map(d => `<th>${I18N[d] || d}</th>`).join(''));

    let html = '';
    let day = new Date(firstDate);
    day.setDate(day.getDate() - ((day.getDay() + 6) % 7 - (firstDayOfWeek - 1)));

    for (let r = 0; r < 6; r++) {
        html += '<tr>';
        for (let c = 0; c < 7; c++) {
            let id = dateToInt(day);
            let diary = list.find(d => d.id === id) || { content: "", weather: "", location: "" };
            let isCurrentMonth = day >= firstDate && day <= lastDate;
            let cellDateHtml = renderCellDate(day);
            if (isCurrentMonth) {
                html += `<td class="month-cell month-active" data-date="${id}">` +
                    `<div class="cell-date">${cellDateHtml}</div>` +
                    `<div class="cell-content" contenteditable="true">${str2contenteditable(diary.content)}</div>` +
                    `</td>`;
            } else {
                html += `<td class="month-cell month-disabled">` +
                    `<div class="cell-date text-muted">${cellDateHtml}</div>` +
                    `<div class="cell-content month-disabled" contenteditable="false"></div>` +
                    `</td>`;
            }

            day.setDate(day.getDate() + 1);
        }
        html += '</tr>';
    }
    $('#diary-table tbody').html(html);
}

$('#date-picker').change(function () {
    let previous = state.date;
    let current = this.value;
    state.date = current;
    if (previous === null || previous.substring(0, 7) !== current.substring(0, 7)) {
        loadDiaries();
    }
    updateWeatherAndLocation();
});
$('#btn-import').click(() => {
    $('#import-file').val('');
    $('#import-file').click();
});

$('#import-file').on('change', function () {
    if (!this.files.length) return;
    const form = new FormData();
    form.append("file", this.files[0]);
    API.upload('/api/diary/import', form,
        () => {
            loadDiaries();
            showSuccess('Import successful');
            this.value = "";
        },
    );
});

$('#btn-export').click(() => {
    console.log("Exporting...");
    window.location.href = `/api/diary/export`;
});
$('#btn-daily').click(() => {
    state.view = VIEW_DAILY;
    updateView();
});
$('#btn-monthly').click(() => {
    state.view = VIEW_MONTHLY;
    updateView();
});
$('#btn-list').click(() => {
    state.view = VIEW_LIST;
    updateView();
});


$('#diary-table').on('click', '.month-active', function () {
    let id = $(this).data('date'); // yyyymmdd
    state.date = dateIntToISOStr(id);
    $('#date-picker').val(state.date);
    updateWeatherAndLocation();
});

function getDiaryByID(id) {
    let diary = list.find(d => d.id === id);
    if (!diary) {
        diary = { id, content: "", weather: "", location: "" };
        list.push(diary);
    }
    return diary;
}

function getCurrentContent(id) {
    let html = '';
    if (state.view === VIEW_DAILY) {
        html = $('#te-content').html();
    } else {
        html = $(`td[data-date="${id}"] .cell-content`).html();
    }
    html = contenteditable2str(html)
    return html.trim();
}

function readCurrentEditor(id) {
    if (state.view === VIEW_LIST) {
        const $tr = $(`#diary-list-table tbody tr[data-id="${id}"]`);
        if (!$tr.length) return null;

        let data = {};

        $tr.find('[data-field]').each(function () {
            const field = $(this).data('field');
            const value = getEditorValue(this);
            if (value !== null) {
                data[field] = value;
            }
        });

        return {
            content: data.content || "",
            weather: data.weather || "",
            location: data.location || "",
        };
    }
    return {
        content: getCurrentContent(id),
        weather: $('#input-weather').val().trim(),
        location: $('#input-location').val().trim(),
    };
}

const updater = createAutoSaver({
  getEntity: getDiaryByID,
  readCurrent: (id) => readCurrentEditor(id),
  save: (id, data, done) => {
    API.post('/api/diary/update', { id, ...data }, done);
  }
});

$("#diary-list-table tbody").on("click", "tr", function () {
    let id = $(this).data('id');
    if (!id) return;
    state.date = dateIntToISOStr(id);
    $('#date-picker').val(state.date);
    $(this).addClass("table-active").siblings().removeClass("table-active");
});

$('#diary-list-table tbody').on('input', 'td[contenteditable]', function () {
    const id = $(this).closest('tr').data('id');
    if (!id) return;
    updater.update(id);
});

$('#diary-table tbody').on('input', '[contenteditable]', function () {
    updater.update(dateToInt(state.date));
});

$('#input-weather, #input-location').on('input', function () {
    updater.update(dateToInt(state.date));
});

$('#input-weather, #input-location').on('keydown', function (e) {
    if (e.key === 'Enter') {
        e.preventDefault();
        updater.update(dateToInt(state.date));
        this.blur();
    }
});


applyNavConfig();
applyState();
setPastePlain("diary-table")
initTable("diary-list-table", null, new Set(["weather", "location"]));
loadDiaries();
addUnloadListener("diary", state)
