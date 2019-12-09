pipeline {
    agent none
    environment {
        GITHUB_TOKEN = credentials('mesosphere-ci-2018-11-08')
    }
    options {
      disableConcurrentBuilds()
    }
    stages {
        stage('Release') {
            agent {
              label "golang112"
            }
            when { tag "v*" }
            steps {
                echo 'Starting release on tag.'
                sh 'wget -O /tmp/goreleaser.tgz https://github.com/goreleaser/goreleaser/releases/download/v0.123.3/goreleaser_Linux_x86_64.tar.gz && tar xzf /tmp/goreleaser.tgz -C /usr/local/bin'
                sh 'goreleaser'
            }
        }
    }
}
