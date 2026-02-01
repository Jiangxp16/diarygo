let list = [];
let filtered = [];
let currentFilter = "";
let selectedId = null;
let state = loadAppState("bill");
let sortState = { key: null, order: null };


function applyState() {
  if (state.month) {
    $('#month-picker').val(state.month);
  } else {
    const today = new Date();
    const monthStr = today.toISOString().slice(0, 7);
    $('#month-picker').val(monthStr);
  }
  if (state.start) {
    $('#date-start').val(state.start);
  } else {
    const today = new Date();
    const firstDay = new Date(Date.UTC(today.getFullYear(), today.getMonth(), 1));
    $('#date-start').val(firstDay.toISOString().split('T')[0]);
  }
  if (state.end) {
    $('#date-end').val(state.end);
  } else {
    const today = new Date();
    const lastDay = new Date(Date.UTC(today.getFullYear(), today.getMonth() + 1, 0));
    $('#date-end').val(lastDay.toISOString().split('T')[0]);
  }
}


function loadBills() {
  const start = date2int($('#date-start').val());
  const end = date2int($('#date-end').val());

  API.get(`/api/bill/list?start=${start}&end=${end}`, data => {
    list = data;
    updateView();
  });
}

function applyFilter() {
  if (!currentFilter) {
    filtered = [...list];
    return;
  }

  const f = currentFilter.trim();
  let lessThan = null;
  let moreThan = null;

  if (f.startsWith("<")) {
    lessThan = parseFloat(f.slice(1));
  } else if (f.startsWith(">")) {
    moreThan = parseFloat(f.slice(1));
  }

  filtered = list.filter(bill => {
    if (lessThan !== null) {
      return bill.amount <= lessThan;
    }
    if (moreThan !== null) {
      return bill.amount >= moreThan;
    }
    return (
      (bill.date + bill.type + bill.item).toLowerCase().includes(f.toLowerCase())
    );
  });
}

function renderTable() {
  const tbody = $("#bill-table tbody");
  tbody.empty();

  filtered.forEach(bill => {
    const tr = $(`
        <tr data-id="${bill.id}">
          <td style="display:none">${bill.id}</td>
          <td contenteditable="true" data-field="date" data-type="int" class="td-center">${bill.date}</td>
          <td>
            <select class="form-select form-select-sm inout-select" data-field="inout" data-type="int">
              <option value="1" ${bill.inout === 1 ? "selected" : ""}>${I18N["In"]}</option>
              <option value="-1" ${bill.inout === -1 ? "selected" : ""}>${I18N["Out"]}</option>
            </select>
          </td>
          <td contenteditable="true" data-field="type" data-type="string" class="td-center">${bill.type}</td>
          <td contenteditable="true" data-field="amount" data-type="float" class="td-center">${bill.amount}</td>
          <td contenteditable="true" data-field="item" data-type="string" class="td-left">${bill.item}</td>
        </tr>
      `);
    if (bill.id == selectedId) { tr.addClass("table-active"); }
    tbody.append(tr);
  });
}


function updateTotal() {
  let total = 0;
  let totalIn = 0;
  let totalOut = 0;

  filtered.forEach(bill => {
    const amount = Number(bill.amount);
    if (bill.inout > 0) {
      total += amount;
      totalIn += amount;
    } else {
      total -= amount;
      totalOut += amount;
    }
  });

  $("#total").text(total.toFixed(2));
  $("#total-in").text(totalIn.toFixed(2));
  $("#total-out").text(totalOut.toFixed(2));
}

function updateView() {
  applyFilter();
  applySort(list, sortState.key, sortState.order);
  if (selectedId && !filtered.some(b => b.id === selectedId)) selectedId = null;
  renderTable();
  updateTotal();
}

/* ====== 事件 ====== */
$('#month-picker').change(() => {
  state.month = $('#month-picker').val()
  if (!state.month) return;
  const [year, month] = state.month.split('-').map(Number);

  const firstDay = new Date(Date.UTC(year, month - 1, 1));
  const startDateStr = firstDay.toISOString().split('T')[0];

  const lastDay = new Date(Date.UTC(year, month, 0));
  const endDateStr = lastDay.toISOString().split('T')[0];

  $('#date-start').val(startDateStr);
  $('#date-end').val(endDateStr);
  state.start = startDateStr;
  state.end = endDateStr;
  loadBills();
});

$("#filter").on("input", function () {
  currentFilter = $(this).val();
  updateView();
});

$('#date-start').change(() => {
  state.start = $('#date-start').val();
  loadBills();
});

$('#date-end').change(() => {
  state.end = $('#date-end').val();
  loadBills();
});

$("#bill-table tbody").on("click", "tr", function () {
  selectedId = $(this).data("id");
  $(this).addClass("table-active").siblings().removeClass("table-active");
});

const updater = createPatchSaver({
  getEntity: id => list.find(o => o.id === id),
  save: (id, patch, done) => {
    API.post('/api/bill/update', { ...patch, id }, () => {
      done();
      updateTotal();
    });
  }
});

$("#bill-table tbody")
  .on("input", "td[contenteditable]", function () {
    const { id, patch } = readTablePatch(this);
    updater.update(id, patch);
  })
  .on("change", ".inout-select", function () {
    const { id, patch } = readTablePatch(this);
    updater.update(id, patch);
  });

$('#btn-add').click(() => {
  API.post('/api/bill/add', {}, loadBills);
});

$("#btn-del").on("click", function () {
  if (!selectedId) return;
  API.delete(`/api/bill/delete?id=${selectedId}`, null, () => {
    list = list.filter(b => b.id !== selectedId);
    selectedId = null;
    updateView();
  });
});

$('#btn-import').click(() => $('#importFile').click());
$('#importFile').change(function () {
  if (!this.files.length) return;
  let form = new FormData();
  form.append("file", this.files[0]);
  API.upload('/api/bill/import', form, () => {
    loadBills();
    showSuccess('Import successful');
    this.value = "";
  });
});

$('#btn-export').click(() => {
  window.location.href = `/api/bill/export`;
});


applyNavConfig();
initTable("bill-table", sortState, "all");
applyState();
loadBills();
addUnloadListener("bill", state)
