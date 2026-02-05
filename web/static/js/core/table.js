function getVisibleThs(table) {
  return [...table.querySelectorAll("thead th")]
    .filter(th => th.offsetParent !== null);
}

function getVisibleCols(table) {
  return [...table.querySelectorAll("colgroup col")]
    .filter(col => col.style.display !== "none");
}


function enableColumnResize(tableId) {
  const table = document.getElementById(tableId);
  if (!table) return;

  const colgroup = table.querySelector("colgroup");
  if (!colgroup) return;

  const ths = getVisibleThs(table);
  const cols = table.querySelectorAll("colgroup col");

  ths.forEach((th, index) => {
    if (th.querySelector(".col-resizer")) return;

    const resizer = document.createElement("div");
    resizer.className = "col-resizer";
    th.appendChild(resizer);

    let startX, startWidth;
    const targetCol = cols[index];

    resizer.addEventListener("mousedown", e => {
      e.preventDefault();
      e.stopPropagation();

      startX = e.pageX;
      startWidth =
        parseInt(targetCol.style.width, 10) ||
        th.previousElementSibling.offsetWidth;

      document.onmousemove = e2 => {
        const dx = e2.pageX - startX;
        const newWidth = Math.max(30, startWidth + dx);
        targetCol.style.width = newWidth + "px";
      };

      document.onmouseup = () => {
        document.onmousemove = null;
        document.onmouseup = null;
        saveColWidths(tableId)
      };
    });
  });
}



function setColWidth(table, index, width) {
  const col = table.querySelectorAll("colgroup col")[index];
  if (col) col.style.width = width + "px";
}


function initColGroup(tableId) {
  const table = document.getElementById(tableId);
  if (!table) return;

  if (table.querySelector("colgroup")) return;

  const thead = table.querySelector("thead");
  if (!thead) return;

  const colgroup = document.createElement("colgroup");
  const ths = getVisibleThs(table);

  ths.forEach((th, i, arr) => {
    const col = document.createElement("col");
    if (i !== arr.length - 1) {
      col.style.width = th.offsetWidth + "px";
    }
    colgroup.appendChild(col);
  });

  table.insertBefore(colgroup, thead);
}

function saveColWidths(tableId) {
  const cols = document
    .getElementById(tableId)
    .querySelectorAll("colgroup col");

  const widths = [...cols].map(c => c.style.width);
  localStorage.setItem(tableId + "_cols", JSON.stringify(widths));
}

function restoreColWidths(tableId) {
  const data = localStorage.getItem(tableId + "_cols");
  if (!data) return;

  const widths = JSON.parse(data);
  const cols = document
    .getElementById(tableId)
    .querySelectorAll("colgroup col");

  widths.forEach((w, i) => {
    if (i !== cols.length - 1 && cols[i] && w) cols[i].style.width = w;
  });
}

function enableTableSort(tableSelector, sortState) {
  const $thead = $(`${tableSelector} thead`);

  $thead.on("click", "th.sortable", function (e) {
    if (e.target.closest(".col-resizer")) return;

    const $th = $(this);
    const key = $th.data("key");
    if (!key) return;

    if (sortState.key !== key) {
      sortState.key = key;
      sortState.order = "asc";
    } else {
      if (sortState.order === "asc") {
        sortState.order = "desc";
      } else if (sortState.order === "desc") {
        sortState.key = null;
        sortState.order = null;
      } else {
        sortState.order = "asc";
      }
    }

    $thead.find("th.sortable")
      .removeClass("active asc desc");

    if (sortState.key === key && sortState.order) {
      $th
        .addClass("active")
        .addClass(sortState.order);
    }

    updateView();
  });
}


function setPastePlain(tableId) {
  const $tableBody = $(`#${tableId} tbody`);
  $tableBody.on("paste", "[contenteditable]", function (e) {
    e.preventDefault();
    const text = (e.originalEvent || e).clipboardData.getData("text/plain");
    const selection = window.getSelection();
    if (!selection.rangeCount) return;
    selection.deleteFromDocument();
    selection.getRangeAt(0).insertNode(document.createTextNode(text));
    selection.collapseToEnd();
    this.dispatchEvent(new Event("input", { bubbles: true }));
  });
}

function disableEnterKey(tableId, fields) {
  const $tableBody = $(`#${tableId} tbody`);
  $tableBody.on("keydown", "[contenteditable]", function (e) {
    if (e.key !== "Enter") return;
    const field = $(this).data("field");
    if (fields==="all" || fields.has(field)){
      e.preventDefault();
      this.blur();
    }
  });
}

function rollBackTableEdits(tableId) {
  const $tableBody = $(`#${tableId} tbody`);
  $tableBody.on("blur", "[contenteditable][data-type!='string']", function () {
    const value = getEditorValue(this);
    if (value === null) {
        const id = $(this).closest("tr").data("id");
        const field = $(this).data("field");
        const original = list.find(b => b.id === id)?.[field];
        $(this).text(original ?? "");
    }
});
}

function initTable(tableId, sortState, fieldsToDisableEnter=null) {
    if (sortState) {
      enableTableSort('#'+tableId, sortState);
    }
    initColGroup(tableId)
    restoreColWidths(tableId)
    enableColumnResize(tableId)
    setPastePlain(tableId)
    if (fieldsToDisableEnter) {
        disableEnterKey(tableId, fieldsToDisableEnter)
    }
    rollBackTableEdits(tableId)
}

function readTablePatch(el) {
  return {
    id: $(el).closest("tr").data("id"),
    patch: { [$(el).data("field")]: getEditorValue(el) }
  };
}
