package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unidoc/unioffice/v2/common/license"
)

// setupDatabase initializes the database connection
func setupDatabase() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("база данных недоступна: %v", err)
	}

	return db, nil
}

// Wait for the database to be available. Pinging it every {backoff} seconds
func waitForDB() *sql.DB {
	var (
		conn *sql.DB
		err  error
	)
	backoff := 2 * time.Second
	for {
		conn, err = setupDatabase()
		if err != nil {
			log.Printf("Ошибка открытия БД: %v. Повтор через %s…", err, backoff)
		} else if pingErr := conn.Ping(); pingErr != nil {
			log.Printf("Ошибка пинга БД: %v. Повтор через %s…", pingErr, backoff)
		} else {
			log.Println("Успешно подключились к БД")
			return conn
		}
		// Если открыли, но не смогли запинговать, закроем
		if conn != nil {
			conn.Close()
		}
		time.Sleep(backoff)
		// можно увеличить backoff (экспоненциально), но не обязательно
	}
}

// dbAliveMiddleware checks if the database connection is alive
func dbAliveMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || db.Ping() != nil {
			// если хочешь, можешь делать c.Redirect(...) на страницу с ошибкой
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   "Service unavailable. Database connection is not available",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// API маршруты для работы с отделами и сотрудниками
func setupAPIRoutes(r *gin.Engine, db *sql.DB) {
	api := r.Group("/api")
	{
		// Отделы
		api.GET("/departments", getDepartments(db))
		api.PUT("/departments/:id", updateDepartmentAPI(db))

		// Служащие
		api.GET("/employees", getEmployees(db))
		api.GET("/employees/:id", getEmployee(db))
		api.GET("/employeesByDepart/:id", getEmployeeByDepartment(db))
		api.POST("/employees", createEmployeeAPI(db))
		api.PUT("/employees/:id", updateEmployeeAPI(db))
		api.DELETE("/employees/:id", deleteEmployeeAPI(db))

		// Return image URL for employee photo
		api.GET("/employees/:id/photo", getEmployeePhotoHandler())

		// Return document MS Word for employee
		api.GET("/employeesByDepart/:id/document", getEmployeeByDepartDocumentHandler(db))
	}
}

func init() {
	// Загружаем API-ключ из переменной окружения
	err := license.SetMeteredKey(unidocKey)
	if err != nil {
		panic(err)
	}

}

func main() {

	// Настройка соединения с базой данных
	db := waitForDB()

	r := gin.Default()
	r.Use(cors.Default())

	// Call Middleware проверки подключения к БД
	r.Use(dbAliveMiddleware(db))

	// Call routes setup function
	setupAPIRoutes(r, db)

	r.RunTLS(listenIP, certPath, keyPath) // Запуск сервера с поддержкой HTTPS
}
