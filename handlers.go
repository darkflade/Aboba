package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/unidoc/unioffice/v2/color"
	"github.com/unidoc/unioffice/v2/document"
	"github.com/unidoc/unioffice/v2/measurement"
	"github.com/unidoc/unioffice/v2/schema/soo/wml"
)

// Декоратор для Run-------------------------------------------------------------
type RunOptions struct {
	FontFamily string
	FontSize   measurement.Distance
	TextColor  color.Color
	Bold       bool
	Italic     bool
}

type RunOption func(*RunOptions)

func FontFamily(f string) RunOption {
	return func(ro *RunOptions) {
		ro.FontFamily = f
	}
}

func FontSize(size measurement.Distance) RunOption {
	return func(ro *RunOptions) {
		ro.FontSize = size
	}
}

func TextColor(c color.Color) RunOption {
	return func(ro *RunOptions) {
		ro.TextColor = c
	}
}

func Bold(b bool) RunOption {
	return func(ro *RunOptions) {
		ro.Bold = b
	}
}
func Italic(i bool) RunOption {
	return func(ro *RunOptions) {
		ro.Italic = i
	}
}

//------------------------------------------------------------

func setupDocument(deptID int, expectedSalary float64, expectedCount int, actualSalary float64, actualCount int, bossID sql.NullInt64, emps []emp) (bytes.Buffer, error) {

	// Styles constants and other
	const (
		defaultFontFamily   = "Arial"
		defaultFontSize     = 10
		cellMarginSize      = 0.08 * measurement.Centimeter // Отступы таблицы
		lineSpacing         = 4 * measurement.Point
		lineSpacingExtended = 5 * measurement.Point
	)
	var (
		backgroundDefault     = color.White
		alternativeBackground = color.FromHex("E7E6E6")
		backgroundHighlight   = color.FromHex("E6F0FA") // светло-синий фон для выделения руководителя

		borderColor = color.FromHex("4F81BD")

		textColorHead           = color.White
		textColorHighlight      = color.FromHex("2E74B5") // цвет текста для выделения руководиителя
		textColorStat           = color.FromHex("888888")
		textColorStatHeading    = color.FromHex("4F81BD") // lightblue
		textColorStatHighlight1 = color.FromHex("4CAF50") // 2E74B5 (Blue)
		textColorStatHighlight2 = color.FromHex("C00000") // Red
	)

	// Bullets Style
	SetupBulletStyle := func(bullets *document.NumberingDefinition) {

		lvl := bullets.AddLevel()
		lvl.SetFormat(wml.ST_NumberFormatBullet)
		lvl.SetAlignment(wml.ST_JcLeft)
		lvl.Properties().SetLeftIndent(0.75 * measurement.Centimeter)
		lvl.SetText("•")

	}

	// Margins in a table
	SetupMargins := func(properties *document.CellProperties, cellMarginSize measurement.Distance) {

		cellMargins := properties.Margins()

		cellMargins.SetTop(cellMarginSize)
		cellMargins.SetBottom(cellMarginSize)
		cellMargins.SetLeft(cellMarginSize)
		cellMargins.SetRight(cellMarginSize)
	}

	// Run Style
	SetupRunProperties := func(Run document.Run, opts ...RunOption) {

		properties := Run.Properties()

		options := RunOptions{
			FontFamily: defaultFontFamily,
			FontSize:   defaultFontSize,
			TextColor:  color.Black,
			Bold:       false,
			Italic:     false,
		}

		for _, opt := range opts {
			opt(&options)
		}

		properties.SetFontFamily(options.FontFamily)
		properties.SetSize(options.FontSize)
		properties.SetColor(options.TextColor)
		properties.SetBold(options.Bold)
		properties.SetItalic(options.Italic)
	}

	// Table Style
	SetupTableCell := func(cell document.Cell, backgroundColor color.Color, fontFamily string, fontSize measurement.Distance, fontColor color.Color, bold bool, text string) {
		cprops := cell.Properties()
		cprops.SetVerticalAlignment(wml.ST_VerticalJcCenter)
		SetupMargins(&cprops, cellMarginSize)
		if backgroundColor != backgroundDefault {
			cprops.SetShading(wml.ST_ShdSolid, backgroundColor, color.Auto)
		}

		p := cell.AddParagraph()
		p.SetAlignment(wml.ST_JcCenter)
		run := p.AddRun()
		run.AddText(text)

		SetupRunProperties(
			run,
			FontFamily(fontFamily),
			FontSize(fontSize),
			TextColor(fontColor),
			Bold(bold),
		)

	}

	//
	// 4) Создаём документ и настраиваем единственную секцию
	doc := document.New()
	defer doc.Close()

	// Bullets для маркированных списков
	bullets := doc.Numbering.AddDefinition()
	SetupBulletStyle(&bullets)

	// Устанавливаем размер страницы и ориентацию
	sec := doc.BodySection()
	sec.SetPageSizeAndOrientation(measurement.Millimeter*210, measurement.Millimeter*297, wml.ST_PageOrientationPortrait)

	// 5) Заголовок (Title 1)
	{
		p := doc.AddParagraph()
		p.Properties().SetHeadingLevel(1)
		p.SetAfterLineSpacing(lineSpacingExtended)

		r := p.AddRun()
		r.AddText(fmt.Sprintf("Сотрудники отдела №%d", deptID))
		SetupRunProperties(
			r,
			Bold(true),
			FontSize(18),
			TextColor(textColorHighlight),
		)
	}

	// 6) Таблица со стилями
	{
		tbl := doc.AddTable()
		props := tbl.Properties()
		props.SetWidthPercent(100) // ширина таблицы 100% от страницы
		props.Borders().SetAll(
			wml.ST_BorderSingle, borderColor, 1*measurement.Point, // Рамка таблицы
		)

		// Шапка
		hdr := tbl.AddRow()
		hdr.Properties().SetHeight(1*measurement.Centimeter, wml.ST_HeightRuleAtLeast)

		for _, txt := range []string{"ID", "Имя", "Статус", "Оклад"} {

			cell := hdr.AddCell()

			SetupTableCell(
				cell,
				borderColor,
				defaultFontFamily,
				12,
				textColorHead,
				true,
				txt,
			)
		}

		// Строки данных: чередуем фон и выделяем босса
		for i, e := range emps {
			row := tbl.AddRow()
			row.Properties().SetHeight(0.9*measurement.Centimeter, wml.ST_HeightRuleAtLeast)
			isBoss := bossID.Valid && int(bossID.Int64) == e.ID

			type CellSpecStyle struct {
				background color.Color
				text       color.Color
				bold       bool
			}

			cellSpecStyle := CellSpecStyle{
				background: color.White,
				text:       color.Black,
				bold:       false,
			}

			if isBoss {
				cellSpecStyle.background = backgroundHighlight // выделяем босса светло-синим фоном
				cellSpecStyle.text = textColorHighlight
				cellSpecStyle.bold = true
			} else if i%2 == 1 {
				cellSpecStyle.background = alternativeBackground
			} else {
				cellSpecStyle.background = backgroundDefault // белый фон для нечетных строк
			}
			for _, txt := range []string{
				strconv.Itoa(e.ID), e.Name, e.Status, fmt.Sprintf("%.2f", e.Salary),
			} {
				cell := row.AddCell()
				SetupTableCell(
					cell,
					cellSpecStyle.background,
					defaultFontFamily,
					10,
					cellSpecStyle.text,
					cellSpecStyle.bold,
					txt,
				)
			}
		}
	}

	// 7) Статистика (Title 2 + списки)
	// Описание шрифта статистики
	statsFont := func(r document.Run) {
		SetupRunProperties(
			r,
			Italic(true),
			TextColor(textColorStat),
			FontSize(11),
		)
	}
	// Empty String
	doc.AddParagraph()

	{
		p := doc.AddParagraph()
		p.SetAfterLineSpacing(lineSpacing)
		run := p.AddRun()
		run.AddText("Статистика по отделу")

		SetupRunProperties(
			run,
			Bold(true),
			FontSize(14),
			TextColor(textColorStatHeading),
		)

	}

	// Ожидаемые
	{
		p := doc.AddParagraph()
		p.SetAfterLineSpacing(lineSpacing)
		run := p.AddRun()
		run.AddText("Ожидаемые (по таблице ОТДЕЛЫ):")

		statsFont(run)

	}
	for _, item := range []struct {
		text  string
		color color.Color
	}{
		{fmt.Sprintf("Суммарная зарплата: %.2f", expectedSalary), textColorStatHighlight1},
		{fmt.Sprintf("Количество сотрудников: %d", expectedCount), textColorStatHighlight1},
	} {
		p := doc.AddParagraph()
		p.SetAfterLineSpacing(lineSpacing)
		p.SetNumberingLevel(0)
		p.SetNumberingDefinition(bullets)

		run := p.AddRun()
		run.AddText(item.text)

		SetupRunProperties(
			run,
			TextColor(item.color),
		)

	}
	// Фактические
	{
		p := doc.AddParagraph()
		p.SetAfterLineSpacing(lineSpacing)
		run := p.AddRun()
		statsFont(run)
		run.AddText("Фактические (по таблице СОТРУДНИКИ):")
	}
	for _, item := range []struct {
		text  string
		color color.Color
	}{
		{fmt.Sprintf("Суммарная зарплата: %.2f", actualSalary), textColorStatHighlight2},
		{fmt.Sprintf("Количество сотрудников: %d", actualCount), textColorStatHighlight2},
	} {
		p := doc.AddParagraph()
		p.SetAfterLineSpacing(lineSpacing)
		p.SetNumberingLevel(0)
		p.SetNumberingDefinition(bullets)

		run := p.AddRun()
		run.AddText(item.text)
		SetupRunProperties(
			run,
			TextColor(item.color),
		)
	}

	buf := &bytes.Buffer{}
	if err := doc.Save(buf); err != nil {
		return bytes.Buffer{}, fmt.Errorf("ошибка сохранения документа: %v", err)
	}
	return *buf, nil
}
