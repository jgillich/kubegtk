package reactive

import (
	"context"
	"reflect"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Box struct {
	Widget
	Orientation gtk.Orientation
	Spacing     int `gtk:"spacing"`
	Children    []Model
}

func (m *Box) Type() reflect.Type {
	return reflect.TypeFor[*gtk.Box]()
}

func (m *Box) Create(ctx context.Context) gtk.Widgetter {
	w := gtk.NewBox(m.Orientation, m.Spacing)
	m.Update(ctx, w)
	return w
}

func (m *Box) Update(ctx context.Context, w gtk.Widgetter) {
	m.update(ctx, m, w, &m.Widget, gtk.BaseWidget(w))

	box := w.(*gtk.Box)
	// does not care about child types
	// mergeChildren(ctx, w, model.Children, func(w gtk.Widgetter) {
	// 	box.Append(w)
	// }, func(w gtk.Widgetter) {
	// 	box.Remove(w)
	// })

	next := box.FirstChild()
	for _, child := range m.Children {
		if next == nil {
			new := createChild(ctx, child)
			box.Append(new)
			continue
		}

		if child.Type() == reflect.TypeOf(next) {
			// child.Update(ctx, next)
			updateChild(next, child)
			next = gtk.BaseWidget(next).NextSibling()
		} else {
			new := createChild(ctx, child)
			gtk.BaseWidget(new).InsertBefore(box, next)
			removeChild(next)
			box.Remove(next)
			next = gtk.BaseWidget(new).NextSibling()
		}
	}

	for {
		if next == nil {
			break
		}
		sibling := gtk.BaseWidget(next).NextSibling()
		box.Remove(next)
		next = sibling
	}
}