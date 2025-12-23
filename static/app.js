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
    // TRIGGER: Shake animation instead of simple alert
    LinkAnimate.shakeElement("longUrl");
    longUrlInput.focus();
    return;
  }

  if (
    !urlValue.startsWith("http://") &&
    !urlValue.startsWith("https://")
  ) {
    showNotify(
      "Định dạng không đúng",
      "URL phải bắt đầu bằng http:// hoặc https://",
      "error"
    );
    return;
  }

  const durationType =
    document.getElementById("expireUnit").value;
  const durationValue =
    Number.parseInt(
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

    const result = await res.json();
    // // DEBUG
    // console.log(
    //   "Cấu trúc JSON nhận được:",
    //   result
    // );

    if (res.ok) {
      LinkAnimate.animateSuccess();
      loadLinks();

      // Get direct link from server
      const finalLink =
        result.short_url ||
        (result.data && result.data.short_url);

      if (finalLink) {
        // Push notification
        showNotify(
          "Tạo link thành công!",
          "Bạn có thể sử dụng link rút gọn dưới đây:",
          "success",
          finalLink
        );
      } else {
        showNotify(
          "Thành công",
          "Link đã được tạo, hãy xem ở danh sách bên dưới.",
          "success"
        );
      }

      longUrlInput.value = "";
      document.getElementById(
        "customAlias"
      ).value = "";
    } else {
      // TRIGGER: Shake on error
      LinkAnimate.shakeElement("longUrl");
      showNotify(
        "Lỗi tạo link",
        data.error ||
          "Vui lòng kiểm tra lại URL",
        "error"
      );
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

    // Check if data null
    if (!data || data.length === 0) {
      list.innerHTML = `<tr><td colspan="6" class="text-center text-muted">Chưa có link nào được tạo</td></tr>`;
      updatePagination(0, 1, 1);
      return;
    }

    // Clear existing list
    list.innerHTML = "";

    // Render rows
    data.forEach((l) => {
      const rowId = `row-${l.id}`;
      const shortFullUrl =
        globalThis.location.origin +
        "/" +
        l.short_code;
      const qrImageUrl = `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${shortFullUrl}`;

      const row = document.createElement("tr");
      row.id = rowId;

      // FIX: Added 'this' to copyToClipboard to enable button bounce animation
      row.innerHTML = `
                <td>
                    <a href="/${
                      l.short_code
                    }" target="_blank" class="fw-bold text-decoration-none">${
        l.short_code
      }</a>
                </td>
                 <td>
                    <button class="btn btn-sm btn-light border shadow-sm" title="Copy link" onclick="copyToClipboard('${shortFullUrl}', this)">
                        📋 Copy
                    </button>
                </td>
                
                <td class="text-truncate" style="max-width: 200px;" title="${
                  l.original_url
                }">
                    ${l.original_url}
                </td>
                
                <td>
                    <div>
                        <span class="fw-bold fs-5 click-count">${
                          l.click_count
                        }</span> clicks 
                        <span class="timer-area">${renderTimerHtml(
                          l
                        )}</span>
                    </div>
                    <div class="analytics-area">${renderAnalyticsHtml(
                      l
                    )}</div>
                </td>
                
                <td>
                    <div class="btn-group">
            <button class="btn btn-sm btn-outline-info" onclick="LinkAnimate.bounceButton(this); viewLargeQR('${qrImageUrl}')">🔍 Xem</button>
            <button class="btn btn-sm btn-outline-primary" onclick="LinkAnimate.bounceButton(this); downloadQR('${qrImageUrl}', '${
        l.short_code
      }')">💾 Lưu</button>
        </div>
                </td>
                
                <td>
                    <div class="btn-group btn-group-sm">
            <button class="btn btn-warning" onclick="LinkAnimate.bounceButton(this); editLink(${
              l.id
            }, '${
        l.original_url
      }')">✏️ Sửa</button>
            <button class="btn btn-danger" onclick="LinkAnimate.bounceButton(this); deleteLink(${
              l.id
            })">🗑️ Xóa</button>
        </div>
                </td>
            `;
      list.appendChild(row);
    });

    // TRIGGER: Staggered entrance animation for all table rows
    LinkAnimate.animateTableRows();

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

// FIX: Update copy function signature to receive the element
function copyToClipboard(text, btnElement) {
  // Check if button is disabled
  if (
    !btnElement ||
    btnElement.classList.contains("disabled") ||
    btnElement.disabled
  )
    return;

  // animate button
  LinkAnimate.bounceButton(btnElement);

  // Disable button
  btnElement.disabled = true;
  btnElement.style.pointerEvents = "none"; // block cursor

  navigator.clipboard
    .writeText(text)
    .then(() => {
      // display success
      const originalHtml = btnElement.innerHTML;
      btnElement.innerHTML = "✅ Copied";

      // change button color is lock
      btnElement.classList.replace(
        "btn-light",
        "btn-success"
      );

      // return original btn
      setTimeout(() => {
        btnElement.innerHTML = originalHtml;
        btnElement.classList.replace(
          "btn-success",
          "btn-light"
        );

        // Unlock button
        btnElement.disabled = false;
        btnElement.style.pointerEvents = "auto";
      }, 1500);
    })
    .catch((err) => {
      btnElement.disabled = false;
      btnElement.style.pointerEvents = "auto";
      console.error("Lỗi copy:", err);
    });
}

// Helper functions
function renderTimerHtml(l) {
  // check expired_at
  if (!l.expired_at) {
    return `<span class="badge bg-light text-muted ms-2">∞</span>`;
  }

  // calc expire time
  const expireTime = new Date(
    l.expired_at
  ).getTime();
  const now = new Date().getTime();
  const isExpired = expireTime - now <= 0;

  if (isExpired) {
    return `<span class="badge bg-danger ms-2">Hết hạn</span>`;
  }

  // return expire data from api
  return `<span class="badge bg-warning text-dark ms-2 countdown" data-expire="${l.expired_at}">⏱ Đang tính...</span>`;
}

function renderAnalyticsHtml(l) {
  if (!l.last_browser && !l.last_os) return "";
  return `
        <div style="font-size: 0.7rem;" class="text-muted mt-1">
            <span class="badge border text-dark fw-normal">🌐 ${
              l.last_browser || "N/A"
            }</span>
            <span class="badge border text-dark fw-normal">💻 ${
              l.last_os || "N/A"
            }</span>
        </div>`;
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

let deleteTargetId = null;

async function deleteLink(id) {
  deleteTargetId = id; // Save ID
  const confirmModal = new bootstrap.Modal(
    document.getElementById(
      "confirmDeleteModal"
    )
  );
  confirmModal.show();

  // Assign an event to the "Delete Now" button in the modal.
  document.getElementById(
    "btnConfirmDelete"
  ).onclick = async function () {
    confirmModal.hide();
    executeDelete();
  };
}

async function executeDelete() {
  try {
    const token =
      localStorage.getItem("jwt_token");
    const res = await fetch(
      `/api/links/${deleteTargetId}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );

    if (res.ok) {
      showNotify(
        "Đã xóa",
        "Liên kết đã được gỡ bỏ."
      );
      loadLinks(currentPage);
    } else {
      showNotify(
        "Lỗi",
        "Không thể xóa liên kết này",
        "error"
      );
    }
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
    Number.parseInt(
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
      // reload links
      showNotify(
        "Cập nhật thành công",
        "Thông tin liên kết đã được thay đổi."
      ); // notify
      await loadLinks(currentPage);

      // update counter
      startRealtimeCounter();
    } else {
      const data = await res.json();
      showNotify(
        "Lỗi cập nhật",
        "Không thể lưu thay đổi",
        "error"
      );
    }
  } catch (err) {
    console.error("Lỗi cập nhật:", err);
  }
}

async function downloadQR(url, code) {
  const response = await fetch(url);
  const blob = await response.blob();
  const downloadUrl =
    globalThis.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = downloadUrl;
  link.download = `QR_${code}.png`;
  document.body.appendChild(link);
  link.click();
  childNode.remove(link);
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
      showNotify(
        "Thành công!",
        "Đang chuyển hướng..."
      );
      setTimeout(() => location.reload(), 1500); // Delay 1.5s and reload
    } else {
      showNotify(
        "Tuyệt vời!",
        "Đăng ký thành công! Hãy đăng nhập ngay."
      );
      toggleAuthMode(); // toggle to login
    }
  } else {
    showNotify(
      "Lỗi xác thực",
      data.error ||
        "Username hoặc mật khẩu không đúng",
      "error"
    );
  }
}

// Change: logout -> logoutConfirm
function logout() {
  const logoutModal = new bootstrap.Modal(
    document.getElementById(
      "logoutConfirmModal"
    )
  );
  logoutModal.show();

  document.getElementById(
    "btnConfirmLogout"
  ).onclick = function () {
    localStorage.clear();
    location.reload();
  };
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

  // // Debug
  // console.log(
  //   "Đang thử kết nối WS với token:",
  //   token.substring(0, 20) + "..."
  // );

  // Fix: get websocket work on deploy
  const protocol =
    globalThis.location.protocol === "https:"
      ? "wss://"
      : "ws://";
  socket = new WebSocket(
    `${protocol}${globalThis.location.host}/ws?token=${token}`
  );

  socket.onopen = function () {
    console.log(
      "✅ Người dùng đã kết nối thành công!"
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
      "❌ User chưa đăng nhập, xin vui lòng đăng nhập"
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

// Cleanup expired links
async function cleanupExpiredLinks() {
  const cleanupModal = new bootstrap.Modal(
    document.getElementById(
      "cleanupConfirmModal"
    )
  );
  cleanupModal.show();

  document.getElementById(
    "btnConfirmCleanup"
  ).onclick = async function () {
    cleanupModal.hide();
    try {
      const token =
        localStorage.getItem("jwt_token");
      const res = await fetch(
        "/api/links/cleanup",
        {
          // endpoint
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (res.ok) {
        const result = await res.json();
        showNotify(
          "Đã dọn dẹp",
          `Đã xóa thành công ${
            result.deleted_count || 0
          } link hết hạn.`
        );
        loadLinks(1);
      } else {
        showNotify(
          "Lỗi",
          "Không thể dọn dẹp vào lúc này.",
          "error"
        );
      }
    } catch (err) {
      console.error("Lỗi Cleanup:", err);
    }
  };
}

// Modal notify
function showNotify(
  title,
  message,
  type = "success",
  shortUrl = null
) {
  const modalElement = document.getElementById(
    "notificationModal"
  );
  const modal = new bootstrap.Modal(
    modalElement
  );

  const iconEl =
    document.getElementById("notifyIcon");
  const titleEl = document.getElementById(
    "notifyTitle"
  );
  const msgEl = document.getElementById(
    "notifyMessage"
  );
  const linkArea = document.getElementById(
    "successLinkArea"
  ); // ID match HTML
  const inputLink = document.getElementById(
    "resultShortUrl"
  ); // ID match HTML

  titleEl.innerText = title;
  msgEl.innerText = message;

  // Hide link area
  linkArea.classList.add("d-none");

  if (type === "success") {
    iconEl.innerHTML = "✅";
    if (shortUrl) {
      linkArea.classList.remove("d-none");
      inputLink.value = shortUrl;

      // Auto select
      setTimeout(() => inputLink.select(), 200);
    }
  } else {
    iconEl.innerHTML = "❌";
    linkArea.classList.add("d-none");
  }
  modal.show();
}

// Copy Link from modal
function copyFromNotify() {
  const copyText = document.getElementById(
    "resultShortUrl"
  );
  const btn = document.getElementById(
    "btnCopyNotify"
  );

  if (!copyText) return;

  navigator.clipboard
    .writeText(copyText.value)
    .then(() => {
      // Change button
      const originalText = btn.innerHTML;
      btn.innerHTML = "✅ Đã chép!";
      btn.classList.replace(
        "btn-primary",
        "btn-success"
      );

      // Delay 2s
      setTimeout(() => {
        btn.innerHTML = originalText;
        btn.classList.replace(
          "btn-success",
          "btn-primary"
        );
      }, 2000);
    })
    .catch((err) => {
      console.error("Lỗi khi copy: ", err);
      // Optional old browser support
      copyText.select();
      document.execCommand("copy");
    });
}

// Modal Cleanup
function showCleanupModal() {
  const modalElement = document.getElementById(
    "cleanupConfirmModal"
  );
  const modal = new bootstrap.Modal(
    modalElement
  );
  modal.show();

  // get Event
  const confirmBtn = document.getElementById(
    "btnConfirmCleanup"
  );
  confirmBtn.onclick = async function () {
    modal.hide(); // Close modal
    await executeCleanup();
  };
}

// Execute Cleanup
async function executeCleanup() {
  const token =
    localStorage.getItem("jwt_token"); // check token from localStorage

  if (!token) {
    showNotify(
      "Lỗi",
      "Phiên làm việc hết hạn, vui lòng đăng nhập lại",
      "error"
    );
    return;
  }

  try {
    const response = await fetch(
      "/api/links/cleanup",
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      }
    );

    const result = await response.json();

    if (response.ok) {
      // Notify
      showNotify(
        "Thành công 🧹",
        `Đã dọn dẹp xong ${result.deleted_count} liên kết hết hạn.`,
        "success"
      );
      // Reload links
      if (typeof loadLinks === "function")
        loadLinks();
    } else {
      showNotify(
        "Lỗi",
        result.error || "Không thể dọn dẹp",
        "error"
      );
    }
  } catch (error) {
    console.error("Cleanup error:", error);
    showNotify(
      "Lỗi",
      "Kết nối server thất bại",
      "error"
    );
  }
}
