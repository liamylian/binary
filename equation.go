package binary

import "errors"

func solveSum(fieldValue map[string]int, totalValue int) (fieldName string, value int, err error) {
	sum := 0
	var toSolveFields []string
	for fieldName, value := range fieldValue {
		if value >= 0 {
			sum += value
		} else {
			toSolveFields = append(toSolveFields, fieldName)
		}
	}

	toSolveFieldsLen := len(toSolveFields)
	if toSolveFieldsLen == 0 {
		return "", 0, nil
	} else if toSolveFieldsLen == 1 {
		fieldName = toSolveFields[0]
		value = totalValue - sum
		return
	} else {
		return "", 0, errors.New("more than 1 field to solve")
	}
}
