package tpl

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/errgo.v1"
)

// Templates is a map of all parsed templates.
type Templates map[string]*template.Template

// Must helper for Load()
func Must(t Templates, err error) Templates {
	if err != nil {
		panic(err)
	}
	return t
}

var bufPool = newBufPool()

// Render the template into a writer. Uses a buffer pool to not be inefficient.
func (t Templates) Render(w http.ResponseWriter, name string, data interface{}) error {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	tmpl, ok := t[name]
	if !ok {
		return errgo.Newf("Template named: %s does not exist", name)
	}

	err := tmpl.ExecuteTemplate(buf, "", data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, buf)
	return err
}

// Load .tpl template files from dir, using layout as a base. Panics on
// failure to parse/load anything.
func Load(dir, partialDir, layout string, funcs template.FuncMap) (Templates, error) {
	tpls := make(Templates)

	b, err := ioutil.ReadFile(filepath.Join(dir, layout))
	if err != nil {
		return nil, errgo.Notef(err, "Could not load layout")
	}

	layoutTpl, err := template.New("").Funcs(funcs).Parse(string(b))
	if err != nil {
		return nil, errgo.Notef(err, "Failed to parse layout")
	}

	err = filepath.Walk(partialDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasPrefix(filepath.Base(path), "_") {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return errgo.Notef(err, "Could not create relative path")
		}
		name := removeExtension(filepath.Base(rel))

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return errgo.Notef(err, "Failed to load partial (%s)", rel)
		}

		_, err = layoutTpl.New(name).Parse(string(b))
		if err != nil {
			return errgo.Notef(err, "Failed to parse partial: (%s)", rel)
		}

		return nil
	})

	if err != nil {
		return nil, errgo.Notef(err, "Failed to load partials")
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if path == filepath.Join(dir, layout) || strings.HasPrefix(filepath.Base(path), "_") {
			return nil
		}

		if err != nil {
			return errgo.Notef(err, "Could not walk directory")
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return errgo.Notef(err, "Could not create relative path")
		}

		name := removeExtension(rel)

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return errgo.Notef(err, "Failed to load template (%s)", rel)
		}

		clone, err := layoutTpl.Clone()
		if err != nil {
			return errgo.Notef(err, "Failed to clone layout")
		}
		t, err := clone.New("yield").Parse(string(b))
		if err != nil {
			return errgo.Notef(err, "Failed to parse template (%s)", rel)
		}

		tpls[name] = t
		return nil
	})

	if err != nil {
		return nil, errgo.Notef(err, "Failed to load templates")
	}

	return tpls, nil
}

func removeExtension(path string) string {
	if dot := strings.Index(path, "."); dot >= 0 {
		return path[:dot]
	}
	return path
}

type bufPoolS struct {
	*sync.Pool
}

func newBufPool() bufPoolS {
	return bufPoolS{Pool: &sync.Pool{
		New: func() interface{} { return bytes.NewBuffer(make([]byte, 4096)) },
	}}
}
func (bp bufPoolS) Get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}
func (bp bufPoolS) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	bp.Pool.Put(buf)
}
