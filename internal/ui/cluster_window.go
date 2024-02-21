package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/behavior"
	"github.com/getseabird/seabird/widget"
)

type ClusterWindow struct {
	*widget.UniversalApplicationWindow
	behavior     *behavior.ClusterBehavior
	prefs        *api.Preferences
	navigation   *Navigation
	listView     *ListView
	detailView   *DetailView
	toastOverlay *adw.ToastOverlay
}

func NewClusterWindow(app *gtk.Application, behavior *behavior.ClusterBehavior) *ClusterWindow {
	w := ClusterWindow{
		UniversalApplicationWindow: widget.NewUniversalApplicationWindow(app),
		behavior:                   behavior,
	}
	w.SetIconName("seabird")
	w.SetTitle(fmt.Sprintf("%s - %s", behavior.ClusterPreferences.Value().Name, ApplicationName))
	w.SetDefaultSize(1280, 720)

	var h glib.SignalHandle
	h = w.ConnectCloseRequest(func() bool {
		prefs := behavior.Preferences.Value()
		if err := prefs.Save(); err != nil {
			d := ShowErrorDialog(&w.Window, "Could not save preferences", err)
			d.ConnectUnrealize(func() {
				w.Close()
			})
			w.HandlerDisconnect(h)
			return true
		}
		return false
	})

	w.toastOverlay = adw.NewToastOverlay()
	w.SetContent(w.toastOverlay)

	lpane := gtk.NewPaned(gtk.OrientationHorizontal)
	lpane.SetShrinkStartChild(false)
	lpane.SetShrinkEndChild(false)
	rpane := gtk.NewPaned(gtk.OrientationHorizontal)
	rpane.SetShrinkStartChild(false)
	rpane.SetShrinkEndChild(false)
	rpane.SetHExpand(true)
	lpane.SetEndChild(rpane)
	w.toastOverlay.SetChild(lpane)

	w.detailView = NewDetailView(&w.Window, behavior.NewRootDetailBehavior())
	nav := adw.NewNavigationView()
	nav.Add(w.detailView.NavigationPage)
	nav.SetSizeRequest(350, 350)
	rpane.SetEndChild(nav)

	w.listView = NewListView(&w.Window, behavior.NewListBehavior())
	rpane.SetStartChild(w.listView)
	sw, _ := w.listView.SizeRequest()
	rpane.SetPosition(sw)

	w.navigation = NewNavigation(behavior)
	lpane.SetStartChild(w.navigation)
	sw, _ = w.navigation.SizeRequest()
	lpane.SetPosition(sw)

	w.createActions()

	return &w
}

func (w *ClusterWindow) createActions() {
	newWindow := gio.NewSimpleAction("newWindow", nil)
	newWindow.ConnectActivate(func(_ *glib.Variant) {
		prefs, err := api.LoadPreferences()
		if err != nil {
			ShowErrorDialog(&w.Window, "Could not load preferences", err)
			return
		}
		prefs.Defaults()
		NewWelcomeWindow(w.Application(), w.behavior.Behavior).Show()
	})
	w.AddAction(newWindow)
	w.Application().SetAccelsForAction("win.newWindow", []string{"<Ctrl>N"})

	disconnect := gio.NewSimpleAction("disconnect", nil)
	disconnect.ConnectActivate(func(_ *glib.Variant) {
		w.ActivateAction("newWindow", nil)
		w.Close()
	})
	w.AddAction(disconnect)
	w.Application().SetAccelsForAction("win.disconnect", []string{"<Ctrl>Q"})

	action := gio.NewSimpleAction("prefs", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		prefs := NewPreferencesWindow(w.behavior)
		prefs.SetTransientFor(&w.Window)
		prefs.Show()
	})
	w.AddAction(action)

	action = gio.NewSimpleAction("about", nil)
	action.ConnectActivate(func(_ *glib.Variant) {
		NewAboutWindow(&w.Window).Show()
	})
	w.AddAction(action)

	filterNamespace := gio.NewSimpleAction("filterNamespace", glib.NewVariantType("s"))
	filterNamespace.ConnectActivate(func(parameter *glib.Variant) {
		text := strings.Trim(fmt.Sprintf("%s ns:%s", w.behavior.SearchText.Value(), parameter.String()), " ")
		w.behavior.SearchText.Update(text)
	})
	actionGroup := gio.NewSimpleActionGroup()
	actionGroup.AddAction(filterNamespace)
	w.InsertActionGroup("list", actionGroup)
}