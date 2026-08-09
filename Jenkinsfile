// =============================================================================
// OpsFlow — Baseline Jenkins Pipeline
// =============================================================================
// Phase 0: Lint, test, and build. No Docker build/push yet (Phase 8).

pipeline {
    agent any

    environment {
        GO_VERSION    = '1.22'
        NODE_VERSION  = '20'
        GOFLAGS       = '-mod=readonly'
    }

    options {
        timeout(time: 15, unit: 'MINUTES')
        timestamps()
        disableConcurrentBuilds()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Go Lint') {
            steps {
                script {
                    def modules = ['pkg/common', 'services/auth', 'services/incident',
                                   'services/registry', 'services/ai-gateway', 'services/gateway']
                    for (mod in modules) {
                        dir(mod) {
                            sh 'golangci-lint run ./...'
                        }
                    }
                }
            }
        }

        stage('Go Test') {
            steps {
                script {
                    def modules = ['pkg/common', 'services/auth', 'services/incident',
                                   'services/registry', 'services/ai-gateway', 'services/gateway']
                    for (mod in modules) {
                        dir(mod) {
                            sh 'go test -v -race -coverprofile=coverage.out ./...'
                        }
                    }
                }
            }
            post {
                always {
                    archiveArtifacts artifacts: '**/coverage.out', allowEmptyArchive: true
                }
            }
        }

        stage('Go Build') {
            steps {
                script {
                    def services = ['auth', 'incident', 'registry', 'ai-gateway', 'gateway']
                    for (svc in services) {
                        dir("services/${svc}") {
                            sh "go build -o bin/${svc}-server ./cmd/server"
                        }
                    }
                }
            }
        }

        stage('Frontend Install') {
            steps {
                dir('frontend') {
                    sh 'npm ci'
                }
            }
        }

        stage('Frontend Lint') {
            steps {
                dir('frontend') {
                    sh 'npx eslint src/'
                }
            }
        }

        stage('Frontend Build') {
            steps {
                dir('frontend') {
                    sh 'npm run build'
                }
            }
        }
    }

    post {
        failure {
            echo 'Pipeline failed. Check stage logs above.'
        }
        success {
            echo 'All stages passed.'
        }
        cleanup {
            cleanWs()
        }
    }
}
