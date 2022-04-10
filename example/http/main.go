package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cuipeiyu/go-i18n"
	"golang.org/x/text/language"
)

func main() {
	bundle := i18n.NewBundle(language.English)
	bundle.LoadMessageFile("locales/en-US.yaml")
	bundle.LoadMessageFile("locales/zh-CN.yaml")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lang := r.FormValue("lang")
		accept := r.Header.Get("Accept-Language")
		localizer := i18n.NewLocalizer(bundle, lang, accept)

		name := r.FormValue("name")
		if name == "" {
			name = "Bob"
		}

		helloPerson := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "HelloPerson",
				Other: "Hello {{.Name}}",
			},
			TemplateData: map[string]string{
				"Name": name,
			},
		})

		w.Write([]byte(helloPerson))
	})

	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
