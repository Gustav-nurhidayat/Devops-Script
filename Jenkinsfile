pipeline {
    agent any

    environment {
        IMAGE_NAME = 'simplejourney-devops'
        TARGET_CONTAINER = 'simplejourney-hotfix'
        TARGET_PORT = '8081'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Test') {
            steps {
                sh '''
                    set -e
                    go test ./...
                '''
            }
        }

        stage('Prepare Version') {
            steps {
                script {
                    env.VERSION = sh(
                        script: 'git rev-parse --short HEAD',
                        returnStdout: true
                    ).trim()
                }

                echo "Build version: ${env.VERSION}"
            }
        }

        stage('Build Docker Image') {
            steps {
                sh '''
                    set -e

                    docker build \
                        --build-arg VERSION="${VERSION}" \
                        -t "${IMAGE_NAME}:${VERSION}" \
                        .
                '''
            }
        }

        stage('Verify Image') {
            steps {
                sh '''
                    set -e

                    docker image inspect "${IMAGE_NAME}:${VERSION}" >/dev/null

                    docker images "${IMAGE_NAME}:${VERSION}"
                '''
            }
        }

        stage('Deploy Hotfix') {
            steps {
                sh '''
                    set -e

                    VERSION="${VERSION}" \
                    CONTAINER="${TARGET_CONTAINER}" \
                    PORT="${TARGET_PORT}" \
                    ./scripts/deploy-hotfix.sh
                '''
            }
        }
    }

    post {
        success {
            echo 'Pipeline completed successfully.'
        }

        failure {
            echo 'Pipeline failed. Deployment was not completed successfully.'
        }

        always {
            sh '''
                rm -f deploy/app
            '''
        }
    }
}