// gLibMenu.go

/*
	Source file auto-generated on Tue, 15 Sep 2020 12:45:12 using Gotk3ObjHandler v1.6.2 ©2018-20 H.F.M
	This software use gotk3 that is licensed under the ISC License:
	https://github.com/gotk3/gotk3/blob/master/LICENSE

	Copyright ©2020 H.F.M:

		- GlibMenuStructure containing a set of [] * glib.Menu
		  and allowing the management of separators and sub-menus.

		- How to use at the bottom ...

	This program comes with absolutely no warranty. See the The MIT License (MIT) for details:
	https://opensource.org/licenses/mit-license.php
*/

package gLibMenu

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// GlibMenuStrut: structure containing a set of []*glib.Menu
// and allowing the management of separators and sub-menus.
// At this time Gotk3 does not handle complex actions correctly
// (checkbox ...). Note: Separator look like blank lines, to make
// them looking like underline, use CSS;
type GlibMenuStruct struct {
	Menus      []sMenu
	Actions    []sAction
	MenuButton *gtk.MenuButton

	app      *gtk.Application
	currMenu int

	sep     string
	maxChar int
}
type sMenu struct {
	Menu       *glib.Menu
	skip       bool
	SubLabel   string
	parentMenu int
}
type sAction struct {
	Label, Action string
	Callback      func()
	GAction       *glib.SimpleAction
	toMenu        int
}

/*
func (gms *GlibMenuStruct) addSep() (out string) {
	for i := 0; i < maxChar; i++ {
		out += gms.sep
	}
}
func (gms *GlibMenuStruct) updMaxChar(inStr string) string {
	if len(inStr) > gms.maxChar {
		gms.maxChar = len(inStr)
	}
	return inStr
}
*/

// GlibMenuStrutNew: Create structure to hold a GMenu.
func GlibMenuStrutNew(application *gtk.Application, menuButton *gtk.MenuButton) (gms *GlibMenuStruct, err error) {
	gms = new(GlibMenuStruct)
	gms.app = application
	gms.MenuButton = menuButton
	gms.sep = "―"

	if menu := glib.MenuNew(); menu != nil {
		gms.Menus = append(gms.Menus, sMenu{
			Menu:     menu,
			skip:     true,
			SubLabel: "",
		})
		gms.currMenu = len(gms.Menus) - 1
	} else {
		return nil, fmt.Errorf("Unable to create glib menu")
	}
	return
}

// SectionAdd: Add a section, sectionLbl may be the title of this
// new section or ignored for a simple blank line.
func (gms *GlibMenuStruct) SectionAdd(label ...string) error {

	title := ""
	if len(label) > 0 {
		title = label[0]
	}
	if menu := glib.MenuNew(); menu != nil {
		gms.Menus = append(gms.Menus, sMenu{
			Menu:     menu,
			skip:     false,
			SubLabel: title,
		})
		gms.currMenu = len(gms.Menus) - 1
	} else {
		return fmt.Errorf("Unable to create glib menu")
	}
	return nil
}

// SubMenuAdd: Add a sub-menu, sectionLbl may be the title of the
// section that hold this sub-menu.
func (gms *GlibMenuStruct) SubMenuAdd(label string /*, sectionLbl ...string*/) (err error) {
	// sectionTitle := ""
	// if len(sectionLbl) > 0 {
	// 	sectionTitle = sectionLbl[0]
	// }
	// New section to change parent menu
	// if err = gms.SectionAdd(sectionTitle); err == nil {

	if menu := glib.MenuNew(); menu != nil {
		gms.Menus = append(gms.Menus, sMenu{
			Menu:       menu,
			skip:       true,
			SubLabel:   label,
			parentMenu: gms.currMenu,
		})

		gms.Menus[gms.currMenu].Menu.AppendSubmenu(label, &menu.MenuModel)
		gms.currMenu++
	} else {
		err = fmt.Errorf("Unable to create glib menu")
	}
	// }
	return
}

// SubMenuPrev: Return to the previous menu, generally used go up one level of a submenu.
func (gms *GlibMenuStruct) SubMenuPrev() {
	gms.currMenu = gms.Menus[gms.currMenu].parentMenu
}

// ActionAdd: Add a *glib.SimpleAction
func (gms *GlibMenuStruct) ActionAdd(label, action string, callback func()) {
	gms.Actions = append(gms.Actions, sAction{
		Label:    label,
		Callback: callback,
		Action:   action,
		GAction:  new(glib.SimpleAction),
		toMenu:   gms.currMenu,
	})
}

// BuilGMenu: Once finalized, this one permit to create GMenu to be displayed.
func (gms *GlibMenuStruct) BuilGMenu() {

	for _, act := range gms.Actions {
		gms.Menus[act.toMenu].Menu.Append(act.Label, "app."+act.Action)
		act.GAction = glib.SimpleActionNew(act.Action, nil)
		act.GAction.Connect("activate", act.Callback)
		gms.app.AddAction(act.GAction)
	}
	for _, menu := range gms.Menus {
		if !menu.skip {
			// gms.Menus[0].Menu.AppendSectionWithoutLabel( &menu.Menu.MenuModel)
			gms.Menus[0].Menu.AppendSection(menu.SubLabel, &menu.Menu.MenuModel)
		}
	}
	gms.MenuButton.SetMenuModel(&gms.Menus[0].Menu.MenuModel)
}

/*								=:= EXAMPLE =:=

if gms, err = GlibMenuStrutNew(app, *gtk.MenuButton); err == nil {
	gms.ActionAdd("Label0", "action0", callback0)

	if err = gms.SubMenuAdd(sts["file"]); err == nil {
		gms.ActionAdd("sub-Label1", "sub-action1", callback1)
		gms.ActionAdd("sub-Label2", "sub-action2", func() { callback2(arg1, arg2, arg3) })
		gms.ActionAdd("sub-Label3", "sub-action3", callback3)
		gms.SubMenuPrev()

		if err = gms.SectionAdd(); err == nil {
			gms.ActionAdd("Label1", "action1", callback4)
			gms.ActionAdd("Label2", "action2", callback5)

			if err = gms.SectionAdd("___"); err == nil {
				gms.ActionAdd("About", "about", about)

				gms.BuilGMenu()
			}
		}
	}
}

*/
