// Pagination & Global Variables
let currentPage = 1;
const limitPerPage = 10;
let isLoading = false;
let socket;

// --- Link Management Functions ---

async function shortenUrl() {
  const longUrlInput =
    document.getElementById("longUrl");
  const urlValue = longUrlInput.value.trim();

  if (!urlValue) {
    alert("Vui lòng nhập URL cần rút gọn!");
    longUrlInput.focus();
    return;
  }

  if (
    !urlValue.startsWith("http://") &&
    !urlValue.startsWith("https://")
  ) {
    alert(
      "URL phải bắt đầu bằng http:// hoặc https://"
    );
    return;
  }

  const durationType =
    document.getElementById("expireUnit").value;
  const durationValue =
    parseInt(
      document.getElementById("expireValue")
        .value
    ) || 1;

  try {
    const token =
      localStorage.getItem("jwt_token");
    const res = await fetch("/api/shorten", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        original_url: urlValue,
        custom_alias: document.getElementById(
          "customAlias"
        ).value,
        duration_type: durationType,
        duration_value: durationValue,
      }),
    });

    const data = await res.json();
    if (res.ok) {
      loadLinks();
      longUrlInput.value = "";
      document.getElementById(
        "customAlias"
      ).value = "";
    } else {
      alert(data.error || "Có lỗi xảy ra");
    }
  } catch (err) {
    console.error("Lỗi:", err);
  }
}

async function loadLinks(page = 1) {
  if (isLoading) return;
  isLoading = true;
  currentPage = page;
  try {
    const sortVal =
      document.getElementById(
        "sortSelect"
      ).value;
    const token =
      localStorage.getItem("jwt_token");
    const res = await fetch(
      `/api/links?sort=${sortVal}&page=${currentPage}&limit=${limitPerPage}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    if (!res.ok)
      throw new Error(
        "Không thể kết nối server"
      );

    const result = await res.json();
    const data = result.data;
    const list =
      document.getElementById("linkList");

    if (!data || data.length === 0) {
      list.innerHTML = `<tr><td colspan="5" class="text-center text-muted">Chưa có link nào được tạo</td></tr>`;
      updatePagination(0, 1, 1);
      return;
    }

    const currentIds = data.map(
      (l) => `row-${l.id}`
    );
    Array.from(list.children).forEach((row) => {
      if (
        row.id &&
        !currentIds.includes(row.id)
      )
        row.remove();
    });

    data.forEach((l, index) => {
      const rowId = `row-${l.id}`;
      let row = document.getElementById(rowId);
      const shortFullUrl =
        window.location.origin +
        "/" +
        l.short_code;
      const qrImageUrl = `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${shortFullUrl}`;

      if (!row) {
        row = document.createElement("tr");
        row.id = rowId;
        if (index === 0) list.prepend(row);
        else list.appendChild(row);
        row.innerHTML = `
                    <td class="col-code"></td>
                    <td class="col-url text-truncate" style="max-width: 200px;"></td>
                    <td class="col-stats"></td>
                    <td class="col-qr text-center"></td>
                    <td class="col-action"></td>
                `;
      }

      const codeCell =
        row.querySelector(".col-code");
      const codeHtml = `
                <div class="d-flex align-items-center gap-2">
                    <a href="/${l.short_code}" target="_blank" class="fw-bold text-decoration-none">${l.short_code}</a>
                    <button class="btn btn-sm btn-light border" onclick="copyToClipboard('${shortFullUrl}')">📋</button>
                </div>`;
      if (codeCell.innerHTML !== codeHtml)
        codeCell.innerHTML = codeHtml;

      const urlCell =
        row.querySelector(".col-url");
      if (
        urlCell.innerText !== l.original_url
      ) {
        urlCell.innerText = l.original_url;
        urlCell.title = l.original_url;
      }

      const statsCell =
        row.querySelector(".col-stats");
      let clickSpan = statsCell.querySelector(
        ".click-count"
      );
      if (!clickSpan) {
        statsCell.innerHTML = `
                    <div><span class="fw-bold fs-5 click-count">${l.click_count}</span> clicks <span class="timer-area"></span></div>
                    <div class="analytics-area"></div>
                `;
        clickSpan = statsCell.querySelector(
          ".click-count"
        );
      }
      if (clickSpan.innerText != l.click_count)
        clickSpan.innerText = l.click_count;

      const timerArea = statsCell.querySelector(
        ".timer-area"
      );
      const currentExpire = timerArea
        .querySelector(".countdown")
        ?.getAttribute("data-expire");

      if (l.expired_at !== currentExpire) {
        if (!l.expired_at) {
          timerArea.innerHTML = `<span class="badge bg-light text-muted ms-2">∞</span>`;
        } else {
          const isExpired =
            new Date(l.expired_at).getTime() -
              new Date().getTime() <=
            0;
          if (isExpired) {
            timerArea.innerHTML = `<span class="badge bg-danger ms-2">Hết hạn</span>`;
          } else {
            timerArea.innerHTML = `<span class="badge bg-warning text-dark ms-2 countdown" data-expire="${l.expired_at}">Đang tính...</span>`;
          }
        }
      }

      const analyticsArea =
        statsCell.querySelector(
          ".analytics-area"
        );
      const analyticsHtml =
        l.last_browser || l.last_os
          ? `
                <div style="font-size: 0.7rem;" class="text-muted mt-1">
                    <span class="badge border text-dark fw-normal">🌐 ${
                      l.last_browser || "N/A"
                    }</span>
                    <span class="badge border text-dark fw-normal">💻 ${
                      l.last_os || "N/A"
                    }</span>
                </div>`
          : "";
      if (
        analyticsArea.innerHTML !==
        analyticsHtml
      )
        analyticsArea.innerHTML = analyticsHtml;

      const qrCell =
        row.querySelector(".col-qr");
      if (!qrCell.innerHTML) {
        qrCell.innerHTML = `
                    <button class="btn btn-sm btn-outline-info" onclick="viewLargeQR('${qrImageUrl}')">🔍 Xem QR</button>
                    <button class="btn btn-sm btn-outline-primary mt-1" onclick="downloadQR('${qrImageUrl}', '${l.short_code}')">💾</button>`;
      }

      const actionCell = row.querySelector(
        ".col-action"
      );
      if (!actionCell.innerHTML) {
        actionCell.innerHTML = `
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-warning" onclick="editLink(${l.id}, '${l.original_url}')">✏️</button>
                        <button class="btn btn-danger" onclick="deleteLink(${l.id})">🗑️</button>
                    </div>`;
      }
    });

    updatePagination(
      result.total,
      result.page,
      result.last_page
    );
    startRealtimeCounter();
  } catch (err) {
    console.error("Lỗi loadLinks:", err);
  } finally {
    isLoading = false;
  }
}

function updatePagination(
  total,
  page,
  lastPage
) {
  const container = document.getElementById(
    "paginationControls"
  );
  const info =
    document.getElementById("pageInfo");
  info.innerText = `Trang ${page} / ${lastPage} (Tổng ${total} links)`;
  let html = "";
  html += `<li class="page-item ${
    page <= 1 ? "disabled" : ""
  }"><a class="page-link" href="#" onclick="loadLinks(${
    page - 1
  })">Trước</a></li>`;
  for (let i = 1; i <= lastPage; i++) {
    if (
      i === 1 ||
      i === lastPage ||
      (i >= page - 2 && i <= page + 2)
    ) {
      html += `<li class="page-item ${
        i === page ? "active" : ""
      }"><a class="page-link" href="#" onclick="loadLinks(${i})">${i}</a></li>`;
    } else if (
      i === page - 3 ||
      i === page + 3
    ) {
      html += `<li class="page-item disabled"><span class="page-link">...</span></li>`;
    }
  }
  html += `<li class="page-item ${
    page >= lastPage ? "disabled" : ""
  }"><a class="page-link" href="#" onclick="loadLinks(${
    page + 1
  })">Sau</a></li>`;
  container.innerHTML = html;
}

function startRealtimeCounter() {
  if (globalThis.cntInterval)
    clearInterval(globalThis.cntInterval);
  const updateDisplay = () => {
    document
      .querySelectorAll(".countdown")
      .forEach((el) => {
        const expireAttr = el.getAttribute(
          "data-expire"
        );
        const expireTime = new Date(
          expireAttr
        ).getTime();
        if (Number.isNaN(expireTime)) {
          el.innerHTML = "Lỗi ngày";
          return;
        }
        const now = new Date().getTime();
        const diff = expireTime - now;
        if (diff <= 0) {
          el.innerHTML = "Hết hạn";
          el.className = "badge bg-danger ms-2";
          el.classList.remove("countdown");
        } else {
          const h = Math.floor(diff / 3600000);
          const m = Math.floor(
            (diff % 3600000) / 60000
          );
          const s = Math.floor(
            (diff % 60000) / 1000
          );
          el.innerHTML = `⏱ ${h
            .toString()
            .padStart(2, "0")}h ${m
            .toString()
            .padStart(2, "0")}m ${s
            .toString()
            .padStart(2, "0")}s`;
        }
      });
  };
  updateDisplay();
  globalThis.cntInterval = setInterval(
    updateDisplay,
    1000
  );
}

function toggleExpireInput() {
  const unit =
    document.getElementById("expireUnit").value;
  const input = document.getElementById(
    "expireValue"
  );
  if (unit === "infinite") {
    input.classList.add("d-none");
    input.value = 0;
  } else {
    input.classList.remove("d-none");
    if (input.value == 0) input.value = 1;
  }
}

async function deleteLink(id) {
  if (
    !confirm(
      "Bạn có chắc chắn muốn xóa liên kết này?"
    )
  )
    return;
  try {
    const token =
      localStorage.getItem("jwt_token");
    const res = await fetch(
      `/api/links/${id}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    if (!res.ok)
      alert("Không thể xóa liên kết");
  } catch (err) {
    console.error(err);
  }
}

function editLink(id, oldUrl) {
  document.getElementById("editId").value = id;
  document.getElementById("editLongUrl").value =
    oldUrl;
  document.getElementById(
    "editExpireUnit"
  ).value = "";
  toggleEditExpireInput();
  const editModal = new bootstrap.Modal(
    document.getElementById("editModal")
  );
  editModal.show();
}

function toggleEditExpireInput() {
  const unit = document.getElementById(
    "editExpireUnit"
  ).value;
  const input = document.getElementById(
    "editExpireValue"
  );
  if (
    unit === "infinite" ||
    unit === "expired" ||
    unit === ""
  ) {
    input.classList.add("d-none");
  } else {
    input.classList.remove("d-none");
  }
}

async function submitEdit() {
  const id =
    document.getElementById("editId").value;
  const newUrl = document.getElementById(
    "editLongUrl"
  ).value;
  const durationType = document.getElementById(
    "editExpireUnit"
  ).value;
  const durationValue =
    parseInt(
      document.getElementById("editExpireValue")
        .value
    ) || 0;

  if (!newUrl.startsWith("http")) {
    alert("URL không hợp lệ");
    return;
  }

  try {
    const res = await fetch(
      `/api/links/${id}`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem(
            "jwt_token"
          )}`,
        },
        body: JSON.stringify({
          original_url: newUrl,
          duration_type: durationType,
          duration_value: durationValue,
        }),
      }
    );

    if (res.ok) {
      const modalElement =
        document.getElementById("editModal");
      const modalInstance =
        bootstrap.Modal.getInstance(
          modalElement
        );
      modalInstance.hide();
    } else {
      const data = await res.json();
      alert(data.error || "Cập nhật thất bại");
    }
  } catch (err) {
    console.error("Lỗi cập nhật:", err);
  }
}

function copyToClipboard(text) {
  navigator.clipboard
    .writeText(text)
    .then(() => {
      alert("Đã copy link thành công: " + text);
    })
    .catch((err) => {
      console.error("Lỗi copy: ", err);
    });
}

async function downloadQR(url, code) {
  const response = await fetch(url);
  const blob = await response.blob();
  const downloadUrl =
    window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = downloadUrl;
  link.download = `QR_${code}.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function viewLargeQR(url) {
  window.open(
    url,
    "_blank",
    "width=300,height=300"
  );
}

// --- Auth Functions ---

async function handleAuth(type) {
  const user =
    document.getElementById("username").value;
  const pass =
    document.getElementById("password").value;
  const retype = document.getElementById(
    "retype_password"
  )?.value;

  // quick check
  if (type === "register" && pass !== retype) {
    alert("Mật khẩu nhập lại không khớp!");
    return;
  }

  const payload = {
    username: user,
    password: pass,
  };

  // if register send retype for backend
  if (type === "register") {
    payload.retype_password = retype;
  }

  const response = await fetch(`/${type}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const data = await response.json();

  if (response.ok) {
    if (type === "login") {
      localStorage.setItem(
        "jwt_token",
        data.token
      );
      localStorage.setItem("username", user);
      localStorage.setItem(
        "user_role",
        data.role
      ); // save role
      alert("Đăng nhập thành công!");
      location.reload();
    } else {
      alert(
        "Đăng ký thành công! Hãy đăng nhập."
      );
      toggleAuthMode(); // toggle to login
    }
  } else {
    alert(data.error || "Lỗi xác thực");
  }
}

function logout() {
  localStorage.clear();
  location.reload();
}

function checkAuth() {
  // check admin
  const role =
    localStorage.getItem("user_role");
  if (role === "admin") {
    document
      .getElementById("adminBtn")
      .classList.remove("d-none");
  }
  const token =
    localStorage.getItem("jwt_token");
  const user = localStorage.getItem("username");
  if (token) {
    document
      .getElementById("loginBtn")
      .classList.add("d-none");
    document
      .getElementById("logoutBtn")
      .classList.remove("d-none");
    document.getElementById(
      "userInfo"
    ).innerText = `Hello, ${user}`;
    // show admin btn
    const adminBtn =
      document.getElementById("adminBtn");
    if (role === "admin" && adminBtn) {
      adminBtn.classList.remove("d-none");
    }
    return true;
  } else {
    document
      .getElementById("loginBtn")
      .classList.remove("d-none");
    document
      .getElementById("logoutBtn")
      .classList.add("d-none");
    document.getElementById(
      "userInfo"
    ).innerText = "";
    return false;
  }
}

// --- WebSocket Function ---

function connectWebSocket() {
  const token =
    localStorage.getItem("jwt_token");
  if (!token || token === "null") {
    console.warn(
      "WS: Chưa có token, bỏ qua kết nối."
    );
    return;
  }

  if (socket) socket.close();

  console.log(
    "Đang thử kết nối WS với token:",
    token.substring(0, 20) + "..."
  );

  socket = new WebSocket(
    `ws://${window.location.host}/ws?token=${token}`
  );

  socket.onopen = function () {
    console.log(
      "✅ WebSocket đã kết nối thành công!"
    );
  };

  socket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    if (data.action === "update_links")
      loadLinks(currentPage);
  };

  socket.onclose = function (e) {
    console.log(
      "❌ WebSocket đóng. Code:",
      e.code,
      "Lý do:",
      e.reason
    );
    if (localStorage.getItem("jwt_token")) {
      setTimeout(connectWebSocket, 5000);
    }
  };

  socket.onerror = function (err) {
    console.error("🔥 WebSocket Error:", err);
  };
}

// --- On Load ---
window.onload = function () {
  if (checkAuth()) {
    loadLinks(1);
    connectWebSocket();
  } else {
    console.log(
      "User chưa đăng nhập, đợi action từ người dùng."
    );
    const list =
      document.getElementById("linkList");
    if (list) {
      list.innerHTML = `<tr><td colspan="5" class="text-center text-muted">Vui lòng đăng nhập để sử dụng dịch vụ</td></tr>`;
    }
  }
};

// auth mode
let isRegisterMode = false;

function toggleAuthMode() {
  isRegisterMode = !isRegisterMode;
  const retypeSection = document.getElementById(
    "retypeSection"
  );
  const authSubmitBtn = document.getElementById(
    "authSubmitBtn"
  );
  const toggleLink = document.getElementById(
    "toggleAuthLink"
  );

  if (isRegisterMode) {
    retypeSection.classList.remove("d-none");
    authSubmitBtn.innerText = "Register Now";
    authSubmitBtn.setAttribute(
      "onclick",
      "handleAuth('register')"
    );
    toggleLink.innerText =
      "Đã có tài khoản? Đăng nhập";
  } else {
    retypeSection.classList.add("d-none");
    authSubmitBtn.innerText = "Login";
    authSubmitBtn.setAttribute(
      "onclick",
      "handleAuth('login')"
    );
    toggleLink.innerText =
      "Chưa có tài khoản? Đăng ký ngay";
  }
}
