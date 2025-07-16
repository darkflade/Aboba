const host = 'https://server.go/api';
//const documentHost = 'https://cdn.experimental.me/api/documents/word.php?id=';
const documentHost = 'https://server.go/api/employeesByDepart/';
let departments = [];
let editEmployeeId = null;

function showSection(section) {
  ['employees', 'departments', 'add'].forEach(s => {
    document.getElementById(s + 'Section').classList.toggle('hidden', s !== section);
  });
}

function clearAddForm() {
  const nameInput = document.getElementById('newName');
  const statusInput = document.getElementById('newStatus');
  const salaryInput = document.getElementById('newSalary');
  const imageInput = document.getElementById('newImage');
  const deptInput = document.getElementById('newDept');
  nameInput.value = "";
  statusInput.value = "";
  salaryInput.value = "";
  imageInput.value = "";
  nameInput.disabled = false;
  statusInput.disabled = false;
  deptInput.disabled = false;
  salaryInput.disabled = false;
  imageInput.disabled = false;

  document.getElementById('addError').textContent = "";
  const warningImage = document.getElementById('fileName');
  warningImage.textContent = "Файл не выбран";
  warningImage.classList.remove('warning')
  document.getElementById('customFileBtn').textContent = "Загрузить фото"

  document.getElementById('addBtn').textContent = "Добавить";

  editEmployeeId = null;
}

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
  document.addEventListener('paste', async (event) => {
    if (!navigator.clipboard || !navigator.clipboard.read) {
      alert("Epta");
      return;
    }

    try {
      const clipboardItems = await navigator.clipboard.read();
      pendingImageBlob = null;

      for (const item of clipboardItems) {
        for (const type of item.types) {
          if (type.startsWith('image/')) {
            pendingImageBlob = await item.getType(type);
            break;
          }
        }
        if (pendingImageBlob) break;
      }

      if (!pendingImageBlob) {
        alert("No bitches");
        //errorDiv.textContent = 'В буфере обмена нет изображения';
        return;
      }

      // Показываем предпросмотр в модальном окне
      const imageUrl = URL.createObjectURL(pendingImageBlob);
      modalPreview = document.getElementById("clipboardimg");
      modalPreview.style.display = "block";
      modalPreview.src = imageUrl;
      modalPreview.style.display = 'block';
      const file = pendingImageBlob.type.includes('png') ? 'image.png' : 'image.jpg';
      const fileObj = new File([pendingImageBlob], file, { type: pendingImageBlob.type });
      const dataTransfer = new DataTransfer();
      dataTransfer.items.add(fileObj);
      document.getElementById("newImage").files = dataTransfer.files;
      document.getElementById("fileName").textContent = file;
    } catch (err) {
      console.error('Ошибка:', err);
      alert("pizdec");
    }
  });

  // Вешаем обработчики на кнопки навигации
  document.querySelectorAll('.nav button').forEach(btn => {
    btn.addEventListener('click', () => {
      const action = btn.dataset.action;
      showSection(action);
      if (action === 'employees') {
        loadDepartments(() => {
          loadEmployeesByDept();
        });
      }
      if (action === 'departments') loadDepartmentList();
      if (action === 'add')         clearAddForm();
    });
  });

  document.getElementById('deptSelect').addEventListener('change', loadEmployeesByDept);


  // Обработчик кнопки добавления
  document.getElementById('addForm').addEventListener('submit', function(e) {
    e.preventDefault();
    if  (editEmployeeId) {
      updateEmployee(editEmployeeId);
    } else {
      addEmployee();
    }
  });
  // Обработчик кнопки cancel
  document.getElementById('addBtnCancel').addEventListener('click', () => {
    clearAddForm();
  });

  // Стартовая загрузка
  loadDepartments(() => {
    showSection('employees');
    loadEmployeesByDept();
  });

  document.getElementById('customFileBtn').onclick = function() {
    document.getElementById('fileName').classList.remove('warning');
    document.getElementById('newImage').click();
  };
  document.getElementById('newImage').onchange = function() {
    const fileName = this.files[0] ? this.files[0].name : 'Файл не выбран';
    document.getElementById('fileName').textContent = fileName;
  };

});
