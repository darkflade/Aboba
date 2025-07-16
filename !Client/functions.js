// Функция для настройки кнопки скачивания документа
function setupDocDownload(btnClass, host) {
  const buttons = document.querySelectorAll(`.${btnClass}`)
  //console.log(`Найдено кнопок: ${buttons.length}`);
  buttons.forEach(btn => {
    const msg = document.getElementById(btn.dataset.msgId);
    if (!msg) return;

    btn.addEventListener('click', async () => {
      msg.textContent = '';
      btn.disabled = true;
      const originalText = btn.textContent;
      btn.textContent = 'Загрузка…';

      const deptId = btn.dataset.deptId;
      try {
        const res = await fetch(host + deptId + '/document');
        const contentType = res.headers.get('Content-Type') || '';

        if (res.ok && contentType.includes('application/json')) {
          const json = await res.json();
          throw new Error(json.error || 'Неизвестная ошибка');
        }
        if (!res.ok) {
          const txt = await res.text();
          throw new Error(txt || `Ошибка ${res.status}`);
        }

        const blob = await res.blob();
        const filename = `department_${deptId}_employees.docx`;
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        a.remove();
        window.URL.revokeObjectURL(url);

      } catch (e) {
        msg.textContent = e.message;
      } finally {
        btn.disabled = false;
        btn.textContent = originalText;
      }
    });
  });

}

function loadDepartments(callback) {
  fetch(`${host}/departments`)
    .then(res => res.json())
    .then(data => {
      if (!data.success) throw new Error(data.error || 'Ошибка загрузки отделов');
      departments = data.data || [];
      const deptSelect = document.getElementById('deptSelect');
      const newDept = document.getElementById('newDept');
      [deptSelect, newDept].forEach(selectEl => selectEl.innerHTML = '');
      departments.forEach(d => {
        const opt = document.createElement('option');
        opt.value = d.id;
        opt.textContent = `#${d.id} - ${d.boss_name}`;
        deptSelect.append(opt);
        newDept.append(opt.cloneNode(true));
      });
      if (callback) callback();
      //loadDepartmentList();
    })
    .catch(e => {
      document.getElementById('departmentList').innerHTML = `<li class="error">${e.message}</li>`;
    });
}

function loadEmployeesByDept() {
  const deptId = document.getElementById('deptSelect').value;
  fetch(`${host}/employeesByDepart/${deptId}`)
    .then(res => res.json())
    .then(data => {
      const tbody = document.getElementById('employeeList');
      const warning = document.getElementById('empWarning');
      tbody.innerHTML = ''; warning.textContent = '';
      if (!data.success) {
        warning.textContent = data.error || 'Ошибка';
        return;
      }
      if (!data.data.length) {
        warning.textContent = 'Нет сотрудников в этом отделе';
        return;
      }
      data.data.forEach(e => {
        const tr = document.createElement('tr');
        const imgCell = `<td><img src="${host}/employees/${e.id}/photo" alt="" width="50" onerror="this.src='no-image.png'"></td>`;
        tr.innerHTML = `
          ${imgCell}
          <td>${e.name}</td>
          <td>${e.status}</td>
          <td>${e.salary}₽</td>
          <td>
            <div class="tableBtnDiv">
              <button class="table-btn delete" onclick="deleteEmployee(${e.id})">Удалить</button>
              <button class="table-btn edit" onclick="prepChangeForm(${e.id})">Изменить</button>
            </div>
          </td>`;
        tbody.append(tr);
      });
    })
    .catch(() => {
      document.getElementById('employeeList').innerHTML = `<tr><td colspan="5" class="error">Ошибка загрузки сотрудников</td></tr>`;
    });
}

function deleteEmployee(id) {
  fetch(`${host}/employees/${id}`, { method: 'DELETE' })
    .then(() => loadEmployeesByDept())
    .catch(() => {
      document.getElementById('employeeList').innerHTML = `<tr><td colspan="5" class="error">Ошибка удаления сотудника #${id}</td></tr>`;
    });
}

function loadDepartmentList() {
  const tbody = document.getElementById('departmentList');
  tbody.innerHTML = '';
  departments.forEach(d => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>#${d.id}</td>
      <td>
        <span id="boss${d.id}">${d.boss_id.toString().padStart(3, '\u00A0')} : ${d.boss_name}</span>
        <select id="bossSelect${d.id}" class="hidden"></select>
      </td>
      <td>${d.total_salary} ₽</td>
      <td>${d.size}</td>
      <td>
        <div class="tableBtnDiv2">
          <button class="boss-btn" id="editBossBtn${d.id}">Изменить</button>
          <button id="saveBoss${d.id}" class="hidden save-boss">Сохранить</button>
          <button id="cancelBoss${d.id}" class="hidden cancel-boss">Отменить</button>
          <button class="downloadDocBtn" data-dept-id="${d.id}" data-msg-id="msg${d.id}">Скачать DOCX</button>
          <div id="msg${d.id}" class="error"></div>
        </div>
      </td>`;
    tbody.append(tr);

    const select = tr.querySelector(`#bossSelect${d.id}`);
    fetch(`${host}/employeesByDepart/${d.id}`)
      .then(res => res.json())
      .then(data => {
        if (data.success) data.data.forEach(emp => {
          if (emp.name !== d.boss_name) {
            const opt = document.createElement('option');
            opt.value = emp.id;
            opt.textContent = `${emp.id.toString().padStart(3, '\u00A0')} : ${emp.name}`;
            select.append(opt);
          }
        });
      });

    tr.querySelector(`#editBossBtn${d.id}`).addEventListener('click', () => {
      tr.querySelector(`#boss${d.id}`).classList.add('hidden');
      select.classList.remove('hidden');
      tr.querySelector(`#saveBoss${d.id}`).classList.remove('hidden');
      tr.querySelector(`#cancelBoss${d.id}`).classList.remove('hidden');
      tr.querySelector(`#editBossBtn${d.id}`).classList.add('hidden');
    });

    tr.querySelector(`#cancelBoss${d.id}`).addEventListener('click', () => {
      select.classList.add('hidden');
      tr.querySelector(`#saveBoss${d.id}`).classList.add('hidden');
      tr.querySelector(`#cancelBoss${d.id}`).classList.add('hidden');
      tr.querySelector(`#boss${d.id}`).classList.remove('hidden');
      tr.querySelector(`#editBossBtn${d.id}`).classList.remove('hidden');
    });

    tr.querySelector(`#saveBoss${d.id}`).addEventListener('click', () => {
      const newBoss = parseInt(select.value, 10);
      fetch(`${host}/departments/${d.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ boss_id: newBoss })
      })
      .then(res => res.json())
      .then(resp => {
      if (resp.success) {
        document.getElementById(`boss${d.id}`).textContent = select.options[select.selectedIndex].textContent;
        tr.querySelector(`#cancelBoss${d.id}`).click();
      } else alert(resp.error || 'Ошибка');
      });
    });
  });

  setupDocDownload('downloadDocBtn', documentHost);
}

// todo

function prepChangeForm(id) {

  editEmployeeId = id;
  const errorDiv = document.getElementById('addError');
  errorDiv.textContent = '';
  const warningImage = document.getElementById('fileName');
  warningImage.textContent = '';
  const imageBtnText = document.getElementById('customFileBtn');
  document.getElementById('addBtn').textContent = "Изменить";

  fetch(`${host}/employees/${id}`)
    .then(res => res.json())
    .then(data => {
      if (!data.success) {
        errorDiv.textContent = data.error;
        return;
      }
      const e = data.data;
      // Получаем элементы формы
      const nameInput = document.getElementById('newName');
      const statusInput = document.getElementById('newStatus');
      const salaryInput = document.getElementById('newSalary');
      const imageInput = document.getElementById('newImage');
      const deptInput = document.getElementById('newDept');

      // Заполняем поля
      nameInput.value = e.name;
      statusInput.value = e.status;
      salaryInput.value = e.salary;
      deptInput.value = e.dept_id;

      if (!e.image_url) {
        warningImage.classList.add('warning')
        warningImage.textContent = "Нет изображения у данного сотрудника"
      } else {
        warningImage.classList.remove('warning')
        imageBtnText.textContent = "Изменить фото"
      }


      // Блокируем все поля кроме картинки и зарплаты
      nameInput.disabled = true;
      statusInput.disabled = true;
      deptInput.disabled = true;
      salaryInput.disabled = false;
      imageInput.disabled = false;
      showSection("add")

      // После изменения и отправки
    });
}

function addEmployee() {
  const name = document.getElementById('newName').value;
  const status = document.getElementById('newStatus').value;
  const salary = parseFloat(document.getElementById('newSalary').value);
  const dept_id = parseInt(document.getElementById('newDept').value);
  const image = document.getElementById('newImage').files[0];
  const errorDiv = document.getElementById('addError');
  errorDiv.textContent = '';

  const formData = new FormData();
  formData.append('name', name);
  formData.append('status', status);
  formData.append('salary', salary);
  formData.append('dept_id', dept_id);
  if (image) formData.append('image', image);

  fetch(`${host}/employees`, {
    method: 'POST',
    body: formData
  })
  .then(res => res.json())
  .then(data => {
    if (!data.success) {
      errorDiv.textContent = data.error || 'Ошибка добавления';
      return;
    }
    document.getElementById('addForm').reset();
    showSection('employees');
    loadEmployeesByDept();
  })
  .catch(e => {
    errorDiv.textContent = e.message;
  });
}

function updateEmployee(id) {
  const name = document.getElementById('newName').value;
  const status = document.getElementById('newStatus').value;
  const salaryInput = document.getElementById('newSalary').value;
  const salary = parseFloat(salaryInput);
  const dept_id = parseInt(document.getElementById('newDept').value);
  const imageInput = document.getElementById('newImage');
  const imageFile = imageInput.files[0] || null;
  const errorDiv = document.getElementById('addError');
  errorDiv.textContent = '';

  const formData = new FormData();
  formData.append('name', name)
  formData.append('status', status);
  formData.append('salary', salary);
  formData.append('dept_id', dept_id);
  if (imageFile) {
    formData.append('image', imageFile);
  } else {
    formData.append('image', null);
  }

  salaryInput.disabled = true;
  imageInput.disabled = true;

  fetch(`${host}/employees/${id}`, {
    method: 'PUT',
    body: formData
  })
    .then(res => res.json())
    .then(resp => {
      if (!resp.success) {
        errorDiv.textContent = resp.error || 'Ошибка обновления';
      } else {
        showSection('employees');
        loadEmployeesByDept();
        editEmployeeId = null; // сбросить флаг
      }
    })
    .catch(() => {
      errorDiv.textContent = 'err';
    })
    .finally(() => {
      salaryInput.disabled = false;
      imageInput.disabled = false;
      window.location.reload(true)
    });
}