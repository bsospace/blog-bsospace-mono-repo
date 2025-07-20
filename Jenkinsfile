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
                            # Create folders if not exist
                            mkdir -p backend/keys
                            mkdir -p frontend
                            mkdir -p backend

                            # Copy PEM key
                            cp "$PUBLIC_KEY_FILE" backend/keys/blogPublicAccess.pem

                            # Copy .env files
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

                    // Check if Discord webhook is set
                    if (env.DISCORD_WEBHOOK == null || env.DISCORD_WEBHOOK == '') {
                        error "DISCORD_WEBHOOK credential is not set"
                    }

                    // Check if PEM key file exists and has content
                    def pemContent = readFile('backend/keys/blogPublicAccess.pem')
                    if (pemContent == null || pemContent.trim() == '') {
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
                    if (frontendEnv == null || frontendEnv.trim() == '') {
                        error "frontend/.env file is empty"
                    }

                    def backendEnv = readFile('backend/.env')
                    if (backendEnv == null || backendEnv.trim() == '') {
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

                    // Make deploy.sh executable and run it
                    sh 'chmod +x deploy.sh'
                    sh './deploy.sh'

                    echo "Deployment completed successfully"
                }
            }
        }

        // stage('Clear Build Cache') {
        //     when {
        //         branch 'master'
        //     }
        //     steps {
        //         script {
        //             echo "Clearing build cache..."

        //             // Clean Go cache
        //             dir('backend') {
        //                 sh 'go clean -cache -modcache -testcache'
        //             }

        //             // Clean pnpm cache (if frontend exists)
        //             dir('frontend') {
        //                 if (fileExists('package.json')) {
        //                     sh 'pnpm store prune'
        //                 }
        //             }

        //             // Clean Docker cache
        //             sh 'docker system prune -f'

        //             echo "Build cache cleared successfully"
        //         }
        //     }
        // }
    }

    post {
        always {
            script {
                // Set build status for notification
                if (currentBuild.result == 'SUCCESS') {
                    env.BUILD_STATUS = 'SUCCESS'
                } else if (currentBuild.result == 'FAILURE') {
                    env.BUILD_STATUS = 'FAILURE'
                } else if (currentBuild.result == 'ABORTED') {
                    env.BUILD_STATUS = 'ABORTED'
                } else {
                    env.BUILD_STATUS = 'UNKNOWN'
                }

                // Send Discord notification
                if (env.BUILD_STATUS != 'UNKNOWN') {
                    sendDiscordNotification()
                }

                // Cleanup workspace
                cleanWs()
            }
        }

        success {
            script {
                echo "Pipeline completed successfully!"
            }
        }

        failure {
            script {
                echo "Pipeline failed!"
            }
        }
    }
}

// Function to send notification to Discord via webhook
def sendDiscordNotification() {
    def color = ''
    def title = ''

    // Set color and title based on build status
    switch(env.BUILD_STATUS) {
        case 'SUCCESS':
            color = '00ff00' // Green
            title = '✅ Deployment Successful'
            break
        case 'FAILURE':
            color = 'ff0000' // Red
            title = '❌ Deployment Failed'
            break
        case 'ABORTED':
            color = 'ffff00' // Yellow
            title = '⚠️ Deployment Aborted'
            break
        default:
            color = '808080' // Gray
            title = '❓ Deployment Status Unknown'
    }

    // Construct Discord embed payload
    def payload = [
        embeds: [
            [
                title: title,
                description: "**Project:** Blog BSO Space\n**Branch:** ${env.BRANCH_NAME}\n**Build:** #${env.BUILD_NUMBER}\n**Duration:** ${currentBuild.durationString}",
                color: color.toInteger(),
                timestamp: new Date().toISOString(),
                footer: [
                    text: "Jenkins Pipeline"
                ]
            ]
        ]
    ]

    def jsonPayload = groovy.json.JsonOutput.toJson(payload)

    // Send notification to Discord
    sh """
        curl -H "Content-Type: application/json" \\
             -X POST \\
             -d '${jsonPayload}' \\
             ${env.DISCORD_WEBHOOK}
    """
}
