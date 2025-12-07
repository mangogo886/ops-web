package statistics

import (
	"html/template"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":       "统计信息",
		"ActiveMenu":  "stats",
		"ContentHtml": template.HTML("<h2>统计功能开发中...</h2>"),
	}

	tmpl.Execute(w, data)
}