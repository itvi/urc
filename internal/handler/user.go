package handler

import (
	"log"
	"net/http"
	"project/internal/model"
	e "project/pkg/error"
	"project/pkg/form"
	"strconv"
)

type UserHandler struct {
	M *model.UserModel
}

func (u *UserHandler) Index(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := u.M.GetUsers()
		if err != nil {
			log.Println("Get users error:", err)
			return
		}
		otherTemplates := []string{
			"./web/template/partial/toolbar_crud.html",
		}
		data := &TemplateData{
			Users: users,
		}
		c.render(w, r, otherTemplates, "./web/template/html/user/index.html", data)
	}
}

func (u *UserHandler) AddView(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := form.New(nil)
		data := &TemplateData{Form: form}
		c.render(w, r, nil, "./web/template/html/user/add.html", data)
	}
}

func (u *UserHandler) Add(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		// validate
		form := form.New(r.PostForm)
		form.Required("sn", "password")
		data := &TemplateData{Form: form}
		page := "./web/template/html/user/add.html"

		if !form.Valid() {
			c.render(w, r, nil, page, data)
			return
		}

		// create user
		var user = &model.User{
			SN:       form.Get("sn"),
			Name:     form.Get("name"),
			Email:    form.Get("email"),
			Password: form.Get("password"),
		}

		err = u.M.Create(user)
		if err == e.ErrDuplicate {
			form.Errors.Add("sn", "用户已存在")
			c.render(w, r, nil, page, data)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/users", http.StatusSeeOther)
	}
}

func (u *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil {
		log.Println("User ID convert error:", err)
		return
	}

	err = u.M.Delete(id)
	if err != nil {
		log.Println(err)
		return
	}
}

func (u *UserHandler) EditView(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get(":id"))
		if err != nil {
			log.Println("User ID convert error:", err)
			return
		}

		// TODO: maybe in middleware
		// don't change others data
		// userID := c.Session.GetInt(r, "userID")
		// if id != userID {
		// 	w.Write([]byte("forbidden"))
		// 	return
		// }

		user, err := u.M.GetUser(id)
		if err == e.ErrNoRecord {
			log.Println("No record found")
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		form := form.New(nil)
		data := &TemplateData{
			Form: form,
			User: user,
		}
		c.render(w, r, nil, "./web/template/html/user/edit.html", data)
	}
}

func (u *UserHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil {
		log.Println("User ID convert error:", err)
		return
	}

	sn := r.PostFormValue("sn")
	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	user := &model.User{
		ID:             id,
		SN:             sn,
		Name:           name,
		Email:          email,
		HashedPassword: []byte(password),
	}

	if err = u.M.Edit(user); err != nil {
		log.Println(err)
		return
	}
}

func (u *UserHandler) LoginView(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := form.New(nil)
		data := &TemplateData{Form: form}
		c.render(w, r, nil, "./web/template/html/user/login.html", data)
	}
}

// Login use configuration as parameter is for set session
func (u *UserHandler) Login(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println("Parse form error:", err)
			return
		}

		form := form.New(r.PostForm)
		sn := form.Get("sn")
		password := form.Get("password")

		user, err := c.User.M.Authenticate(sn, password)
		if err == e.ErrInvalidCredentials {
			form.Errors.Add("generic", "用户或密码不正确！")
			data := &TemplateData{
				Form: form,
			}
			c.render(w, r, nil, "./web/template/html/user/login.html", data)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		// session
		c.Session.Put(r, "userID", user.ID)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Logout remove session
func (u *UserHandler) Logout(c *Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Session.Remove(r, "userID")
		http.Redirect(w, r, "/", 303)
	}
}
