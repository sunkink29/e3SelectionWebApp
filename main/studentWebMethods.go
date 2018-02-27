package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/sunkink29/e3webapp/errors"
	"github.com/sunkink29/e3webapp/student"
	"github.com/sunkink29/e3webapp/teacher"
	"github.com/sunkink29/e3webapp/user"
)

func addStudentMethods() {
	addToWebMethods("setTeacher", setTeacher)
	addToWebMethods("getCurrentStudentBlocks", getCurrentStudentBlocks)
	addToWebMethods("newStudent", newStudent)
	addToWebMethods("editStudent", editStudent)
	addToWebMethods("deleteStudent", deleteStudent)
	addToWebMethods("getAllStudents", getAllStudents)
	addToWebMethods("studentClassOpen", studentClassOpen)
}

func setTeacher(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	var variables struct {
		Teacher string
		Block   int
	}
	if err := dec.Decode(&variables); err != nil {
		return errors.New(err.Error())
	}

	stdnt, err := student.Current(ctx, false, debug)
	if err != nil {
		return err
	}
	newTeacher, err := teacher.WithEmail(ctx, variables.Teacher, false, debug)
	if err != nil {
		return err
	}
	if variables.Block == 0 {
		prevTeacher, err := teacher.WithEmail(ctx, stdnt.Teacher1, false, debug)
		prevOpen := true
		if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
			return err
		} else if err == nil {
			prevOpen = prevTeacher.Block1.BlockOpen
		}
		newOpen := newTeacher.Block1.BlockOpen
		newFull := newTeacher.Block1.CurSize >= newTeacher.Block1.MaxSize

		if prevOpen && newOpen && !newFull {
			stdnt.Teacher1 = variables.Teacher
		}
	} else {
		prevTeacher, err := teacher.WithEmail(ctx, stdnt.Teacher2, false, debug)
		prevOpen := true
		if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
			return err
		} else if err == nil {
			prevOpen = prevTeacher.Block2.BlockOpen
		}
		newOpen := newTeacher.Block2.BlockOpen
		newFull := newTeacher.Block2.CurSize >= newTeacher.Block2.MaxSize

		if prevOpen && newOpen && !newFull {
			stdnt.Teacher2 = variables.Teacher
		}
	}

	err = stdnt.Edit(ctx)
	return err
}

func getCurrentStudentBlocks(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	var current bool
	if err := dec.Decode(&current); err != nil {
		return errors.New(err.Error())
	}

	stdnt, err := student.Current(ctx, current, debug)
	if err != nil {
		return err
	}
	block1, err := teacher.WithEmail(ctx, stdnt.Teacher1, current, debug)
	if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
		return err
	}
	block2, err := teacher.WithEmail(ctx, stdnt.Teacher2, current, debug)
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

func newStudent(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	curU, err := user.Current(ctx, debug)
	if err != nil {
		return err
	}
	if curU.Admin {
		newS := new(student.Student)
		if err := dec.Decode(newS); err != nil {
			return errors.New(err.Error())
		}
		err := newS.New(ctx, debug)
		return err
	}
	return errors.New(errors.AccessDenied)
}

func editStudent(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	curU, err := user.Current(ctx, debug)
	if err != nil {
		return err
	}
	if curU.Admin {
		stdnt := new(student.Student)
		if err := dec.Decode(stdnt); err != nil {
			return errors.New(err.Error())
		}
		err := stdnt.Edit(ctx)
		return err
	}
	return errors.New(errors.AccessDenied)
}

func deleteStudent(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	curU, err := user.Current(ctx, debug)
	if err != nil {
		return err
	}
	if curU.Admin {
		sKey := new(string)
		if err := dec.Decode(sKey); err != nil {
			return errors.New(err.Error())
		}
		stdnt := student.Student{ID: *sKey}
		err = stdnt.Delete(ctx)
		return err
	}
	return errors.New(errors.AccessDenied)
}

var studentList []*student.Student

func getAllStudents(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	debug := r.Form.Get("debug") == "true"
	if len(studentList) == 0 {
		var err error
		studentList, err = student.All(ctx, false, debug)
		if err != nil {
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

func studentClassOpen(dec *json.Decoder, w http.ResponseWriter, r *http.Request) error {
	ctx := appengine.NewContext(r)
	var variables struct{StdntID string; Block int}
	if err := dec.Decode(&variables); err != nil {
		return errors.New(err.Error())
	}
	
	key, err := datastore.DecodeKey(variables.StdntID)
	if err != nil {
		return errors.New(err.Error())
	}
	stdnt, err := student.WithKey(ctx, key)
	if err != nil {
		return err
	}
	
	open := true
	if variables.Block == 0 {
		Teacher, err := teacher.WithEmail(ctx, stdnt.Teacher1, false, false)
		
		if err != nil && err.(errors.Error).Message != teacher.TeacherNotFound {
			return err
		} else if err == nil {
			open = Teacher.Block1.BlockOpen
		}
	} else {
		Teacher, err := teacher.WithEmail(ctx, stdnt.Teacher2, false, false)
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