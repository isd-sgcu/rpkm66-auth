package utils

import (
	"github.com/isd-sgcu/rnkm65-auth/src/constant"
	"github.com/pkg/errors"
	"strconv"
)

const CurrentYear = 65

func GetFacultyFromID(sid string) (*constant.Faculty, error) {
	if len(sid) != 10 {
		return nil, errors.New("Invalid faculty id")
	}

	result, ok := constant.Faculties[sid[8:10]]
	if !ok {
		return nil, errors.New("Invalid faculty id")
	}
	return &result, nil
}

func CalYearFromID(sid string) (string, error) {
	if len(sid) != 10 {
		return "", errors.New("Invalid student id")
	}

	yearIn, err := strconv.Atoi(sid[:2])
	if err != nil {
		return "", errors.New("Invalid student id")
	}

	studYear := CurrentYear - yearIn + 1
	if studYear <= 0 {
		return "", errors.New("Invalid student ID")
	}

	return strconv.Itoa(studYear), nil
}
