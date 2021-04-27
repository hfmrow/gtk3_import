// gladeXmlParser.go

/*
	Copyright ©2018-19 H.F.M. MIT license - GladeXmlParser v2.3 Library
	This program comes with absolutely no warranty. See the The MIT License (MIT) for details:
	https://opensource.org/licenses/mit-license.php

	parseGladeXmlFile: Create a parsed glade structure containing
	all objects with property, signals and packing information.
*/

package gtk3_import

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	glsg "github.com/hfmrow/gen_lib/strings"
)

var Name = "GladeXmlParser"
var Vers = "v2.0"
var Descr = "Golang parser for glade xml files"
var Creat = "H.F.M"
var YearCreat = "2019"
var LicenseShort = "This program comes with absolutely no warranty.\nSee the The MIT License (MIT) for details:\nhttps://opensource.org/licenses/mit-license.php"
var LicenseAbrv = "License (MIT)"
var LicenseUrl = "https://opensource.org/licenses/mit-license.php"
var Repository = "github.com/..."

type GtkInterface struct {
	GXPVersion    string
	ObjectsCount  int
	UpdatedOn     string
	GladeFilename string
	Requires      requires
	Objects       []GtkObject
	Comments      []string

	NamingCSS          bool
	NamingCSSLowerCase bool
	NamingCSSForce     bool
	NamingCSSClear     bool
	SkipObjectNaming   []string

	// private
	objectsLoaded   bool
	skipNoId        bool
	skipLoweAtFirst bool
	idxLine         int
	prevIdxLine     int
	xmlSource       []string
	getValuesReg    *regexp.Regexp
	eol             string
}

type GtkObject struct {
	Class        string
	Id           string
	Property     []GtkProps
	Signal       []GtkProps
	Packing      []GtkProps
	alreadyNamed bool
}

type GtkProps struct {
	Name         string
	Value        string
	Swapped      string
	Translatable string
	line         int
}
type requires struct {
	Lib     string
	Version string
}

// GladeXmlParserNew: Create new parsed glade structure containing
// all objects with property, signals, packing information.
func GladeXmlParserNew(filename string, skipObjectNaming []string, skipNoId, skipLoweAtFirst bool) (iFace *GtkInterface, err error) {
	iFace = new(GtkInterface)

	iFace.SkipObjectNaming = skipObjectNaming

	err = iFace.StructureSetup(filename, skipNoId, skipLoweAtFirst)
	return
}

// StructureSetup: Setup a parser structure containing all objects with property,
// signals, packing information.
// * Alternative use of GladeXmlParserNew() * in the case where we need
// to import the structure rather than using it via the imported library.
func (iFace *GtkInterface) StructureSetup(gladeFilename string, skipNoId, skipLoweAtFirst bool) (err error) {
	var data []byte

	if len(iFace.SkipObjectNaming) == 0 {
		log.Fatal(`GtkInterface.SkipObjectNaming is empty: not loaded ...
*gladeParserProj / ObjectFiltersNew (filename string) to load it. Must
be assigned to this structure from the calling function..`)
	}

	iFace.GladeFilename = gladeFilename
	iFace.GXPVersion = fmt.Sprintf("%s %s", Name, Vers)
	iFace.getValuesReg = regexp.MustCompile(`"(.*?)"|>(.*?)<`)
	iFace.skipNoId = skipNoId
	iFace.skipLoweAtFirst = skipLoweAtFirst

	if data, err = iFace.readGladeXmlFile(); err == nil {
		iFace.sanitizeXml(data) // and store lines to structure
		iFace.parseGladeXmlFile()
		iFace.ObjectsCount = len(iFace.Objects)
		iFace.objectsLoaded = true
	}
	return
}

// readGladeXmlFile:
func (iFace *GtkInterface) readGladeXmlFile() (data []byte, err error) {
	var ok bool
	data, err = ioutil.ReadFile(iFace.GladeFilename)
	if err == nil {
		iFace.eol = getTextEOL(data)
		for _, line := range strings.Split(string(data), iFace.eol) {
			if strings.Contains(line, `<requires lib="gtk+"`) { // Check for gtk+ xml file format ...
				ok = true
				break
			}
		}
		if !ok {
			return data, errors.New("Bad file format: " + filepath.Base(iFace.GladeFilename))
		}
	}
	return
}

// parseGladeXmlFile:
func (iFace *GtkInterface) parseGladeXmlFile() (err error) {
	var line, tmpID string
	var values []string

	// To match single line object declaration
	regSingleLineObj := regexp.MustCompile(`<object class=.*id=.*\/>`)
	// To match multi lines object declaration
	regMultiLinesObj := regexp.MustCompile(`<object class=.*id=.*>`)

	for iFace.idxLine = 0; iFace.idxLine < len(iFace.xmlSource); iFace.idxLine++ {
		values, line = iFace.getValues()
		if len(values) > 1 { // Check for ID
			tmpID = values[1]
		} else {
			tmpID = ""
		}
		if !(iFace.skipNoId && len(tmpID) == 0) { // Exclude no name objects if requires
			if !(iFace.skipLoweAtFirst && glsg.LowercaseAtFirst(tmpID)) { // Exclude lower at first char objects if requires
				switch {
				case strings.Contains(line, `<requires lib="`): // requires
					iFace.Requires.Lib = values[0]
					iFace.Requires.Version = tmpID

				case regSingleLineObj.MatchString(line): // Single line object declaration
					tmpGtkObj := GtkObject{
						Class:    values[0],
						Id:       tmpID,
						Property: []GtkProps{},
						Signal:   []GtkProps{},
						Packing:  []GtkProps{}}
					iFace.Objects = append(iFace.Objects, tmpGtkObj)

				case regMultiLinesObj.MatchString(line): // object declaration with properties
					iFace.idxLine++ // Jump to start prop line
					tmpGtkObj := GtkObject{
						Class:    values[0],
						Id:       tmpID,
						Property: []GtkProps{},
						Signal:   []GtkProps{},
						Packing:  []GtkProps{}}
					iFace.readObject(&tmpGtkObj)
					iFace.Objects = append(iFace.Objects, tmpGtkObj)
				}
			}

			/* DEBUG purpose: In a case where DLV debugger can not recover the value of some slice structures,here:
			"Property", "Signal", "Packing" look like empty, it is really boring. But values still available in code ...*/
			// var tmpGtkObj GtkObject
			// fmt.Printf("%#v\n", iFace.Objects[len(iFace.Objects)-1].Property)
			// fmt.Printf("%#v\n", iFace.Objects[len(iFace.Objects)-1].Packing)
			// theObject := iFace.readObject(GtkObject{Class: values[0], Id: values[1], Property: []GtkProps{}, Signal: []GtkProps{}, Packing: []GtkProps{}})
			// tmpGtkObj.Class = theObject.Class
			// tmpGtkObj.Id = theObject.Id
			// tmpGtkObj.Property = make([]GtkProps, len(theObject.Property))
			// tmpGtkObj.Signal = make([]GtkProps, len(theObject.Signal))
			// tmpGtkObj.Packing = make([]GtkProps, len(theObject.Packing))
			// copy(tmpGtkObj.Property, theObject.Property)
			// copy(tmpGtkObj.Signal, theObject.Signal)
			// copy(tmpGtkObj.Packing, theObject.Packing)
			// iFace.Objects = append(iFace.Objects, tmpGtkObj)

		}
	}
	if iFace.NamingCSS {
		if !iFace.NamingCSSClear {
			iFace.namingCss()
		}
		var fi os.FileInfo
		if fi, err = os.Stat(iFace.GladeFilename); err == nil {
			if err = os.Rename(iFace.GladeFilename, iFace.GladeFilename+".goh~"); err == nil {
				err = ioutil.WriteFile(iFace.GladeFilename, []byte(strings.Join(iFace.xmlSource, iFace.eol)), fi.Mode().Perm())
			}
		}
	}
	return
}

// namingCss: Add a widget' Name for css usage. (the object need to have a non empty ID)
// This action backup the glade file and modify the original ...
func (iFace *GtkInterface) namingCss() {
	var nameCSS string
	var addCount int
	var tmpGtkObj GtkObject
	for idxObj := 0; idxObj < len(iFace.Objects); idxObj++ {
		tmpGtkObj = iFace.Objects[idxObj]

		// Check if object can be named
		if iFace.proceedNamingObj(&tmpGtkObj) {

			// widget Naming for CSS usage
			if (!tmpGtkObj.alreadyNamed || iFace.NamingCSSForce) && len(tmpGtkObj.Id) != 0 {
				if iFace.NamingCSSLowerCase {
					nameCSS = strings.ToLower(tmpGtkObj.Id)
				} else {
					nameCSS = tmpGtkObj.Id
				}
			}
			// TODO debug
			// if tmpGtkObj.Id == "WindowInfos" {
			// 	fmt.Println(idxObj)
			// }

			if len(tmpGtkObj.Property) > 0 {
				indent := strings.Split(iFace.xmlSource[tmpGtkObj.Property[0].line+addCount+1], "<")
				addedLine := []string{indent[0] + `<property name="name">` + nameCSS + `</property>`}
				// Insert line
				iFace.xmlSource = append(iFace.xmlSource[:tmpGtkObj.Property[0].line+addCount],
					append(addedLine, iFace.xmlSource[tmpGtkObj.Property[0].line+addCount:]...)...)
				addCount++
			} else {
				fmt.Printf("Warning, Object: %s, %s. Cannot be named.\n", tmpGtkObj.Class, tmpGtkObj.Id)
			}
		}
	}
}

// readObject: read object, signal, property and packing.
func (iFace *GtkInterface) readObject(inObj *GtkObject) {
	var newObjectCount int
	ok := true
	var line string
	var values []string

	// To match single line object declaration
	regSingleLineObj := regexp.MustCompile(`<object class=.*id=.*\/>`)
	// To match multi lines object declaration
	regMultiLinesObj := regexp.MustCompile(`<object class=.*id=.*>`)
	// Obj declaration end
	regEndObj := regexp.MustCompile(`<\/object>`)

	iFace.prevIdxLine = iFace.idxLine
	for iFace.idxLine = iFace.idxLine; iFace.idxLine < len(iFace.xmlSource); iFace.idxLine++ {
		values, line = iFace.getValues()
		switch ok {
		case true:
			ok = iFace.readProps(line, values, inObj)
		case false:
			switch {
			case regSingleLineObj.MatchString(line): // single line object declaration
				// values, line = iFace.getValues()
				// // inObj = new(GtkObject)
				// *inObj = GtkObject{Class: values[0], Id: values[1]}
				iFace.idxLine = iFace.prevIdxLine
				return
			case regMultiLinesObj.MatchString(line): // multi-lines object declaration
				newObjectCount++
			// case strings.Contains(line, `<child>`): // New object to scan
			// ok = true
			case regEndObj.MatchString(line): // object declaration ending
				newObjectCount--
			case newObjectCount == 0: // no more object
				// TODO Previous value: newObjectCount == -1: // Cause missing 1st object in xml file ...
				iFace.idxLine = iFace.prevIdxLine
				return
			case strings.Contains(line, `<packing>`): // <packing>
				iFace.idxLine++
				for iFace.idxLine = iFace.idxLine; iFace.idxLine < len(iFace.xmlSource); iFace.idxLine++ {
					values, line = iFace.getValues()
					ok = readPacking(line, values, inObj)
					if !ok {
						iFace.idxLine = iFace.prevIdxLine
						return
					}
				}
			}
		}
	}
}

// getValues: get all values from a line and clean them
func (iFace *GtkInterface) getValues() (values []string, line string) {
	line = iFace.xmlSource[iFace.idxLine]
	tmpValues := iFace.getValuesReg.FindAllStringSubmatch(line, -1)
	for _, v := range tmpValues {
		values = append(values, strings.Trim(strings.Trim(strings.Trim(v[0], `"`), `>`), `<`))
	}
	return values, line
}

// proceedNamingObj:
func (iFace *GtkInterface) proceedNamingObj(inObj *GtkObject) (ok bool) {

	for _, forbiddenObj := range iFace.SkipObjectNaming {
		if inObj.Class == forbiddenObj {
			return false
		}
	}
	return true
}

// readProps:
func (iFace *GtkInterface) readProps(line string, values []string, inObj *GtkObject) (ok bool) {
	switch {
	case strings.Contains(line, `<property name="`): // property
		ok = true
		if len(values) > 0 {
			if iFace.NamingCSS && iFace.proceedNamingObj(inObj) { // In case were we desire naming object for CSS usage
				if values[0] == "name" {
					if iFace.NamingCSSClear || iFace.NamingCSSForce {
						iFace.xmlSource = append(iFace.xmlSource[:iFace.idxLine], iFace.xmlSource[iFace.idxLine+1:]...)
						// iFace.idxLine--
					} else {
						inObj.alreadyNamed = true
					}
				}
			}
		}
		switch {
		case strings.Contains(line, `translatable="`): // property and translatable
			if len(values) > 2 {
				inObj.Property = append(inObj.Property, GtkProps{Name: values[0], Translatable: values[1], Value: values[2], line: iFace.idxLine})
			} else {
				fmt.Printf("[translatable] %s: %d was skipped.\n", "only 2 args, line", iFace.idxLine+1)
			}
		default: // property only
			if len(values) > 1 {
				inObj.Property = append(inObj.Property, GtkProps{Name: values[0], Value: values[1], line: iFace.idxLine})
			} else {
				fmt.Printf("[property] %s: %d was skipped.\n", "only 1 arg, line", iFace.idxLine+1)
			}
		}
	case strings.Contains(line, `<signal name="`): // signal
		ok = true
		if len(values) > 2 {
			inObj.Signal = append(inObj.Signal, GtkProps{Name: values[0], Value: values[1], Swapped: values[2]})
		} else {
			fmt.Printf("[signal name] %s: %d was skipped.\n", "only 2 args, line", iFace.idxLine+1)
		}
	}
	if !ok {
		iFace.prevIdxLine = iFace.idxLine
	}
	return ok
}

// readPacking:
func readPacking(line string, values []string, inObj *GtkObject) (ok bool) {
	switch {
	case strings.Contains(line, `<property name="`): // property
		ok = true
		inObj.Packing = append(inObj.Packing, GtkProps{Name: values[0], Value: values[1]})
	}
	return
}

// SanitizeXml: Escape eol when multilines text are found,
// give an error if file is not a valid glade xml format.
func (iFace *GtkInterface) sanitizeXml(inBytes []byte) {
	var row string
	regComments := regexp.MustCompile(`<!--(.*|\n+)+-->`)
	regCommentLine := regexp.MustCompile(`<!--(.*?)-->`)
	commentsBytes := regComments.FindAll(inBytes, -1)
	commentLinesBytes := regCommentLine.FindAll(inBytes, -1)
	for _, com := range commentsBytes {
		iFace.Comments = append(iFace.Comments, string(com))
	}
	for _, com := range commentLinesBytes {
		iFace.Comments = append(iFace.Comments, string(com))
	}

	// remove comments from XML file
	inBytes = regCommentLine.ReplaceAll(inBytes, []byte(""))
	inBytes = regComments.ReplaceAll(inBytes, []byte(""))

	regStart := regexp.MustCompile(`^(<)`)
	regPropertyAtStart := regexp.MustCompile(`^(</property>)`)
	regSpaceTab := regexp.MustCompile(`\s`)

	iFace.xmlSource = strings.Split(string(inBytes), iFace.eol)
	// Replace LF inside labels or hints with "\n"
	for idx := len(iFace.xmlSource) - 1; idx >= 0; idx-- {
		row = iFace.xmlSource[idx]
		if len(row) != 0 {
			if !regStart.MatchString(regSpaceTab.ReplaceAllString(row, "")) || regPropertyAtStart.MatchString(regSpaceTab.ReplaceAllString(row, "")) {
				iFace.xmlSource = append(iFace.xmlSource[:idx], iFace.xmlSource[idx+1:]...)
				iFace.xmlSource[idx-1] += `\n` + row
			}
		}
	}
}

// Read Text Controls from file
func (iFace *GtkInterface) ReadFile(filename string) (err error) {
	err = jsonRead(filename, iFace)
	if err == nil {
		iFace.objectsLoaded = true
		return err
	} else {
		iFace.objectsLoaded = false
		return err
	}
}

// Write Text Controls to file
func (iFace *GtkInterface) WriteFile(filename string) error {
	iFace.ObjectsCount = len(iFace.Objects)
	iFace.UpdatedOn = timestamp().Full
	return jsonWrite(filename, iFace)
}

// getTextEOL: Get EOL from text bytes (CR, LF, CRLF)
func getTextEOL(inTextBytes []byte) (outString string) {
	bCR := []byte{0x0D}
	bLF := []byte{0x0A}
	bCRLF := []byte{0x0D, 0x0A}
	switch {
	case bytes.Contains(inTextBytes, bCRLF):
		outString = string(bCRLF)
	case bytes.Contains(inTextBytes, bCR):
		outString = string(bCR)
	default:
		outString = string(bLF)
	}
	return
}

type timeStamp struct {
	Year          string
	YearCopyRight string
	Month         string
	MonthWord     string
	Day           string
	DayWord       string
	Date          string
	Time          string
	Full          string
}

// timestamp: Get current timestamp
func timestamp() *timeStamp {
	ts := new(timeStamp)
	timed := time.Now()
	regD := regexp.MustCompile("([^[:digit:]])")
	regA := regexp.MustCompile("([^[:alpha:]])")
	splitedNum := regD.Split(timed.Format(time.RFC3339), -1)
	splitedWrd := regA.Split(timed.Format(time.RFC850), -1)
	ts.Year = splitedNum[0]
	ts.Month = splitedNum[1]
	ts.Day = splitedNum[2]
	ts.Time = splitedNum[3] + `:` + splitedNum[4] + `:` + splitedNum[5]
	ts.DayWord = splitedWrd[0]
	ts.MonthWord = splitedWrd[5]
	ts.YearCopyRight = `©` + ts.Year
	ts.Full = strings.Join(strings.Split(timed.Format(time.RFC1123), " ")[:5], " ")
	return ts
}

// jsonRead: datas from file to given interface / structure
// i.e: err := ReadJson(filename, &person)
// remember to put upper char at left of variables names to be saved.
func jsonRead(filename string, interf interface{}) (err error) {
	var textFileBytes []byte
	if textFileBytes, err = ioutil.ReadFile(filename); err == nil {
		err = json.Unmarshal(textFileBytes, &interf)
	}
	return err
}

// jsonWrite: datas to file from given interface / structure
// i.e: err := WriteJson(filename, &person)
// remember to put upper char at left of variables names to be saved.
func jsonWrite(filename string, interf interface{}) (err error) {
	var out bytes.Buffer
	var jsonData []byte
	if jsonData, err = json.Marshal(&interf); err == nil {
		if err = json.Indent(&out, jsonData, "", "\t"); err == nil {
			if err = ioutil.WriteFile(filename, out.Bytes(), 0644); err == nil {
			}
		}
	}
	return err
}
