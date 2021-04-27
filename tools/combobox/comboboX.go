// combobox-text.go

/*
	Copyright ©2020 H.F.M - Gotk3 GtkComboBox/Text handling structure library v1.0 https://github.com/hfmrow
	This program comes with absolutely no warranty. See the The MIT License (MIT) for details:
	https://opensource.org/licenses/mit-license.php
*/

package gtk3_import

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	gimc "github.com/hfmrow/gtk3_import/misc"
)

// TODO Add an internal logger for errors ?

type ComboBoXStruct struct {
	Options   *ComboBoXOpt
	ListStore *gtk.ListStore
	TreeModel *gtk.TreeModel

	// Store the type of GtkComboBox/Text for further usages
	ComboBoxX interface{}

	CurrentEntry string

	signHdl glib.SignalHandle
}

// ComboBoXOpt: Hold available options for 'ComboBoXStruct'
type ComboBoXOpt struct {
	Sort,
	Prepend,
	Descending bool // Sort direction
	ColActive int

	// concern only GtkComboBoxText
	Editable,
	PopupRemove,
	AddOnEnter bool

	// CurrentEntryPtr: Will be updated every time the 'combobox' or the entry 'control'
	// receives the 'modified' signal. Usefull for auto updating a caller variable as &ptr
	CurrentEntryPtr *string
	// callback function for 'enter' event
	AddOnEnterCallback func(item *string) bool
	// callback function for Combobox 'changed' event
	CallbackChanged func(obj interface{})
	// callback function for Entry 'changed' event
	CallbackEntryChanged func(obj interface{})

	PopupLabelAll,
	PopupLabelEntry string

	// Realtime recording flag
	setRealTimeRecordDone bool
}

func ComboBoXOptNew() *ComboBoXOpt {
	cbxo := new(ComboBoXOpt)
	cbxo.CurrentEntryPtr = new(string)
	cbxo.PopupLabelAll = "_All"
	cbxo.PopupLabelEntry = "_Entry"
	return cbxo
}

// ComboBoXStructNew: Create a new structure that handle the methods to process
// GtkComboBox & GtkComboBoxText objects transparently, some options are not
// available for simple ComboBox like 'Editable', 'AddOnEnter', 'PopupRemove'.
// A realtime recorded modification is available. This method should be followed
// by 'Setup()' & 'Fill()' methods for a correct structure's initialization.
// 'Options' structure can be retrieved using 'GetOptions()' and filled as your
// wishes before passing it to 'Setup()' (optional).
func ComboBoXStructNew(comboBoX interface{}) (cbxs *ComboBoXStruct) {

	cbxs = new(ComboBoXStruct)
	cbxs.ComboBoxX = comboBoX
	cbxs.Options = ComboBoXOptNew()

	return
}

// ComboBoXFillWithOptionsNew: all in one function
// - Create the structure.
// - Initialization with the given options.
// - Fill comboBox / Text with entries.
func ComboBoXFillWithOptionsNew(comboBoX interface{},
	inList []string,
	activeId string,
	options ...*ComboBoXOpt) (cbxs *ComboBoXStruct, err error) {

	cbxs = ComboBoXStructNew(comboBoX)

	if len(options) > 0 {
		cbxs.Options = options[0]
	} else {
		cbxs.Options = ComboBoXOptNew()
	}

	if err = cbxs.Setup(); err == nil {
		err = cbxs.Fill(inList, activeId)
	}
	return
}

// GetOptions: get option structure to be completed (whether somthings need to be
// specified / modified), before using 'Init'
func (cbxs *ComboBoXStruct) GetOptions() *ComboBoXOpt {
	return cbxs.Options
}

// Setup: Applying options before setting up structure.
// NOTE: "has_entry" property if used, must be set at creation (glade)
func (cbxs *ComboBoXStruct) Setup(options ...*ComboBoXOpt) (err error) {

	var (
		entry *gtk.Entry

		getEntry = func(e *gtk.Entry) string {
			txt, err := e.GetText()
			if err != nil {
				err = errors.New("Unable to GetText")
				log.Printf("*ComboBoXStruct/Setup: %v\n", err)
			}
			return txt
		}
	)
	if len(options) > 0 {
		cbxs.Options = options[0]
	}
	// Get ListStore depending of ComboBox type,
	// allow to have unified methods for the two kind of ComboBox.
	cbxs.toTreeX()
	switch cb := cbxs.ComboBoxX.(type) {

	case *gtk.ComboBox:

		cb.SetIDColumn(cbxs.Options.ColActive)
	case *gtk.ComboBoxText:

		cb.SetEntryTextColumn(cbxs.Options.ColActive)
		cb.SetIDColumn(cbxs.Options.ColActive)
		cb.SetCanFocus(cbxs.Options.Editable)
		if cb.GetHasEntry() {

			if entry, err = cb.GetEntry(); err == nil {
				if cbxs.Options.Editable {
					// Entry changed signal
					entry.Connect("changed", func(entry *gtk.Entry) {

						txt := getEntry(entry)
						cbxs.CurrentEntry = txt
						*cbxs.Options.CurrentEntryPtr = txt

						if cbxs.Options.CallbackEntryChanged != nil {
							cbxs.Options.CallbackEntryChanged(entry)
						}
					})
					if cbxs.Options.AddOnEnter {
						// Enter pressed signal
						entry.Connect("activate", func(entry *gtk.Entry) {
							txt := getEntry(entry)
							if cbxs.Options.AddOnEnterCallback != nil {
								if !cbxs.Options.AddOnEnterCallback(&txt) {
									return
								}
							}
							// Add if not exist
							cbxs.AddSetEntry(txt)
							cbxs.CurrentEntry = txt
							*cbxs.Options.CurrentEntryPtr = cbxs.CurrentEntry
							cbxs.Sort()
							cb.SetActiveID(txt)
						})
					}
				}
				if cbxs.Options.PopupRemove {
					// Popup menu (populate) signal
					entry.Connect("populate-popup", func(e *gtk.Entry, w *gtk.Widget) {

						pop := gimc.PopupMenuStructNew()
						pop.AddItem("Remove", nil, pop.OPT_SEPARATOR)
						pop.AddItem(cbxs.Options.PopupLabelEntry, func() {

							cbxs.RemoveEntry(cb.GetActiveText())
						}, pop.OPT_NORMAL)
						pop.AddItem(cbxs.Options.PopupLabelAll, func() {

							cbxs.Clear()
						}, pop.OPT_NORMAL)

						menu := &gtk.Menu{gtk.MenuShell{gtk.Container{*w}}}
						pop.AppendToExistingMenu(menu)
					})
				}
				entry.SetCanFocus(cbxs.Options.Editable)
				entry.Editable.SetEditable(cbxs.Options.Editable)
			}
		}
	default:
		err = errors.New("Unable to setup ComboBox/Text")
	}
	if err != nil {
		err = fmt.Errorf("Setup: %v", err)
	}
	return
}

// Fill: should only be used after calling the 'Setup' method.
// All previous entries will be deleted before the operation.
// The 'activeId' argument define current selected entry (ID).
func (cbxs *ComboBoXStruct) Fill(inList []string, activeId string) (err error) {

	cbxs.Clear()

	// Fill cbx
	if inList != nil {
		if len(inList) > 0 {
			for _, item := range inList {
				cbxs.ListStoreAdd(item)
			}
			cbxs.Sort()
		}
	}

	switch cb := cbxs.ComboBoxX.(type) {
	case *gtk.ComboBox:

		cbxs.signHdl = cb.Connect("changed",
			func(c *gtk.ComboBox) {
				cbxs.CurrentEntry = c.GetActiveID()
				*cbxs.Options.CurrentEntryPtr = cbxs.CurrentEntry
				if cbxs.Options.CallbackChanged != nil {
					cbxs.Options.CallbackChanged(c)
				}
			})
		cb.HandlerBlock(cbxs.signHdl)
		defer cb.HandlerUnblock(cbxs.signHdl)
		cbxs.CurrentEntry = activeId
		*cbxs.Options.CurrentEntryPtr = cbxs.CurrentEntry

		cb.SetActiveID(activeId)

	case *gtk.ComboBoxText:

		cbxs.signHdl = cb.Connect("changed",
			func(c *gtk.ComboBoxText) {
				cbxs.CurrentEntry = c.GetActiveID()
				*cbxs.Options.CurrentEntryPtr = cbxs.CurrentEntry
				if cbxs.Options.CallbackChanged != nil {
					cbxs.Options.CallbackChanged(c)
				}
			})
		cb.HandlerBlock(cbxs.signHdl)
		defer cb.HandlerUnblock(cbxs.signHdl)
		cbxs.CurrentEntry = activeId
		*cbxs.Options.CurrentEntryPtr = cbxs.CurrentEntry

		cb.SetActiveID(activeId)
	}
	return
}

// Clear: Remove content of ComboBox/Text
func (cbxs *ComboBoXStruct) Clear() {

	cbxs.ListStoreAdd("")
	cbxs.ListStore.Clear()
}

// RemoveEntry:
func (cbxs *ComboBoXStruct) RemoveEntry(item string) {

	var idx int

	switch cb := cbxs.ComboBoxX.(type) {
	case *gtk.ComboBox:

		idx = cb.GetActive()
		defer cb.SetActive(idx)
	case *gtk.ComboBoxText:

		idx = cb.GetActive()
		defer cb.SetActive(idx)
	}

	list := cbxs.GetAllEntries()
	cbxs.Clear()
	for _, i := range list {
		if item != i {
			cbxs.ListStoreAdd(i)
		}
	}
	cbxs.Sort()
}

// GetActive: Return the active cell.
func (cbxs *ComboBoXStruct) GetActive() (index int, id string) {

	switch cb := cbxs.ComboBoxX.(type) {
	case *gtk.ComboBox:

		index = cb.GetActive()
		id = cb.GetActiveID()
	case *gtk.ComboBoxText:

		index = cb.GetActive()
		id = cb.GetActiveID()
	}
	return
}

// Sort: Order can be set in 'Options'
func (cbxs *ComboBoXStruct) Sort() {

	if cbxs.Options.Sort {

		in := cbxs.GetAllEntries()

		cbxs.Clear()

		if cbxs.Options.Descending {
			sort.SliceStable(in,
				func(i, j int) bool {
					return strings.ToLower(in[i]) >
						strings.ToLower(in[j])
				})

		} else {
			sort.SliceStable(in,
				func(i, j int) bool {
					return strings.ToLower(in[i]) <
						strings.ToLower(in[j])
				})
		}

		for _, item := range in {
			cbxs.ListStoreAdd(item)
		}
	}
}

// ListStoreAdd: Append or Prepend (via 'ComboBoXOpt') an 'item', then
// set ActiveId to it. Unlike 'AddSetEntry',this one, does not verify
// wether 'item' already exist or not. 'Sort' method is not used.
func (cbxs *ComboBoXStruct) ListStoreAdd(item string) {

	var (
		err  error
		iter *gtk.TreeIter
	)

	if cbxs.Options.Prepend {
		iter = cbxs.ListStore.Prepend()
	} else {
		iter = cbxs.ListStore.Append()
	}

	err = cbxs.ListStore.SetValue(iter, cbxs.Options.ColActive, item)
	if err != nil {
		log.Printf("ListStoreAdd: %v\n", err)
	}

	cbxs.setActiveIter(iter)
}

// AddSetEntry: Adds a new entry if it does not exist to ComboBox / Text.
// Returns -1 if 'item' does not exist or, set 'ActiveId' to it and return
// the position of the existing element. 'Sort' method is not used.
func (cbxs *ComboBoXStruct) AddSetEntry(item string, position ...int) int {

	var (
		err error

		// idx ,
		existAtPos,
		posCounter int
		pos = -1

		iter *gtk.TreeIter
	)

	if len(position) > 0 {
		pos = position[0]
	}
	if existAtPos = cbxs.Find(item); existAtPos == -1 {

		if pos > -1 {
			// Append or prepend at specific position
			cbxs.ListStore.ForEach(
				func(model *gtk.TreeModel, path *gtk.TreePath, inIter *gtk.TreeIter) bool {

					if posCounter != pos {
						posCounter++
						return false
					}
					switch cbxs.Options.Prepend {
					case true:

						iter = cbxs.ListStore.InsertAfter(inIter)
						err = cbxs.ListStore.SetValue(iter, cbxs.Options.ColActive, item)
					case false:

						iter = cbxs.ListStore.InsertBefore(inIter)
						err = cbxs.ListStore.SetValue(iter, cbxs.Options.ColActive, item)
					}
					return true
				})

			if pos > posCounter {
				err = fmt.Errorf("The requested position is not within the available ranges: %v > %v", pos, posCounter)
			}

		} else {
			// Append or prepend
			switch cbxs.Options.Prepend {
			case true:

				iter = cbxs.ListStore.Prepend()
				err = cbxs.ListStore.SetValue(iter, cbxs.Options.ColActive, item)
			case false:

				iter = cbxs.ListStore.Append()
				err = cbxs.ListStore.SetValue(iter, cbxs.Options.ColActive, item)
			}
		}

		if err != nil {
			log.Printf("AddSetEntry: %v\n", err.Error())
		} else {
			cbxs.setActiveIter(iter)
		}
	}

	return existAtPos
}

// GetAllEntries: get all entries
// handle both, GtkComboBox or GtkComboBoxText
func (cbxs *ComboBoXStruct) GetAllEntries() (out []string) {

	var (
		err    error
		val    *glib.Value
		valStr string
	)

	cbxs.TreeModel.ForEach(
		func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter) bool {
			if val, err = model.GetValue(iter, cbxs.Options.ColActive); err == nil {
				if valStr, err = val.GetString(); err == nil {
					out = append(out, valStr)
					return false
				}
			}
			if err != nil {
				log.Printf("GetAllEntries/ForEach: %v\n", err)
			}
			return true
		})

	return
}

// Find: find string value. Return -1 if nothing found
// handle both, GtkComboBox or GtkComboBoxText
func (cbxs *ComboBoXStruct) Find(item string) int {

	var (
		err       error
		glibValue *glib.Value
		valStr    string
		count     int
	)

	iter, ok := cbxs.TreeModel.GetIterFirst()
	for ok {
		if glibValue, err = cbxs.TreeModel.GetValue(iter, cbxs.Options.ColActive); err == nil {
			if valStr, err = glibValue.GetString(); err == nil {
				if valStr == item {
					return count
					break
				}
				count++
				ok = cbxs.TreeModel.IterNext(iter)
			}
		}
		if err != nil {
			log.Printf("Find: %v\n", err.Error())
		}
	}

	return -1
}

// setActiveIter:
func (cbxs *ComboBoXStruct) setActiveIter(iter *gtk.TreeIter) {

	switch cb := cbxs.ComboBoxX.(type) {
	case *gtk.ComboBox:

		cb.SetActiveIter(iter)

	case *gtk.ComboBoxText:

		cb.SetActiveIter(iter)
	}
}

// toTreeX: retrieve GtkTreeModel and GtkListStore, handle both,
// GtkComboBox or GtkComboBoxText transparently. Not exported
// because executed during the creation of the structure, the
// objects obtained are stored there and are used internally,
// you can use (but not change) them if you wish.
func (cbxs *ComboBoXStruct) toTreeX() {

	var (
		err      error
		iMdl     gtk.ITreeModel
		cRndrTxt *gtk.CellRendererText
	)

	switch cb := cbxs.ComboBoxX.(type) {

	case *gtk.ComboBox:

		if cbxs.ListStore, err = gtk.ListStoreNew(glib.TYPE_STRING); err == nil {
			cb.SetModel(cbxs.ListStore)
			if iMdl, err = cb.GetModel(); err == nil {
				cbxs.TreeModel = iMdl.ToTreeModel()

				if cRndrTxt, err = gtk.CellRendererTextNew(); err == nil {
					cb.PackStart(cRndrTxt, true)
					cb.AddAttribute(cRndrTxt, "text", cbxs.Options.ColActive)

				}
			}
		}

	case *gtk.ComboBoxText:

		if iMdl, err = cb.GetModel(); err == nil {
			cbxs.TreeModel = iMdl.ToTreeModel()
			cbxs.ListStore = (iMdl).(*gtk.ListStore)
		}
	}

	if err != nil {
		log.Printf("toTreeX: %v\n", err)
	}
}

/*
 * Old function to preserve compatibility and not break previous code.
 */

// ComboBoxTextAddSetEntry: Add newEntry if not exist to ComboBoxText, Option: prepend:bool.
// Get index and set cbxText at it if already exist.
func ComboBoxTextAddSetEntry(cbxEntry *gtk.ComboBoxText, newEntry string, prepend ...bool) (existAtPos int) {
	var prependEntry bool
	var count int
	var iter *gtk.TreeIter
	var ok bool
	existAtPos = -1
	if len(prepend) > 0 {
		prependEntry = prepend[0]
	}

	log.Printf("[ComboBoxTextAddSetEntry], This function is outdated, please use %s instead\n", "'ComboBoXStruct'")

	iTreeModel, err := cbxEntry.GetModel()
	model := iTreeModel.ToTreeModel()
	iter, ok = model.GetIterFirst()
	for ok {
		if glibValue, err := model.GetValue(iter, 0); err == nil {
			if entry, err := glibValue.GetString(); err == nil {
				if entry == newEntry {
					existAtPos = count
					break
				}
				count++
				ok = model.IterNext(iter)
			}
		}
		if err != nil {
			fmt.Errorf("ComboBoxTextAddSetEntry: %s", err.Error())
		}
	}
	if existAtPos == -1 {
		switch {
		case prependEntry:
			cbxEntry.PrependText(newEntry)
		default:
			cbxEntry.AppendText(newEntry)
		}
	} else {
		cbxEntry.SetActiveIter(iter)
	}
	return existAtPos
}
