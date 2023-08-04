package utils

import (
	"testing"

	"github.com/isd-sgcu/rpkm66-auth/constant/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChulaUtilTest struct {
	suite.Suite
}

func TestChulaUtil(t *testing.T) {
	suite.Run(t, new(ChulaUtilTest))
}

func (t *ChulaUtilTest) TestGetFacultySuccess() {
	testGetFacultySuccess(t.T(), "xxxxxxxx21", "21")
	testGetFacultySuccess(t.T(), "xxxxxxxx22", "22")
	testGetFacultySuccess(t.T(), "xxxxxxxx23", "23")
	testGetFacultySuccess(t.T(), "xxxxxxxx24", "24")
}

func testGetFacultySuccess(t *testing.T, sid string, expect string) {
	want := utils.Faculties[expect]

	actual, err := GetFacultyFromID(sid)

	assert.Nil(t, err)
	assert.Equal(t, &want, actual)
}

func (t *ChulaUtilTest) TestGetFacultyInvalidFormat() {
	testGetFacultyInvalidInput(t.T(), "")
	testGetFacultyInvalidInput(t.T(), "xxxxxxx")
	testGetFacultyInvalidInput(t.T(), "xxxxxxxx21xxxxxxxxxx")
}

func (t *ChulaUtilTest) TestInvalidFacultyID() {
	testGetFacultyInvalidInput(t.T(), "xxxxxxxx80")
	testGetFacultyInvalidInput(t.T(), "xxxxxxxx00")
	testGetFacultyInvalidInput(t.T(), "xxxxxxxx12")
}

func testGetFacultyInvalidInput(t *testing.T, sid string) {
	want := "Invalid faculty id"

	actual, err := GetFacultyFromID(sid)

	assert.Nil(t, actual)
	assert.Equal(t, want, err.Error())
}

func (t *ChulaUtilTest) TestGetStudyYearSuccess() {
	testGetStudyYearSuccess(t.T(), "63xxxxxxxx", "4")
	testGetStudyYearSuccess(t.T(), "64xxxxxxxx", "3")
	testGetStudyYearSuccess(t.T(), "65xxxxxxxx", "2")
	testGetStudyYearSuccess(t.T(), "66xxxxxxxx", "1")
}

func testGetStudyYearSuccess(t *testing.T, sid string, expect string) {
	want := expect

	actual, err := CalYearFromID(sid)

	assert.Nil(t, err)
	assert.Equal(t, want, actual)
}

func (t *ChulaUtilTest) TestCalStudyYearInvalidFormat() {
	testCalStudyYearInvalidInput(t.T(), "")
	testCalStudyYearInvalidInput(t.T(), "66xxx")
	testCalStudyYearInvalidInput(t.T(), "xx24xxxxxx")
	testCalStudyYearInvalidInput(t.T(), "66xxxxxxxxxxx")
}

func (t *ChulaUtilTest) TestCalStudyYearInvalidYear() {
	testCalStudyYearInvalidInput(t.T(), "67xxxxxxxxxxx")
	testCalStudyYearInvalidInput(t.T(), "68xxxxxxxxxxx")
	testCalStudyYearInvalidInput(t.T(), "99xxxxxxxxxxx")
}

func testCalStudyYearInvalidInput(t *testing.T, sid string) {
	want := "Invalid student id"

	actual, err := CalYearFromID(sid)

	assert.Equal(t, actual, "")
	assert.Equal(t, want, err.Error())
}

func (t *ChulaUtilTest) TestValidStudentGmail() {
	gmail := "6530000021@student.chula.ac.th"

	ouid, err := GetOuidFromGmail(gmail)

	t.Nil(err)
	t.Equal("6530000021", ouid)
}

func (t *ChulaUtilTest) TestInvalidStudentGmail_InvalidOuid() {
	gmail := "notstudent@student.chula.ac.th"

	expected := "Invalid student email"

	_, err := GetOuidFromGmail(gmail)

	t.Equal(expected, err.Error())
}
