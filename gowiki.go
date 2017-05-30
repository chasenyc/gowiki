package main

import (
    "html/template"
    "net/http"
    "regexp"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "os"
    "fmt"
    "strings"
)

var pageTemplates = template.Must(template.ParseFiles("tmpl/page/edit.html", "tmpl/page/view.html"))
var tagsTemplates = template.Must(template.ParseFiles("tmpl/tags/tags.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|tags)/([a-zA-Z0-9]+)$")

func loadPage(title string) (*Page, error) {
    session, err := mgo.Dial(getMongo())
    if err != nil {
        return nil, err
    }

    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)
    c := session.DB("testwiki").C("pages")
    result := Page{}
    err = c.Find(bson.M{"title": title}).One(&result)

    if err != nil {
        return nil, err
    }
    return &Page{Title: result.Title, Body: result.Body, Tags: result.Tags, Timestamp: result.Timestamp}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    tagsRaw := r.FormValue("tags")
    tags := strings.Split(tagsRaw, ", ")
    p := &Page{Title: title, Body: []byte(body), Tags: tags}
    _, err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := pageTemplates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func redirectFront(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/tags/", makeHandler(tagHandler))
    http.HandleFunc("/", redirectFront)
    http.ListenAndServe(getPort(), nil)
}

func tagHandler(w http.ResponseWriter, r *http.Request, title string) {
    tags, err := loadTags()
    fmt.Println("got tags.")
    if err != nil {
        redirectFront(w, r)
        return
    }

    errTwo := tagsTemplates.ExecuteTemplate(w, "tags.html", tags)
    if errTwo != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func loadTags() ([]string, error) {
    session, err := mgo.Dial(getMongo())
    if err != nil {
        return nil, err
    }

    defer session.Close()

    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)
    c := session.DB("testwiki").C("pages")
    result := []string{}
    err = c.Find(nil).Distinct("tags", &result)

    if err != nil {
        return nil, err
    }
    return result, nil
}

// Get the Port from the environment so we can run on Heroku
func getPort() string {
	port := os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func getMongo() string {
    uri := os.Getenv("MONGOLAB_URL")

    if uri == "" {
        fmt.Printf("Can't connect to mongo, no uri provided.")
        os.Exit(1)
    }

    return uri
}
