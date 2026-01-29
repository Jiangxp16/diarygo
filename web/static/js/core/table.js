
function enableColumnResize(tableId) {
  const table = document.getElementById(tableId);
  if (!table) return;

  const colgroup = table.querySelector("colgroup");
  if (!colgroup) return;

  const cols = colgroup.querySelectorAll("col");
  const ths = table.querySelectorAll("thead th");

  ths.forEach((th, index) => {
    // 第 0 列没有左分隔线，直接跳过
    if (index === 0) return;

    if (th.querySelector(".col-resizer")) return;

    const resizer = document.createElement("div");
    resizer.className = "col-resizer";
    th.appendChild(resizer);

    let startX, startWidth;
    const targetCol = cols[index - 1]; // ⭐ 核心

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

  // 已存在就不重复加
  if (table.querySelector("colgroup")) return;

  const thead = table.querySelector("thead");
  if (!thead) return;

  const colgroup = document.createElement("colgroup");

  thead.querySelectorAll("th").forEach(th => {
    const col = document.createElement("col");
    col.style.width = th.offsetWidth
      ? th.offsetWidth + "px"
      : "120px";
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
    if (cols[i] && w) cols[i].style.width = w;
  });
}

function enableTableSort(tableSelector, sortState, onSortChange) {
  const $thead = $(`${tableSelector} thead`);

  $thead.on("click", "th.sortable", function (e) {
    // ⭐ 屏蔽列宽拖拽手柄
    if (e.target.closest(".col-resizer")) return;

    const key = $(this).data("key");
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

    onSortChange && onSortChange(sortState);
  });
}

function initTable(tabelId, sortState, onSortChange) {
    enableTableSort('#'+tabelId, sortState, onSortChange);
    initColGroup(tabelId)
    restoreColWidths(tabelId)
    enableColumnResize(tabelId)
}