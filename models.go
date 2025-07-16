package main

// Department модель отдела
type Department struct {
	ID          int     `json:"id"`
	BossID      *int    `json:"boss_id,omitempty"` // nil если NULL
	BossName    *string `json:"boss_name"`         // имя руководителя или "Не установлен"
	TotalSalary float64 `json:"total_salary"`
	Size        int     `json:"size"`
}

// Employee модель сотрудника
type Employee struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"` // "active", "inactive", "fired"
	Salary   float64 `json:"salary"`
	DeptID   int     `json:"dept_id"`             // ID отдела, к которому принадлежит сотрудник
	ImageURL *string `json:"image_url,omitempty"` // URL для изображения сотрудника
}

// NavItem модель элемента навигации
type NavItem struct {
	Label string
	URL   string
}

// emp модель сотрудника для внутреннего использования создания документов
type emp struct {
	ID     int
	Name   string
	Status string
	Salary float64
}
