package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Структура для стандартных ответов API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// Функция для JSON ответов
func sendAPIResponse(c *gin.Context, data interface{}, err error, code int) {
	resp := APIResponse{
		Success: err == nil,
		Data:    data,
		Code:    code,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	c.JSON(code, resp)
}

// Получить список отделов
func getDepartments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		const q = `
        SELECT
            d.ОТД_НОМЕР,
            d.ОТД_РУК,
            e.СЛУ_ИМЯ,
            d.ОТД_СОТР_ЗАРП,
            d.ОТД_РАЗМ
        FROM ОТДЕЛЫ d
        LEFT JOIN СЛУЖАЩИЕ e
          	ON d.ОТД_РУК = e.СЛУ_НОМЕР`
		rows, err := db.Query(q)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка при получении отделов: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var deps []Department
		for rows.Next() {
			var (
				d      Department
				bossId sql.NullInt64
				name   sql.NullString
			)
			if err := rows.Scan(&d.ID, &bossId, &name, &d.TotalSalary, &d.Size); err != nil {
				sendAPIResponse(c, nil, fmt.Errorf("ошибка парсинга данных: %v", err), http.StatusInternalServerError)
				return
			}

			// Устанавливаем имя руководителя
			if bossId.Valid {
				id := int(bossId.Int64)
				d.BossID = &id
			} else {
				id := 0
				d.BossID = &id // 0 означает, что руководитель не установлен
			}
			// BossName: если есть имя — ставим его, иначе "Не установлен"
			if name.Valid && name.String != "" {
				d.BossName = &name.String
			} else {
				placeholder := "Не установлен"
				d.BossName = &placeholder
			}

			deps = append(deps, d)
		}

		sendAPIResponse(c, deps, nil, http.StatusOK)
	}
}

// Получить список сотрудников
func getEmployees(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT СЛУ_НОМЕР, СЛУ_ИМЯ, СЛУ_СТАТ, СЛУ_ЗАРП, СЛУ_ОТД_НОМЕР FROM СЛУЖАЩИЕ")
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка при получении сотрудников: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var emps []Employee
		for rows.Next() {
			var e Employee
			if err := rows.Scan(&e.ID, &e.Name, &e.Status, &e.Salary, &e.DeptID); err != nil {
				sendAPIResponse(c, nil, fmt.Errorf("ошибка парсинга данных: %v", err), http.StatusInternalServerError)
				return
			}
			emps = append(emps, e)
		}

		sendAPIResponse(c, emps, nil, http.StatusOK)
	}
}

// Получить сотрудника по ID
func getEmployee(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный ID сотрудника"), http.StatusBadRequest)
			return
		}

		var e Employee
		row := db.QueryRow("SELECT СЛУ_НОМЕР, СЛУ_ИМЯ, СЛУ_СТАТ, СЛУ_ЗАРП, СЛУ_ОТД_НОМЕР, IMAGE_URL FROM СЛУЖАЩИЕ WHERE СЛУ_НОМЕР = ?", id)
		if err := row.Scan(&e.ID, &e.Name, &e.Status, &e.Salary, &e.DeptID, &e.ImageURL); err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("сотрудник не найден: %v", err), http.StatusNotFound)
			return
		}

		sendAPIResponse(c, e, nil, http.StatusOK)
	}
}

// Получить сотрудника по ID отдела
func getEmployeeByDepartment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		DeptID := c.Param("id")
		if DeptID == "" {
			// Получаем первый существующий ID отдела
			row := db.QueryRow("SELECT ОТД_НОМЕР FROM ОТДЕЛЫ ORDER BY ОТД_НОМЕР LIMIT 1")
			var firstDeptID int
			if err := row.Scan(&firstDeptID); err != nil {
				//showError(c, "Не удалось получить отделы")
				return
			}
			DeptID = strconv.Itoa(firstDeptID)
		}

		rows, err := db.Query("SELECT СЛУ_НОМЕР, СЛУ_ИМЯ, СЛУ_СТАТ, СЛУ_ЗАРП, СЛУ_ОТД_НОМЕР FROM СЛУЖАЩИЕ WHERE СЛУ_ОТД_НОМЕР = ?", DeptID)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка при получении данных о сотрудниках по данному отделу: %v", err), http.StatusNotFound)
			return
		}
		defer rows.Close()

		var emps []Employee
		for rows.Next() {
			var e Employee
			if err := rows.Scan(&e.ID, &e.Name, &e.Status, &e.Salary, &e.DeptID); err != nil {
				//showError(c, fmt.Sprintf("Ошибка парсинга данных сотрудника: %v", err))
				return
			}
			emps = append(emps, e)
		}

		if len(emps) == 0 {
			sendAPIResponse(c, nil, fmt.Errorf("нет сотрудников в отделе с ID %s", DeptID), http.StatusNotFound)
			return
		}

		sendAPIResponse(c, emps, nil, http.StatusOK)
	}
}

// Добавить нового сотрудника
func createEmployeeAPI(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		name := c.PostForm("name")
		status := c.PostForm("status")
		salaryStr := c.PostForm("salary")
		deptIDStr := c.PostForm("dept_id")
		photoHeader, _ := c.FormFile("image")

		salary, err := strconv.ParseFloat(salaryStr, 64)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректная зарплата: %v", err), http.StatusBadRequest)
			return
		}
		deptID, err := strconv.Atoi(deptIDStr)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный dept_id: %v", err), http.StatusBadRequest)
			return
		}
		if name == "" || status == "" {
			sendAPIResponse(c, nil, fmt.Errorf("недостаточно данных"), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка транзакции: %v", err), http.StatusInternalServerError)
			return
		}

		// Вставка сотрудника
		res, err := tx.Exec(
			`INSERT INTO СЛУЖАЩИЕ (СЛУ_ИМЯ, СЛУ_СТАТ, СЛУ_ЗАРП, СЛУ_ОТД_НОМЕР)
       VALUES (?, ?, ?, ?)`,
			name, status, salary, deptID,
		)
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка создания сотрудника: %v", err), http.StatusInternalServerError)
			return
		}
		lastID, _ := res.LastInsertId()
		empID := int(lastID)

		// Если есть изображение, сохраняем его
		var filename *string
		if photoHeader != nil {
			fn, err := saveUploadedPhoto(photoHeader, empID)
			if err != nil {
				// файл не сохранился, но сам сотрудник уже в БД — просто логируем
				log.Printf("warning: не удалось сохранить фото для %d: %v", empID, err)
			} else {
				filename = &fn
				if _, err := tx.Exec(
					"UPDATE СЛУЖАЩИЕ SET IMAGE_URL = ? WHERE СЛУ_НОМЕР = ?",
					fn, empID,
				); err != nil {
					log.Printf("warning: не удалось обновить photo path для %d: %v", empID, err)
				}
			}
		}

		// Обновление данных отдела
		_, err = tx.Exec(
			`UPDATE ОТДЕЛЫ
                 SET ОТД_РАЗМ = ОТД_РАЗМ + 1,
                 ОТД_СОТР_ЗАРП = ОТД_СОТР_ЗАРП + ?
               WHERE ОТД_НОМЕР = ?`,
			salary, deptID,
		)
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка обновления отдела: %v", err), http.StatusInternalServerError)
			return
		}

		tx.Commit()
		sendAPIResponse(c, map[string]interface{}{
			"id":        lastID,
			"image_url": filename,
		}, nil, http.StatusCreated)
	}
}

// Обновить данные сотрудника
func updateEmployeeAPI(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		empID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный ID сотрудника"), http.StatusBadRequest)
			return
		}

		name := c.PostForm("name")
		status := c.PostForm("status")
		salaryStr := c.PostForm("salary")
		deptIDStr := c.PostForm("dept_id")
		if name == "" || status == "" || salaryStr == "" || deptIDStr == "" {
			sendAPIResponse(c, nil, fmt.Errorf("недостаточно данных"), http.StatusBadRequest)
			return
		}
		newSalary, err := strconv.ParseFloat(salaryStr, 64)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректная зарплата"), http.StatusBadRequest)
			return
		}
		newDeptID, err := strconv.Atoi(deptIDStr)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный dept_id"), http.StatusBadRequest)
			return
		}

		photoHeader, _ := c.FormFile("image")

		// Сначала узнаём старую зарплату и отдел
		var oldSalary float64
		var oldDeptID int
		var oldPhoto sql.NullString
		row := db.QueryRow(`
            SELECT СЛУ_ЗАРП, СЛУ_ОТД_НОМЕР, IMAGE_URL
            FROM СЛУЖАЩИЕ WHERE СЛУ_НОМЕР = ?`, empID)
		if err := row.Scan(&oldSalary, &oldDeptID, &oldPhoto); err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("сотрудник не найден: %v", err), http.StatusNotFound)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка транзакции: %v", err), http.StatusInternalServerError)
			return
		}

		// Обновляем  данные сотрудника
		_, err = tx.Exec(`
            UPDATE СЛУЖАЩИЕ
            SET СЛУ_ИМЯ     = ?,
                СЛУ_СТАТ    = ?,
                СЛУ_ЗАРП    = ?,
                СЛУ_ОТД_НОМЕР = ?
            WHERE СЛУ_НОМЕР = ?`,
			name, status, newSalary, newDeptID, empID,
		)
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка обновления сотрудника: %v", err), http.StatusInternalServerError)
			return
		}

		// Обновляем сумму зарплат отдела
		deltaSalary := newSalary - oldSalary
		if oldDeptID == newDeptID {
			// те же отдел: меняем только сумму зарплат
			_, err = tx.Exec(`
                UPDATE ОТДЕЛЫ
                SET ОТД_СОТР_ЗАРП = ОТД_СОТР_ЗАРП + ?
                WHERE ОТД_НОМЕР = ?`,
				deltaSalary, oldDeptID,
			)
		} else {
			// минус из старого
			_, err = tx.Exec(`
                UPDATE ОТДЕЛЫ
                SET ОТД_РАЗМ       = ОТД_РАЗМ - 1,
                    ОТД_СОТР_ЗАРП = ОТД_СОТР_ЗАРП - ?
                WHERE ОТД_НОМЕР = ?`,
				oldSalary, oldDeptID,
			)
			if err == nil {
				// плюс в новый
				_, err = tx.Exec(`
                    UPDATE ОТДЕЛЫ
                    SET ОТД_РАЗМ       = ОТД_РАЗМ + 1,
                        ОТД_СОТР_ЗАРП = ОТД_СОТР_ЗАРП + ?
                    WHERE ОТД_НОМЕР = ?`,
					newSalary, newDeptID,
				)
			}
		}
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка обновления отдела: %v", err), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка фиксации транзакции: %v", err), http.StatusInternalServerError)
			return
		}

		if photoHeader != nil {
			// удаляем старое
			if oldPhoto.Valid {
				if err := deletePhoto(oldPhoto.String); err != nil {
					sendAPIResponse(c, nil, fmt.Errorf("warning: не удалось удалить старое фото для %d: %v", empID, err), 400)
					log.Printf("warning: не удалось удалить старое фото для %d: %v", empID, err)
				}
			}
			// сохраняем новое
			fn, err := saveUploadedPhoto(photoHeader, empID)
			if err != nil {
				sendAPIResponse(c, nil, fmt.Errorf("warning: не удалось сохранить новое фото для %d: %v", empID, err), 400)
				log.Printf("warning: не удалось сохранить новое фото для %d: %v", empID, err)
			} else {
				if _, err := db.Exec(
					"UPDATE СЛУЖАЩИЕ SET IMAGE_URL = ? WHERE СЛУ_НОМЕР = ?",
					fn, empID,
				); err != nil {
					sendAPIResponse(c, nil, fmt.Errorf("warning: не удалось обновить PHOTO_PATH для %d: %v", empID, err), 400)
					log.Printf("warning: не удалось обновить PHOTO_PATH для %d: %v", empID, err)
				}
			}
		}

		sendAPIResponse(c, nil, nil, http.StatusOK)
	}
}

// Удалить сотрудника
func deleteEmployeeAPI(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный ID сотрудника"), http.StatusBadRequest)
			return
		}

		// Нам нужно знать, из какого отдела и с какой зарплатой удаляем
		var deptID int
		var salary float64
		var imageURL sql.NullString
		row := db.QueryRow(`
            SELECT СЛУ_ОТД_НОМЕР, СЛУ_ЗАРП, IMAGE_URL
            FROM СЛУЖАЩИЕ WHERE СЛУ_НОМЕР = ?`, id)
		if err := row.Scan(&deptID, &salary, &imageURL); err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("сотрудник не найден: %v", err), http.StatusNotFound)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка транзакции: %v", err), http.StatusInternalServerError)
			return
		}

		// Удаляем сотрудника
		_, err = tx.Exec("DELETE FROM СЛУЖАЩИЕ WHERE СЛУ_НОМЕР = ?", id)
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка удаления сотрудника: %v", err), http.StatusInternalServerError)
			return
		}

		// Удаляем фото, если есть
		if imageURL.Valid && imageURL.String != "" {
			if err := deletePhoto(imageURL.String); err != nil {
				log.Printf("warning: не удалось удалить фото для %d: %v", id, err)
			}
		}

		// Обновляем данные отдела
		_, err = tx.Exec(`
            UPDATE ОТДЕЛЫ
            SET ОТД_РАЗМ = ОТД_РАЗМ - 1,
            ОТД_СОТР_ЗАРП = ОТД_СОТР_ЗАРП - ?
            WHERE ОТД_НОМЕР = ?`,
			salary, deptID,
		)

		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка обновления отдела: %v", err), http.StatusInternalServerError)
			return
		}

		tx.Commit()
		sendAPIResponse(c, nil, nil, http.StatusOK)
	}
}

// Обновить данные отдела (руководителя)
func updateDepartmentAPI(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный ID отдела"), http.StatusBadRequest)
			return
		}

		// Проверяем, существует ли отдел с таким ID
		var req struct {
			BossID *int `json:"boss_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректные данные: %v", err), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка транзакции: %v", err), http.StatusInternalServerError)
			return
		}
		var res sql.Result
		if req.BossID == nil {
			// Сброс руководителя в NULL
			res, err = tx.Exec("UPDATE ОТДЕЛЫ SET ОТД_РУК = NULL WHERE ОТД_НОМЕР = ?", id)
		} else {
			// Установка нового ID руководителя
			res, err = tx.Exec("UPDATE ОТДЕЛЫ SET ОТД_РУК = ? WHERE ОТД_НОМЕР = ?", *req.BossID, id)
		}
		if err != nil {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("ошибка обновления отдела: %v", err), http.StatusInternalServerError)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			tx.Rollback()
			sendAPIResponse(c, nil, fmt.Errorf("отдел с ID %d не найден", id), http.StatusNotFound)
			return
		}

		tx.Commit()
		sendAPIResponse(c, nil, nil, http.StatusOK)
	}
}

// GET /employees/:id/photo
func getEmployeePhotoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Парсим ID сотрудника
		empID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
			return
		}

		// Проверяем все возможные расширения
		extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
		var fullPath string
		var foundExt string
		for _, ext := range extensions {
			filename := fmt.Sprintf("emp_%d%s", empID, ext)
			tryPath := path.Join(storageEmployeeImages, filename)
			if _, err := os.Stat(tryPath); err == nil {
				fullPath = tryPath
				foundExt = ext
				break
			}
		}

		if fullPath == "" {
			//c.JSON(http.StatusNotFound, gin.H{"error": "does not exist"})
			//return
			fullPath = path.Join(storageEmployeeImages, "default.png")
		}

		// Определяем Content-Type по расширению
		contentTypes := map[string]string{
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".png":  "image/png",
			".gif":  "image/gif",
			".bmp":  "image/bmp",
			".webp": "image/webp",
		}
		contentType := contentTypes[foundExt]
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Content-Type", contentType)
		c.File(fullPath)
	}
}

// getEmployeeByDepartDocumentHandler возвращает DOCX со списком сотрудников отдела.
func getEmployeeByDepartDocumentHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Парсим ID отдела
		deptID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("некорректный ID отдела"), http.StatusBadRequest)
			return
		}

		// 2) Берём из ОТДЕЛЫ ожидаемые значения и boss_id
		var expectedSalary float64
		var expectedCount int
		var bossID sql.NullInt64
		err = db.QueryRow(`
            SELECT ОТД_СОТР_ЗАРП, ОТД_РАЗМ, ОТД_РУК
            FROM ОТДЕЛЫ WHERE ОТД_НОМЕР = ?`, deptID,
		).Scan(&expectedSalary, &expectedCount, &bossID)
		if err == sql.ErrNoRows {
			sendAPIResponse(c, nil, fmt.Errorf("отдел №%d не найден", deptID), http.StatusNotFound)
			return
		} else if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка запроса отдела: %v", err), http.StatusInternalServerError)
			return
		}

		// 3) Запрашиваем сотрудников
		rows, err := db.Query(`
            SELECT СЛУ_НОМЕР, СЛУ_ИМЯ, СЛУ_СТАТ, СЛУ_ЗАРП
            FROM СЛУЖАЩИЕ WHERE СЛУ_ОТД_НОМЕР = ?`, deptID,
		)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка запроса сотрудников: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var emps []emp
		var actualSalary float64
		for rows.Next() {
			var e emp
			if err := rows.Scan(&e.ID, &e.Name, &e.Status, &e.Salary); err != nil {
				sendAPIResponse(c, nil, fmt.Errorf("ошибка чтения сотрудника: %v", err), http.StatusInternalServerError)
				return
			}
			emps = append(emps, e)
			actualSalary += e.Salary
		}
		if len(emps) == 0 {
			sendAPIResponse(c, nil, fmt.Errorf("нет сотрудников в отделе №%d", deptID), http.StatusNotFound)
			return
		}
		actualCount := len(emps)

		// 8) Сохраняем и отдаем
		buf, err := setupDocument(deptID, expectedSalary, expectedCount, actualSalary, actualCount, bossID, emps)
		if err != nil {
			sendAPIResponse(c, nil, fmt.Errorf("ошибка создания документа: %v", err), http.StatusInternalServerError)
			return
		}
		filename := fmt.Sprintf("department_%d_employees.docx", deptID)
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Data(http.StatusOK,
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			buf.Bytes(),
		)
	}
}
