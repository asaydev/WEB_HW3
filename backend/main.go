package main

import (

	"fmt"
	"github.com/gobuffalo/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var port = "8080"
var users []User
var posts []Post
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
var dbSessions = map[string]Session{} // session ID, session
const sessionLength int = 5 * 60

func main() {
	var admin User
	admin.Type="admin"
	admin.Email="m.sab@yahoo.com"
	bs, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	admin.Password=string(bs)
	users=append(users,admin)
	fmt.Println("Server started on port :" + port)
	handleRequests()
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/signup", signup)
	router.HandleFunc("/api/signin", signin)
	router.HandleFunc("/api/logout", authorized(signout))
	router.HandleFunc("/api/admin/post/crud", checkadmin(crud))
	router.HandleFunc("/api/admin/post/crud/{id}", checkadmin(updatedelete))
	router.HandleFunc("/api/admin/user/crud/{id}", checkadmin(read))
	router.HandleFunc("/api/post", checkuser(getposts))
	http.ListenAndServe(":"+port, router)
}

func getposts(w http.ResponseWriter, req *http.Request)  {
	if req.Method == http.MethodGet{
		json.NewEncoder(w).Encode(posts)
	}

}

func read(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet{
		id:= mux.Vars(req)["id"]
		fmt.Println(id)
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "url id is not valid", http.StatusBadRequest)
			return // don't call original handler
		}
		int64Id:=int64(intid)
		ok,postid := getpostbyid(int64Id)
		if !ok {
			http.Error(w, "post with this id is not found!", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(posts[postid])

	}
}

func signup(w http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(w, req) {
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		email := req.FormValue("email")
		password := req.FormValue("password")

		//null value
		if email == "" ||  password== "" {
			http.Error(w, "Request Length should be 2", http.StatusBadRequest)
			return
		}

		if !isEmailValid(email) {
			http.Error(w, "filed `email` is not valid", http.StatusBadRequest)
			return
		}

		flag := false
		for i := 0; i < len(users); i++ {
			if users[i].Email == email {
				flag = true
			}
		}

		// username taken?
		if flag{
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}


		// create session
		sID, _ := uuid.NewV4()
		cookie := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		cookie.MaxAge = sessionLength
		http.SetCookie(w, cookie)
		dbSessions[cookie.Value] = Session{email, time.Now()}
		// store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		user := User{
			Email:email,
			Password:string(bs),
			Type:"user",
		}

		users = append(users,user)
		w.WriteHeader(http.StatusOK)
		return
	}else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func signin(w http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(w, req) {
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {
		email := req.FormValue("email")
		password := req.FormValue("password")

		if email == "" ||  password== "" {
			http.Error(w, "Request Length should be 2", http.StatusBadRequest)
			return
		}

		if !isEmailValid(email) {
			http.Error(w, "filed `email` is not valid", http.StatusBadRequest)
			return
		}

		// is there a username?
		var user User
		flag := false
		for i := 0; i < len(users); i++ {
			if users[i].Email == email {
				user = users[i]
				flag = true
			}
		}

		if !flag {
			//w.WriteHeader(http.StatusOK)
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// does the entered password match the stored password?

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		//create session
		sID, _ := uuid.NewV4()
		cookie := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		cookie.MaxAge = sessionLength
		http.SetCookie(w, cookie)
		dbSessions[cookie.Value] = Session{email, time.Now()}
		w.WriteHeader(http.StatusOK)

		return
	}else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func crud(w http.ResponseWriter, req *http.Request)  {

	if req.Method == http.MethodPost {
		title := req.FormValue("title")
		content := req.FormValue("content")

		if title == "" ||  content == "" {
			http.Error(w, "Request Length should be 2", http.StatusBadRequest)
			return
		}
		fmt.Println(title)
		isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
		if !isAlpha(title) {
			http.Error(w, "filed `title` is not valid", http.StatusBadRequest)
			return // don't call original handler
		}


		postId:=len(posts)+1
		currentTime:=time.ANSIC
		cookie, err := req.Cookie("session")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		session, _ := dbSessions[cookie.Value]
		var post Post
		post.Id = int64(postId)
		post.Content = content
		post.Created_by = session.Un
		post.Created_at = currentTime
		post.Title = title
		posts=append(posts , post)
		var pos PostID
		pos.Id=post.Id
		json.NewEncoder(w).Encode(pos)
	}else if req.Method == http.MethodGet{
		json.NewEncoder(w).Encode(posts)
	}
}

func updatedelete(w http.ResponseWriter, req *http.Request)  {

	if req.Method == http.MethodPut{
		id:= mux.Vars(req)["id"]
		isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
		title := req.FormValue("title")
		content:= req.FormValue("content")
		if !isAlpha(title) {
			http.Error(w, "filed `title` is not valid", http.StatusBadRequest)
			return // don't call original handler
		}
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "url id is not valid", http.StatusBadRequest)
			return // don't call original handler
		}
		int64Id:=int64(intid)
		ok,postid := getpostbyid(int64Id)
		if !ok {
			http.Error(w, "post with this id is not found!", http.StatusNotFound)
			return
		}
		posts[postid].Content = content
		posts[postid].Title = title
		w.WriteHeader(http.StatusOK)
	}else if req.Method == http.MethodDelete{
		id:= mux.Vars(req)["id"]
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "url id is not valid", http.StatusBadRequest)
			return // don't call original handler
		}
		int64Id:=int64(intid)
		ok,postid := getpostbyid(int64Id)
		if !ok {
			http.Error(w, "post with this id is not found!", http.StatusNotFound)
			return
		}
		posts=removepost(posts,postid)
		w.WriteHeader(http.StatusOK)
	}else if req.Method == http.MethodGet {
		id:= mux.Vars(req)["id"]
		intid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			http.Error(w, "url id is not valid", http.StatusBadRequest)
			return // don't call original handler
		}
		int64Id:=int64(intid)
		ok,postid := getpostbyid(int64Id)
		if !ok {
			http.Error(w, "post with this id is not found!", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(posts[postid])
	}
}


func authorized(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !alreadyLoggedIn(w, r) {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return // don't call original handler
		}
		h.ServeHTTP(w, r)
	})
}

func checkadmin(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !alreadyLoggedIn(w, r) {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return // don't call original handler
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return // don't call original handler
		}
		session, _ := dbSessions[cookie.Value]

		for i:=0;i<len(users) ;i++  {
			if session.Un ==users[i].Email {
				if users[i].Type == "admin" {
					h.ServeHTTP(w, r)
					return
				}else {
					http.Error(w, "Only admin can access to this service", http.StatusUnauthorized)
					return // don't call original handler
				}
			}
		}
	//	h.ServeHTTP(w, r)
	})
}

func checkuser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !alreadyLoggedIn(w, r) {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return // don't call original handler
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return // don't call original handler
		}
		session, _ := dbSessions[cookie.Value]

		for i:=0;i<len(users) ;i++  {
			if session.Un ==users[i].Email {
				if users[i].Type == "user" {
					h.ServeHTTP(w, r)
					return
				}else {
					http.Error(w, "Only user can access to this service", http.StatusUnauthorized)
					return // don't call original handler
				}
			}
		}
		//	h.ServeHTTP(w, r)
	})
}


func signout(w http.ResponseWriter, req *http.Request) {
	cookie, _ := req.Cookie("session")
	// delete the session
	delete(dbSessions, cookie.Value)
	// remove the cookie
	cookie = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}


func alreadyLoggedIn(w http.ResponseWriter, req *http.Request) bool {
	cookie, err := req.Cookie("session")
	if err != nil {
		return false
	}
	session, ok := dbSessions[cookie.Value]
	if ok {
		//session.LastActivity = time.Now()
		//dbSessions[cookie.Value] = session
	} else {
		return false
	}

	ok = contains(users, session.Un)

	// refresh session
	//cookie.MaxAge = sessionLength
	///http.SetCookie(w, cookie)
	return ok
}

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	if !emailRegex.MatchString(e) {
		return false
	}
	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

func contains(slice []User, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s.Email] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func removepost(slice []Post, s int) []Post {
	return append(slice[:s], slice[s+1:]...)
}


func getpostbyid(id int64)(bool,int){
	for i:=0;i<len(posts);i++  {
		if (id) == posts[i].Id {
			return true ,i
		}
	}
	return false,-1
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type Session struct {
	Un           string
	LastActivity time.Time
}

type Post struct {
	Id int64 `json:"id"`
	Title string `json:"title"`
	Content string `json:"content"`
	Created_by string `json:"created_by"`
	Created_at string `json:"created_at"`
}


type PostID struct {
	Id int64 `json:"id"`
}
