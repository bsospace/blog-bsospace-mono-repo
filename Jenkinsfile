pipeline {
    agent any

    environment {
        DISCORD_WEBHOOK = credentials('discord-webhook')
        BUILD_STATUS = 'UNKNOWN'
    }

    options {
        skipStagesAfterUnstable()
        timeout(time: 30, unit: 'MINUTES')
    }

    triggers {
        pollSCM('H/5 * * * *')
    }

    stages {
        stage('Checkout') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Starting deployment pipeline for master branch..."
                    checkout scm
                }
            }
        }

        stage('Setup Credentials') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Setting up credentials..."

                    withCredentials([
                        file(credentialsId: 'blog-access-public-key-pem', variable: 'PUBLIC_KEY_FILE'),
                        file(credentialsId: 'blog-frontend-env', variable: 'FRONTEND_ENV_FILE'),
                        file(credentialsId: 'blog-backend-env', variable: 'BACKEND_ENV_FILE')
                    ]) {
                        sh '''
                            mkdir -p backend/keys frontend backend
                            cp "$PUBLIC_KEY_FILE" backend/keys/blogPublicAccess.pem
                            cp "$FRONTEND_ENV_FILE" frontend/.env
                            cp "$BACKEND_ENV_FILE" backend/.env
                        '''
                    }

                    echo "Credentials setup completed"
                }
            }
        }

        stage('Validate Credentials') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Validating credentials..."

                    if (env.DISCORD_WEBHOOK == null || env.DISCORD_WEBHOOK == '') {
                        error "DISCORD_WEBHOOK credential is not set"
                    }

                    def pemContent = readFile('backend/keys/blogPublicAccess.pem')
                    if (!pemContent?.trim()) {
                        error "PEM key file is empty or invalid"
                    }

                    echo "All credentials validated successfully"
                }
            }
        }

        stage('Validate Environment Files') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Validating environment files..."

                    if (!fileExists('frontend/.env')) {
                        error "frontend/.env file is missing"
                    }

                    if (!fileExists('backend/.env')) {
                        error "backend/.env file is missing"
                    }

                    def frontendEnv = readFile('frontend/.env')
                    if (!frontendEnv?.trim()) {
                        error "frontend/.env file is empty"
                    }

                    def backendEnv = readFile('backend/.env')
                    if (!backendEnv?.trim()) {
                        error "backend/.env file is empty"
                    }

                    echo "All environment files validated successfully"
                }
            }
        }

        stage('Deploy') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Starting deployment..."
                    sh 'chmod +x deploy.sh'
                    sh './deploy.sh'
                    echo "Deployment completed successfully"
                }
            }
        }
    }

    post {
        always {
            script {
                // Capture build status safely
                env.BUILD_STATUS = currentBuild.currentResult ?: 'UNKNOWN'

                if (env.BUILD_STATUS != 'UNKNOWN') {
                    sendDiscordNotification()
                }

                cleanWs()
            }
        }

        success {
            echo "Pipeline completed successfully!"
        }

        failure {
            echo "Pipeline failed!"
        }
    }
}

// ======= Discord Webhook Function =========
def sendDiscordNotification() {
    def color = 0x808080
    def title = '❓ Deployment Status Unknown'

    switch(env.BUILD_STATUS) {
        case 'SUCCESS':
            color = 0x00ff00
            title = '✅ Deployment Successful'
            break
        case 'FAILURE':
            color = 0xff0000
            title = '❌ Deployment Failed'
            break
        case 'ABORTED':
            color = 0xffff00
            title = '⚠️ Deployment Aborted'
            break
    }

    def payload = [
        username: "Jenkins CI",
        embeds: [
            [
                title: title,
                description: """**Project:** Blog BSO Space
**Branch:** ${env.BRANCH_NAME}
**Build Number:** #${env.BUILD_NUMBER}
**Duration:** ${currentBuild.durationString}
**Build URL:** ${env.BUILD_URL}""",
                color: color,
                timestamp: new Date().format("yyyy-MM-dd'T'HH:mm:ss'Z'", TimeZone.getTimeZone('UTC')),
                footer: [ text: "Jenkins Pipeline" ]
            ]
        ]
    ]

    def jsonPayload = groovy.json.JsonOutput.toJson(payload)

    // Send to Discord
    sh """
        curl -H "Content-Type: application/json" \\
             -X POST \\
             -d '${jsonPayload}' \\
             '${env.DISCORD_WEBHOOK}'
    """
}
