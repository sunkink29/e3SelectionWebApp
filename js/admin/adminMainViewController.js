app.controller('adminMainViewController', function($scope, $rootScope, objService, $mdDialog) {
  var controller = this;
  
  controller.showAddDialog = function(ev) {
    $mdDialog.show({
      contentElement: '#addUser',
      parent: angular.element(document.body),
      targetEvent: ev,
      clickOutsideToClose: true,
      onRemoving: function() {
        $rootScope.addUser.resetUser();
      }
    });
  }
  
  controller.showEditDialog = function(ev) {
    $mdDialog.show({
      contentElement: '#editUser',
      parent: angular.element(document.body),
      targetEvent: ev,
      clickOutsideToClose: true,
      onRemoving: function() {
        $rootScope.editUser.resetUser();
      }
    });
  }
  
  controller.showRemoveDialog = function(ev) {
    $mdDialog.show({
      contentElement: '#removeUser',
      parent: angular.element(document.body),
      targetEvent: ev,
      clickOutsideToClose: true,
      onRemoving: function() {
        $rootScope.removeUser.resetUser();
      }
    });
  }
  
  controller.showScheduleDialog = function(ev) {
    $mdDialog.show({
      contentElement: '#viewSchedule',
      parent: angular.element(document.body),
      targetEvent: ev,
      clickOutsideToClose: true,
      onRemoving: function() {
        
      }
    });
  }
})