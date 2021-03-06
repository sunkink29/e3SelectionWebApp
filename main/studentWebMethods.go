package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/sunkink29/e3webapp/errors"
	"github.com/sunkink29/e3webapp/messaging"
	"github.com/sunkink29/e3webapp/student"
	"github.com/sunkink29/e3webapp/teacher"
	"github.com/sunkink29/e3webapp/user"
)

func addStudentHandle(url string, handle appHandler) {
	http.Handle("/api/student/"+url, handle)
}

func addStudentMethods() {
	addStudentHandle("setteacher", appHandler(setTeacher))
	addStudentHandle("getteachers", appHandler(getCurrentStudentBlocks))
	addStudentHandle("new", appHandler(newStudent))
	addStudentHandle("edit", appHandler(editStudent))
	addStudentHandle("delete", appHandler(deleteStudent))
	addStudentHandle("getall", appHandler(getAllStudents))
	addStudentHandle("open", appHandler(studentClassOpen))
}

func setTeacher(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	decoder := json.NewDecoder(r.Body)
	var variables struct {
		ID    string
		Block int
	}
	if err := decoder.Decode(&variables); err != nil {
		return errors.New(err.Error())
	}

	stdnt, err := student.Current(ctx, false)
	if err != nil {
		return err
	}
	key, err := datastore.DecodeKey(variables.ID)
	if err != nil {
		return errors.New(err.Error())
	}
	newTeacher, err := teacher.WithKey(ctx, key)
	if err != nil {
		return err
	}

	var tchr string
	if variables.Block == 0 {
		tchr = stdnt.Teacher1
	} else {
		tchr = stdnt.Teacher2
	}
	prevTeacher, err := teacher.WithEmail(ctx, tchr, false)
	prevOpen := true
	if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
		return err
	} else if err == nil {
		if variables.Block == 0 {
			prevOpen = prevTeacher.Block1.BlockOpen
		} else {
			prevOpen = prevTeacher.Block2.BlockOpen
		}
	}

	var newBlock teacher.Block
	if variables.Block == 0 {
		newBlock = newTeacher.Block1
	} else {
		newBlock = newTeacher.Block2
	}
	newOpen := newBlock.BlockOpen
	newFull := newBlock.CurSize >= newBlock.MaxSize

	var stdntBlock *string
	if variables.Block == 0 {
		stdntBlock = &stdnt.Teacher1
	} else {
		stdntBlock = &stdnt.Teacher2
	}
	if prevOpen && newOpen && !newFull {
		*stdntBlock = newTeacher.Email
	}

	if err = stdnt.Edit(ctx); err != nil {
		return err
	}

	newtchrUsr, err := user.WithEmail(ctx, newTeacher.Email)
	if err != nil {
		return err
	}

	changeStudent := struct {
		Block   int
		Student *student.Student
		Method  string
	}{variables.Block, stdnt, "add"}
	jStudent, err := json.Marshal(changeStudent)
	if err != nil {
		return errors.New(err.Error())
	}
	message := string(jStudent[:])
	if newtchrUsr.RToken != "" {
		err = messaging.SendUserEvent(ctx, messaging.EventTypes.StudentUpdate, message, newtchrUsr.RToken)
	}
	if err != nil {
		return err
	}
	if prevTeacher != nil {
		prevtchrUsr, err := user.WithEmail(ctx, prevTeacher.Email)
		if err != nil {
			return err
		}

		changeStudent.Method = "remove"
		jStudent, err = json.Marshal(changeStudent)
		if err != nil {
			return errors.New(err.Error())
		}
		message = string(jStudent[:])
		err = messaging.SendUserEvent(ctx, messaging.EventTypes.StudentUpdate, message, prevtchrUsr.RToken)
		if err != nil {
			return err
		}
	}
	teachers := []*teacher.Teacher{newTeacher}
	if prevTeacher != nil {
		teachers = append(teachers, prevTeacher)
	}
	jTeachers, err := json.Marshal(teachers)
	if err != nil {
		return errors.New(err.Error())
	}
	message = string(jTeachers[:])
	err = messaging.SendEvent(ctx, messaging.EventTypes.ClassEdit, message, messaging.Topics.Student)
	if err != nil {
		return err
	}

	return nil
}

func getCurrentStudentBlocks(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	current := r.Form.Get("current") == "true"

	stdnt, err := student.Current(ctx, current)
	if err != nil {
		return err
	}
	block1, err := teacher.WithEmail(ctx, stdnt.Teacher1, current)
	if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
		return err
	}
	block2, err := teacher.WithEmail(ctx, stdnt.Teacher2, current)
	if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
		return err
	}
	blocks := []*teacher.Teacher{block1, block2}
	jBlocks, err := json.Marshal(blocks)
	if err != nil {
		return errors.New(err.Error())
	}
	s := string(jBlocks[:])
	fmt.Fprintln(w, s)
	return nil
}

func newStudent(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	curU, err := user.Current(ctx)
	if err != nil {
		return err
	}
	if curU.Admin {
		decoder := json.NewDecoder(r.Body)
		newS := new(student.Student)
		if err := decoder.Decode(newS); err != nil {
			return errors.New(err.Error())
		}
		err := newS.New(ctx)
		studentList = append(studentList, newS)
		return err
	}
	return errors.New(errors.AccessDenied)
}

func editStudent(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	curU, err := user.Current(ctx)
	if err != nil {
		return err
	}
	if curU.Admin {
		decoder := json.NewDecoder(r.Body)
		stdnt := new(student.Student)
		if err := decoder.Decode(stdnt); err != nil {
			return errors.New(err.Error())
		}
		for i, j := range studentList {
			if j.ID == stdnt.ID {
				studentList[i] = stdnt
			}
		}
		err := stdnt.Edit(ctx)
		return err
	}
	return errors.New(errors.AccessDenied)
}

func deleteStudent(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	curU, err := user.Current(ctx)
	if err != nil {
		return err
	}
	if curU.Admin {
		decoder := json.NewDecoder(r.Body)
		sKey := new(string)
		if err := decoder.Decode(sKey); err != nil {
			return errors.New(err.Error())
		}
		stdnt := student.Student{ID: *sKey}
		for i, j := range studentList {
			if j.ID == stdnt.ID {
				studentList[i] = nil
			}
		}
		err = stdnt.Delete(ctx)
		return err
	}
	return errors.New(errors.AccessDenied)
}

var studentList []*student.Student

func getAllStudents(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	if len(studentList) == 0 {
		var err error
		if studentList, err = student.All(ctx, false); err != nil {
			return err
		}
	}

	jStudents, err := json.Marshal(studentList)
	if err != nil {
		return errors.New(err.Error())
	}
	s := string(jStudents[:])

	fmt.Fprintln(w, s)
	return nil
}

func studentClassOpen(w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	stdntID := r.Form.Get("id")
	block, _ := strconv.Atoi(r.Form.Get("Block"))

	key, err := datastore.DecodeKey(stdntID)
	if err != nil {
		return errors.New(err.Error())
	}
	stdnt, err := student.WithKey(ctx, key)
	if err != nil {
		return err
	}

	open := true
	if block == 0 {
		Teacher, err := teacher.WithEmail(ctx, stdnt.Teacher1, false)

		if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
			return err
		} else if err == nil {
			open = Teacher.Block1.BlockOpen
		}
	} else {
		Teacher, err := teacher.WithEmail(ctx, stdnt.Teacher2, false)
		if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
			return err
		} else if err == nil {
			open = Teacher.Block2.BlockOpen
		}
	}

	jOutput, err := json.Marshal(open)
	if err != nil {
		return errors.New(err.Error())
	}
	s := string(jOutput[:])

	fmt.Fprintln(w, s)
	return nil
}
