package main

import (
	"app/connection"
	"app/middleware"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type TestimonialResponse struct {
	Image  string `json:"image"`
	Quote  string `json:"quote"`
	Author string `json:"author"`
	Rating int    `json:"rating"`
}

type ProjectFormRequest struct {
	ProjectName  string   `form:"project_name"`
	StartDate    string   `form:"start_date"`
	EndDate      string   `form:"end_date"`
	Description  string   `form:"description"`
	Technologies []string `form:"technologies"`
}

type Project struct {
	Id           int
	ProjectName  string `db:"name"`
	StartDate    time.Time
	EndDate      time.Time
	Description  string
	Technologies []string
	Image        string
	FullName     sql.NullString
}

type ProjectUpdateResponse struct {
	Id           int
	ProjectName  string
	StartDate    string
	EndDate      string
	Description  string
	Technologies []string
	Image        string
}

type ProjectResponse struct {
	Id           int
	ProjectName  string
	Duration     string
	Description  string
	Technologies []string
	Image        string
	FullName     sql.NullString
}

type ProjectDetailResponse struct {
	Id           int
	ProjectName  string
	StartDate    string
	EndDate      string
	Duration     string
	Description  string
	Technologies []string
	Image        string
	FullName     sql.NullString
}

type User struct {
	Id       int
	FullName string
	Email    string
	Password string
}

var dataProject = []Project{
	{
		ProjectName:  "Marketing Dashboard",
		StartDate:    time.Now(),
		EndDate:      time.Now(),
		Description:  "My First Project Is Marketing Dashboard.",
		Technologies: []string{"node.js", "react.js", "next.js"},
	},
	{
		ProjectName:  "Job Seeker",
		StartDate:    time.Now(),
		EndDate:      time.Now(),
		Description:  "My Second Project Is Marketing Dashboard.",
		Technologies: []string{"node.js", "next.js", "typescript"},
	},
}

func main() {
	e := echo.New()

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Static("/assets", "assets")
	e.Static("/uploads", "uploads")

	e.GET("/", home)
	e.GET("/contact/me", contactMe)
	e.GET("/project", project)
	e.GET("/project/detail/:id", projectDetail)
	e.GET("/testimonial", testimonial)
	e.POST("/project", middleware.UploadFile(addProject))
	e.GET("/project/delete/:id", deleteProject)
	e.GET("/project/:id", updateProjectView)
	e.POST("/project/update/:id", middleware.UploadFile(updateProject))

	e.GET("/page/register", formRegister)
	e.GET("/page/login", formLogin)

	e.POST("/action/register", register)
	e.POST("/action/login", login)
	e.GET("/action/logout", logout)

	e.Logger.Fatal(e.Start(":8000"))
}

func formLogin(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/login.html")
	if err != nil {

		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["status"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	session.Values["message"] = ""

	session.Save(c.Request(), c.Response())

	return tmpl.Execute(c.Response(), flash)
}

func formRegister(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/register.html")
	if err != nil {

		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	return tmpl.Execute(c.Response(), flash)
}

func register(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	ctx := context.Background()

	client := connection.DBClient(ctx)

	insertUsersQuery := "INSERT INTO public.tb_users(name, email, password) VALUES ($1, $2, $3)"

	_, err = client.Exec(ctx, insertUsersQuery, name, email, passwordHash)

	if err != nil {
		redirectWithMessage(c, "Register Failed", false, "/page/register")
	}

	return redirectWithMessage(c, "Register Success", true, "/page/login")
}

func login(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := c.FormValue("email")
	password := c.FormValue("password")

	ctx := context.Background()

	client := connection.DBClient(ctx)

	selectUser := "SELECT * FROM public.tb_users WHERE email=$1"

	user := User{}

	err = client.QueryRow(ctx, selectUser, email).Scan(&user.Id, &user.FullName, &user.Email, &user.Password)

	if err != nil {
		return redirectWithMessage(c, "Email Salah", false, "/page/login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return redirectWithMessage(c, "Password Salah", false, "/page/login")
	}

	session, _ := session.Get("session", c)
	session.Options.MaxAge = 10800 // 3 jam
	session.Values["status"] = true
	session.Values["message"] = "Login Success"
	session.Values["name"] = user.FullName
	session.Values["id"] = user.Id
	session.Values["isLogin"] = true
	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func logout(c echo.Context) error {
	session, _ := session.Get("session", c)
	session.Options.MaxAge = -1
	session.Values["isLogin"] = false
	session.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusTemporaryRedirect, "/page/login")
}

func redirectWithMessage(c echo.Context, message string, status bool, path string) error {
	session, _ := session.Get("session", c)
	session.Values["status"] = status
	session.Values["message"] = message
	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, path)
}

func home(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {

		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	session, _ := session.Get("session", c)

	selectAllProject := ""

	var results []Project

	var newResults []ProjectResponse

	var idUser = session.Values["id"]

	if idUser != nil {
		selectAllProject = "SELECT tp.id, tp.name, tp.start_date, tp.end_date, tp.description, tp.technologies, tp.image, tu.name as fullname FROM public.tb_projects tp LEFT JOIN public.tb_users tu ON tp.users_id = tu.id WHERE tu.id = $1 ORDER BY id DESC"
		data, errQuery := client.Query(ctx, selectAllProject, idUser)

		if errQuery != nil {
			data := map[string]string{"message": errQuery.Error()}

			return c.JSON(http.StatusInternalServerError, data)
		}

		for data.Next() {
			var object Project
			err := data.Scan(&object.Id, &object.ProjectName, &object.StartDate, &object.EndDate, &object.Description, &object.Technologies, &object.Image, &object.FullName)
			if err != nil {
				data := map[string]string{"message": err.Error()}
				return c.JSON(http.StatusInternalServerError, data)
			}

			results = append(results, object)
		}

		newResults = make([]ProjectResponse, len(results))

		for i, val := range results {
			newResults[i].Id = val.Id
			newResults[i].ProjectName = val.ProjectName
			newResults[i].Duration = distanceDate(val.StartDate, val.EndDate)
			newResults[i].Description = val.Description
			newResults[i].Technologies = val.Technologies
			newResults[i].Image = val.Image
			newResults[i].FullName = val.FullName
		}

	} else {
		selectAllProject = "SELECT tp.id, tp.name, tp.start_date, tp.end_date, tp.description, tp.technologies, tp.image, tu.name as fullname FROM public.tb_projects tp LEFT JOIN public.tb_users tu ON tp.users_id = tu.id ORDER BY id DESC"
		data, errQuery := client.Query(ctx, selectAllProject)

		if errQuery != nil {
			data := map[string]string{"message": errQuery.Error()}

			return c.JSON(http.StatusInternalServerError, data)
		}

		for data.Next() {
			var object Project
			err := data.Scan(&object.Id, &object.ProjectName, &object.StartDate, &object.EndDate, &object.Description, &object.Technologies, &object.Image, &object.FullName)
			if err != nil {
				data := map[string]string{"message": err.Error()}
				return c.JSON(http.StatusInternalServerError, data)
			}

			results = append(results, object)
		}

		newResults = make([]ProjectResponse, len(results))

		for i, val := range results {
			newResults[i].Id = val.Id
			newResults[i].ProjectName = val.ProjectName
			newResults[i].Duration = distanceDate(val.StartDate, val.EndDate)
			newResults[i].Description = val.Description
			newResults[i].Technologies = val.Technologies
			newResults[i].Image = val.Image
			newResults[i].FullName = val.FullName
		}
	}

	flash := map[string]interface{}{
		"data":         newResults,
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	session.Values["message"] = ""

	session.Save(c.Request(), c.Response())

	return tmpl.Execute(c.Response(), flash)
}

func contactMe(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/contact-me.html")
	if err != nil {
		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	return tmpl.Execute(c.Response(), flash)
}

func distanceDate(startDate time.Time, endDate time.Time) string {
	diff := endDate.Sub(startDate)

	var yearDistance float64 = math.Floor(float64(diff.Milliseconds()) / (12 * 30 * 24 * 60 * 60 * 1000))
	if yearDistance > 0 {
		year := fmt.Sprintf("%d year", int(yearDistance))
		return year
	} else {
		var monthDistance float64 = math.Floor(float64(diff.Milliseconds()) / (30 * 24 * 60 * 60 * 1000))
		if monthDistance > 0 {
			month := fmt.Sprintf("%d month", int(monthDistance))
			return month
		} else {
			var dayDistance float64 = math.Floor(float64(diff.Milliseconds()) / (24 * 60 * 60 * 1000))
			if dayDistance > 0 {
				day := fmt.Sprintf("%d day", int(dayDistance))
				return day
			} else {
				return "1 day"
			}
		}
	}
}

func project(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/project.html")
	if err != nil {
		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	project := map[string]interface{}{
		"FlashValue": flash,
	}

	return tmpl.Execute(c.Response(), project)
}

func projectDetail(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var tmpl, err = template.ParseFiles("views/project-detail.html")
	if err != nil {
		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	selectProjectQuery := "SELECT tp.id, tp.name, tp.start_date, tp.end_date, tp.description, tp.technologies, tp.image, tu.name as fullname FROM public.tb_projects tp LEFT JOIN public.tb_users tu ON tp.users_id = tu.id WHERE tp.id=$1"

	var projectDetail = Project{}

	errQuery := client.QueryRow(ctx, selectProjectQuery, id).Scan(&projectDetail.Id, &projectDetail.ProjectName, &projectDetail.StartDate, &projectDetail.EndDate, &projectDetail.Description, &projectDetail.Technologies, &projectDetail.Image, &projectDetail.FullName)

	if errQuery != nil {
		data := map[string]string{"message": errQuery.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	var newResults = ProjectDetailResponse{}

	newResults = ProjectDetailResponse{
		Id:           projectDetail.Id,
		ProjectName:  projectDetail.ProjectName,
		StartDate:    dateFormat(projectDetail.StartDate, "RFC822"),
		EndDate:      dateFormat(projectDetail.EndDate, "RFC822"),
		Duration:     distanceDate(projectDetail.StartDate, projectDetail.EndDate),
		Description:  projectDetail.Description,
		Technologies: projectDetail.Technologies,
		Image:        projectDetail.Image,
		FullName:     projectDetail.FullName,
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	data := map[string]interface{}{
		"data":       newResults,
		"FlashValue": flash,
	}

	return tmpl.Execute(c.Response(), data)
}

func addProject(c echo.Context) error {

	session, _ := session.Get("session", c)

	var tech ProjectFormRequest
	image := c.Get("dataFile").(string)
	var userId = session.Values["id"]

	err := c.Bind(&tech)
	if err != nil {
		data := map[string]string{"message": err.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	insertBlogQuery := "INSERT INTO public.tb_projects(name,start_date,end_date,description,technologies,image,users_id) VALUES($1,$2,$3,$4,$5,$6,$7)"

	_, errQuery := client.Exec(ctx, insertBlogQuery, tech.ProjectName, tech.StartDate, tech.EndDate, tech.Description, tech.Technologies, image, userId)

	if errQuery != nil {
		data := map[string]string{"message": errQuery.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func deleteProject(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		data := map[string]string{"message": err.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	deleteProjectQuery := "DELETE FROM public.tb_projects WHERE id=$1"

	_, errQuery := client.Exec(ctx, deleteProjectQuery, id)

	if errQuery != nil {
		data := map[string]string{"message": errQuery.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func dateFormat(date time.Time, dateType string) string {
	if dateType == "RFC822" {
		return date.Format("02 Jan 2006")
	} else {
		return fmt.Sprintf("%d-%02d-%02d", date.Year(), int(date.Month()), date.Day())
	}
}

func updateProjectView(c echo.Context) error {
	var tmpl, errTmpl = template.ParseFiles("views/project-update.html")
	if errTmpl != nil {
		data := map[string]string{"message": errTmpl.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}
	id, errId := strconv.Atoi(c.Param("id"))
	if errId != nil {
		data := map[string]string{"message": errId.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	selectProjectQuery := "SELECT tp.id, tp.name, tp.start_date, tp.end_date, tp.description, tp.technologies, tp.image, tu.name as fullname FROM public.tb_projects tp LEFT JOIN public.tb_users tu ON tp.users_id = tu.id WHERE tp.id = $1"

	var projectDetail = Project{}

	errQuery := client.QueryRow(ctx, selectProjectQuery, id).Scan(&projectDetail.Id, &projectDetail.ProjectName, &projectDetail.StartDate, &projectDetail.EndDate, &projectDetail.Description, &projectDetail.Technologies, &projectDetail.Image, &projectDetail.FullName)

	var projectUpdateDetail = ProjectUpdateResponse{}

	if errQuery != nil {
		data := map[string]string{"message": errQuery.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	} else {
		projectUpdateDetail = ProjectUpdateResponse{
			Id:           projectDetail.Id,
			ProjectName:  projectDetail.ProjectName,
			StartDate:    dateFormat(projectDetail.StartDate, ""),
			EndDate:      dateFormat(projectDetail.EndDate, ""),
			Description:  projectDetail.Description,
			Technologies: projectDetail.Technologies,
			Image:        projectDetail.Image,
		}
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	data := map[string]interface{}{
		"data":       projectUpdateDetail,
		"FlashValue": flash,
	}

	return tmpl.Execute(c.Response(), data)
}

func updateProject(c echo.Context) error {
	id, errId := strconv.Atoi(c.Param("id"))

	if errId != nil {
		data := map[string]string{"message": errId.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	var tech ProjectFormRequest
	session, _ := session.Get("session", c)
	image := c.Get("dataFile").(string)
	var userId = session.Values["id"]

	errForm := c.Bind(&tech)
	if errForm != nil {
		data := map[string]string{"message": errForm.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	ctx := context.Background()

	client := connection.DBClient(ctx)

	updateBlogQuery := "UPDATE public.tb_projects SET name=$1, start_date=$2, end_date=$3, description=$4, technologies=$5 ,image=$6, users_id=$7 WHERE id=$8"

	_, errQuery := client.Exec(ctx, updateBlogQuery, tech.ProjectName, tech.StartDate, tech.EndDate, tech.Description, tech.Technologies, image, userId, id)

	if errQuery != nil {
		data := map[string]string{"message": errQuery.Error()}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func testimonial(c echo.Context) error {
	var tmpl, errTmpl = template.ParseFiles("views/testimonial.html")
	if errTmpl != nil {
		data := map[string]string{"message": errTmpl.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	response, errCall := http.Get("https://api.npoint.io/4d27242ce954114b709d")

	if errCall != nil {
		data := map[string]string{"message": errCall.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	responseData, errData := ioutil.ReadAll(response.Body)

	if errData != nil {
		data := map[string]string{"message": errData.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	var responseObject []TestimonialResponse

	errJson := json.Unmarshal(responseData, &responseObject)

	if errJson != nil {
		data := map[string]string{"message": errJson.Error()}

		return c.JSON(http.StatusInternalServerError, data)
	}

	session, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  session.Values["isLogin"],
		"FlashMessage": session.Values["message"],
		"FlashName":    session.Values["name"],
	}

	data := map[string]interface{}{
		"Data":       responseObject,
		"FlashValue": flash,
	}

	return tmpl.Execute(c.Response(), data)
}
