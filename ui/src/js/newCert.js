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

app.controller('NewCertForm', function( $scope, $http ) {
  $scope.moduleState = 'form';
  $scope.submit = function() {

    var hostname = [];
    hostname[ 0 ] = $scope.form.server + $scope.form.subzone;

    $http( {
      method  : 'POST',
      url     : apiEndpoint + "/certificates",
      data    : { name : hostname, validFor : 30 }
    }).success( function( data ) {
      console.log( data );
      $scope.moduleState = 'result';
      $scope.certificate = data.certificate;
      $scope.key = data.key;
    } );
  };

  $http.get( apiEndpoint + "/permittedDomains" )
    .success( function( data ){
      $scope.subzones = data.domains;
      $scope.form = { subzone : "." + data.domains[ 0 ] };
      $scope.server = "";
    })
});
