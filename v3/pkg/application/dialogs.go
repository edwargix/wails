package application

import "C"
import (
	"strings"
	"sync"
)

type DialogType int

var dialogMapID = make(map[uint]struct{})
var dialogIDLock sync.RWMutex

func getDialogID() uint {
	dialogIDLock.Lock()
	defer dialogIDLock.Unlock()
	var dialogID uint
	for {
		if _, ok := dialogMapID[dialogID]; !ok {
			dialogMapID[dialogID] = struct{}{}
			break
		}
		dialogID++
		if dialogID == 0 {
			panic("no more dialog IDs")
		}
	}
	return dialogID
}

func freeDialogID(id uint) {
	dialogIDLock.Lock()
	defer dialogIDLock.Unlock()
	delete(dialogMapID, id)
}

var openFileResponses = make(map[uint]chan string)
var saveFileResponses = make(map[uint]chan string)

const (
	InfoDialog DialogType = iota
	QuestionDialog
	WarningDialog
	ErrorDialog
	OpenDirectoryDialog
)

type Button struct {
	label     string
	isCancel  bool
	isDefault bool
	callback  func()
}

func (b *Button) OnClick(callback func()) {
	b.callback = callback
}

type messageDialogImpl interface {
	show()
}

type MessageDialog struct {
	dialogType DialogType
	title      string
	message    string
	buttons    []*Button
	icon       []byte

	// platform independent
	impl messageDialogImpl
}

var defaultTitles = map[DialogType]string{
	InfoDialog:     "Information",
	QuestionDialog: "Question",
	WarningDialog:  "Warning",
	ErrorDialog:    "Error",
}

func newMessageDialog(dialogType DialogType) *MessageDialog {
	return &MessageDialog{
		dialogType: dialogType,
		title:      defaultTitles[dialogType],
	}
}

func (d *MessageDialog) SetTitle(title string) *MessageDialog {
	d.title = title
	return d
}

func (d *MessageDialog) Show() {
	if d.impl == nil {
		d.impl = newDialogImpl(d)
	}
	d.impl.show()
}

func (d *MessageDialog) SetIcon(icon []byte) *MessageDialog {
	d.icon = icon
	return d
}

func (d *MessageDialog) AddButton(s string) *Button {
	result := &Button{
		label: s,
	}
	d.buttons = append(d.buttons, result)
	return result
}

func (d *MessageDialog) SetDefaultButton(button *Button) *MessageDialog {
	for _, b := range d.buttons {
		b.isDefault = false
	}
	button.isDefault = true
	return d
}

func (d *MessageDialog) SetCancelButton(button *Button) *MessageDialog {
	for _, b := range d.buttons {
		b.isCancel = false
	}
	button.isCancel = true
	return d
}

func (d *MessageDialog) SetMessage(title string) *MessageDialog {
	d.title = title
	return d
}

type openFileDialogImpl interface {
	show() ([]string, error)
}

type fileFilter struct {
	displayName string // Filter information EG: "Image Files (*.jpg, *.png)"
	pattern     string // semicolon separated list of extensions, EG: "*.jpg;*.png"
}

type OpenFileDialog struct {
	id                              uint
	canChooseDirectories            bool
	canChooseFiles                  bool
	canCreateDirectories            bool
	showHiddenFiles                 bool
	resolvesAliases                 bool
	allowsMultipleSelection         bool
	hideExtension                   bool
	canSelectHiddenExtension        bool
	treatsFilePackagesAsDirectories bool
	allowsOtherFileTypes            bool
	filters                         []fileFilter

	title      string
	message    string
	buttonText string
	directory  string
	window     *WebviewWindow

	impl openFileDialogImpl
}

func (d *OpenFileDialog) CanChooseFiles(canChooseFiles bool) *OpenFileDialog {
	d.canChooseFiles = canChooseFiles
	return d
}

func (d *OpenFileDialog) CanChooseDirectories(canChooseDirectories bool) *OpenFileDialog {
	d.canChooseDirectories = canChooseDirectories
	return d
}

func (d *OpenFileDialog) CanCreateDirectories(canCreateDirectories bool) *OpenFileDialog {
	d.canCreateDirectories = canCreateDirectories
	return d
}

func (d *OpenFileDialog) AllowsOtherFileTypes(allowsOtherFileTypes bool) *OpenFileDialog {
	d.allowsOtherFileTypes = allowsOtherFileTypes
	return d
}

func (d *OpenFileDialog) ShowHiddenFiles(showHiddenFiles bool) *OpenFileDialog {
	d.showHiddenFiles = showHiddenFiles
	return d
}

func (d *OpenFileDialog) HideExtension(hideExtension bool) *OpenFileDialog {
	d.hideExtension = hideExtension
	return d
}

func (d *OpenFileDialog) TreatsFilePackagesAsDirectories(treatsFilePackagesAsDirectories bool) *OpenFileDialog {
	d.treatsFilePackagesAsDirectories = treatsFilePackagesAsDirectories
	return d
}

func (d *OpenFileDialog) AttachToWindow(window *WebviewWindow) *OpenFileDialog {
	d.window = window
	return d
}

func (d *OpenFileDialog) ResolvesAliases(resolvesAliases bool) *OpenFileDialog {
	d.resolvesAliases = resolvesAliases
	return d
}

func (d *OpenFileDialog) SetTitle(title string) *OpenFileDialog {
	d.title = title
	return d
}

func (d *OpenFileDialog) PromptForSingleSelection() (string, error) {
	d.allowsMultipleSelection = false
	if d.impl == nil {
		d.impl = newOpenFileDialogImpl(d)
	}
	selection, err := d.impl.show()
	var result string
	if len(selection) > 0 {
		result = selection[0]
	}

	return result, err
}

// AddFilter adds a filter to the dialog. The filter is a display name and a semicolon separated list of extensions.
// EG: AddFilter("Image Files", "*.jpg;*.png")
func (d *OpenFileDialog) AddFilter(displayName, pattern string) *OpenFileDialog {
	d.filters = append(d.filters, fileFilter{
		displayName: strings.TrimSpace(displayName),
		pattern:     strings.TrimSpace(pattern),
	})
	return d
}

func (d *OpenFileDialog) PromptForMultipleSelection() ([]string, error) {
	d.allowsMultipleSelection = true
	if d.impl == nil {
		d.impl = newOpenFileDialogImpl(d)
	}
	return d.impl.show()
}

func (d *OpenFileDialog) SetMessage(message string) *OpenFileDialog {
	d.message = message
	return d
}

func (d *OpenFileDialog) SetButtonText(text string) *OpenFileDialog {
	d.buttonText = text
	return d
}

func (d *OpenFileDialog) SetDirectory(directory string) *OpenFileDialog {
	d.directory = directory
	return d
}

func (d *OpenFileDialog) CanSelectHiddenExtension(canSelectHiddenExtension bool) *OpenFileDialog {
	d.canSelectHiddenExtension = canSelectHiddenExtension
	return d
}

func newOpenFileDialog() *OpenFileDialog {
	return &OpenFileDialog{
		id:                   getDialogID(),
		canChooseDirectories: false,
		canChooseFiles:       true,
		canCreateDirectories: true,
		resolvesAliases:      false,
	}
}

func newSaveFileDialog() *SaveFileDialog {
	return &SaveFileDialog{
		id:                   getDialogID(),
		canCreateDirectories: true,
	}
}

type SaveFileDialog struct {
	id                              uint
	canCreateDirectories            bool
	showHiddenFiles                 bool
	canSelectHiddenExtension        bool
	allowOtherFileTypes             bool
	hideExtension                   bool
	treatsFilePackagesAsDirectories bool
	message                         string
	directory                       string
	filename                        string
	buttonText                      string

	window *WebviewWindow

	impl saveFileDialogImpl
}

type saveFileDialogImpl interface {
	show() (string, error)
}

func (d *SaveFileDialog) CanCreateDirectories(canCreateDirectories bool) *SaveFileDialog {
	d.canCreateDirectories = canCreateDirectories
	return d
}

func (d *SaveFileDialog) CanSelectHiddenExtension(canSelectHiddenExtension bool) *SaveFileDialog {
	d.canSelectHiddenExtension = canSelectHiddenExtension
	return d
}

func (d *SaveFileDialog) ShowHiddenFiles(showHiddenFiles bool) *SaveFileDialog {
	d.showHiddenFiles = showHiddenFiles
	return d
}

func (d *SaveFileDialog) SetMessage(message string) *SaveFileDialog {
	d.message = message
	return d
}

func (d *SaveFileDialog) SetDirectory(directory string) *SaveFileDialog {
	d.directory = directory
	return d
}

func (d *SaveFileDialog) AttachToWindow(window *WebviewWindow) *SaveFileDialog {
	d.window = window
	return d
}

func (d *SaveFileDialog) PromptForSingleSelection() (string, error) {
	if d.impl == nil {
		d.impl = newSaveFileDialogImpl(d)
	}
	return d.impl.show()
}

func (d *SaveFileDialog) SetButtonText(text string) *SaveFileDialog {
	d.buttonText = text
	return d
}

func (d *SaveFileDialog) SetFilename(filename string) *SaveFileDialog {
	d.filename = filename
	return d
}

func (d *SaveFileDialog) AllowsOtherFileTypes(allowOtherFileTypes bool) *SaveFileDialog {
	d.allowOtherFileTypes = allowOtherFileTypes
	return d
}

func (d *SaveFileDialog) HideExtension(hideExtension bool) *SaveFileDialog {
	d.hideExtension = hideExtension
	return d
}

func (d *SaveFileDialog) TreatsFilePackagesAsDirectories(treatsFilePackagesAsDirectories bool) *SaveFileDialog {
	d.treatsFilePackagesAsDirectories = treatsFilePackagesAsDirectories
	return d
}
