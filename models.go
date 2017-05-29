package main

import (
    "regexp"
    "html/template"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "time"
    "fmt"
)

type Page struct {
    Title string
    Body  []byte
    Tags  []string
    Timestamp time.Time
}

func (p Page) ConvertedBody() template.HTML {
    search := regexp.MustCompile("\\[([a-zA-Z]+)\\]")

    body := search.ReplaceAllFunc(p.Body, func(s []byte) []byte {
        m := string(s[1 : len(s)-1])
        return []byte("<a href=\"/view/" + m + "\">" + m + "</a>")
    })

    return template.HTML(string(body[:]))
}

func (p Page) ConvertedTime() template.HTML {
    const layout = "Jan 2, 2006 at 3:04pm (MST)"
    return template.HTML(p.Timestamp.Format(layout))
}

func (p *Page) save() (info *mgo.ChangeInfo, err error) {
    session, _ := mgo.Dial(getMongo())
    defer session.Close()
    p.Timestamp = time.Now()
    fmt.Println(p.Timestamp.String())
    session.SetMode(mgo.Monotonic, true)

    c := session.DB("testwiki").C("pages")

    return c.Upsert(bson.M{"title": p.Title}, p)
}
