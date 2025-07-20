pipeline {
    agent any
    
    environment {
        DISCORD_WEBHOOK = credentials('discord-webhook')
        BLOG_ACCESS_PUBLIC_KEY_PEM_PROD = credentials('blog-access-public-key-pem')
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
                    
                    // Create backend/keys directory if it doesn't exist
                    sh 'mkdir -p backend/keys'
                    
                    // Write the PEM key to backend/keys/blogPublicAccess.pem
                    writeFile file: 'backend/keys/blogPublicAccess.pem', text: BLOG_ACCESS_PUBLIC_KEY_PEM_PROD
                    
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
                        error "BLOG_ACCESS_PUBLIC_KEY_PEM_PROD credential is empty or invalid"
                    }
                    
                    // Validate PEM format (basic check)
                    if (!pemContent.contains('-----BEGIN') || !pemContent.contains('-----END')) {
                        error "PEM key format appears to be invalid"
                    }
                    
                    echo "All credentials validated successfully"
                }
            }
        }
        
        stage('Build') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Starting build process..."
                    
                    // Build backend (Go)
                    dir('backend') {
                        sh 'go mod download'
                        sh 'go build -o bin/server ./cmd/server'
                        echo "Backend build completed"
                    }
                    
                    // Build frontend (if it's a Node.js project)
                    dir('frontend') {
                        if (fileExists('package.json')) {
                            sh 'pnpm install --frozen-lockfile'
                            sh 'pnpm run build'
                            echo "Frontend build completed"
                        } else {
                            echo "No package.json found in frontend, skipping frontend build"
                        }
                    }
                    
                    echo "Build process completed successfully"
                }
            }
        }
        
        stage('Test') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Running tests..."
                    
                    // Run Go tests
                    dir('backend') {
                        sh 'go test ./... -v'
                        echo "Go tests completed"
                    }
                    
                    echo "All tests completed"
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
        
        stage('Clear Build Cache') {
            when {
                branch 'master'
            }
            steps {
                script {
                    echo "Clearing build cache..."
                    
                    // Clean Go cache
                    dir('backend') {
                        sh 'go clean -cache -modcache -testcache'
                    }
                    
                    // Clean pnpm cache (if frontend exists)
                    dir('frontend') {
                        if (fileExists('package.json')) {
                            sh 'pnpm store prune'
                        }
                    }
                    
                    // Clean Docker cache
                    sh 'docker system prune -f'
                    
                    echo "Build cache cleared successfully"
                }
            }
        }
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
        
        always {
            script {
                // Send Discord notification
                if (env.BUILD_STATUS != 'UNKNOWN') {
                    sendDiscordNotification()
                }
                
                // Cleanup
                cleanWs()
            }
        }
    }
}

def sendDiscordNotification() {
    def color = ''
    def title = ''
    
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
    
    sh """
        curl -H "Content-Type: application/json" \
             -X POST \
             -d '${jsonPayload}' \
             ${env.DISCORD_WEBHOOK}
    """
}