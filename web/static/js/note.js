let list = [];
let filtered = [];
let currentFilter = "";
let selectedId = null;
let state = loadAppState("note")
let stateFilter = 0;
let sortState = { key: null, order: null };

function applyState() {
    flag = state.flag || 0
    $("#flag-select").val(flag)
}

function loadNotes() {
    API.get('/api/note/list', data => {
        list = data;
        updateView();
    });
}

function applyFilter() {
    filtered = list.filter(note => {
        let matchFilter = !currentFilter || String(note).toLowerCase().includes(currentFilter.toLowerCase());
        let matchState = stateFilter === 0 ||
            (stateFilter === 1 && note.process < 100) ||
            (stateFilter === 2 && note.process >= 100);
        return matchFilter && matchState;
    });
}

function renderTable() {
    const tbody = $("#note-table tbody");
    tbody.empty();
    filtered.forEach(note => {
        const tr = $(`
      <tr data-id="${note.id}">
        <td style="display:none">${note.id}</td>
        <td contenteditable="true" data-field="begin" data-type="int" class="td-center">${note.begin}</td>
        <td contenteditable="true" data-field="last" data-type="int" class="td-center">${note.last}</td>
        <td contenteditable="true" data-field="process" data-type="int" class="td-center">${note.process}</td>
        <td contenteditable="true" data-field="desire" data-type="int" class="td-center">${note.desire}</td>
        <td contenteditable="true" data-field="priority" data-type="int" class="td-center">${note.priority}</td>
        <td contenteditable="true" data-field="content" data-type="string" class="td-left">${str2contenteditable(note.content)}</td>
      </tr>
    `);
        if (note.id == selectedId) tr.addClass('table-active');
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

$("#note-table tbody").on("click", "tr", function () {
    selectedId = $(this).data("id");
    $(this).addClass("table-active").siblings().removeClass("table-active");
});

function handleUpdate() {
    const id = $(this).closest("tr").data("id");
    const note = list.find(n => n.id === id);
    if (!note) return;
    const field = $(this).data("field");
    const value = getEditorValue(this);
    if (note[field] === value) return;
    note[field] = value;
    API.post('/api/note/update', { id, [field]: value });
}

$("#note-table tbody").on("blur", "td[contenteditable]", handleUpdate);

$("#btn-add").click(() => {
    API.post('/api/note/add', {}, () => {
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
    API.delete(`/api/note/delete?id=${selectedId}`, null, () => {
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
    API.upload('/api/note/import', form, () => {
        loadNotes();
        showSuccess('Import successful');
        this.value = "";
    });
});

$("#btn-expprt").click(() => {
    window.location.href = `/api/note/export`;
});

initTable("note-table", sortState, updateView);
applyState();
loadNotes();
addUnloadListener("note", state)

const noEnterFields = new Set([
    "begin",
    "last",
    "progress",
    "desire",
    "priority",
]);

$("#note-table tbody").on("keydown", "td[contenteditable]", function (e) {
    if (e.key !== "Enter") return;

    const field = $(this).data("field");
    if (!noEnterFields.has(field)) return;

    e.preventDefault();
    this.blur();
});
