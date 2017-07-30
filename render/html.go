// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"html/template"
	"net/http"
)

type (
	HTML struct {
		Template *template.Template
		Data     interface{}
	}
)

var htmlContentType = []string{"text/html; charset=utf-8"}

func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return r.Template.Execute(w, r.Data)
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
