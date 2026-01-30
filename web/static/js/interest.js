let list = [];
let filtered = [];
let currentFilter = "";
let selectedId = null;
let state = loadAppState("interest");
let sortState = { key: null, order: null };


function applyState() {
    const sel = $("#sort-select");
    INTEREST_SORTS.forEach((s, i) => {
        sel.append(`<option value="${i}">${I18N[s]}</option>`);
    });
    sel.val(state.sort);
}

function loadInterests() {
    API.get(`/api/interest/list?sort=${state.sort}`, data => {
        list = data;
        updateView();
    });
}

function applyFilter() {
    if (!currentFilter) {
        filtered = [...list];
        return;
    }
    const f = currentFilter.toLowerCase();
    filtered = list.filter(i =>
        (JSON.stringify(i)).toLowerCase().includes(f)
    );
}

function renderTable() {
    const tbody = $("#interest-table tbody");
    tbody.empty();

    filtered.forEach(i => {
        const tr = $(`
        <tr data-id="${i.id}">
            <td style="display:none">${i.id}</td>
            <td contenteditable data-field="added" data-type="int" class="td-center">${i.added}</td>
            <td contenteditable data-field="name" data-type="string" class="td-left">${i.name}</td>
            <td>
            <select class="form-select form-select-sm sort-select" data-field="sort" data-type="int">
                ${INTEREST_SORTS.map((s, idx) => `
                <option value="${idx}" ${idx === i.sort ? "selected" : ""}>${I18N[s]}</option>
                `).join("")}
            </select>
            </td>
            <td contenteditable data-field="progress" data-type="string" class="td-center">${i.progress}</td>
            <td contenteditable data-field="publish" data-type="int" class="td-center">${i.publish}</td>
            <td contenteditable data-field="date" data-type="int" class="td-center">${i.date}</td>
            <td contenteditable data-field="score_db" data-type="float" class="td-center">${i.score_db}</td>
            <td contenteditable data-field="score_imdb" data-type="float" class="td-center">${i.score_imdb}</td>
            <td contenteditable data-field="score" data-type="float" class="td-center">${i.score}</td>
            <td contenteditable data-field="remark" data-type="string" class="td-left">${i.remark}</td>
        </tr>
        `);
        if (i.id == selectedId) { tr.addClass("table-active"); }
        tbody.append(tr);
    });
}

function updateView() {
    applyFilter();
    applySort(list, sortState.key, sortState.order);
    if (selectedId && !filtered.some(b => b.id === selectedId)) selectedId = null;
    renderTable();
}

/* ====== 事件 ====== */
$("#sort-select").on("change", function () {
    state.sort = Number(this.value);
    loadInterests();
});

$("#filter").on("input", function () {
    currentFilter = this.value;
    updateView();
});

$("#interest-table tbody").on("click", "tr", function () {
    selectedId = $(this).data("id");
    $(this).addClass("table-active").siblings().removeClass("table-active");
});

function handleUpdate() {
    const id = $(this).closest("tr").data("id");
    const interest = list.find(i => i.id === id)
    if (!interest) return;
    const field = $(this).data("field");
    const value = getEditorValue(this);
    if (interest[field] === value) return;
    interest[field] = value;
    API.post("/api/interest/update", { id, [field]: value });
}
$("#interest-table tbody")
    .on("change", ".sort-select", handleUpdate)
    .on("blur", "td[contenteditable]", handleUpdate);


$("#btn-add").click(() => {
    API.post("/api/interest/add", { sort: state.sort }, loadInterests);
});

$("#btn-del").click(async function() {
    if (!selectedId) return;
    const ok = await showConfirm(
        I18N["Delete selected record?"],
        'info'
    );
    if (!ok) return;
    API.delete(`/api/interest/delete?id=${selectedId}`, null, () => {
        list = list.filter(i => i.id !== selectedId);
        selectedId = null;
        updateView();
    });
});

$("#btn-import").click(() => $("#importFile").click());
$("#importFile").change(function () {
    if (!this.files.length) return;
    const form = new FormData();
    form.append("file", this.files[0]);
    API.upload('/api/interest/import', form,
        () => {
            loadInterests();
            showSuccess('Import successful');
            this.value = "";
        });
});

$("#btn-export").click(() => {
    window.location.href = `/api/interest/export`;
});


initTable("interest-table", sortState, updateView);
applyState();
loadInterests();
addUnloadListener("interest", state)

// 1. 粘贴时只允许纯文本
$("#interest-table tbody").on("paste", "td[contenteditable]", function (e) {
    e.preventDefault();
    const text = (e.originalEvent || e)
        .clipboardData
        .getData("text/plain");
    document.execCommand("insertText", false, text);
});

// 2. 禁止回车
$("#interest-table tbody").on("keydown", "td[contenteditable]", function (e) {
    if (e.key === "Enter") {
        e.preventDefault();
        this.blur();
    }
});
