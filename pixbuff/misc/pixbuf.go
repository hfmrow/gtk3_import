// pixbuf.go

/*
	Source file auto-generated on Fri, 07 Aug 2020 18:54:48 using Gotk3ObjHandler v1.5 ©2018-20 H.F.M
	This software use gotk3 that is licensed under the ISC License:
	https://github.com/gotk3/gotk3/blob/master/LICENSE

	Copyright ©2020-21 hfmrow - used as bridge to gotk3 library for gen_lib 'package installer'
	This program comes with absolutely no warranty. See the The MIT License (MIT) for details:
	https://opensource.org/licenses/mit-license.php
*/

package misc

import "github.com/gotk3/gotk3/gdk"

// PixbufNewFromFileAtScale:
func PixbufNewFromFileAtScale(inFilename string, width, height int, presAR bool) (*gdk.Pixbuf, error) {
	return gdk.PixbufNewFromFileAtScale(inFilename, width, height, presAR)
}

// PixbufNewFromFile:
func PixbufNewFromFile(inFilename string) (*gdk.Pixbuf, error) {
	return gdk.PixbufNewFromFile(inFilename)
}
