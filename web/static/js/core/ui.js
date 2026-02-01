
const APP_KEY = 'diarygo_state';
const VIEW_DAILY = 0;
const VIEW_MONTHLY = 1;
const WEEK_DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const INTEREST_SORTS = ["All", "Movie", "TV", "Comic", "Game", "Book", "Music", "Others"]
const SAVE_DELAY = 500;

window.appState = {
    diary: {
        view: VIEW_DAILY,
        date: null,
        autosave: false,
    },
    bill: {
        start: null,
        end: null,
        month: null,

    },
    interest: {
        sort: 0,
    },
    note: {
        flag: 0,
    },
    sport: {},
}

function castValue(value, type) {
    switch (type) {
        case "int":
            return parseInt(value, 10);
        case "float":
            return parseFloat(value);
        case "bool":
            return value === "true" || value === "1";
        default:
            return value;
    }
}

function str2contenteditable(s) {
    return s ? s.replace(/\n/g, '<br>') : ""
}

function contenteditable2str(s) {
    if (!s) return '';

    return s
        .replace(/<\/div>\s*<div>/gi, '\n')
        .replace(/^<div>/i, '')
        .replace(/<\/div>$/i, '')
        .replace(/<\/?div>/gi, '\n')
        .replace(/<br\s*\/?>/gi, '\n')
        .replace(/\n{3,}/g, '\n\n');
}


function getEditorValue(el) {
    const $el = $(el);
    const type = $el.data('type')
    if ($el.is("select, input, textarea")) {
        return castValue($el.val(), type);
    }
    if ($el.is("[contenteditable]")) {
        if (type === "string") {
            let html = $el.html();
            html = contenteditable2str(html)
            return castValue(html.trim(), type);
        } else {
            return castValue($el.text().trim(), type);
        }
    }
    return null;
}

function date2int(dateInput) {
    let d;
    if (typeof dateInput === 'string') {
        d = new Date(dateInput);
    } else if (dateInput instanceof Date) {
        d = dateInput;
    } else {
        d = new Date();
    }
    return d.getFullYear() * 10000 + (d.getMonth() + 1) * 100 + d.getDate();
}

function date2str(dateInput) {
    let d;
    if (typeof dateInput === 'string') {
        d = new Date(dateInput);
    } else if (dateInput instanceof Date) {
        d = dateInput;
    } else {
        d = new Date();
    }
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}${m}${day}`;
}

function str2date(idStr) {
    if (!/^\d{8}$/.test(idStr)) return null;
    const y = parseInt(idStr.substring(0, 4));
    const m = parseInt(idStr.substring(4, 6)) - 1;
    const d = parseInt(idStr.substring(6, 8));
    return new Date(y, m, d);
}

function showMsg(message, type = 'danger', delay = 3000, redirect = null) {
    const infoBox = document.getElementById('info-box');

    infoBox.classList.remove('alert-danger', 'alert-success', 'alert-warning', 'alert-info', 'alert-primary');

    infoBox.classList.add(`alert-${type}`);

    infoBox.textContent = message;
    infoBox.classList.remove('d-none');

    setTimeout(() => {
        if (redirect) {
            window.location.href = redirect;
        }
        infoBox.classList.add('d-none');
    }, delay);
}

function showInfo(msg, delay = 3000) {
    showMsg(msg, 'info', delay);
}

function showError(msg, delay = 5000) {
    showMsg(msg, 'danger', delay);
}

function showErrorAndRedirect(msg, delay = 5000, redirect = "/") {
    showMsg(msg, 'danger', delay, redirect);
}

function showSuccess(msg, delay = 3000) {
    showMsg(msg, 'success', delay);
}

function showConfirm(message, type = 'danger') {
    return new Promise(resolve => {
        const modalEl = document.getElementById('confirmModal');
        const msgEl = document.getElementById('confirm-message');
        const yesBtn = document.getElementById('confirm-yes');

        msgEl.textContent = message;
        yesBtn.className = `btn btn-sm btn-${type}`;

        const modal = new bootstrap.Modal(modalEl);

        const cleanup = (result) => {
            modal.hide();
            yesBtn.onclick = null;
            document.getElementById('confirm-no').onclick = null;
            resolve(result);
        };

        yesBtn.onclick = () => cleanup(true);
        document.getElementById('confirm-no').onclick = () => cleanup(false);

        modal.show();
    });
}

function applyNavConfig() {
    if (!window.APP_CONFIG) {
        console.warn('APP_CONFIG not loaded yet');
        return;
    }

    if (!APP_CONFIG.show_bill) {
        $('#nav-bill').remove();
    }
    if (!APP_CONFIG.show_note) {
        $('#nav-note').remove();
    }
    if (!APP_CONFIG.show_interest) {
        $('#nav-interest').remove();
    }
    if (!APP_CONFIG.show_sport) {
        $('#nav-sport').remove();
    }
}

function loadAppState(module) {
    let saved = localStorage.getItem(APP_KEY);
    if (saved) {
        try {
            Object.assign(window.appState, JSON.parse(saved));
        } catch { }
    }
    return window.appState[module]
}

function saveAppState(module, state) {
    let fullState = {};
    const saved = localStorage.getItem(APP_KEY);
    if (saved) {
        try {
            fullState = JSON.parse(saved);
        } catch (err) {
            console.warn("Failed to parse existing app state", err);
        }
    }
    fullState[module] = { ...state };
    localStorage.setItem(APP_KEY, JSON.stringify(fullState));
}

function addUnloadListener(module, state, fn = null) {
    window.addEventListener('beforeunload', () => {
        if (fn !== null) { fn() }
        saveAppState(module, state);
    })
}

function debounce(fn, delay = 600) {
    let timer = null;
    return function (...args) {
        if (timer) clearTimeout(timer);
        timer = setTimeout(() => {
            fn.apply(this, args);
        }, delay);
    };
}

function applySort(list, key, order) {
    if (!key || !order) return;
    const factor = order === 'asc' ? 1 : -1;
    list.sort((a, b) => {
        let v1 = a[key], v2 = b[key];
        if (typeof v1 === 'number' && typeof v2 === 'number') return (v1 - v2) * factor;
        return String(v1).localeCompare(String(v2)) * factor;
    });
}

function serializeValue(val) {
    if (typeof val === 'boolean') {
        return val ? '1' : '0';
    }
    if (typeof val === 'number') {
        return String(val);
    }
    if (val == null) {
        return '';
    }
    return String(val);
}

function shallowEqual(a, b) {
    if (a === b) return true;
    if (!a || !b) return false;

    const aKeys = Object.keys(a);
    const bKeys = Object.keys(b);

    if (aKeys.length !== bKeys.length) return false;

    return aKeys.every(k => a[k] === b[k]);
}

function shallowEqualCommon(a, b) {
    if (!a || !b) return false;
    const commonKeys = Object.keys(a).filter(k =>
        Object.prototype.hasOwnProperty.call(b, k)
    );
    return commonKeys.every(k => a[k] === b[k]);
}

function createAutoSaver({ getEntity, readCurrent, save }) {
    const draftCache = new Map();
    let timer = null;

    function update(id) {
        const entity = getEntity(id);
        const current = readCurrent();

        if (shallowEqualCommon(entity, current)) {
            draftCache.delete(id);
            return;
        }

        draftCache.set(id, current);

        if (timer) clearTimeout(timer);
        timer = setTimeout(() => {
            timer = null;
            const draft = draftCache.get(id);
            if (!draft) return;
            save(id, draft, () => {
                Object.assign(entity, draft);
                draftCache.delete(id);
                console.log('Saved', id);
            });
        }, SAVE_DELAY);
    }

    return { update };
}

function createPatchSaver({
    getEntity,   // id => persisted object
    save        // (id, patch, done) => void
}) {
    const draftCache = new Map();
    let timer = null;

    function update(id, patch) {
        if (!patch || typeof patch !== 'object') return;

        const entity = getEntity(id);
        if (!entity) return;

        let draft = draftCache.get(id) || {};

        Object.assign(draft, patch);

        for (const k in patch) {
            if (entity[k] === patch[k]) {
                delete draft[k];
            }
        }

        if (Object.keys(draft).length === 0) {
            draftCache.delete(id);
            return;
        }

        draftCache.set(id, draft);

        if (timer) clearTimeout(timer);
        timer = setTimeout(() => {
            timer = null;
            const data = draftCache.get(id);
            if (!data) return;

            save(id, data, () => {
                Object.assign(entity, data);
                draftCache.delete(id);
                console.log('Saved', id);
            });
        }, SAVE_DELAY);
    }

    return { update };
}

function readTablePatch(el) {
  return {
    id: $(el).closest("tr").data("id"),
    patch: { [$(el).data("field")]: getEditorValue(el) }
  };
}
