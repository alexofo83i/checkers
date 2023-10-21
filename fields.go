package main

import "fmt"

type Field struct {
	y            int
	yx           string
	fieldsAround []*Field
}

func (field *Field) Yx() string {
	return field.yx
}

func NewField(i int, yx string) *Field {
	fieldNew := Field{
		y:  i,
		yx: yx,
	}
	fieldNew.fieldsAround = make([]*Field, 0)
	return &fieldNew
}

var fieldsOnBoard map[string]*Field = make(map[string]*Field, 32)

func getFieldOnBoard(yx string) *Field {
	field, exist := fieldsOnBoard[yx]
	if !exist {
		return nil
	}
	return field
}

func initFieldsOnBoard() {
	for i := 8; i >= 1; i-- {
		var x byte = 'a'
		for k := 1; k <= 8; k++ {
			if i%2+1 == k%2+1 || i%2 == k%2 {
				yx := fmt.Sprint(i) + string(x)
				fieldNew := NewField(i, yx)
				fieldsOnBoard[yx] = fieldNew
			}
			x++
		}
	}
	for key := range fieldsOnBoard {
		field := getFieldOnBoard(key)
		fieldsAroundStr := getFieldsAroundStr(field.Yx())
		for _, fieldStr := range fieldsAroundStr {
			fieldNear := getFieldOnBoard(fieldStr)
			field.fieldsAround = append(field.fieldsAround, fieldNear)
		}
	}
}

func getFieldsAroundStr(field string) []string {
	y := field[0]
	x := field[1]
	fieldsAround := []string{string(y+1) + string(x+1), string(y+1) + string(x-1), string(y-1) + string(x+1), string(y-1) + string(x-1)}

	res := make([]string, 0, 4)
	for i := len(fieldsAround) - 1; i >= 0; i-- {
		fieldPossible := fieldsAround[i]
		if isFieldExist(fieldPossible) {
			res = append(res, fieldPossible)
		}
	}
	return res
}

func isFieldExist(field string) bool {
	y := field[0]
	x := field[1]
	return (x >= 'a' && x <= 'h' && y >= '1' && y <= '8')
}

func getFieldForKick(fieldSrc *Field, fieldTgt *Field) *Field {
	y1 := fieldSrc.Yx()[0]
	x1 := fieldSrc.Yx()[1]
	y2 := fieldTgt.Yx()[0]
	x2 := fieldTgt.Yx()[1]
	y3 := y2 + (y2 - y1)
	x3 := x2 + (x2 - x1)
	fieldForKick := getFieldOnBoard(string(y3) + string(x3))
	return fieldForKick
}

func canMakeStep(fieldFrom *Field, fieldTo *Field, whodo byte) bool {
	if whodo == 'w' {
		return fieldTo.y-fieldFrom.y > 0
	} else {
		return fieldTo.y-fieldFrom.y < 0
	}
}
