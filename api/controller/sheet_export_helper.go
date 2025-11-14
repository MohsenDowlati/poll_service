package controller

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/xuri/excelize/v2"
)

func buildSheetWorkbook(sheet domain.Sheet, polls []domain.Poll) (*excelize.File, error) {
	workbook := excelize.NewFile()

	summarySheetName := "Summary"
	defaultSheetName := workbook.GetSheetName(workbook.GetActiveSheetIndex())
	if err := workbook.SetSheetName(defaultSheetName, summarySheetName); err != nil {
		_ = workbook.Close()
		return nil, err
	}

	_ = workbook.SetColWidth(summarySheetName, "A", "A", 24)
	_ = workbook.SetColWidth(summarySheetName, "B", "B", 80)

	row := 1
	if !sheet.ID.IsZero() {
		row = writeLabelValueRow(workbook, summarySheetName, row, "ID", sheet.ID.Hex())
	}
	row = writeLabelValueRow(workbook, summarySheetName, row, "عنوان", sheet.Title)
	row = writeLabelValueRow(workbook, summarySheetName, row, "سالن", sheet.Venue)
	if sheet.Description != "" {
		row = writeLabelValueRow(workbook, summarySheetName, row, "توضیح", sheet.Description)
	}
	row = writeLabelValueRow(workbook, summarySheetName, row, "وضعیت", string(sheet.Status))
	row = writeLabelValueRow(workbook, summarySheetName, row, "شماره تلفن", yesNo(sheet.IsPhoneRequired))
	if !sheet.CreatedAt.IsZero() {
		row = writeLabelValueRow(workbook, summarySheetName, row, "زمان ساخت", formatDateTime(sheet.CreatedAt))
	}
	if !sheet.UpdatedAt.IsZero() && !sheet.UpdatedAt.Equal(sheet.CreatedAt) {
		row = writeLabelValueRow(workbook, summarySheetName, row, "زمان بروزرسانی", formatDateTime(sheet.UpdatedAt))
	}
	if !sheet.ApprovedAt.IsZero() {
		row = writeLabelValueRow(workbook, summarySheetName, row, "زمان تایید شدن", formatDateTime(sheet.ApprovedAt))
	}

	if row > 1 {
		row++
	}

	maxParticipants := 0
	opinionResponses := 0
	for _, poll := range polls {
		if poll.Participant > maxParticipants {
			maxParticipants = poll.Participant
		}
		if len(poll.Responses) > 0 {
			opinionResponses += len(poll.Responses)
		}
	}

	row = writeLabelValueRow(workbook, summarySheetName, row, "تعداد نظرسنجی", len(polls))
	row = writeLabelValueRow(workbook, summarySheetName, row, "تعداد شرکت‌کننده‌ها", maxParticipants)
	if opinionResponses > 0 {
		row = writeLabelValueRow(workbook, summarySheetName, row, "تعداد نظرات", opinionResponses)
	}
	row = writeLabelValueRow(workbook, summarySheetName, row, "زمان ساخت فایل اکسل", formatDateTime(time.Now()))

	usedSheetNames := map[string]int{summarySheetName: 1}

	for idx, poll := range polls {
		fallback := fmt.Sprintf("Poll %d", idx+1)
		sheetName := uniqueSheetName(poll.Title, fallback, usedSheetNames)
		if _, err := workbook.NewSheet(sheetName); err != nil {
			_ = workbook.Close()
			return nil, err
		}

		_ = workbook.SetColWidth(sheetName, "A", "A", 24)
		_ = workbook.SetColWidth(sheetName, "B", "C", 80)

		row := 1
		row = writeLabelValueRow(workbook, sheetName, row, "سوال", poll.Title)
		if poll.Description != "" {
			row = writeLabelValueRow(workbook, sheetName, row, "توضیحات", poll.Description)
		}
		row = writeLabelValueRow(workbook, sheetName, row, "نوع سوال", handleSheetType(poll.PollType))
		if len(poll.Category) > 0 {
			row = writeLabelValueRow(workbook, sheetName, row, "دسته‌بندی‌ها", strings.Join(poll.Category, ", "))
		}
		row = writeLabelValueRow(workbook, sheetName, row, "تعداد شرکت‌کننده‌ها", poll.Participant)

		row++

		if len(poll.Options) > 0 {
			_ = workbook.SetCellValue(sheetName, cellRef("A", row), "گزینه")
			_ = workbook.SetCellValue(sheetName, cellRef("B", row), "آرا")
			row++
			for optIndex, option := range poll.Options {
				vote := 0
				if optIndex < len(poll.Votes) {
					vote = poll.Votes[optIndex]
					if poll.PollType == domain.PollTypeSlide {
						vote /= len(poll.Votes)
					}
				}
				_ = workbook.SetCellValue(sheetName, cellRef("A", row), option)
				_ = workbook.SetCellValue(sheetName, cellRef("B", row), vote)
				row++
			}
		}

		if len(poll.Responses) > 0 {
			row++
			_ = workbook.SetCellValue(sheetName, cellRef("A", row), "پاسخ #")
			_ = workbook.SetCellValue(sheetName, cellRef("B", row), "متن")
			row++
			for respIndex, response := range poll.Responses {
				_ = workbook.SetCellValue(sheetName, cellRef("A", row), respIndex+1)
				_ = workbook.SetCellValue(sheetName, cellRef("B", row), response)
				row++
			}
		}
	}

	if idx, err := workbook.GetSheetIndex(summarySheetName); err == nil {
		workbook.SetActiveSheet(idx)
	}
	return workbook, nil
}

func writeLabelValueRow(workbook *excelize.File, sheetName string, row int, label string, value interface{}) int {
	_ = workbook.SetCellValue(sheetName, cellRef("A", row), label)
	_ = workbook.SetCellValue(sheetName, cellRef("B", row), value)
	return row + 1
}

func cellRef(column string, row int) string {
	return fmt.Sprintf("%s%d", column, row)
}

func yesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.In(time.Local).Format("2006-01-02 15:04:05 MST")
}

func uniqueSheetName(title, fallback string, used map[string]int) string {
	base := sanitizeSheetName(title)
	if base == "" {
		base = sanitizeSheetName(fallback)
	}
	if base == "" {
		base = "Sheet"
	}

	if _, exists := used[base]; !exists {
		used[base] = 1
		return base
	}

	baseRunes := []rune(base)
	counter := used[base]
	for {
		counter++
		suffix := fmt.Sprintf(" (%d)", counter)
		available := 31 - utf8.RuneCountInString(suffix)
		if available < 1 {
			available = 1
		}
		trimmed := base
		if len(baseRunes) > available {
			trimmed = string(baseRunes[:available])
		}
		candidate := trimmed + suffix
		if _, exists := used[candidate]; exists {
			continue
		}
		used[base] = counter
		used[candidate] = 1
		return candidate
	}
}

func sanitizeSheetName(name string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', '?', '*', '[', ']', ':':
			return -1
		case '\r', '\n', '\t':
			return ' '
		default:
			return r
		}
	}, strings.TrimSpace(name))

	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return ""
	}

	runes := []rune(cleaned)
	if len(runes) > 31 {
		cleaned = string(runes[:31])
	}

	return cleaned
}

func handleSheetType(ty domain.PollType) string {
	if ty == domain.PollTypeOpinion {
		return "نظر"
	}
	if ty == domain.PollTypeSingleChoice {
		return "تک گزینه‌ای"
	}
	if ty == domain.PollTypeMultiChoice {
		return "چند گزینه‌ای"
	}
	if ty == domain.PollTypeSlide {
		return "امتیاز‌دهی"
	}
	return ""
}
