package xlsx_template

import (
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"regexp"
)

var (
	rangeRgx    = regexp.MustCompile(`{{\s*range\s+.(\w+)\s*}}`)
	rangeEndRgx = regexp.MustCompile(`{{\s*end\s*}}`)
)

type rangeRows struct {
	PropName string
	BRow     int
	ERow     int
}

func getRangeRows(iRow int, rows []*xlsx.Row) (*rangeRows, error) {

	// check begin range
	retValue := &rangeRows{
		PropName: getRangeRgx(rows[iRow]),
		BRow:     iRow + 1,
	}

	if retValue.PropName == "" {
		return nil, nil
	}

	// find end range
	if ri, err := getRangeEndRgx(rows[retValue.BRow:]); err != nil {
		return nil, fmt.Errorf("Range '%s' error: %s", retValue.PropName, err.Error())
	} else {
		retValue.ERow = retValue.BRow + ri
	}

	return retValue, nil
}

func getRangeRgx(in *xlsx.Row) string {

	if len(in.Cells) != 0 {
		match := rangeRgx.FindAllStringSubmatch(in.Cells[0].Value, -1)
		if match != nil {
			return match[0][1]
		}
	}

	return ""
}

func getRangeEndRgx(rows []*xlsx.Row) (int, error) {
	var nesting int

	for idx, row := range rows {
		if len(row.Cells) == 0 {
			continue
		}

		if rangeEndRgx.MatchString(rows[idx].Cells[0].Value) {
			if nesting == 0 {
				return idx, nil
			}

			nesting--
			continue
		}

		if rangeRgx.MatchString(rows[idx].Cells[0].Value) {
			nesting++
		}
	}

	return -1, errors.New("Not found end of range.")
}

func renderRange(iRow *int, sheet *xlsx.Sheet, rows []*xlsx.Row, ctx interface{}) (IsRender bool, err error) {

	val, err := getRangeRows(*iRow, rows)
	if err != nil || val == nil {
		return false, err
	}

	var flg bool
	flg, err = isSliceOrArray(ctx, val.PropName)
	if err != nil {
		return false, err
	}
	if !flg {
		return false, fmt.Errorf("Range '%s' error: field '%s' in ctx is not slice or array", val.PropName, val.PropName)
	}

	var rangeCtx []interface{}
	rangeCtx, err = getField(ctx, val.PropName)
	if err != nil {
		return false, err
	}

	for _, rnCtx := range rangeCtx {
		err = renderRows(sheet, rows[val.BRow:val.ERow], rnCtx)
		if err != nil {
			return false, err
		}
	}
	*iRow = val.ERow
	return true, nil
}
