package utils

import (
	"regexp"
	"strconv"

	"github.com/isd-sgcu/rpkm66-auth/constant/utils"
	"github.com/pkg/errors"
)

const CurrentYear = 66

var (
	pattern = regexp.MustCompile("(\\d{10})@student.chula.ac.th")
)

func GetFacultyFromID(sid string) (*utils.Faculty, error) {
	if len(sid) != 10 {
		return nil, errors.New("Invalid faculty id")
	}

	result, ok := utils.Faculties[sid[8:10]]
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

func GetOuidFromGmail(email string) (string, error) {
	found := pattern.FindAllStringSubmatch(email, -1)

	if len(found) > 0 && len(found[0]) > 1 {
		return found[0][1], nil
	} else {
		return "", errors.New("Invalid student email")
	}
}
