app.controller('assignStudentController', function($scope, $rootScope, objService, $timeout, $mdDialog) {
  var controller = this;
  $rootScope.assign = this;
  controller.studentList = [{id:-1,name:'Loading'}];
  controller.rootScope = $rootScope;
  controller.grade;
  controller.selectedStudent;
  controller.submitDisabled = true;
  controller.selectDisabled = true;
  
  controller.querySearch = function(query) {
    return query ? controller.studentList.filter( controller.createFilterFor(query) ) : controller.studentList;
  }
  
  controller.createFilterFor = function(query) {
    var lowercaseQuery = angular.lowercase(query);

    return function filterFn(student) {
      return (angular.lowercase(student.name).indexOf(lowercaseQuery) === 0);
    };
  }
  
  controller.updateStudents = function() {
    controller.selectDisabled = true;
    controller.selectedStudent = null;
    google.script.run.withSuccessHandler(controller.showStudents).getStudents(true, controller.grade);
  }
  
  controller.showStudents = function(students) {
    $scope.$apply(function() {
      var list = [];
      students.forEach(function(item, index) {
        var studentObj = objService.getStudentObjFromList(item);
        studentObj.curBlock = index;
        list.push(studentObj);
      })
      controller.studentList = list;
      if (controller.grade != '') {
        controller.selectDisabled = false;
      }
    })
  }
  
  controller.addStudent = function() {
    var selectedStudent = controller.selectedStudent;
    var studentId = parseInt(selectedStudent.id);
    var block = $rootScope.block;
    google.script.run.withSuccessHandler($rootScope.mainView.handleError).addStudentToClass(studentId, block);
    var studentTable = $rootScope.mainView.NextStudents;
    studentTable[block][studentTable[block].length] = selectedStudent;
    controller.closeDialog();
  }
  
  controller.changeGrade = function() {
    controller.updateStudents();
    controller.submitDisabled = true;
  }
  
  controller.changeStudent = function() {
    var selectedStudent = controller.selectedStudent;
    var block = $rootScope.block;
    controller.submitDisabled = true;
    var classFull = $rootScope.edit.currentBlockInfo[block].curSize >= $rootScope.edit.currentBlockInfo[block].maxSize;
    if (!controller.selectDisabled && selectedStudent != null) {
      controller.submitDisabled = selectedStudent.blocks[block].blocks[block].status == 'closed' || classFull;
    }
  }
  
  controller.closeDialog = function() {
    $mdDialog.hide();
  }
  
  controller.onClose = function() {
    controller.grade = null;
    controller.selectDisabled = true;
    controller.selectedStudent = null;
    controller.submitDisabled = true;
  }
  
  controller.updateStudents()
});