package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/yosssi/ace"
	"golang.org/x/net/context"

	proto "github.com/micro/go-platform/trace/proto"
	"github.com/micro/trace-srv/proto/trace"
)

var (
	opts        *ace.Options
	traceClient trace.TraceClient
)

func Init(dir string, t trace.TraceClient) {
	traceClient = t

	opts = ace.InitializeOptions(nil)
	opts.BaseDir = dir
	opts.DynamicReload = true
	opts.FuncMap = template.FuncMap{
		"Data": func(spans []*proto.Span) string {
			b, _ := json.Marshal(spans)
			return string(b)
		},
		"Delta": func(i int, a []*proto.Annotation) string {
			if i == 0 {
				return "0ms"
			}
			j := a[i].Timestamp
			k := a[i-1].Timestamp
			return fmt.Sprintf("%.3fms", float64(j-k)/1000.0)
		},
		"Duration": func(t int64) string {
			return fmt.Sprintf("%.3fms", float64(t)/1000.0)
		},
		"Offset": func(t int64, s []*proto.Span) string {
			d := t - s[0].Timestamp
			if d == 0 {
				return "0px"
			}
			w := float64(d) / float64(s[0].Duration)
			return fmt.Sprintf("%.0f%%", w*100)
		},
		"Service": func(s *proto.Service) string {
			if s == nil {
				return "n/a"
			}
			return s.Name
		},
		"Scale": func(t int64, s []*proto.Span) string {
			if len(s) == 0 {
				return "100%"
			}

			w := float64(t) / float64(s[0].Duration)

			return fmt.Sprintf("%.0f%%", w*100)
		},
		"TimeAgo": func(t int64) string {
			return timeAgo(t)
		},
		"Timestamp": func(t int64) string {
			return time.Unix(t/1e6, 0).Format(time.RFC822)
		},
		"Colour": func(s string) string {
			return colour(s)
		},
	}
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	basePath := hostPath(r)

	opts.FuncMap["URL"] = func(path string) string {
		return filepath.Join(basePath, path)
	}

	tpl, err := ace.Load("layout", tmpl, opts)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 302)
		return
	}

	if err := tpl.Execute(w, data); err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 302)
	}
}

// The index page
func Index(w http.ResponseWriter, r *http.Request) {
	rsp, err := traceClient.Search(context.TODO(), &trace.SearchRequest{
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedSpans{spans: rsp.Spans, reverse: false})

	for _, span := range rsp.Spans {
		if len(span.Annotations) == 0 {
			continue
		}
		sort.Sort(sortedAnns{span.Annotations})
	}

	render(w, r, "index", map[string]interface{}{
		"Latest": rsp.Spans,
	})
}

func Latest(w http.ResponseWriter, r *http.Request) {
	rsp, err := traceClient.Search(context.TODO(), &trace.SearchRequest{
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedSpans{spans: rsp.Spans, reverse: false})

	for _, span := range rsp.Spans {
		if len(span.Annotations) == 0 {
			continue
		}
		sort.Sort(sortedAnns{span.Annotations})
	}

	render(w, r, "latest", map[string]interface{}{
		"Latest": rsp.Spans,
	})
}

func Search(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		id := r.Form.Get("id")

		if len(id) > 0 {
			http.Redirect(w, r, filepath.Join(hostPath(r), "trace/"+id), 302)
			return
		}

		q := r.Form.Get("q")

		if len(q) == 0 {
			http.Redirect(w, r, filepath.Join(hostPath(r), "search"), 302)
			return
		}

		rsp, err := traceClient.Search(context.TODO(), &trace.SearchRequest{
			Name:    q,
			Reverse: true,
		})
		if err != nil {
			http.Redirect(w, r, filepath.Join(hostPath(r), "search"), 302)
			return
		}
		render(w, r, "results", map[string]interface{}{
			"Name":    q,
			"Results": rsp.Spans,
		})

		return
	}
	render(w, r, "search", map[string]interface{}{})
}

func Trace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) == 0 {
		http.Redirect(w, r, "/", 302)
		return
	}
	// TODO: limit/offset
	rsp, err := traceClient.Read(context.TODO(), &trace.ReadRequest{
		Id: id,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedSpans{spans: rsp.Spans, reverse: true})

	for _, span := range rsp.Spans {
		if len(span.Annotations) == 0 {
			continue
		}
		sort.Sort(sortedAnns{span.Annotations})
	}

	render(w, r, "trace", map[string]interface{}{
		"Id":    id,
		"Spans": rsp.Spans,
	})
}
