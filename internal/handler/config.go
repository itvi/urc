package handler

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"project/internal/model"
	e "project/pkg/error"
	"project/pkg/form"
	tmp "project/pkg/template"
	"time"

	"github.com/golangcollege/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type TemplateData struct {
	Now               time.Time
	AuthenticatedUser *model.User
	Form              *form.Form
	User              *model.User
	Users             []*model.User
	Role              *model.Role
	Roles             []*model.Role
	RolesForUser      []string
	Policy            *model.CasbinPolicy
	Policies          []*model.CasbinPolicy
}

type Configuration struct {
	Session *sessions.Session
	Home    *HomeHandler
	User    *UserHandler
	Role    *RoleHandler
	Casbin  *CasbinHandler
}

type contextKey string

var contextKeyUser = contextKey("user")

func Config() (*Configuration, *sql.DB) {
	// database
	db, err := openDB("./db.db")
	if err != nil {
		err = fmt.Errorf("open db error: %w -> from open db", err)
		log.Panic(err)
	}

	// session
	session := sessions.New([]byte("afkkjfkajf!23234324#@#$"))
	session.Lifetime = 1 * time.Hour

	c := &Configuration{
		Session: session,
		Home:    &HomeHandler{},
		User:    &UserHandler{M: &model.UserModel{DB: db}},
		Role:    &RoleHandler{M: &model.RoleModel{DB: db}},
		Casbin:  &CasbinHandler{M: &model.CasbinModel{DB: db}},
	}
	return c, db
}

func (c *Configuration) authenticatedUser(r *http.Request) *model.User {
	user, ok := r.Context().Value(contextKeyUser).(*model.User)
	if !ok {
		return nil
	}
	return user
}

func (c *Configuration) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.authenticatedUser(r) == nil {
			http.Redirect(w, r, "/users/login", 302)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (c *Configuration) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if a "userID" exists in the session.
		exists := c.Session.Exists(r, "userID")
		if !exists {
			log.Println("session not exist.")
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the details of the current user from the database.
		userID := c.Session.GetInt(r, "userID")
		fmt.Println("session userid:", userID)
		user, err := c.User.M.GetUser(userID)
		if err == e.ErrNoRecord {
			c.Session.Remove(r, "userID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		// The request is coming form a valid, authenticated user.
		// Create a new copy of the request with the user
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *Configuration) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enforcer, err := c.Casbin.M.Init()
		if err != nil {
			log.Println("casbin init error")
		}

		// get current logined user
		user := c.authenticatedUser(r)

		var sub string
		if user == nil {
			sub = "anonymous"
		} else {
			sub = user.SN
		}
		obj := r.URL.Path
		act := r.Method

		ok, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			log.Println("casbin enforce error:", err)
		}
		if ok {
			next.ServeHTTP(w, r)
		} else {
			log.Println("forbidden [", sub, obj, act+" ]")
			w.Write([]byte("forbidden"))
		}
	})
}

// database
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (c *Configuration) addDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}
	td.AuthenticatedUser = c.authenticatedUser(r)

	return td
}

// Render render templates with preset layouts
func (c *Configuration) render(w http.ResponseWriter, r *http.Request,
	otherTemplates []string, templateName string, data *TemplateData) {

	layouts := []string{
		"./web/template/layout.html",
		"./web/template/partial/menu.html",
	}
	layouts = append(layouts, otherTemplates...)

	templateData := c.addDefaultData(data, r)
	tmp.RenderTemplates(w, r, layouts, templateName, "layout", funcMaps, templateData)
}

var funcMaps = template.FuncMap{
	"safe": func(s string) template.HTMLAttr {
		return template.HTMLAttr(s)
	},
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"ownedRoles": func(role string, roles []string) string {
		for _, r := range roles {
			if r == role {
				return "checked"
			}
		}
		return ""
	},
}
