let diaries = [];
let state = loadAppState("diary");
const draftCache = new Map(); // key: yyyymmdd, value: {content, weather, location}
let saveTimer = null;
const SAVE_DELAY = 300;


function applyState() {
    if (state.date) {
        $('#date-picker').val(state.date);
    } else {
        let today = new Date().toISOString().split('T')[0];
        $('#date-picker').val(today);
        state.date = today;
    }
}


function loadDiaries() {
    let d = new Date(state.date);
    let month = d.getMonth() + 1;
    let year = d.getFullYear()
    $.getJSON(`/api/diary/list?month=${month}&year=${year}`, function (data) {
        diaries = data;
        updateView();
    });
}

function updateView() {
    if (state.view === VIEW_DAILY) {
        updateDailyView();
    } else {
        updateMonthlyView();
    }
}

function updateWeatherAndLocation() {
    const diary = diaries.find(d => d.id === date2int(state.date)) || { content: "", weather: "", location: "" };
    $('#le-weather').val(diary.weather);
    $('#le-location').val(diary.location);
}

function getLunarText(date, withMonth = false) {
    if (typeof solarlunar === 'undefined') {
        console.error('solarlunar-es 库未正确加载。');
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
    const date = new Date(state.date);
    const diary = diaries.find(d => d.id === date2int(date)) || { content: "", weather: "", location: "" };
    $('#tw-header').html(`<th>${renderDailyHeader(date)}</th>`);
    $('#tw-body').html(`<tr><td class="daily-cell"><div id="te-content"
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
    let dateStr = $('#date-picker').val();
    let date = new Date(dateStr);
    updateWeatherAndLocation()

    let year = date.getFullYear();
    let month = date.getMonth();

    let firstDate = new Date(year, month, 1);
    let lastDate = new Date(year, month + 1, 0);

    let firstDayOfWeek = APP_CONFIG.first_day_of_week % 7; // 默认为周一
    let headers = WEEK_DAYS.slice(firstDayOfWeek - 1)
        .concat(WEEK_DAYS.slice(0, firstDayOfWeek - 1));
    $('#tw-header').html(headers.map(d => `<th>${I18N[d] || d}</th>`).join(''));

    let html = '';
    let day = new Date(firstDate);
    day.setDate(day.getDate() - ((day.getDay() + 6) % 7 - (firstDayOfWeek - 1)));

    for (let r = 0; r < 6; r++) {
        html += '<tr>';
        for (let c = 0; c < 7; c++) {
            let id = date2int(day);
            let diary = diaries.find(d => d.id === id) || { content: "", weather: "", location: "" };
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

    $('#tw-body').html(html);
}

// -------------------- 事件绑定 --------------------
$('#date-picker').change(function () {
    let previous = state.date
    let current = $('#date-picker').val()
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


$('#tw-content').on('click', '.month-active', function () {
    let id = $(this).data('date'); // yyyymmdd
    let y = id.toString().substring(0, 4);
    let m = id.toString().substring(4, 6);
    let d = id.toString().substring(6, 8);

    let dateStr = `${y}-${m}-${d}`;
    $('#date-picker').val(dateStr);
    state.date = dateStr;
    updateWeatherAndLocation();
});

function getDiaryByDate(dateStr) {
    const id = date2int(dateStr);
    let diary = diaries.find(d => d.id === id);
    if (!diary) {
        diary = { id, content: "", weather: "", location: "" };
        diaries.push(diary);
    }
    return diary;
}

function getCurrentContent() {
    let html = '';
    if (state.view === VIEW_DAILY) {
        html = $('#te-content').html();
    } else {
        const id = date2int(state.date);
        html = $(`td[data-date="${id}"] .cell-content`).html();
    }
    html = contenteditable2str(html)
    return html.trim();
}

function readCurrentEditor() {
    return {
        content: getCurrentContent(),
        weather: $('#le-weather').val().trim(),
        location: $('#le-location').val().trim(),
    };
}

function shallowEqual(a, b) {
    return a &&
        b &&
        a.content === b.content &&
        a.weather === b.weather &&
        a.location === b.location;
}

function updateDraft(dateStr) {
    const diary = getDiaryByDate(dateStr);
    const current = readCurrentEditor();
    if (
        diary.content === current.content &&
        diary.weather === current.weather &&
        diary.location === current.location
    ) {
        draftCache.delete(dateStr);
        return;
    }
    draftCache.set(dateStr, current);
    scheduleSave(dateStr);
}

function scheduleSave(dateStr) {
    if (saveTimer) clearTimeout(saveTimer);

    saveTimer = setTimeout(() => {
        saveTimer = null;
        saveDraft(dateStr);
    }, SAVE_DELAY);
}

function saveDraft(dateStr) {
    const draft = draftCache.get(dateStr);
    if (!draft) return;
    const diary = getDiaryByDate(dateStr);
    API.post('/api/diary/update', {
        id: diary.id,
        content: draft.content,
        weather: draft.weather,
        location: draft.location,
    }, () => {
        diary.content = draft.content;
        diary.weather = draft.weather;
        diary.location = draft.location;
        draftCache.delete(dateStr);
        console.log('Saved', dateStr);
    });
}

$('#tw-body').on('input', '[contenteditable]', function () {
    updateDraft(state.date);
});

$('#le-weather, #le-location').on('input', function () {
    updateDraft(state.date);
});

$('#le-weather, #le-location').on('keydown', function (e) {
    if (e.key === 'Enter') {
        e.preventDefault();
        updateDraft(state.date);
        this.blur();
    }
});


applyState();    // 1. 加载页面状态
loadDiaries();   // 2. 加载数据
updateView();    // 3. 渲染视图
addUnloadListener("diary", state)
