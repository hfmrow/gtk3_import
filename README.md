# gtk3_import v0.1.1

### Personal library, [gotk3](https://golang.org/) helpers

##### This is a private library, not intended for distribution.

> Most of my GUI projects use it and are based on the features it offers.
> It is composed of a multitude of functions, structures, methods designed as aids to the development of software using the [Golang](https://golang.org/) and [gotk3](https://github.com/gotk3/gotk3) library, many aspects of gotk3 possibilities are covered.

---

> There is no documentation available, only code comments are present (in most cases but not always).
> If you are curious or have met him through a Google search, you are free to browse and use his content, with a view to participating in the free world of information.
> All of this is governed by the [MIT License](https://opensource.org/licenses/MIT), unless otherwise specified.
> 
> Comments and participation are welcome.

```bash
$ go install github.com/hfmrow/gtk3_import/...
```

### Gtk3 resources:

- [GTK3](https://developer.gnome.org/gtk3/stable/)
- [GDK3](https://developer.gnome.org/gdk3/stable/)
- [GLib](https://developer.gnome.org/glib/)
- [Pango](https://developer.gnome.org/pango/stable/)
- [Cairo](https://www.cairographics.org/documentation/)
- [Gtk Inspector](https://blog.gtk.org/2017/04/05/the-gtk-inspector/), a great helper to show the construction of an underlying gtk3 interface, control and produce CSS code easily and efficiently with instant visual, view / modify object settings and many other features.

*This list is not exhaustive, with some exceptions due to Golang's specifications. These documentations can help understand the implementation of gotk3 with Golang.*

---

# Important Notice:

This library will be constantly updated, it is not necessary to include it manually, it will eventually be downloaded automatically if the software requires it. Actually when I need to add / correct / modify something, I will. So this library will never have a definitive stable version until I dance with angels or deamonds (hopefully for many long years to come ^^).

It's just there for my own convenience. All my projects that use it are distributed with a **vendor** directory which includes a specific version and which will allow the program to be built without worrying about anything.

If you want to play around with the source code of my software, I recommend that you only use the **vendor** version and not try to update to a newer version (only these), as the current version may no longer be compatible with the software version available in the repository. Since Go modules exist, normally only the compatible version will be used (unless you modify the 'go.mod' file).

In this library some functions are deprecated because I started it a few years ago and unfortunately haven't cleaned up all unnecessary functions yet.

*The previous explanations do not include (of course), the use for your personal use (in your own programs, at your own risks). Each majors changes will be tagged as specific version respecting golang semantic versioning usage, as stated before, you can use any part of it as long as the definition of the MIT license is met as well as for all other types of licenses used by third party libraries used. And you are always welcome to offer ideas or fixes.*

---
