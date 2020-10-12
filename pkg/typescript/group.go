package typescript

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
)

// Group represents a list of Code items, separated by tokens with an optional
// open and close token.
type Group struct {
	name      string
	items     []Code
	open      string
	close     string
	separator string
	multi     bool
}

func (g *Group) isNull(f *File) bool {
	if g == nil {
		return true
	}
	if g.open != "" || g.close != "" {
		return false
	}
	for _, c := range g.items {
		if !c.isNull(f) {
			return false
		}
	}
	return true
}

func (g *Group) render(f *File, w io.Writer, s *Statement) error {
	if g.name == "block" && s != nil {
		// Special CaseBlock format for then the previous item in the statement
		// is a Case group or the default keyword.
		prev := s.previous(g)
		grp, isGrp := prev.(*Group)
		tkn, isTkn := prev.(token)
		if isGrp && grp.name == "case" || isTkn && tkn.content == "default" {
			g.open = ""
			g.close = ""
		}
	}
	if g.open != "" {
		if _, err := w.Write([]byte(g.open)); err != nil {
			return err
		}
	}
	isNull, err := g.renderItems(f, w)
	if err != nil {
		return err
	}
	if !isNull && g.multi && g.close != "" {
		s := "\n"
		if g.separator == "," {
			s = ",\n"
		}
		if _, err := w.Write([]byte(s)); err != nil {
			return err
		}
	}
	if g.close != "" {
		if _, err := w.Write([]byte(g.close)); err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) renderItems(f *File, w io.Writer) (isNull bool, err error) {
	first := true
	for _, code := range g.items {

		if code == nil || code.isNull(f) {
			continue
		}
		if g.name == "values" {
			if _, ok := code.(Dict); ok && len(g.items) > 1 {
				panic("Error in Values: if Dict is used, must be one item only")
			}
		}
		if !first && g.separator != "" {
			if _, err := w.Write([]byte(g.separator)); err != nil {
				return false, err
			}
		}
		if g.multi {
			if _, err := w.Write([]byte("\n")); err != nil {
				return false, err
			}
		}
		if err := code.render(f, w, nil); err != nil {
			return false, err
		}
		first = false
	}
	return first, nil
}

// Render renders the Group to the provided writer.
func (g *Group) Render(writer io.Writer) error {
	return g.RenderWithFile(writer, NewFile())
}

// GoString renders the Group for testing. Any error will cause a panic.
func (g *Group) GoString() string {
	buf := bytes.Buffer{}
	if err := g.Render(&buf); err != nil {
		panic(err)
	}
	return buf.String()
}

// RenderWithFile renders the Group to the provided writer, using imports from the provided file.
func (g *Group) RenderWithFile(writer io.Writer, file *File) error {
	buf := &bytes.Buffer{}
	if err := g.render(file, buf, nil); err != nil {
		return err
	}
	b, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("Error %s while formatting source:\n%s", err, buf.String())
	}
	if _, err := writer.Write(b); err != nil {
		return err
	}
	return nil
}
