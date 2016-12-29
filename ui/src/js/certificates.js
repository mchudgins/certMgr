"use strict";

/*
https://104.154.46.50/api/v1/proxy/namespaces/development/services/fabric8
*/

var apiEndpoint = $("meta[name='apiEndpoint']").attr("content");

var app = angular.module( 'app', [ 'ngRoute' ] )
            .config(["$interpolateProvider", function($interpolateProvider){
                            $interpolateProvider.startSymbol("[[");
                            $interpolateProvider.endSymbol("]]");
              }]);

app.controller('ListControl', function( $scope, $http ) {
  $http.get( apiEndpoint ).success(function(data){
    $scope.list = data;
  })
  $scope.x = apiEndpoint;
});
