pipeline {
    agent none
    environment {
      GITHUB_TOKEN = credentials('gh-token-mesosphere-ci-dcos-deploy')
      APPLE_DEVACC = credentials('APPLE_DEVELOPER_ACCOUNT')
      GOLANG_VER = "1.13.7"
      GORELEASER_VER = "0.140.1"
      GON_VER = "0.2.2"
    }
    options {
      disableConcurrentBuilds()
    }
    stages {
        stage('Docs') {
            environment {
              GIT_COMMITTER_NAME = 'dcos-sre-robot'
              GIT_COMMITTER_EMAIL = 'sre@mesosphere.io'
            }
            agent {
              label "golang112"
            }
            when {
                branch "master"
                changeset "docs/*"
            }
            steps {
                echo 'Deleting old publication'
                sh """
                    rm -rf public
                    mkdir public
                    git worktree prune
                    rm -rf .git/worktrees/public/

                    git remote set-branches --add origin gh-pages
                    git fetch origin gh-pages
                    git worktree add -B gh-pages public origin/gh-pages
                    rm -rf public/*
                """

                echo 'Building documentation'
                sh """
                    wget -O /tmp/hugo.tgz https://github.com/gohugoio/hugo/releases/download/v0.65.0/hugo_extended_0.65.0_Linux-64bit.tar.gz
                    tar xzf /tmp/hugo.tgz -C /usr/local/bin
                    git submodule init
                    git submodule update
                    ( cd docs && hugo -d ../public)
                    ( cd public && git status )
                """

                echo 'Uploading documentation'
                sh """
                    git config --local --add credential.helper 'store --file=${WORKSPACE}/.git-credentials'
                    GITDOMAIN=\$(git config --get remote.origin.url | cut -f 3 -d "/")
                    echo "https://${GIT_COMMITTER_NAME}:${GITHUB_TOKEN}@\${GITDOMAIN}" >> ${WORKSPACE}/.git-credentials
                    git config --local user.name '${GIT_COMMITTER_NAME}'
                    git config --local user.email '${GIT_COMMITTER_EMAIL}'

                    ( cd public && git add --all && git commit -m "Automated Jenkins deployment" && git push)
                """
            }
        }
        stage('Release') {
            agent {
              label "mac"
            }
            when { tag "*" }
            steps {
              withCredentials(bindings: [certificate(credentialsId: 'APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE', \
                                             keystoreVariable: 'SIGNING_CERTIFICATE', \
                                             passwordVariable: 'SIGNING_CERTIFICATE_PASSWORD')]) {
                  ansiColor('xterm') {
                      sh '''
                        security delete-keychain jenkins-${JOB_NAME} || :
                        security create-keychain -p test jenkins-${JOB_NAME}
                        security unlock-keychain -p test jenkins-${JOB_NAME}
                        security list-keychains -d user -s jenkins-${JOB_NAME}
                        security default-keychain -s jenkins-${JOB_NAME}
                        cat ${SIGNING_CERTIFICATE} > cert.p12
                        security import cert.p12 -k jenkins-${JOB_NAME} -P ${SIGNING_CERTIFICATE_PASSWORD} -T /usr/bin/codesign
                        security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k test jenkins-${JOB_NAME}
                        xcrun altool --notarization-history 0 -u "${APPLE_DEVACC_USR}" -p "${APPLE_DEVACC_PSW}"
                      '''
                      sh 'security find-identity -v | grep -q 204250D9ADA4A6CDB12C5E8BA168E48F5043CFDE'
                      echo 'Starting release on tag.'
                      sh 'mkdir -p ${WORKSPACE}/usr/local/bin'
                      sh 'wget -O /tmp/golang.tgz https://dl.google.com/go/go\${GOLANG_VER}.darwin-amd64.tar.gz && tar xzf /tmp/golang.tgz -C ${WORKSPACE}/usr/local'
                      sh 'wget -O /tmp/goreleaser.tgz https://github.com/goreleaser/goreleaser/releases/download/v\${GORELEASER_VER}/goreleaser_Darwin_x86_64.tar.gz && tar xzf /tmp/goreleaser.tgz -C ${WORKSPACE}/usr/local/bin'
                      sh 'wget -O /tmp/gon.zip https://github.com/mitchellh/gon/releases/download/v\${GON_VER}/gon_\${GON_VER}_macos.zip && unzip /tmp/gon.zip -d ${WORKSPACE}/usr/local/bin'
                      sh '''
                        set +xe
                        export PATH=$PATH:${WORKSPACE}/usr/local/go/bin:${WORKSPACE}/usr/local/bin
                        export AC_USERNAME="${APPLE_DEVACC_USR}"
                        export AC_PASSWORD="${APPLE_DEVACC_PSW}"
                        export GITHUB_TOKEN="${GITHUB_TOKEN}"
                        goreleaser --rm-dist
                      '''
                  }
              }
            }
        }
    }
}
