package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gy-kim/creating-web-application-with-go/webapp/middleware"
	"github.com/gy-kim/creating-web-application-with-go/webapp/model"

	"github.com/gy-kim/creating-web-application-with-go/webapp/controller"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

func main() {
	/* // Custom File Server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("../public" + r.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		defer f.Close()
		var contentType string
		switch {
		case strings.HasSuffix(r.URL.Path, "css"):
			contentType = "text/css"
		case strings.HasSuffix(r.URL.Path, "html"):
			contentType = "text/html"
		case strings.HasSuffix(r.URL.Path, "png"):
			contentType = "text/plain"
		}
		w.Header().Add("Content-Type", contentType)
		io.Copy(w, f)
	})
	http.ListenAndServe(":8000", nil)
	*/
	/*
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "../public"+r.URL.Path)
		})
		http.ListenAndServe(":8000", nil)
	*/
	// http.ListenAndServe(":8000", http.FileServer(http.Dir("../public")))

	/*
		templates := populateTemplates()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			requestedFile := r.URL.Path[1:]
			t := templates.Lookup(requestedFile + ".html")
			if t != nil {
				err := t.Execute(w, nil)
				if err != nil {
					log.Println(err)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		})
		http.Handle("/img/", http.FileServer(http.Dir("../public")))
		http.Handle("/css/", http.FileServer(http.Dir("../public")))
		http.ListenAndServe(":8000", nil) // http://localhost:8000/home
	*/
	/*
		templates := populateTemplates()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			requestedFile := r.URL.Path[1:]
			template := templates[requestedFile+".html"]
			var context interface{}
			switch requestedFile {
			case "shop":
				context = viewmodel.NewShop()
			default:
				context = viewmodel.NewHome()
			}
			if template != nil {
				err := template.Execute(w, context)
				if err != nil {
					log.Println(err)
				}
			} else {
				w.WriteHeader(404)
			}
		})

		http.ListenAndServe(":8000", nil) // http://localhost:8000/home
	*/
	templates := populateTemplates()
	controller.Startup(templates)
	db := connectToDatabase()
	defer db.Close()
	// http.ListenAndServe(":8000", nil)
	// http.ListenAndServe(":8000", new(middleware.GzipMiddleware))
	// http.ListenAndServe(":8000", &middleware.TimeoutMiddleware{new(middleware.GzipMiddleware)})
	go http.ListenAndServe(":8080", nil)
	http.ListenAndServeTLS(":8000", "../cert.pem", "../key.pem", &middleware.TimeoutMiddleware{new(middleware.GzipMiddleware)})
}

func connectToDatabase() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:example@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(fmt.Errorf("Unable to connect to database: %v", err))
	}
	model.SetDatabase(db)
	return db
}

// func populateTemplates() *template.Template {
// 	result := template.New("templates")
// 	const basePath = "../templates"
// 	template.Must(result.ParseGlob(basePath + "/*.html"))
// 	return result
// }

func populateTemplates() map[string]*template.Template {
	/*
		result := template.New("templates")
		const basePath = "../templates"
		template.Must(result.ParseGlob(basePath + "/*.html"))
		return result
	*/

	result := make(map[string]*template.Template)
	const basePath = "../templates"
	layout := template.Must(template.ParseFiles(basePath + "/_layout.html"))
	template.Must(layout.ParseFiles(basePath+"/_header.html", basePath+"/_footer.html"))
	dir, err := os.Open(basePath + "/content")
	if err != nil {
		panic("Failed to open template blocks directory: " + err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		panic("Failed to read contents of content directory: " + err.Error())
	}
	for _, fi := range fis {
		f, err := os.Open(basePath + "/content/" + fi.Name())
		if err != nil {
			panic("Failed to open template '" + fi.Name() + "'")
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			panic("Failed to read content from file '" + fi.Name() + "'")
		}
		f.Close()
		tmpl := template.Must(layout.Clone())
		_, err = tmpl.Parse(string(content))
		if err != nil {
			panic("Failed to parse contents of '" + fi.Name() + "' as template")
		}
		result[fi.Name()] = tmpl
	}
	return result
}
