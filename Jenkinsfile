node { // <1>
    stage( 'Checkout' ) {
        git( branch: 'jenkinsfile', url: 'http://git-mch.192.168.1.14.xip.io/certMgr.git' )
    }
    stage('Build') { // <2>
        echo "Build stage"
    }
    stage('Test') {
        echo "Test stage"
    }
    stage('Deploy') {
        echo "Deploy stage"
    }
}

