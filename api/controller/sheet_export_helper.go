package controller

import (
	"fmt"
	"sort"
	"strings"
	"sync"
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

	categoryPolls := make(map[string][]domain.Poll)
	for _, poll := range polls {
		categories := poll.Category
		if len(categories) == 0 {
			categories = []string{"Uncategorized"}
		}

		seen := make(map[string]struct{})
		for _, rawCat := range categories {
			category := strings.TrimSpace(rawCat)
			if category == "" {
				category = "Uncategorized"
			}
			if _, ok := seen[category]; ok {
				continue
			}
			seen[category] = struct{}{}
			categoryPolls[category] = append(categoryPolls[category], poll)
		}
	}

	categoryNames := make([]string, 0, len(categoryPolls))
	for name := range categoryPolls {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	for idx, category := range categoryNames {
		fallback := fmt.Sprintf("Category %d", idx+1)
		sheetName := uniqueSheetName(category, fallback, usedSheetNames)
		if _, err := workbook.NewSheet(sheetName); err != nil {
			_ = workbook.Close()
			return nil, err
		}

		_ = workbook.SetColWidth(sheetName, "A", "A", 24)
		_ = workbook.SetColWidth(sheetName, "B", "C", 80)

		row := 1
		row = writeLabelValueRow(workbook, sheetName, row, "Category", category)
		for pollIdx, poll := range categoryPolls[category] {
			row = writePollBlock(workbook, sheetName, row, poll)
			if pollIdx < len(categoryPolls[category])-1 {
				row = insertPollSeparator(workbook, sheetName, row)
			}
		}
	}

	if idx, err := workbook.GetSheetIndex(summarySheetName); err == nil {
		workbook.SetActiveSheet(idx)
	}

	addSubmissionSheet(workbook, polls, usedSheetNames)
	return workbook, nil
}

func writeLabelValueRow(workbook *excelize.File, sheetName string, row int, label string, value interface{}) int {
	_ = workbook.SetCellValue(sheetName, cellRef("A", row), label)
	_ = workbook.SetCellValue(sheetName, cellRef("B", row), value)
	return row + 1
}

func writePollBlock(workbook *excelize.File, sheetName string, row int, poll domain.Poll) int {
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

	return row
}

func insertPollSeparator(workbook *excelize.File, sheetName string, row int) int {
	_ = workbook.SetCellValue(sheetName, cellRef("A", row), strings.Repeat("-", 24))
	return row + 2
}

func addSubmissionSheet(workbook *excelize.File, polls []domain.Poll, usedSheetNames map[string]int) {
	type entry struct {
		pollTitle  string
		categories string
		name       string
		phone      string
		submitted  time.Time
	}

	var data []entry
	for _, poll := range polls {
		if len(poll.Submissions) == 0 {
			continue
		}

		cats := strings.Join(poll.Category, ", ")
		for _, sub := range poll.Submissions {
			data = append(data, entry{
				pollTitle:  poll.Title,
				categories: cats,
				name:       sub.Name,
				phone:      sub.Phone,
				submitted:  sub.SubmittedAt,
			})
		}
	}

	if len(data) == 0 {
		return
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i].pollTitle == data[j].pollTitle {
			return data[i].submitted.Before(data[j].submitted)
		}
		return data[i].pollTitle < data[j].pollTitle
	})

	sheetName := uniqueSheetName("Submissions", "Submissions", usedSheetNames)
	if _, err := workbook.NewSheet(sheetName); err != nil {
		// Fail silently to avoid breaking export if we cannot create the sheet.
		return
	}

	_ = workbook.SetColWidth(sheetName, "A", "A", 32)
	_ = workbook.SetColWidth(sheetName, "B", "B", 28)
	_ = workbook.SetColWidth(sheetName, "C", "D", 22)
	_ = workbook.SetColWidth(sheetName, "E", "E", 26)

	headers := []string{"سوال", "دسته‌بندی‌ها", "نام", "تلفن", "زمان ثبت"}
	for colIdx, header := range headers {
		col := string(rune('A' + colIdx))
		_ = workbook.SetCellValue(sheetName, cellRef(col, 1), header)
	}

	row := 2
	for _, item := range data {
		_ = workbook.SetCellValue(sheetName, cellRef("A", row), item.pollTitle)
		_ = workbook.SetCellValue(sheetName, cellRef("B", row), item.categories)
		_ = workbook.SetCellValue(sheetName, cellRef("C", row), item.name)
		_ = workbook.SetCellValue(sheetName, cellRef("D", row), item.phone)
		_ = workbook.SetCellValue(sheetName, cellRef("E", row), formatDateTime(item.submitted))
		row++
	}
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
	return value.In(tehranLocation()).Format("2006-01-02 15:04:05 MST")
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

var (
	tehranLoc  *time.Location
	tehranOnce sync.Once
)

func tehranLocation() *time.Location {
	tehranOnce.Do(func() {
		loc, err := time.LoadLocation("Asia/Tehran")
		if err != nil {
			// Fallback to fixed GMT+03:30 if timezone database is unavailable.
			loc = time.FixedZone("Asia/Tehran", 3*3600+1800)
		}
		tehranLoc = loc
	})
	return tehranLoc
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
