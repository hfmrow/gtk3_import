// textTag.go

/*
	This software use gotk3 that is licensed under the ISC License:
	https://github.com/gotk3/gotk3/blob/master/LICENSE

	Copyright ©2019 H.F.M
	This program comes with absolutely no warranty. See the The MIT License (MIT) for details:
	https://opensource.org/licenses/mit-license.php
*/

package textView

import (
	"fmt"
	"log"
	"strings"

	glsg "github.com/hfmrow/gen_lib/strings"

	"github.com/gotk3/gotk3/gtk"
)

// TextViewScrollToLine: Scroll to line and highligth it
// Independant function ...
func TextViewScrollToLine(textView *gtk.TextView, line int, highLight ...bool) {
	var doHighLight bool
	if len(highLight) > 0 {
		doHighLight = highLight[0]
	}
	var err error
	if line > 0 {
		if buf, err := textView.GetBuffer(); err == nil {

			iterTxt0 := buf.GetIterAtLine(line)
			iterTxt1 := buf.GetIterAtOffset(buf.GetIterAtLine(line).GetOffset() - 1)

			buf.PlaceCursor(iterTxt0)
			for gtk.EventsPending() {
				gtk.MainIterationDo(false)
			}
			textView.ScrollToIter(iterTxt0, 0.0, true, 0.5, 0.5)

			if doHighLight {
				buf.SelectRange(iterTxt0, iterTxt1) // HighLight current line.
			}
		}
	}
	if err != nil {
		log.Fatalf("TextViewScrollToLine: %s\n", err.Error())
	}
}

// GetTextView: Retrieve text from TextView as []string
func GetTextView(tv *gtk.TextView, removeEmpty ...bool) (out []string) {
	var re bool
	var tmpTxt string
	var err error
	var buff *gtk.TextBuffer

	if len(removeEmpty) > 0 {
		re = removeEmpty[0]
	}
	if buff, err = tv.GetBuffer(); err == nil {
		if tmpTxt, err = buff.GetText(buff.GetStartIter(), buff.GetEndIter(), false); err == nil {
			out = strings.Split(tmpTxt, glsg.GetTextEOL([]byte(tmpTxt)))
			if re {
				for idx, line := range out {
					if len(line) == 0 {
						out = append(out[:idx], out[idx+1:]...)
					}
				}
			}
		}
	}
	if err != nil {
		fmt.Printf("GetTextView: %s", err.Error())
	}
	return
}

// SetTextView: Set []string to TextView
func SetTextView(tv *gtk.TextView, in []string, removeEmpty ...bool) {
	var re bool
	var err error
	var buff *gtk.TextBuffer

	if len(removeEmpty) > 0 {
		re = removeEmpty[0]
	}
	if buff, err = tv.GetBuffer(); err == nil {
		if re {
			for idx, line := range in {
				if len(line) == 0 {
					in = append(in[:idx], in[idx+1:]...)
				}
			}
		}
		buff.SetText(strings.Join(in, glsg.GetOsLineEnd()))
	}
	if err != nil {
		fmt.Printf("SetTextView: %s", err.Error())
	}
	return
}
