package main

import (
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id = 0

type Contact struct {
	Name  string
	Email string
	Id    int
}

func NewContact(name, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}
}

type Contacts = []Contact

type PageData struct {
	Contacts Contacts
}

func (pd PageData) indexOf(id int) int {
	for i, c := range pd.Contacts {
		if c.Id == id {
			return i
		}
	}
	return -1
}

func NewPageData() PageData {
	return PageData{
		Contacts: []Contact{
			NewContact("john stevenson", "js@example.com"),
			NewContact("bill stevenson", "bs@example.com"),
			NewContact("lang horford", "lh@example.com"),
		},
	}
}

func (pd PageData) HasEmail(email string) bool {
	for _, e := range pd.Contacts {
		if e.Email == email {
			return true
		}
	}
	return false
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func NewFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	PageData PageData
	FormData FormData
}

func NewPage() Page {
	return Page{
		PageData: NewPageData(),
		FormData: NewFormData(),
	}
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = NewTemplates()

	page := NewPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name, email := c.FormValue("name"), c.FormValue("email")
		if page.PageData.HasEmail(email) {
			formData := NewFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(422, "form", formData)
		}
		contact := NewContact(name, email)
		page.PageData.Contacts = append(page.PageData.Contacts, contact)
		c.Render(200, "form", NewFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(3 * time.Second)
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "invalid id")
		}
		index := page.PageData.indexOf(id)
		if index == -1 {
			return c.String(404, "contact not found")
		}
		page.PageData.Contacts = append(page.PageData.Contacts[:index], page.PageData.Contacts[index+1:]...)
		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
