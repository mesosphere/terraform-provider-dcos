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
            agent {
              label "golang112"
            }
            // when {
            //     branch "master"
            //     changeset "docs/*"
            // }
            steps {
                echo 'Deleting old publication'
                sh 'rm -rf public && mkdir public && git worktree prune && rm -rf .git/worktrees/public/'
                sh 'git worktree add -B gh-pages public origin/gh-pages && rm -rf public/*'

                echo 'Building documentation'
                sh 'wget -O /tmp/hugo.tgz https://github.com/gohugoio/hugo/releases/download/v0.65.0/hugo_extended_0.65.0_Linux-64bit.tar.gz && tar xzf /tmp/hugo.tgz -C /usr/local/bin'
                sh 'git submodule init'
                sh 'cd docs && hugo -d ../public'

                echo 'Uploading documentation'
                sh 'cd public && git add --all && git commit -m "Automated Jenkins deployment"'
                //sh 'git push'
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
