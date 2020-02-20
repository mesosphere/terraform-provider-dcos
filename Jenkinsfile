pipeline {
    agent none
    environment {
        GITHUB_TOKEN = credentials('gh-token-mesosphere-ci-dcos-deploy')
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
              label "golang112"
            }
            when { tag "v*" }
            steps {
                echo 'Starting release on tag.'
                sh 'wget -O /tmp/goreleaser.tgz https://github.com/goreleaser/goreleaser/releases/download/v0.123.3/goreleaser_Linux_x86_64.tar.gz && tar xzf /tmp/goreleaser.tgz -C /usr/local/bin'
                sh 'goreleaser --rm-dist'
            }
        }
    }
}
