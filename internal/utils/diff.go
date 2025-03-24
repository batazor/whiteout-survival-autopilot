package diffutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	changedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)   // красный
	normalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))              // белый
	fieldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true) // серый
)

func DiffStruct(oldVal, newVal any) string {
	var builder strings.Builder

	oldV := reflect.ValueOf(oldVal)
	newV := reflect.ValueOf(newVal)

	// В случае указателей
	if oldV.Kind() == reflect.Ptr {
		oldV = oldV.Elem()
	}
	if newV.Kind() == reflect.Ptr {
		newV = newV.Elem()
	}

	if oldV.Type() != newV.Type() {
		return changedStyle.Render("Incompatible types for diff")
	}

	t := oldV.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		oldF := oldV.Field(i)
		newF := newV.Field(i)

		// Пропускаем несравнимые поля
		if !oldF.CanInterface() || !newF.CanInterface() {
			continue
		}

		oldStr := fmt.Sprintf("%v", oldF.Interface())
		newStr := fmt.Sprintf("%v", newF.Interface())

		if oldStr != newStr {
			builder.WriteString(fmt.Sprintf("%s: %s → %s\n",
				fieldStyle.Render(field.Name),
				normalStyle.Render(oldStr),
				changedStyle.Render(newStr),
			))
		}
	}

	if builder.Len() == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("No changes")
	}

	return builder.String()
}
