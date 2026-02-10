let list = [];
let filtered = [];
let currentFilter = "";
let selectedId = null;
let state = loadAppState("sport")
let stateFilter = 0;
let sortState = { key: null, order: null };

function applyState() {
}

function loadNotes() {
    API.get('/api/sport/list', data => {
        list = data;
        updateView();
    });
}

function applyFilter() {
    filtered = list.filter(sport => {
        let matchFilter = !currentFilter || String(sport).toLowerCase().includes(currentFilter.toLowerCase());
        let matchState = stateFilter === 0 ||
            (stateFilter === 1 && sport.process < 100) ||
            (stateFilter === 2 && sport.process >= 100);
        return matchFilter && matchState;
    });
}

function renderTable() {
    const tbody = $("#sport-table tbody");
    tbody.empty();
    filtered.forEach(sport => {
        const tr = $(`
      <tr data-id="${sport.id}">
        <td style="display:none">${sport.id}</td>
        <td contenteditable="true" data-field="date" data-type="int" class="td-center">${sport.date}</td>
        <td contenteditable="true" data-field="content" data-type="string" class="td-left">${str2contenteditable(sport.content)}</td>
      </tr>
    `);
        if (sport.id == selectedId) tr.addClass('table-active');
        tbody.append(tr);
    });
}

function updateView() {
    applyFilter();
    applySort(list, sortState.key, sortState.order);
    if (selectedId && !filtered.some(n => n.id === selectedId)) selectedId = null;
    renderTable();
}

$("#flag-select").on("change", function () {
    stateFilter = parseInt($(this).val());
    updateView();
});

$("#filter").on("input", function () {
    currentFilter = $(this).val();
    updateView();
});

$("#sport-table tbody").on("click", "tr", function () {
    selectedId = $(this).data("id");
    $(this).addClass("table-active").siblings().removeClass("table-active");
});

const updater = createPatchSaver({
    getEntity: id => list.find(o => o.id === id),
    save: (id, patch, done) => {
        API.post('/api/sport/update', { ...patch, id }, () => {
            done();
        });
    }
});

$("#sport-table tbody")
    .on("input", "td[contenteditable]", function () {
        const { id, patch } = readTablePatch(this);
        updater.update(id, patch);
    })

$("#btn-add").click(() => {
    API.post('/api/sport/add', {}, () => {
        loadNotes()
    });
});

$("#btn-del").click(async function() {
    if (!selectedId) return;
    const ok = await showConfirm(
        I18N["Delete selected record?"],
        'info'
    );
    if (!ok) return;
    API.delete(`/api/sport/delete?id=${selectedId}`, null, () => {
        list = list.filter(n => n.id !== selectedId);
        selectedId = null;
        updateView();
    });
});

$("#btn-import").click(() => $('#importFile').click());
$("#importFile").change(function () {
    if (!this.files.length) return;
    const form = new FormData();
    form.append("file", this.files[0]);
    API.upload('/api/sport/import', form, () => {
        loadNotes();
        showSuccess('Import successful');
        this.value = "";
    });
});

$("#btn-export").click(() => {
    window.location.href = `/api/sport/export`;
});

applyNavConfig();
initTable("sport-table", sortState, new Set(["date"]));
applyState();
loadNotes();
addUnloadListener("sport", state)
addHeartbeat();
